package outlook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

type ListMessagesArgs struct {
	Page *int `json:"page,omitempty"`
}

type GetMessageArgs struct {
	MessageID string `json:"message_id"`
}

type SearchMessagesArgs struct {
	Query string `json:"query"`
}

// ListMessagesHandler handles the list_messages tool
func ListMessagesHandler(manager *Manager) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args ListMessagesArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		page := 1
		if args.Page != nil {
			page = *args.Page
		}

		response, err := manager.ListMessages(page)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list messages: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(`Messages (Page %d of %d):

Total Messages: %d
Current Page: %d messages

`, response.Pagination.Page,
			(response.Pagination.Total+response.Pagination.PageSize-1)/response.Pagination.PageSize,
			response.Pagination.Total,
			len(response.Messages)) +
			formatMessageList(response.Messages)), nil
	}
}

// GetMessageHandler handles the get_message tool
func GetMessageHandler(manager *Manager) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args GetMessageArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		if args.MessageID == "" {
			return mcp.NewToolResultError("message_id parameter is required"), nil
		}

		message, err := manager.GetMessage(args.MessageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get message: %v", err)), nil
		}

		result := fmt.Sprintf(`Message Details:

Subject: %s
From: %s <%s>
Received: %s
Size: %d bytes
Unread: %t
Has Attachments: %t (%d attachments)
Importance: %s

Preview:
%s`, message.Subject, message.Sender, message.SenderEmail,
			message.ReceivedTime.Format("2006-01-02 15:04:05"),
			message.Size, message.Unread, message.HasAttachments, message.AttachmentCount,
			getImportanceString(message.Importance), message.BodyPreview)

		return mcp.NewToolResultText(result), nil
	}
}

// GetMessageBodyHandler handles the get_message_body tool
func GetMessageBodyHandler(manager *Manager) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args GetMessageArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		if args.MessageID == "" {
			return mcp.NewToolResultError("message_id parameter is required"), nil
		}

		response, err := manager.GetMessageBody(args.MessageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get message body: %v", err)), nil
		}

		result := fmt.Sprintf(`Message Body (Readable Text):

Word Count: %d
Character Count: %d

Content:
%s`, response.WordCount, response.CharCount, response.BodyText)

		return mcp.NewToolResultText(result), nil
	}
}

// GetMessageBodyRawHandler handles the get_message_body_raw tool
func GetMessageBodyRawHandler(manager *Manager) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args GetMessageArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		if args.MessageID == "" {
			return mcp.NewToolResultError("message_id parameter is required"), nil
		}

		response, err := manager.GetMessageBodyRaw(args.MessageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get raw message body: %v", err)), nil
		}

		result := fmt.Sprintf(`Message Body (Raw):

Format: %s

Plain Text Body:
%s

HTML Body:
%s`, response.Format, response.BodyText, response.BodyHTML)

		return mcp.NewToolResultText(result), nil
	}
}

// SearchMessagesHandler handles the search_messages tool
func SearchMessagesHandler(manager *Manager) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args SearchMessagesArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		if args.Query == "" {
			return mcp.NewToolResultError("query parameter is required"), nil
		}

		response, err := manager.SearchMessages(args.Query)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to search messages: %v", err)), nil
		}

		result := fmt.Sprintf(`Search Results for "%s":

Found %d messages:

%s`, response.Query, response.Count, formatMessageList(response.Results))

		return mcp.NewToolResultText(result), nil
	}
}

// Helper function to format a list of messages
func formatMessageList(messages []Message) string {
	if len(messages) == 0 {
		return "No messages found."
	}

	result := ""
	for i, msg := range messages {
		unreadStatus := ""
		if msg.Unread {
			unreadStatus = " [UNREAD]"
		}

		attachmentInfo := ""
		if msg.HasAttachments {
			attachmentInfo = fmt.Sprintf(" 📎(%d)", msg.AttachmentCount)
		}

		result += fmt.Sprintf(`%d. %s%s%s
   From: %s <%s>
   Received: %s
   Size: %d bytes
   ID: %s

`, i+1, msg.Subject, unreadStatus, attachmentInfo,
			msg.Sender, msg.SenderEmail,
			msg.ReceivedTime.Format("2006-01-02 15:04:05"),
			msg.Size, msg.ID)
	}

	return result
}

// Helper function to convert importance number to string
func getImportanceString(importance int) string {
	switch importance {
	case 0:
		return "Low"
	case 1:
		return "Normal"
	case 2:
		return "High"
	default:
		return "Unknown"
	}
}
