package outlook

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetToolDefinitions() []mcp.Tool {
	return []mcp.Tool{
		mcp.NewTool("list_messages",
			mcp.WithDescription("List messages from Outlook inbox with pagination"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithNumber("page",
				mcp.Description("Page number (default: 1)"),
			),
		),
		mcp.NewTool("get_message",
			mcp.WithDescription("Get full details of a specific message by ID"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("message_id",
				mcp.Description("The message ID (EntryID from Outlook)"),
				mcp.Required(),
			),
		),
		mcp.NewTool("get_message_body",
			mcp.WithDescription("Get the readable text content of a message"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("message_id",
				mcp.Description("The message ID (EntryID from Outlook)"),
				mcp.Required(),
			),
		),
		mcp.NewTool("get_message_body_raw",
			mcp.WithDescription("Get the raw body content (HTML and plain text) of a message"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("message_id",
				mcp.Description("The message ID (EntryID from Outlook)"),
				mcp.Required(),
			),
		),
		mcp.NewTool("search_messages",
			mcp.WithDescription("Search messages in Outlook inbox by subject, body, or sender"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("query",
				mcp.Description("Search query to match against subject, body, or sender"),
				mcp.Required(),
			),
		),
	}
}
