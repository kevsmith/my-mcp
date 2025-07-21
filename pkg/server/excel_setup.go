package server

import (
	"github.com/kevsmith/my-mcp/pkg/excel"
	"github.com/mark3labs/mcp-go/server"
)

// ExcelSetup creates and configures the MCP server with all excel tools
func ExcelSetup() *server.MCPServer {
	// Create Excel manager
	excelManager := excel.NewManager()

	// Create tool handlers
	handlers := excel.NewHandlers(excelManager)

	// Create MCP server
	mcpServer := server.NewMCPServer("excel-mcp", "1.0.0", server.WithToolCapabilities(true))

	// Get tool definitions
	toolDefs := excel.GetToolDefinitions()

	// Register all tools with their handlers
	mcpServer.AddTool(toolDefs[0], handlers.EnumerateColumns)
	mcpServer.AddTool(toolDefs[1], handlers.EnumerateRows)
	mcpServer.AddTool(toolDefs[2], handlers.GetCellValue)
	mcpServer.AddTool(toolDefs[3], handlers.GetRangeValues)
	mcpServer.AddTool(toolDefs[4], handlers.ListSheets)
	mcpServer.AddTool(toolDefs[5], handlers.SetCurrentSheet)

	return mcpServer
}
