package outlook

import "time"

// Message represents an Outlook message with metadata
type Message struct {
	ID              string     `json:"id"`
	Subject         string     `json:"subject"`
	Sender          string     `json:"sender"`
	SenderEmail     string     `json:"senderEmail"`
	ReceivedTime    time.Time  `json:"receivedTime"`
	SentOn          *time.Time `json:"sentOn,omitempty"`
	Size            int        `json:"size"`
	Unread          bool       `json:"unread"`
	Importance      int        `json:"importance"`
	HasAttachments  bool       `json:"hasAttachments"`
	AttachmentCount int        `json:"attachmentCount"`
	BodyPreview     string     `json:"bodyPreview,omitempty"`
}

// MessageListResponse represents the response from the /messages endpoint
type MessageListResponse struct {
	Messages   []Message  `json:"messages"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination information
type Pagination struct {
	Page        int  `json:"page"`
	PageSize    int  `json:"pageSize"`
	Total       int  `json:"total"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}

// MessageBodyResponse represents the response from the /messages/{id}/body endpoint
type MessageBodyResponse struct {
	ID        string `json:"id"`
	BodyText  string `json:"bodyText"`
	WordCount int    `json:"wordCount"`
	CharCount int    `json:"charCount"`
}

// MessageBodyRawResponse represents the response from the /messages/{id}/body/raw endpoint
type MessageBodyRawResponse struct {
	ID       string `json:"id"`
	BodyText string `json:"bodyText"`
	BodyHTML string `json:"bodyHtml"`
	Format   string `json:"format"`
}

// SearchResponse represents the response from the /search endpoint
type SearchResponse struct {
	Query   string    `json:"query"`
	Results []Message `json:"results"`
	Count   int       `json:"count"`
}

// ErrorResponse represents an error response from the PowerShell server
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
