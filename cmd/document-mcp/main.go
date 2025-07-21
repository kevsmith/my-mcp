package main

import (
	"github.com/kevsmith/my-mcp/pkg/server"
	mcpServer "github.com/mark3labs/mcp-go/server"
)

func main() {
	srv := server.DocumentSetup()

	mcpServer.ServeStdio(srv)
}
