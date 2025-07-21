package filesystem

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func ListDirectoryHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args ListDirectoryArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		files, err := handler.ListDirectory(args.Path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list directory: %v", err)), nil
		}

		content, err := json.Marshal(files)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(content)), nil
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

		files, err := handler.Glob(args.Pattern)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to glob pattern: %v", err)), nil
		}

		content, err := json.Marshal(files)
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

func GetAbsolutePathHandler(handler *Handler) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args GetAbsolutePathArgs
		argBytes, err := json.Marshal(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal arguments"), nil
		}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return mcp.NewToolResultError("Invalid arguments"), nil
		}

		absPath, err := handler.GetAbsolutePath(args.Path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get absolute path: %v", err)), nil
		}

		return mcp.NewToolResultText(absPath), nil
	}
}
