package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	outlookserver "github.com/kevsmith/my-mcp/pkg/server"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Check if running on Windows
	if runtime.GOOS != "windows" {
		log.Fatal("outlook-mcp server is only supported on Windows")
	}

	s, err := outlookserver.NewOutlookMCPServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Fprintf(os.Stderr, "\nShutting down outlook-mcp server...\n")
		outlookserver.ShutdownOutlookManager()
		os.Exit(0)
	}()

	fmt.Fprintf(os.Stderr, "Starting outlook-mcp server for Windows...\n")

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
