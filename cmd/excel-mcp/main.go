package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/kevsmith/my-mcp/pkg/server"
	mcpServer "github.com/mark3labs/mcp-go/server"
)

func main() {
	var cacheSize int
	var cacheTTLMinutes int

	// Parse command line flags
	flag.IntVar(&cacheSize, "cache-size", 0, "Maximum number of Excel files to cache (default: 10, env: EXCEL_CACHE_MAX_SIZE)")
	flag.IntVar(&cacheTTLMinutes, "cache-ttl", 0, "Cache TTL in minutes (default: 5, env: EXCEL_CACHE_TTL_MINUTES)")
	flag.Parse()

	// Override environment variables if command line args are provided
	if cacheSize > 0 {
		os.Setenv("EXCEL_CACHE_MAX_SIZE", strconv.Itoa(cacheSize))
	}
	if cacheTTLMinutes > 0 {
		os.Setenv("EXCEL_CACHE_TTL_MINUTES", strconv.Itoa(cacheTTLMinutes))
	}

	// Setup the MCP server with all tools and handlers
	srv := server.ExcelSetup()

	// Start serving via stdio
	mcpServer.ServeStdio(srv)
}
