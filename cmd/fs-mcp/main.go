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
		log.Fatal("Usage: fs-mcp <root-dir1> [root-dir2] [root-dir3] ...")
	}

	allowedRoots := os.Args[1:]

	s, err := mcpserver.NewMCPServer(allowedRoots)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Starting fs-mcp server v2.0 with allowed roots: %v\n", allowedRoots)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
