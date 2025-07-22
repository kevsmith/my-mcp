package excel

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Manager handles Excel file operations and maintains file state
type Manager struct {
	cache         *FileCache
	currentSheet  map[string]string
	cleanupTicker *time.Ticker
}

// NewManager creates a new Excel file manager
func NewManager() *Manager {
	config := GetCacheConfig()
	cache := NewFileCache(config)

	manager := &Manager{
		cache:        cache,
		currentSheet: make(map[string]string),
	}

	// Start cleanup ticker to remove expired entries every minute
	manager.cleanupTicker = cache.StartCleanupTicker(time.Minute)

	return manager
}

// NewManagerWithConfig creates a new Excel file manager with custom cache config
func NewManagerWithConfig(config CacheConfig) *Manager {
	cache := NewFileCache(config)

	manager := &Manager{
		cache:        cache,
		currentSheet: make(map[string]string),
	}

	// Start cleanup ticker to remove expired entries every minute
	manager.cleanupTicker = cache.StartCleanupTicker(time.Minute)

	return manager
}

// Close closes the manager and cleans up resources
func (m *Manager) Close() {
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}
	if m.cache != nil {
		m.cache.Clear()
	}
}

// FlushCache flushes the file cache and returns cache statistics
func (m *Manager) FlushCache() (int, error) {
	if m.cache == nil {
		return 0, fmt.Errorf("cache not initialized")
	}

	// Get current cache size before clearing
	cacheSize := m.cache.Size()

	// Clear the cache (closes all files)
	m.cache.Clear()

	// Also clear current sheet mappings since files are closed
	m.currentSheet = make(map[string]string)

	return cacheSize, nil
}

// ExplainFormulas extracts and explains all formulas from all sheets
func (m *Manager) ExplainFormulas(filePath string) ([]FormulaInfo, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	extractor := NewFormulaExtractor(file)
	return extractor.ExtractFormulas()
}

// ExplainFormulasFromSheet extracts and explains formulas from a specific sheet
func (m *Manager) ExplainFormulasFromSheet(filePath, sheetName string) ([]FormulaInfo, error) {
	file, err := m.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	extractor := NewFormulaExtractor(file)
	return extractor.ExtractFormulasFromSheet(sheetName)
}

// ExplainFormula extracts and explains a formula from a specific cell
func (m *Manager) ExplainFormula(filePath, cell, sheetName string) (*FormulaInfo, error) {
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

	// Get the formula from the cell
	formula, err := file.GetCellFormula(sheetName, cell)
	if err != nil {
		return nil, fmt.Errorf("failed to get cell formula: %v", err)
	}

	if formula == "" {
		return nil, fmt.Errorf("cell %s does not contain a formula", cell)
	}

	// Get the cell value
	value, err := file.GetCellValue(sheetName, cell)
	if err != nil {
		value = ""
	}

	// Create extractor and translate the formula
	extractor := NewFormulaExtractor(file)
	translatedFormula := extractor.translateFormula(sheetName, formula)
	label := extractor.getCellLabel(sheetName, cell)

	return &FormulaInfo{
		Sheet:             sheetName,
		Cell:              cell,
		Formula:           formula,
		Value:             value,
		TranslatedFormula: translatedFormula,
		Label:             label,
	}, nil
}

// OpenFile opens an Excel file and caches it for future operations
func (m *Manager) OpenFile(filePath string) (*excelize.File, error) {
	// Try to get from cache first
	if file, found := m.cache.Get(filePath); found {
		return file, nil
	}

	// Open the file
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	// Store in cache
	m.cache.Put(filePath, file)
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

	// Pre-allocate slices with known capacity for performance
	rowCount := endRow - startRow + 1
	colCount := endCol - startCol + 1
	values := make([][]string, 0, rowCount)
	
	for row := startRow; row <= endRow; row++ {
		rowValues := make([]string, 0, colCount)
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

	// Pre-allocate slice with known capacity
	values := make([]string, 0, len(rows))
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

// SheetStats represents statistical information about an Excel sheet
type SheetStats struct {
	RowCount      int            `json:"row_count"`
	ColumnCount   int            `json:"column_count"`
	NonEmptyRows  int            `json:"non_empty_rows"`
	NonEmptyCells int            `json:"non_empty_cells"`
	DataTypes     map[string]int `json:"data_types"`
	FirstDataRow  int            `json:"first_data_row"`
	LastDataRow   int            `json:"last_data_row"`
	FirstDataCol  string         `json:"first_data_col"`
	LastDataCol   string         `json:"last_data_col"`
}

// GetSheetStats returns statistical information about a sheet
func (m *Manager) GetSheetStats(filePath, sheetName string) (*SheetStats, error) {
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

	stats := &SheetStats{
		RowCount:      len(rows),
		ColumnCount:   0,
		NonEmptyRows:  0,
		NonEmptyCells: 0,
		DataTypes:     make(map[string]int),
		FirstDataRow:  0,
		LastDataRow:   0,
		FirstDataCol:  "",
		LastDataCol:   "",
	}

	if len(rows) == 0 {
		return stats, nil
	}

	maxColumns := 0
	firstDataRowFound := false
	var firstDataCol, lastDataCol int

	for rowIdx, row := range rows {
		if len(row) > maxColumns {
			maxColumns = len(row)
		}

		hasData := false
		rowFirstCol, rowLastCol := -1, -1

		for colIdx, cell := range row {
			if strings.TrimSpace(cell) != "" {
				stats.NonEmptyCells++
				hasData = true

				if rowFirstCol == -1 {
					rowFirstCol = colIdx
				}
				rowLastCol = colIdx

				if !firstDataRowFound {
					stats.FirstDataRow = rowIdx + 1
					firstDataRowFound = true
				}
				stats.LastDataRow = rowIdx + 1

				dataType := classifyDataType(cell)
				stats.DataTypes[dataType]++
			}
		}

		if hasData {
			stats.NonEmptyRows++

			if firstDataCol == 0 || (rowFirstCol >= 0 && rowFirstCol < firstDataCol) {
				firstDataCol = rowFirstCol
			}
			if rowLastCol > lastDataCol {
				lastDataCol = rowLastCol
			}
		}
	}

	stats.ColumnCount = maxColumns

	if firstDataCol >= 0 {
		stats.FirstDataCol, _ = excelize.ColumnNumberToName(firstDataCol + 1)
	}
	if lastDataCol >= 0 {
		stats.LastDataCol, _ = excelize.ColumnNumberToName(lastDataCol + 1)
	}

	return stats, nil
}

// classifyDataType determines the data type of a cell value
func classifyDataType(value string) string {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return "empty"
	}

	if _, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return "integer"
	}

	if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return "number"
	}

	if strings.ToLower(trimmed) == "true" || strings.ToLower(trimmed) == "false" {
		return "boolean"
	}

	if len(trimmed) >= 8 && (strings.Contains(trimmed, "/") || strings.Contains(trimmed, "-")) {
		return "date"
	}

	return "text"
}
