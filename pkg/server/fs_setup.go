package server

import (
	"fmt"
	"os"

	"github.com/kevsmith/my-mcp/pkg/filesystem"
	"github.com/mark3labs/mcp-go/server"
)

func NewMCPServer(basePath string) (*server.MCPServer, error) {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("base directory does not exist: %s", basePath)
	}

	handler := filesystem.NewHandler(basePath)

	s := server.NewMCPServer(
		"fs-mcp",
		"1.0.0",
		server.WithLogging(),
	)

	toolDefinitions := filesystem.GetToolDefinitions()
	toolHandlers := []func(){
		func() {
			s.AddTool(toolDefinitions[0], filesystem.ListDirectoryHandler(handler))
		},
		func() {
			s.AddTool(toolDefinitions[1], filesystem.GlobHandler(handler))
		},
		func() {
			s.AddTool(toolDefinitions[2], filesystem.GetFileInfoHandler(handler))
		},
		func() {
			s.AddTool(toolDefinitions[3], filesystem.ReadFileHandler(handler))
		},
		func() {
			s.AddTool(toolDefinitions[4], filesystem.GetAbsolutePathHandler(handler))
		},
	}

	for _, addTool := range toolHandlers {
		addTool()
	}

	return s, nil
}
