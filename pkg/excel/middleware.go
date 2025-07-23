package excel

import (
	"context"
	"fmt"

	"github.com/kevsmith/my-mcp/pkg/shared"
	"github.com/mark3labs/mcp-go/mcp"
)

// HandlerFunc represents an Excel handler function with pre-processed context
type HandlerFunc func(ctx context.Context, hctx *HandlerContext) (*mcp.CallToolResult, error)

// HandlerContext provides pre-processed common data for handlers
type HandlerContext struct {
	Request   mcp.CallToolRequest
	Manager   *Manager
	FilePath  string
	SheetName string      // Resolved sheet name (never empty after middleware)
	File      interface{} // Cached file reference
}

// Middleware wraps common Excel handler operations
func (h *Handlers) withMiddleware(handler HandlerFunc) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Create handler context
		hctx := &HandlerContext{
			Request: request,
			Manager: h.excelManager,
		}

		// Validate and extract file path
		hctx.FilePath = request.GetString("file_path", "")
		if hctx.FilePath == "" {
			return mcp.NewToolResultError("file_path parameter is required"), nil
		}

		// Open file once for reuse
		file, err := h.excelManager.OpenFile(hctx.FilePath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		hctx.File = file

		// Resolve sheet name if needed
		hctx.SheetName = request.GetString("sheet_name", "")
		if hctx.SheetName == "" {
			resolvedSheet, err := h.excelManager.GetCurrentSheet(hctx.FilePath, file)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			hctx.SheetName = resolvedSheet
		}

		// Call the actual handler
		return handler(ctx, hctx)
	}
}

// withMiddlewareNoSheet wraps handlers that don't need sheet resolution
func (h *Handlers) withMiddlewareNoSheet(handler HandlerFunc) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Create handler context
		hctx := &HandlerContext{
			Request: request,
			Manager: h.excelManager,
		}

		// Validate and extract file path
		hctx.FilePath = request.GetString("file_path", "")
		if hctx.FilePath == "" {
			return mcp.NewToolResultError("file_path parameter is required"), nil
		}

		// Open file once for reuse
		file, err := h.excelManager.OpenFile(hctx.FilePath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		hctx.File = file

		// Call the actual handler
		return handler(ctx, hctx)
	}
}

// Helper functions for common response patterns

// NewJSONResponse creates a JSON response with optimized marshaling
func NewJSONResponse(data interface{}) (*mcp.CallToolResult, error) {
	// Use optimized JSON marshaling from shared package
	return shared.OptimizedToolResultJSON(data)
}

// NewFormattedTextResponse creates a formatted text response
func NewFormattedTextResponse(format string, args ...interface{}) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(fmt.Sprintf(format, args...)), nil
}

// ValidateRequiredParam validates that a required parameter exists
func ValidateRequiredParam(hctx *HandlerContext, paramName string) (string, *mcp.CallToolResult) {
	value := hctx.Request.GetString(paramName, "")
	if value == "" {
		return "", mcp.NewToolResultError(paramName + " parameter is required")
	}
	return value, nil
}

// ValidateRequiredParamWithExample validates a parameter with an example
func ValidateRequiredParamWithExample(hctx *HandlerContext, paramName, example string) (string, *mcp.CallToolResult) {
	value := hctx.Request.GetString(paramName, "")
	if value == "" {
		return "", mcp.NewToolResultError(paramName + " parameter is required (e.g., '" + example + "')")
	}
	return value, nil
}
