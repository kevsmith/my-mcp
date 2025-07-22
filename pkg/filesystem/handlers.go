package filesystem

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kevsmith/my-mcp/pkg/shared"
	"github.com/mark3labs/mcp-go/mcp"
)

// Navigation handlers
func ChangeDirectoryHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args ChangeDirectoryArgs
		if err := shared.OptimizedUnmarshalRequest(request, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments: " + err.Error()), nil
		}

		err := handler.ChangeDirectory(args.Path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to change directory: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Changed directory to: %s", handler.GetCurrentDirectory())), nil
	}
}

func GetCurrentDirectoryHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		currentDir := handler.GetCurrentDirectory()
		return mcp.NewToolResultText(currentDir), nil
	}
}

func GetDirectoryInfoHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dirInfo := handler.GetDirectoryInfo()

		return shared.OptimizedToolResultJSON(dirInfo)
	}
}

// File operation handlers
func ListDirectoryHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args ListDirectoryArgs
		if err := shared.OptimizedUnmarshalRequest(request, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments: " + err.Error()), nil
		}

		files, err := handler.ListDirectory(args.Path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list directory: %v", err)), nil
		}

		return shared.OptimizedToolResultJSON(files)
	}
}

func GlobHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args GlobArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		result, err := handler.Glob(args.Pattern)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to glob pattern: %v", err)), nil
		}

		content, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(content)), nil
	}
}

func GetFileInfoHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args GetFileInfoArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		fileInfo, err := handler.GetFileInfo(args.Path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get file info: %v", err)), nil
		}

		content, err := json.Marshal(fileInfo)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(content)), nil
	}
}

func ReadFileHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args ReadFileArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		content, err := handler.ReadFile(args.Path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
		}

		return mcp.NewToolResultText(content), nil
	}
}
