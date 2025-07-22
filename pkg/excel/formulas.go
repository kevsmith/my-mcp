package excel

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Precompiled regex patterns for performance
var (
	cellRefRegex = regexp.MustCompile(`\$?[A-Z]+\$?\d+`)
)

// Constants for header search optimization
const (
	maxHeaderSearchDepth = 10 // Maximum rows/columns to search for headers
)

// FormulaInfo represents a formula with its translation and context
type FormulaInfo struct {
	Sheet             string `json:"sheet"`
	Cell              string `json:"cell"`
	Formula           string `json:"formula"`
	Value             string `json:"value"`
	TranslatedFormula string `json:"translated_formula"`
	Label             string `json:"label"`
}

// FormulaExtractor handles formula extraction and translation
type FormulaExtractor struct {
	file        *excelize.File
	headerCache map[string]map[string]string
}

// NewFormulaExtractor creates a new formula extractor for the given Excel file
func NewFormulaExtractor(file *excelize.File) *FormulaExtractor {
	return &FormulaExtractor{
		file:        file,
		headerCache: make(map[string]map[string]string),
	}
}

// ExtractFormulas extracts all formulas from all sheets with translations
func (fe *FormulaExtractor) ExtractFormulas() ([]FormulaInfo, error) {
	var formulas []FormulaInfo

	sheetList := fe.file.GetSheetList()
	for _, sheetName := range sheetList {
		sheetFormulas, err := fe.ExtractFormulasFromSheet(sheetName)
		if err != nil {
			return nil, err
		}
		formulas = append(formulas, sheetFormulas...)
	}

	return formulas, nil
}

// ExtractFormulasFromSheet extracts formulas from a specific sheet
func (fe *FormulaExtractor) ExtractFormulasFromSheet(sheetName string) ([]FormulaInfo, error) {
	var formulas []FormulaInfo

	rows, err := fe.file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows for sheet %s: %w", sheetName, err)
	}

	for rowIndex, row := range rows {
		for colIndex := range row {
			cellName, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				continue
			}

			formula, err := fe.file.GetCellFormula(sheetName, cellName)
			if err != nil {
				continue
			}

			if formula != "" {
				value, err := fe.file.GetCellValue(sheetName, cellName)
				if err != nil {
					value = ""
				}

				translatedFormula := fe.translateFormula(sheetName, formula)
				label := fe.getCellLabel(sheetName, cellName)

				formulas = append(formulas, FormulaInfo{
					Sheet:             sheetName,
					Cell:              cellName,
					Formula:           formula,
					Value:             value,
					TranslatedFormula: translatedFormula,
					Label:             label,
				})
			}
		}
	}

	return formulas, nil
}

// ExtractFormulasFromRange extracts formulas from a specific range
func (fe *FormulaExtractor) ExtractFormulasFromRange(sheetName, rangeRef string) ([]FormulaInfo, error) {
	var formulas []FormulaInfo

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

	for row := startRow; row <= endRow; row++ {
		for col := startCol; col <= endCol; col++ {
			cellName, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				continue
			}

			formula, err := fe.file.GetCellFormula(sheetName, cellName)
			if err != nil {
				continue
			}

			if formula != "" {
				value, err := fe.file.GetCellValue(sheetName, cellName)
				if err != nil {
					value = ""
				}

				translatedFormula := fe.translateFormula(sheetName, formula)
				label := fe.getCellLabel(sheetName, cellName)

				formulas = append(formulas, FormulaInfo{
					Sheet:             sheetName,
					Cell:              cellName,
					Formula:           formula,
					Value:             value,
					TranslatedFormula: translatedFormula,
					Label:             label,
				})
			}
		}
	}

	return formulas, nil
}

// findColumnHeader searches upward from a given cell to find the column header
func (fe *FormulaExtractor) findColumnHeader(sheetName string, col, row int) string {
	key := fmt.Sprintf("%s:%d:%d", sheetName, col, row)
	if fe.headerCache[sheetName] == nil {
		fe.headerCache[sheetName] = make(map[string]string)
	}

	if header, exists := fe.headerCache[sheetName][key]; exists {
		return header
	}

	// Search upward with bounded depth
	searchStart := row - 1
	searchEnd := max(1, row-maxHeaderSearchDepth)
	for r := searchStart; r >= searchEnd; r-- {
		cellName, err := excelize.CoordinatesToCellName(col, r)
		if err != nil {
			continue
		}

		value, err := fe.file.GetCellValue(sheetName, cellName)
		if err != nil {
			continue
		}

		value = strings.TrimSpace(value)
		if value != "" && !isNumeric(value) {
			fe.headerCache[sheetName][key] = value
			return value
		}
	}

	fe.headerCache[sheetName][key] = ""
	return ""
}

