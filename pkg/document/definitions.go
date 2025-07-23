package document

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetToolDefinitions() []mcp.Tool {
	return []mcp.Tool{
		mcp.NewTool("extract_text",
			mcp.WithDescription("Extract clean prose text from document files (.pdf, .docx, .pptx) - removes XML markup and formatting"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("file_path",
				mcp.Description("Absolute path to the document file"),
				mcp.Required(),
			),
		),
		mcp.NewTool("get_document_info",
			mcp.WithDescription("Get metadata and information about a document file"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("file_path",
				mcp.Description("Absolute path to the document file"),
				mcp.Required(),
			),
		),
	}
}
