package main

import (
	"github.com/kevsmith/my-mcp/pkg/server"
	mcpServer "github.com/mark3labs/mcp-go/server"
)

func main() {
	// Setup the MCP server with all tools and handlers
	srv := server.ExcelSetup()

	// Start serving via stdio
	mcpServer.ServeStdio(srv)
}
