package filesystem

import "github.com/mark3labs/mcp-go/mcp"

func GetToolDefinitions() []mcp.Tool {
	return []mcp.Tool{
		// Navigation tools
		mcp.NewTool("change_directory",
			mcp.WithDescription("Change current working directory (like 'cd' command)"),
			mcp.WithReadOnlyHintAnnotation(false), // Changes internal state
			mcp.WithString("path",
				mcp.Description("Directory path to change to (relative to CWD or absolute within allowed roots)"),
				mcp.Required(),
			),
		),
		mcp.NewTool("get_current_directory",
			mcp.WithDescription("Get current working directory (like 'pwd' command)"),
			mcp.WithReadOnlyHintAnnotation(true),
		),
		mcp.NewTool("get_directory_info",
			mcp.WithDescription("Get current directory and list of allowed root directories"),
			mcp.WithReadOnlyHintAnnotation(true),
		),

		// File operations
		mcp.NewTool("list_directory",
			mcp.WithDescription("List files and directories (like 'ls' command)"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("path",
				mcp.Description("Directory path to list (optional, defaults to current directory)"),
			),
		),
		mcp.NewTool("read_file",
			mcp.WithDescription("Read the contents of a text file"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("path",
				mcp.Description("File path to read (relative to CWD or absolute within allowed roots)"),
				mcp.Required(),
			),
		),
		mcp.NewTool("get_file_info",
			mcp.WithDescription("Get metadata for a specific file or directory"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("path",
				mcp.Description("File or directory path to get info for"),
				mcp.Required(),
			),
		),
		mcp.NewTool("glob",
			mcp.WithDescription("Find files matching a wildcard pattern (like shell globbing)"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("pattern",
				mcp.Description("Glob pattern to match (e.g., '*.go', '**/test_*.py')"),
				mcp.Required(),
			),
		),
	}
}
