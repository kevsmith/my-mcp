param(
    [int]$Port = 8080
)

# Get port from environment variable if not provided as parameter
if ($env:OUTLOOK_SERVER_PORT) {
    $Port = [int]$env:OUTLOOK_SERVER_PORT
}

Write-Host "Starting Outlook REST API server on localhost:$Port"

# Initialize Outlook COM object with error handling
$outlook = $null
$outlookAvailable = $false

try {
    $outlook = New-Object -ComObject Outlook.Application
    $namespace = $outlook.GetNamespace("MAPI")
    $inbox = $namespace.GetDefaultFolder(6) # olFolderInbox = 6
    $outlookAvailable = $true
    Write-Host "Successfully connected to Outlook"
} catch {
    Write-Warning "Could not connect to Outlook: $($_.Exception.Message)"
    Write-Host "Server will start but return errors for all requests"
}

# HTTP Listener setup
$listener = New-Object System.Net.HttpListener
$listener.Prefixes.Add("http://localhost:$Port/")
$listener.Start()

Write-Host "Outlook REST API Server listening on http://localhost:$Port"
Write-Host "Press Ctrl+C to stop the server"

# Helper function to convert Outlook item to JSON-compatible object
function Convert-OutlookItemToObject {
    param($item)
    
    $obj = @{
        id = $item.EntryID
        subject = $item.Subject
        sender = $item.SenderName
        senderEmail = $item.SenderEmailAddress
        receivedTime = $item.ReceivedTime.ToString("yyyy-MM-ddTHH:mm:ss.fffZ")
        sentOn = if ($item.SentOn) { $item.SentOn.ToString("yyyy-MM-ddTHH:mm:ss.fffZ") } else { $null }
        size = $item.Size
        unread = $item.UnRead
        importance = $item.Importance
        hasAttachments = $item.Attachments.Count -gt 0
        attachmentCount = $item.Attachments.Count
    }
    
    return $obj
}

# Helper function to get message body text (cooked)
function Get-MessageBodyText {
    param($item)
    
    # Try to get plain text body first
    if ($item.Body) {
        return $item.Body.Trim()
    }
    
    # Fallback to HTML body with basic cleanup
    if ($item.HTMLBody) {
        $htmlBody = $item.HTMLBody
        # Basic HTML tag removal - this is simplified
        $htmlBody = $htmlBody -replace '<[^>]+>', ''
        $htmlBody = $htmlBody -replace '&nbsp;', ' '
        $htmlBody = $htmlBody -replace '&lt;', '<'
        $htmlBody = $htmlBody -replace '&gt;', '>'
        $htmlBody = $htmlBody -replace '&amp;', '&'
        return $htmlBody.Trim()
    }
    
    return ""
}