// findRowHeader searches leftward from a given cell to find the row header
func (fe *FormulaExtractor) findRowHeader(sheetName string, col, row int) string {
	key := fmt.Sprintf("%s:r:%d:%d", sheetName, col, row)
	if fe.headerCache[sheetName] == nil {
		fe.headerCache[sheetName] = make(map[string]string)
	}

	if header, exists := fe.headerCache[sheetName][key]; exists {
		return header
	}

	// Search leftward with bounded depth
	searchStart := col - 1
	searchEnd := max(1, col-maxHeaderSearchDepth)
	for c := searchStart; c >= searchEnd; c-- {
		cellName, err := excelize.CoordinatesToCellName(c, row)
		if err != nil {
			continue
		}

		value, err := fe.file.GetCellValue(sheetName, cellName)
		if err != nil {
			continue
		}

		value = strings.TrimSpace(value)
		if value != "" && !isNumeric(value) {
			fe.headerCache[sheetName][key] = value
			return value
		}
	}

	fe.headerCache[sheetName][key] = ""
	return ""
}

// getCellLabel returns the human-readable label for a cell
func (fe *FormulaExtractor) getCellLabel(sheetName, cellName string) string {
	col, row, err := excelize.CellNameToCoordinates(cellName)
	if err != nil {
		return ""
	}

	colHeader := fe.findColumnHeader(sheetName, col, row)
	if colHeader != "" {
		return colHeader
	}

	rowHeader := fe.findRowHeader(sheetName, col, row)
	return rowHeader
}

// translateFormula translates cell references in a formula to human-readable names
func (fe *FormulaExtractor) translateFormula(sheetName, formula string) string {
	return cellRefRegex.ReplaceAllStringFunc(formula, func(cellRef string) string {
		cleanRef := strings.ReplaceAll(strings.ReplaceAll(cellRef, "$", ""), " ", "")
		header := fe.getCellHeaderByReference(sheetName, cleanRef)
		if header != "" {
			return header
		}

		return cellRef
	})
}

// getCellHeaderByReference gets the header for a cell reference
func (fe *FormulaExtractor) getCellHeaderByReference(sheetName, cellRef string) string {
	col, row, err := excelize.CellNameToCoordinates(cellRef)
	if err != nil {
		return ""
	}

	header := fe.findColumnHeader(sheetName, col, row)
	if header != "" {
		return header
	}
	return fe.findRowHeader(sheetName, col, row)
}

// isNumeric checks if a string represents a numeric value
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// String returns a formatted string representation of FormulaInfo
func (fi FormulaInfo) String() string {
	result := fmt.Sprintf("%s!%s: %s = %s", fi.Sheet, fi.Cell, fi.Formula, fi.Value)
	if fi.Label != "" {
		result = fmt.Sprintf("%s (%s)", result, fi.Label)
	}
	if fi.TranslatedFormula != "" && fi.TranslatedFormula != fi.Formula {
		result = fmt.Sprintf("%s | Translated: %s", result, fi.TranslatedFormula)
	}
	return result
}

// FilterFormulas filters formulas based on a predicate function
func FilterFormulas(formulas []FormulaInfo, filter func(FormulaInfo) bool) []FormulaInfo {
	var filtered []FormulaInfo
	for _, formula := range formulas {
		if filter(formula) {
			filtered = append(filtered, formula)
		}
	}
	return filtered
}

// GetTranslatedFormulas returns only formulas that were successfully translated
func GetTranslatedFormulas(formulas []FormulaInfo) []FormulaInfo {
	return FilterFormulas(formulas, func(fi FormulaInfo) bool {
		return fi.TranslatedFormula != "" && fi.TranslatedFormula != fi.Formula
	})
}

// GetLabeledFormulas returns only formulas that have labels
func GetLabeledFormulas(formulas []FormulaInfo) []FormulaInfo {
	return FilterFormulas(formulas, func(fi FormulaInfo) bool {
		return fi.Label != ""
	})
}

// FindFormulasContaining finds formulas containing a specific substring
func FindFormulasContaining(formulas []FormulaInfo, substring string) []FormulaInfo {
	return FilterFormulas(formulas, func(fi FormulaInfo) bool {
		return strings.Contains(strings.ToUpper(fi.Formula), strings.ToUpper(substring))
	})
}
