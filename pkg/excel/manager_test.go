package excel

import (
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func createTestExcelFile(t *testing.T) string {
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
	file.SetCellValue("Sheet1", "A3", "Jane")
	file.SetCellValue("Sheet1", "B3", 25)
	file.SetCellValue("Sheet1", "C3", "Los Angeles")

	// Create a second sheet
	file.NewSheet("Sheet2")
	file.SetCellValue("Sheet2", "A1", "Product")
	file.SetCellValue("Sheet2", "B1", "Price")
	file.SetCellValue("Sheet2", "A2", "Laptop")
	file.SetCellValue("Sheet2", "B2", 999.99)

	// Save to temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.xlsx")
	if err := file.SaveAs(filePath); err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}

	return filePath
}

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.cache == nil {
		t.Error("cache not initialized")
	}
	if manager.currentSheet == nil {
		t.Error("currentSheet map not initialized")
	}
	if manager.cleanupTicker == nil {
		t.Error("cleanup ticker not started")
	}

	// Clean up
	manager.Close()
}

func TestOpenFile(t *testing.T) {
	manager := NewManager()
	filePath := createTestExcelFile(t)

	// Test opening a valid file
	file, err := manager.OpenFile(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	if file == nil {
		t.Fatal("OpenFile returned nil file")
	}

	// Test that file is cached
	file2, err := manager.OpenFile(filePath)
	if err != nil {
		t.Fatalf("Failed to open cached file: %v", err)
	}
	if file != file2 {
		t.Error("File not properly cached")
	}

	// Test opening non-existent file
	_, err = manager.OpenFile("nonexistent.xlsx")
	if err == nil {
		t.Error("Expected error when opening non-existent file")
	}
}

func TestGetCurrentSheet(t *testing.T) {
	manager := NewManager()
	filePath := createTestExcelFile(t)
	file, _ := manager.OpenFile(filePath)

	// Test getting current sheet when none is set (should return first sheet)
	sheet, err := manager.GetCurrentSheet(filePath, file)
	if err != nil {
		t.Fatalf("Failed to get current sheet: %v", err)
	}
	if sheet != "Sheet1" {
		t.Errorf("Expected 'Sheet1', got '%s'", sheet)
	}

	// Test after setting current sheet
	manager.SetCurrentSheet(filePath, "Sheet2")
	sheet, err = manager.GetCurrentSheet(filePath, file)
	if err != nil {
		t.Fatalf("Failed to get current sheet: %v", err)
	}
	if sheet != "Sheet2" {
		t.Errorf("Expected 'Sheet2', got '%s'", sheet)
	}
}

func TestSetCurrentSheet(t *testing.T) {
	manager := NewManager()
	filePath := createTestExcelFile(t)

	// Test setting valid sheet
	err := manager.SetCurrentSheet(filePath, "Sheet2")
	if err != nil {
		t.Fatalf("Failed to set current sheet: %v", err)
	}

	// Test setting invalid sheet
	err = manager.SetCurrentSheet(filePath, "NonExistentSheet")
	if err == nil {
		t.Error("Expected error when setting non-existent sheet")
	}
}

func TestGetColumns(t *testing.T) {
	manager := NewManager()
	filePath := createTestExcelFile(t)

	columns, err := manager.GetColumns(filePath, "Sheet1")
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	expected := []string{"A", "B", "C"}
	if len(columns) != len(expected) {
		t.Errorf("Expected %d columns, got %d", len(expected), len(columns))
	}

	for i, col := range expected {
		if i >= len(columns) || columns[i] != col {
			t.Errorf("Expected column %s at index %d, got %s", col, i, columns[i])
		}
	}
}

func TestGetRowCount(t *testing.T) {
	manager := NewManager()
	filePath := createTestExcelFile(t)

	count, err := manager.GetRowCount(filePath, "Sheet1")
	if err != nil {
		t.Fatalf("Failed to get row count: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 rows, got %d", count)
	}
}

func TestGetCellValue(t *testing.T) {
	manager := NewManager()
	filePath := createTestExcelFile(t)

	value, err := manager.GetCellValue(filePath, "A1", "Sheet1")
	if err != nil {
		t.Fatalf("Failed to get cell value: %v", err)
	}

	if value != "Name" {
		t.Errorf("Expected 'Name', got '%s'", value)
	}
}

func TestGetRangeValues(t *testing.T) {
	manager := NewManager()
	filePath := createTestExcelFile(t)

	values, err := manager.GetRangeValues(filePath, "A1:C2", "Sheet1")
	if err != nil {
		t.Fatalf("Failed to get range values: %v", err)
	}

	if len(values) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(values))
	}

	if len(values[0]) != 3 {
		t.Errorf("Expected 3 columns in first row, got %d", len(values[0]))
	}

	if values[0][0] != "Name" {
		t.Errorf("Expected 'Name' at [0][0], got '%s'", values[0][0])
	}
}

func TestGetSheetList(t *testing.T) {
	manager := NewManager()
	filePath := createTestExcelFile(t)

	sheets, err := manager.GetSheetList(filePath)
	if err != nil {
		t.Fatalf("Failed to get sheet list: %v", err)
	}

	if len(sheets) != 2 {
		t.Errorf("Expected 2 sheets, got %d", len(sheets))
	}

	expectedSheets := []string{"Sheet1", "Sheet2"}
	for i, expected := range expectedSheets {
		if i >= len(sheets) || sheets[i] != expected {
			t.Errorf("Expected sheet '%s' at index %d, got '%s'", expected, i, sheets[i])
		}
	}
}
