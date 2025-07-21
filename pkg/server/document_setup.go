package server

import (
	"github.com/kevsmith/my-mcp/pkg/document"
	"github.com/mark3labs/mcp-go/server"
)

func DocumentSetup() *server.MCPServer {
	documentManager := document.NewManager()

	handlers := document.NewHandlers(documentManager)

	mcpServer := server.NewMCPServer("document-mcp", "1.0.0", server.WithToolCapabilities(true))

	toolDefs := document.GetToolDefinitions()

	mcpServer.AddTool(toolDefs[0], handlers.ExtractText)
	mcpServer.AddTool(toolDefs[1], handlers.GetDocumentInfo)

	return mcpServer
}
