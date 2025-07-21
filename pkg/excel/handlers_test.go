package excel

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/xuri/excelize/v2"
)

func createTestExcelFileForHandlers(t *testing.T) string {
	// Create a temporary Excel file for testing
	file := excelize.NewFile()
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("Failed to close test file: %v", err)
		}
	}()

	// Add some test data
	file.SetCellValue("Sheet1", "A1", "Name")
	file.SetCellValue("Sheet1", "B1", "Age")
	file.SetCellValue("Sheet1", "C1", "City")
	file.SetCellValue("Sheet1", "A2", "John")
	file.SetCellValue("Sheet1", "B2", 30)
	file.SetCellValue("Sheet1", "C2", "New York")

	// Create a second sheet
	file.NewSheet("Sheet2")
	file.SetCellValue("Sheet2", "A1", "Product")
	file.SetCellValue("Sheet2", "B1", "Price")

	// Save to temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.xlsx")
	if err := file.SaveAs(filePath); err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}

	return filePath
}

func TestNewHandlers(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)

	if handlers == nil {
		t.Fatal("NewHandlers returned nil")
	}
	if handlers.excelManager != manager {
		t.Error("Excel manager not properly set")
	}
}

func TestEnumerateColumns(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)
	filePath := createTestExcelFileForHandlers(t)
	ctx := context.Background()

	// Test with valid file
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
			},
		},
	}

	result, err := handlers.EnumerateColumns(ctx, request)
	if err != nil {
		t.Fatalf("EnumerateColumns failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Errorf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Error("Expected success, got error")
		}
	}

	// Test with missing file_path
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}

	result, err = handlers.EnumerateColumns(ctx, request)
	if err != nil {
		t.Fatalf("EnumerateColumns failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing file_path")
	}
}

func TestEnumerateRows(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)
	filePath := createTestExcelFileForHandlers(t)
	ctx := context.Background()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
			},
		},
	}

	result, err := handlers.EnumerateRows(ctx, request)
	if err != nil {
		t.Fatalf("EnumerateRows failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Errorf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Error("Expected success, got error")
		}
	}
}

func TestHandlerGetCellValue(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)
	filePath := createTestExcelFileForHandlers(t)
	ctx := context.Background()

	// Test with valid parameters
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
				"cell":      "A1",
			},
		},
	}

	result, err := handlers.GetCellValue(ctx, request)
	if err != nil {
		t.Fatalf("GetCellValue failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Errorf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Error("Expected success, got error")
		}
	}

	// Test with missing cell parameter
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
			},
		},
	}

	result, err = handlers.GetCellValue(ctx, request)
	if err != nil {
		t.Fatalf("GetCellValue failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing cell parameter")
	}
}

func TestHandlerGetRangeValues(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)
	filePath := createTestExcelFileForHandlers(t)
	ctx := context.Background()

	// Test with valid parameters
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
				"range":     "A1:C2",
			},
		},
	}

	result, err := handlers.GetRangeValues(ctx, request)
	if err != nil {
		t.Fatalf("GetRangeValues failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Errorf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Error("Expected success, got error")
		}
	}

	// Test with missing range parameter
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
			},
		},
	}

	result, err = handlers.GetRangeValues(ctx, request)
	if err != nil {
		t.Fatalf("GetRangeValues failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing range parameter")
	}
}

func TestListSheets(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)
	filePath := createTestExcelFileForHandlers(t)
	ctx := context.Background()

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
			},
		},
	}

	result, err := handlers.ListSheets(ctx, request)
	if err != nil {
		t.Fatalf("ListSheets failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Errorf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Error("Expected success, got error")
		}
	}
}

func TestHandlerSetCurrentSheet(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)
	filePath := createTestExcelFileForHandlers(t)
	ctx := context.Background()

	// Test with valid parameters
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path":  filePath,
				"sheet_name": "Sheet2",
			},
		},
	}

	result, err := handlers.SetCurrentSheet(ctx, request)
	if err != nil {
		t.Fatalf("SetCurrentSheet failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Errorf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Error("Expected success, got error")
		}
	}

	// Test with invalid sheet name
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path":  filePath,
				"sheet_name": "NonExistentSheet",
			},
		},
	}

	result, err = handlers.SetCurrentSheet(ctx, request)
	if err != nil {
		t.Fatalf("SetCurrentSheet failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for invalid sheet name")
	}

	// Test with missing sheet_name parameter
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
			},
		},
	}

	result, err = handlers.SetCurrentSheet(ctx, request)
	if err != nil {
		t.Fatalf("SetCurrentSheet failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing sheet_name parameter")
	}
}

func TestGetColumn(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)
	filePath := createTestExcelFileForHandlers(t)
	ctx := context.Background()

	// Test with valid parameters
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
				"column":    "A",
			},
		},
	}

	result, err := handlers.GetColumn(ctx, request)
	if err != nil {
		t.Fatalf("GetColumn failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Errorf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Error("Expected success, got error")
		}
	}

	// Test with missing column parameter
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
			},
		},
	}

	result, err = handlers.GetColumn(ctx, request)
	if err != nil {
		t.Fatalf("GetColumn failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing column parameter")
	}

	// Test with missing file_path
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"column": "A",
			},
		},
	}

	result, err = handlers.GetColumn(ctx, request)
	if err != nil {
		t.Fatalf("GetColumn failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing file_path")
	}
}

func TestGetRow(t *testing.T) {
	manager := NewManager()
	handlers := NewHandlers(manager)
	filePath := createTestExcelFileForHandlers(t)
	ctx := context.Background()

	// Test with valid parameters
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path":  filePath,
				"row_number": float64(1), // JSON numbers are float64
			},
		},
	}

	result, err := handlers.GetRow(ctx, request)
	if err != nil {
		t.Fatalf("GetRow failed: %v", err)
	}

	if result.IsError {
		if textContent, ok := mcp.AsTextContent(result.Content[0]); ok {
			t.Errorf("Expected success, got error: %s", textContent.Text)
		} else {
			t.Error("Expected success, got error")
		}
	}

	// Test with missing row_number parameter
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": filePath,
			},
		},
	}

	result, err = handlers.GetRow(ctx, request)
	if err != nil {
		t.Fatalf("GetRow failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing row_number parameter")
	}

	// Test with missing file_path
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"row_number": float64(1),
			},
		},
	}

	result, err = handlers.GetRow(ctx, request)
	if err != nil {
		t.Fatalf("GetRow failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing file_path")
	}

	// Test with invalid row number (0)
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path":  filePath,
				"row_number": float64(0),
			},
		},
	}

	result, err = handlers.GetRow(ctx, request)
	if err != nil {
		t.Fatalf("GetRow failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for row_number 0")
	}

	// Test with row number that doesn't exist
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path":  filePath,
				"row_number": float64(100),
			},
		},
	}

	result, err = handlers.GetRow(ctx, request)
	if err != nil {
		t.Fatalf("GetRow failed: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for non-existent row number")
	}
}
