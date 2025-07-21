package server

import (
	"fmt"

	"github.com/kevsmith/my-mcp/pkg/filesystem"
	"github.com/mark3labs/mcp-go/server"
)

func NewMCPServer(allowedRoots []string) (*server.MCPServer, error) {
	if len(allowedRoots) == 0 {
		return nil, fmt.Errorf("at least one allowed root directory is required")
	}

	handler, err := filesystem.NewHandler(allowedRoots)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem handler: %w", err)
	}

	s := server.NewMCPServer(
		"fs-mcp",
		"2.0.0", // Version bump for new interface
		server.WithLogging(),
	)

	toolDefinitions := filesystem.GetToolDefinitions()

	// Navigation tools
	s.AddTool(toolDefinitions[0], filesystem.ChangeDirectoryHandler(handler))     // change_directory
	s.AddTool(toolDefinitions[1], filesystem.GetCurrentDirectoryHandler(handler)) // get_current_directory
	s.AddTool(toolDefinitions[2], filesystem.GetDirectoryInfoHandler(handler))    // get_directory_info

	// File operation tools
	s.AddTool(toolDefinitions[3], filesystem.ListDirectoryHandler(handler)) // list_directory
	s.AddTool(toolDefinitions[4], filesystem.ReadFileHandler(handler))      // read_file
	s.AddTool(toolDefinitions[5], filesystem.GetFileInfoHandler(handler))   // get_file_info
	s.AddTool(toolDefinitions[6], filesystem.GlobHandler(handler))          // glob

	return s, nil
}
