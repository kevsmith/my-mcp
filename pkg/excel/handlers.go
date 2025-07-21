package excel

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// Handlers contains all MCP tool handlers
type Handlers struct {
	excelManager *Manager
}

// NewHandlers creates a new handlers instance
func NewHandlers(excelManager *Manager) *Handlers {
	return &Handlers{
		excelManager: excelManager,
	}
}

// EnumerateColumns handles the enumerate_columns tool
func (h *Handlers) EnumerateColumns(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	sheetName := request.GetString("sheet_name", "")

	columns, err := h.excelManager.GetColumns(filePath, sheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get the actual sheet name used
	file, err := h.excelManager.OpenFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if sheetName == "" {
		sheetName, err = h.excelManager.GetCurrentSheet(filePath, file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Columns: %v, Sheet: %s", columns, sheetName)), nil
}

// EnumerateRows handles the enumerate_rows tool
func (h *Handlers) EnumerateRows(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	sheetName := request.GetString("sheet_name", "")

	rowCount, err := h.excelManager.GetRowCount(filePath, sheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get the actual sheet name used
	file, err := h.excelManager.OpenFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if sheetName == "" {
		sheetName, err = h.excelManager.GetCurrentSheet(filePath, file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Rows: %d rows found in sheet %s", rowCount, sheetName)), nil
}

// GetCellValue handles the get_cell_value tool
func (h *Handlers) GetCellValue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	cell := request.GetString("cell", "")
	if cell == "" {
		return mcp.NewToolResultError("cell parameter is required (e.g., 'A1')"), nil
	}

	sheetName := request.GetString("sheet_name", "")

	value, err := h.excelManager.GetCellValue(filePath, cell, sheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get the actual sheet name used
	file, err := h.excelManager.OpenFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if sheetName == "" {
		sheetName, err = h.excelManager.GetCurrentSheet(filePath, file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Cell %s in sheet %s: %s", cell, sheetName, value)), nil
}

// GetRangeValues handles the get_range_values tool
func (h *Handlers) GetRangeValues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	rangeRef := request.GetString("range", "")
	if rangeRef == "" {
		return mcp.NewToolResultError("range parameter is required (e.g., 'A1:C3')"), nil
	}

	sheetName := request.GetString("sheet_name", "")

	values, err := h.excelManager.GetRangeValues(filePath, rangeRef, sheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get the actual sheet name used
	file, err := h.excelManager.OpenFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if sheetName == "" {
		sheetName, err = h.excelManager.GetCurrentSheet(filePath, file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Range %s in sheet %s contains %d rows", rangeRef, sheetName, len(values))), nil
}

// ListSheets handles the list_sheets tool
func (h *Handlers) ListSheets(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	sheets, err := h.excelManager.GetSheetList(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Available sheets: %v", sheets)), nil
}

// SetCurrentSheet handles the set_current_sheet tool
func (h *Handlers) SetCurrentSheet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	sheetName := request.GetString("sheet_name", "")
	if sheetName == "" {
		return mcp.NewToolResultError("sheet_name parameter is required"), nil
	}

	err := h.excelManager.SetCurrentSheet(filePath, sheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Current sheet set to '%s' for file %s", sheetName, filePath)), nil
}

// GetColumn handles the get_column tool
func (h *Handlers) GetColumn(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	column := request.GetString("column", "")
	if column == "" {
		return mcp.NewToolResultError("column parameter is required (e.g., 'A', 'B', 'Z')"), nil
	}

	sheetName := request.GetString("sheet_name", "")

	values, err := h.excelManager.GetColumnValues(filePath, column, sheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get the actual sheet name used
	file, err := h.excelManager.OpenFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if sheetName == "" {
		sheetName, err = h.excelManager.GetCurrentSheet(filePath, file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Column %s in sheet %s: %v", column, sheetName, values)), nil
}

// GetRow handles the get_row tool
func (h *Handlers) GetRow(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	rowNumber := request.GetInt("row_number", 0)
	if rowNumber == 0 {
		return mcp.NewToolResultError("row_number parameter is required (1-based)"), nil
	}

	sheetName := request.GetString("sheet_name", "")

	values, err := h.excelManager.GetRowValues(filePath, int(rowNumber), sheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get the actual sheet name used
	file, err := h.excelManager.OpenFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if sheetName == "" {
		sheetName, err = h.excelManager.GetCurrentSheet(filePath, file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Row %d in sheet %s: %v", int(rowNumber), sheetName, values)), nil
}
