package excel

import (
	"testing"
)

func TestGetToolDefinitions(t *testing.T) {
	tools := GetToolDefinitions()

	if len(tools) != 8 {
		t.Errorf("Expected 8 tools, got %d", len(tools))
	}

	expectedTools := []string{
		"enumerate_columns",
		"enumerate_rows",
		"get_cell_value",
		"get_range_values",
		"list_sheets",
		"set_current_sheet",
		"get_column",
		"get_row",
	}

	for i, expectedName := range expectedTools {
		if i >= len(tools) {
			t.Errorf("Missing tool: %s", expectedName)
			continue
		}

		tool := tools[i]
		if tool.Name != expectedName {
			t.Errorf("Expected tool name '%s' at index %d, got '%s'", expectedName, i, tool.Name)
		}

		if tool.Description == "" {
			t.Errorf("Tool '%s' has empty description", expectedName)
		}
	}
}

func TestToolNamesAndDescriptions(t *testing.T) {
	tools := GetToolDefinitions()

	// Test specific tools have expected names and non-empty descriptions
	testCases := []struct {
		index       int
		name        string
		description string
	}{
		{0, "enumerate_columns", "Enumerate all columns in an Excel spreadsheet"},
		{1, "enumerate_rows", "Enumerate all rows in an Excel spreadsheet"},
		{2, "get_cell_value", "Get the value of a specific cell in an Excel spreadsheet"},
		{3, "get_range_values", "Get values from a range of cells in an Excel spreadsheet"},
		{4, "list_sheets", "List all available sheets in an Excel spreadsheet"},
		{5, "set_current_sheet", "Set the current active sheet for subsequent operations on a file"},
		{6, "get_column", "Get all values in a specific column from an Excel spreadsheet"},
		{7, "get_row", "Get all values in a specific row from an Excel spreadsheet"},
	}

	for _, tc := range testCases {
		if tc.index >= len(tools) {
			t.Errorf("Tool at index %d not found", tc.index)
			continue
		}

		tool := tools[tc.index]
		if tool.Name != tc.name {
			t.Errorf("Expected tool name '%s' at index %d, got '%s'", tc.name, tc.index, tool.Name)
		}

		if tool.Description != tc.description {
			t.Errorf("Expected tool description '%s' for '%s', got '%s'", tc.description, tc.name, tool.Description)
		}
	}
}
