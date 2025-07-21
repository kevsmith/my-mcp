package document

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

type Handlers struct {
	documentManager *Manager
}

func NewHandlers(documentManager *Manager) *Handlers {
	return &Handlers{
		documentManager: documentManager,
	}
}

func (h *Handlers) ExtractText(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	text, err := h.documentManager.ExtractText(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if text == "" {
		return mcp.NewToolResultText("No text content found in the document"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Extracted text from %s:\n\n%s", filePath, text)), nil
}

func (h *Handlers) GetDocumentInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	info, err := h.documentManager.GetDocumentInfo(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	supportedText := "No"
	if info.IsSupported {
		supportedText = "Yes"
	}

	result := fmt.Sprintf(`Document Information:
File: %s
Size: %d bytes
Modified: %s
Extension: %s
Supported: %s`,
		info.FilePath,
		info.FileSize,
		info.ModTime.Format("2006-01-02 15:04:05"),
		info.Extension,
		supportedText,
	)

	return mcp.NewToolResultText(result), nil
}
