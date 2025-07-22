package excel

import (
	"context"
	"fmt"

	"github.com/kevsmith/my-mcp/pkg/shared"
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
	return h.withMiddleware(h.getCellValueHandler)(ctx, request)
}

// getCellValueHandler is the optimized implementation using middleware
func (h *Handlers) getCellValueHandler(ctx context.Context, hctx *HandlerContext) (*mcp.CallToolResult, error) {
	// Validate cell parameter
	cell, errResult := ValidateRequiredParamWithExample(hctx, "cell", "A1")
	if errResult != nil {
		return errResult, nil
	}

	// Get cell value using cached file and resolved sheet
	value, err := hctx.Manager.GetCellValue(hctx.FilePath, cell, hctx.SheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Return formatted response
	return NewFormattedTextResponse("Cell %s in sheet %s: %s", cell, hctx.SheetName, value)
}

// GetRangeValues handles the get_range_values tool
func (h *Handlers) GetRangeValues(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.withMiddleware(h.getRangeValuesHandler)(ctx, request)
}

// getRangeValuesHandler is the optimized implementation using middleware
func (h *Handlers) getRangeValuesHandler(ctx context.Context, hctx *HandlerContext) (*mcp.CallToolResult, error) {
	// Validate range parameter
	rangeRef, errResult := ValidateRequiredParamWithExample(hctx, "range", "A1:C3")
	if errResult != nil {
		return errResult, nil
	}

	// Get range values using cached file and resolved sheet
	values, err := hctx.Manager.GetRangeValues(hctx.FilePath, rangeRef, hctx.SheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Return formatted response
	return NewFormattedTextResponse("Range %s in sheet %s contains %d rows", rangeRef, hctx.SheetName, len(values))
}

// ListSheets handles the list_sheets tool
func (h *Handlers) ListSheets(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.withMiddlewareNoSheet(h.listSheetsHandler)(ctx, request)
}

// listSheetsHandler is the optimized implementation using middleware
func (h *Handlers) listSheetsHandler(ctx context.Context, hctx *HandlerContext) (*mcp.CallToolResult, error) {
	// Get sheets using cached file
	sheets, err := hctx.Manager.GetSheetList(hctx.FilePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Return formatted response
	return NewFormattedTextResponse("Available sheets: %v", sheets)
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

// GetSheetStats handles the get_sheet_stats tool
func (h *Handlers) GetSheetStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	sheetName := request.GetString("sheet_name", "")

	stats, err := h.excelManager.GetSheetStats(filePath, sheetName)
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

	// Format the response as JSON for better readability using optimized marshaling
	statsJSON, err := shared.OptimizedMarshalIndent(stats, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format stats: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Sheet '%s' statistics:\n%s", sheetName, string(statsJSON))), nil
}

// FlushCache handles the flush_cache tool
func (h *Handlers) FlushCache(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filesCleared, err := h.excelManager.FlushCache()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Cache flushed successfully. %d files were closed and removed from cache.", filesCleared)), nil
}

// ExplainFormula handles the explain_formula tool
func (h *Handlers) ExplainFormula(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	cell := request.GetString("cell", "")
	if cell == "" {
		return mcp.NewToolResultError("cell parameter is required (e.g., 'A1')"), nil
	}

	sheetName := request.GetString("sheet_name", "")

	formula, err := h.excelManager.ExplainFormula(filePath, cell, sheetName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Format the response as JSON for better readability using optimized marshaling
	formulaJSON, err := shared.OptimizedMarshalIndent(formula, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format formula: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Formula explanation for cell %s:\n%s", cell, string(formulaJSON))), nil
}