# Main request processing loop
try {
    while ($listener.IsListening) {
        $context = $listener.GetContext()
        $request = $context.Request
        $response = $context.Response
        
        Write-Host "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') - $($request.HttpMethod) $($request.Url.PathAndQuery)"
        
        # Set CORS headers for localhost
        $response.Headers.Add("Access-Control-Allow-Origin", "http://localhost:*")
        $response.Headers.Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        $response.Headers.Add("Access-Control-Allow-Headers", "Content-Type")
        $response.ContentType = "application/json"
        
        $responseObj = $null
        $statusCode = 200
        
        try {
            if (-not $outlookAvailable) {
                $responseObj = @{
                    error = "Outlook is not available. Please ensure Outlook is installed and running."
                    code = "OUTLOOK_UNAVAILABLE"
                }
                $statusCode = 503
            } else {
                $path = $request.Url.AbsolutePath
                $query = $request.Url.Query
                
                switch -Regex ($path) {
                    "^/messages$" {
                        # GET /messages - list inbox messages with pagination
                        $pageParam = [System.Web.HttpUtility]::ParseQueryString($query)["page"]
                        $page = if ($pageParam) { [int]$pageParam } else { 1 }
                        $pageSize = 10
                        $skip = ($page - 1) * $pageSize
                        
                        $totalCount = $inbox.Items.Count
                        $items = $inbox.Items | Sort-Object ReceivedTime -Descending | Select-Object -Skip $skip -First $pageSize
                        
                        $messages = @()
                        foreach ($item in $items) {
                            if ($item.Class -eq 43) { # olMail = 43
                                $messages += Convert-OutlookItemToObject $item
                            }
                        }
                        
                        $responseObj = @{
                            messages = $messages
                            pagination = @{
                                page = $page
                                pageSize = $pageSize
                                total = $totalCount
                                hasNext = ($skip + $pageSize) -lt $totalCount
                                hasPrevious = $page -gt 1
                            }
                        }
                    }
                    
                    "^/messages/([^/]+)$" {
                        # GET /messages/{id} - full message details
                        $messageId = $matches[1]
                        
                        try {
                            $item = $namespace.GetItemFromID($messageId)
                            if ($item.Class -eq 43) { # olMail = 43
                                $messageObj = Convert-OutlookItemToObject $item
                                $messageObj.bodyPreview = (Get-MessageBodyText $item).Substring(0, [Math]::Min(200, (Get-MessageBodyText $item).Length))
                                $responseObj = $messageObj
                            } else {
                                $responseObj = @{ error = "Item is not a mail message"; code = "NOT_MAIL_ITEM" }
                                $statusCode = 400
                            }
                        } catch {
                            $responseObj = @{ error = "Message not found"; code = "MESSAGE_NOT_FOUND" }
                            $statusCode = 404
                        }
                    }
                    
                    "^/messages/([^/]+)/body/raw$" {
                        # GET /messages/{id}/body/raw - raw message body
                        $messageId = $matches[1]
                        
                        try {
                            $item = $namespace.GetItemFromID($messageId)
                            if ($item.Class -eq 43) { # olMail = 43
                                $responseObj = @{
                                    id = $messageId
                                    bodyText = $item.Body
                                    bodyHtml = $item.HTMLBody
                                    format = if ($item.BodyFormat -eq 2) { "HTML" } elseif ($item.BodyFormat -eq 3) { "RichText" } else { "PlainText" }
                                }
                            } else {
                                $responseObj = @{ error = "Item is not a mail message"; code = "NOT_MAIL_ITEM" }
                                $statusCode = 400
                            }
                        } catch {
                            $responseObj = @{ error = "Message not found"; code = "MESSAGE_NOT_FOUND" }
                            $statusCode = 404
                        }
                    }
                    
                    "^/messages/([^/]+)/body$" {
                        # GET /messages/{id}/body - cooked message body (readable text)
                        $messageId = $matches[1]
                        
                        try {
                            $item = $namespace.GetItemFromID($messageId)
                            if ($item.Class -eq 43) { # olMail = 43
                                $bodyText = Get-MessageBodyText $item
                                $responseObj = @{
                                    id = $messageId
                                    bodyText = $bodyText
                                    wordCount = ($bodyText -split '\s+').Count
                                    charCount = $bodyText.Length
                                }
                            } else {
                                $responseObj = @{ error = "Item is not a mail message"; code = "NOT_MAIL_ITEM" }
                                $statusCode = 400
                            }
                        } catch {
                            $responseObj = @{ error = "Message not found"; code = "MESSAGE_NOT_FOUND" }
                            $statusCode = 404
                        }
                    }
                    
                    "^/search$" {
                        # GET /search?q={query} - search within inbox
                        $searchQuery = [System.Web.HttpUtility]::ParseQueryString($query)["q"]
                        
                        if (-not $searchQuery) {
                            $responseObj = @{ error = "Query parameter 'q' is required"; code = "MISSING_QUERY" }
                            $statusCode = 400
                        } else {
                            # Use Outlook's search functionality
                            $searchResults = $inbox.Items.Restrict("[Subject] LIKE '%$searchQuery%' OR [Body] LIKE '%$searchQuery%' OR [SenderName] LIKE '%$searchQuery%'")
                            
                            $messages = @()
                            foreach ($item in $searchResults) {
                                if ($item.Class -eq 43) { # olMail = 43
                                    $messages += Convert-OutlookItemToObject $item
                                }
                            }
                            
                            # Sort by received time descending
                            $messages = $messages | Sort-Object receivedTime -Descending
                            
                            $responseObj = @{
                                query = $searchQuery
                                results = $messages
                                count = $messages.Count
                            }
                        }
                    }
                    
                    default {
                        $responseObj = @{ error = "Endpoint not found"; code = "NOT_FOUND" }
                        $statusCode = 404
                    }
                }
            }
        } catch {
            Write-Error "Error processing request: $($_.Exception.Message)"
            $responseObj = @{ 
                error = "Internal server error: $($_.Exception.Message)"
                code = "INTERNAL_ERROR"
            }
            $statusCode = 500
        }
        
        # Send response
        $response.StatusCode = $statusCode
        $jsonResponse = $responseObj | ConvertTo-Json -Depth 10
        $buffer = [System.Text.Encoding]::UTF8.GetBytes($jsonResponse)
        $response.ContentLength64 = $buffer.Length
        $response.OutputStream.Write($buffer, 0, $buffer.Length)
        $response.OutputStream.Close()
    }
} catch {
    Write-Error "Server error: $($_.Exception.Message)"
} finally {
    if ($listener) {
        $listener.Stop()
    }
    if ($outlook) {
        [System.Runtime.Interopservices.Marshal]::ReleaseComObject($outlook) | Out-Null
    }
    Write-Host "Server stopped"
}