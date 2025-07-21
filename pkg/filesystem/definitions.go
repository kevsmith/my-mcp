package filesystem

import "github.com/mark3labs/mcp-go/mcp"

func GetToolDefinitions() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "list_directory",
			Description: "List files and directories in a given path",
			Annotations: mcp.ToolAnnotation{
				ReadOnlyHint: &[]bool{true}[0],
			},
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The directory path to list (relative to base directory)",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "glob",
			Description: "Find files matching a wildcard pattern",
			Annotations: mcp.ToolAnnotation{
				ReadOnlyHint: &[]bool{true}[0],
			},
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "The glob pattern to match (relative to base directory)",
					},
				},
				Required: []string{"pattern"},
			},
		},
		{
			Name:        "get_file_info",
			Description: "Get metadata for a specific file or directory",
			Annotations: mcp.ToolAnnotation{
				ReadOnlyHint: &[]bool{true}[0],
			},
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The file path to get info for (relative to base directory)",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read the contents of a file",
			Annotations: mcp.ToolAnnotation{
				ReadOnlyHint: &[]bool{true}[0],
			},
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The file path to read (relative to base directory)",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "get_absolute_path",
			Description: "Get the absolute path for a given file or directory",
			Annotations: mcp.ToolAnnotation{
				ReadOnlyHint: &[]bool{true}[0],
			},
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The file or directory path to get absolute path for (relative to base directory)",
					},
				},
				Required: []string{"path"},
			},
		},
	}
}
