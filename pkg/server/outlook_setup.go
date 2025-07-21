package server

import (
	"github.com/kevsmith/my-mcp/pkg/outlook"
	"github.com/mark3labs/mcp-go/server"
)

// Global manager reference for cleanup
var outlookManager *outlook.Manager

// OutlookMCPServer extends MCPServer with Outlook-specific functionality
type OutlookMCPServer struct {
	*server.MCPServer
	manager *outlook.Manager
}

// NewOutlookMCPServer creates a new Outlook MCP server
func NewOutlookMCPServer() (*server.MCPServer, error) {
	manager, err := outlook.NewManager()
	if err != nil {
		return nil, err
	}

	s := server.NewMCPServer(
		"outlook-mcp",
		"1.0.0",
		server.WithLogging(),
	)

	toolDefinitions := outlook.GetToolDefinitions()

	// Add all Outlook tools
	s.AddTool(toolDefinitions[0], outlook.ListMessagesHandler(manager))      // list_messages
	s.AddTool(toolDefinitions[1], outlook.GetMessageHandler(manager))        // get_message
	s.AddTool(toolDefinitions[2], outlook.GetMessageBodyHandler(manager))    // get_message_body
	s.AddTool(toolDefinitions[3], outlook.GetMessageBodyRawHandler(manager)) // get_message_body_raw
	s.AddTool(toolDefinitions[4], outlook.SearchMessagesHandler(manager))    // search_messages

	// Store manager reference for cleanup (using a global or context as needed)
	outlookManager = manager

	return s, nil
}

// ShutdownOutlookManager gracefully shuts down the global Outlook manager
func ShutdownOutlookManager() error {
	if outlookManager != nil {
		return outlookManager.Stop()
	}
	return nil
}
