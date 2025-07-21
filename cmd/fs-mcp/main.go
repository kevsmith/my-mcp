package main

import (
	"fmt"
	"log"
	"os"

	mcpserver "github.com/kevsmith/my-mcp/pkg/server"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: fs-mcp <base-directory>")
	}

	basePath := os.Args[1]

	s, err := mcpserver.NewMCPServer(basePath)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Starting fs-mcp server with base directory: %s\n", basePath)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
