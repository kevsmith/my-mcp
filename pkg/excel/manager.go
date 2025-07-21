package excel

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Manager handles Excel file operations and maintains file state
type Manager struct {
	files        map[string]*excelize.File
	currentSheet map[string]string
}

// NewManager creates a new Excel file manager
func NewManager() *Manager {
	return &Manager{
		files:        make(map[string]*excelize.File),
		currentSheet: make(map[string]string),
	}
}

// OpenFile opens an Excel file and caches it for future operations
func (m *Manager) OpenFile(filePath string) (*excelize.File, error) {
	if file, exists := m.files[filePath]; exists {
		return file, nil
	}

	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	m.files[filePath] = file
	return file, nil
}

// GetCurrentSheet returns the current sheet for a file, or the first sheet if none is set
func (m *Manager) GetCurrentSheet(filePath string, file *excelize.File) (string, error) {
	if currentSheet, exists := m.currentSheet[filePath]; exists {
		return currentSheet, nil
	}

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return "", fmt.Errorf("no sheets found in file")
	}

	return sheets[0], nil
}

// SetCurrentSheet sets the current active sheet for a file
func (m *Manager) SetCurrentSheet(filePath, sheetName string) error {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	sheets := file.GetSheetList()
	found := false
	for _, sheet := range sheets {
		if sheet == sheetName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("sheet '%s' not found. Available sheets: %v", sheetName, sheets)
	}

	m.currentSheet[filePath] = sheetName
	return nil
}

// GetColumns returns all column names for a sheet
func (m *Manager) GetColumns(filePath, sheetName string) ([]string, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	if sheetName == "" {
		sheetName, err = m.GetCurrentSheet(filePath, file)
		if err != nil {
			return nil, err
		}
	}

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %v", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows found in spreadsheet")
	}

	columns := make([]string, len(rows[0]))
	for i := range rows[0] {
		col, _ := excelize.ColumnNumberToName(i + 1)
		columns[i] = col
	}

	return columns, nil
}

// GetRowCount returns the number of rows in a sheet
func (m *Manager) GetRowCount(filePath, sheetName string) (int, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %v", err)
	}

	if sheetName == "" {
		sheetName, err = m.GetCurrentSheet(filePath, file)
		if err != nil {
			return 0, err
		}
	}

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to get rows: %v", err)
	}

	return len(rows), nil
}

// GetCellValue returns the value of a specific cell
func (m *Manager) GetCellValue(filePath, cell, sheetName string) (string, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}

	if sheetName == "" {
		sheetName, err = m.GetCurrentSheet(filePath, file)
		if err != nil {
			return "", err
		}
	}

	value, err := file.GetCellValue(sheetName, cell)
	if err != nil {
		return "", fmt.Errorf("failed to get cell value: %v", err)
	}

	return value, nil
}

// GetRangeValues returns values from a range of cells
func (m *Manager) GetRangeValues(filePath, rangeRef, sheetName string) ([][]string, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	if sheetName == "" {
		sheetName, err = m.GetCurrentSheet(filePath, file)
		if err != nil {
			return nil, err
		}
	}

	rangeParts := strings.Split(rangeRef, ":")
	if len(rangeParts) != 2 {
		return nil, fmt.Errorf("invalid range format, expected 'A1:C3'")
	}

	startCol, startRow, err := excelize.CellNameToCoordinates(rangeParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid start cell: %v", err)
	}

	endCol, endRow, err := excelize.CellNameToCoordinates(rangeParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid end cell: %v", err)
	}

	var values [][]string
	for row := startRow; row <= endRow; row++ {
		var rowValues []string
		for col := startCol; col <= endCol; col++ {
			cellName, _ := excelize.CoordinatesToCellName(col, row)
			value, _ := file.GetCellValue(sheetName, cellName)
			rowValues = append(rowValues, value)
		}
		values = append(values, rowValues)
	}

	return values, nil
}

// GetSheetList returns all available sheets in a file
func (m *Manager) GetSheetList(filePath string) ([]string, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in file")
	}

	return sheets, nil
}

// GetColumnValues returns all values in a specific column
func (m *Manager) GetColumnValues(filePath, column, sheetName string) ([]string, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	if sheetName == "" {
		sheetName, err = m.GetCurrentSheet(filePath, file)
		if err != nil {
			return nil, err
		}
	}

	colNum, err := excelize.ColumnNameToNumber(column)
	if err != nil {
		return nil, fmt.Errorf("invalid column name '%s': %v", column, err)
	}

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %v", err)
	}

	var values []string
	for _, row := range rows {
		if colNum <= len(row) {
			values = append(values, row[colNum-1])
		} else {
			values = append(values, "")
		}
	}

	return values, nil
}

// GetRowValues returns all values in a specific row
func (m *Manager) GetRowValues(filePath string, rowNum int, sheetName string) ([]string, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	if sheetName == "" {
		sheetName, err = m.GetCurrentSheet(filePath, file)
		if err != nil {
			return nil, err
		}
	}

	if rowNum < 1 {
		return nil, fmt.Errorf("row number must be greater than 0")
	}

	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %v", err)
	}

	if rowNum > len(rows) {
		return nil, fmt.Errorf("row %d does not exist (sheet has %d rows)", rowNum, len(rows))
	}

	return rows[rowNum-1], nil
}
