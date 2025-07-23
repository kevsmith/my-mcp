package document

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"code.sajari.com/docconv"
	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

// Precompiled regex patterns for performance
var (
	xmlPattern        = regexp.MustCompile(`<[^>]*>`)
	controlPattern    = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	whitespacePattern = regexp.MustCompile(`\s+`)
)

// Magic number signatures for file format detection
var (
	pdfMagic = []byte{0x25, 0x50, 0x44, 0x46}             // %PDF
	zipMagic = []byte{0x50, 0x4B, 0x03, 0x04}             // PK.. (ZIP-based formats)
	oleDoc   = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1} // Old DOC/PPT files
)

// DocumentType represents the detected file type
type DocumentType int

const (
	DocumentTypeUnknown DocumentType = iota
	DocumentTypePDF
	DocumentTypeDOCX
	DocumentTypePPTX
	DocumentTypeDOC
	DocumentTypePPT
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

// detectFileType detects file type using magic numbers for better accuracy
func (m *Manager) detectFileType(filePath string) DocumentType {
	file, err := os.Open(filePath)
	if err != nil {
		return DocumentTypeUnknown
	}
	defer file.Close()

	// Read first 512 bytes for magic number detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil || n < 4 {
		return DocumentTypeUnknown
	}

	// Check for PDF magic number
	if len(buffer) >= len(pdfMagic) && bytesEqual(buffer[:len(pdfMagic)], pdfMagic) {
		return DocumentTypePDF
	}

	// Check for ZIP-based formats (DOCX, PPTX)
	if len(buffer) >= len(zipMagic) && bytesEqual(buffer[:len(zipMagic)], zipMagic) {
		// Differentiate between DOCX and PPTX by checking internal structure
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".docx":
			return DocumentTypeDOCX
		case ".pptx":
			return DocumentTypePPTX
		}
		return DocumentTypeUnknown
	}

	// Check for older Office formats (DOC, PPT)
	if len(buffer) >= len(oleDoc) && bytesEqual(buffer[:len(oleDoc)], oleDoc) {
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".doc":
			return DocumentTypeDOC
		case ".ppt":
			return DocumentTypePPT
		}
		return DocumentTypeUnknown
	}

	return DocumentTypeUnknown
}

// bytesEqual compares two byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type DocumentInfo struct {
	FilePath    string
	FileSize    int64
	ModTime     time.Time
	Extension   string
	IsSupported bool
}

func (m *Manager) ExtractText(filePath string) (string, error) {
	// Use magic number detection for more accurate file type identification
	docType := m.detectFileType(filePath)

	switch docType {
	case DocumentTypePDF:
		return m.extractPDFText(filePath)
	case DocumentTypeDOCX:
		return m.extractDocxText(filePath)
	case DocumentTypePPTX:
		return m.extractPptxText(filePath)
	case DocumentTypeDOC:
		return "", fmt.Errorf("DOC files are not yet supported, please convert to DOCX format")
	case DocumentTypePPT:
		return "", fmt.Errorf("PPT files are not yet supported, please convert to PPTX format")
	default:
		// Fall back to extension-based detection if magic number fails
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".pdf", ".docx", ".pptx":
			return "", fmt.Errorf("file appears to be corrupted or invalid %s format", ext)
		case ".doc":
			return "", fmt.Errorf("DOC files are not yet supported, please convert to DOCX format")
		case ".ppt":
			return "", fmt.Errorf("PPT files are not yet supported, please convert to PPTX format")
		default:
			return "", fmt.Errorf("unsupported file format: %s", ext)
		}
	}
}

func (m *Manager) GetDocumentInfo(filePath string) (*DocumentInfo, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	isSupported := ext == ".pdf" || ext == ".docx" || ext == ".pptx" || ext == ".doc" || ext == ".ppt"

	return &DocumentInfo{
		FilePath:    filePath,
		FileSize:    stat.Size(),
		ModTime:     stat.ModTime(),
		Extension:   ext,
		IsSupported: isSupported,
	}, nil
}

func (m *Manager) extractPDFText(filePath string) (string, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	totalPages := reader.NumPage()
	// Pre-allocate string builder with estimated capacity (avg 2KB per page)
	var text strings.Builder
	text.Grow(totalPages * 2048)

	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		page := reader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			continue // Skip pages that can't be read
		}

		text.WriteString(pageText)
		text.WriteString("\n")
	}

	return strings.TrimSpace(text.String()), nil
}

func (m *Manager) extractDocxText(filePath string) (string, error) {
	reader, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open DOCX file: %w", err)
	}
	defer reader.Close()

	document := reader.Editable()
	rawContent := document.GetContent()

	// Extract clean prose text from XML content
	cleanText, err := m.extractCleanTextFromXML(rawContent)
	if err != nil {
		return "", fmt.Errorf("failed to extract clean text: %w", err)
	}

	return strings.TrimSpace(cleanText), nil
}

func (m *Manager) extractPptxText(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PPTX file: %w", err)
	}
	defer file.Close()

	// Use docconv to extract text from PPTX
	plainText, _, err := docconv.ConvertPptx(file)
	if err != nil {
		return "", fmt.Errorf("failed to extract text from PPTX: %w", err)
	}

	// Clean the extracted text to ensure it's clean prose
	cleanText := m.cleanExtractedText(plainText)

	return strings.TrimSpace(cleanText), nil
}

// extractCleanTextFromXML parses XML content and extracts only the readable text
func (m *Manager) extractCleanTextFromXML(xmlContent string) (string, error) {
	var result strings.Builder
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))

	for {
		token, err := decoder.Token()
		if err != nil {
			break // End of document or error
		}

		switch t := token.(type) {
		case xml.CharData:
			// Extract character data (actual text content)
			text := strings.TrimSpace(string(t))
			if text != "" {
				if result.Len() > 0 {
					// Add space between text segments to maintain readability
					result.WriteString(" ")
				}
				result.WriteString(text)
			}
		}
	}

	text := result.String()

	// Clean up the extracted text
	text = m.cleanExtractedText(text)

	return text, nil
}

// cleanExtractedText performs additional cleanup on extracted text
func (m *Manager) cleanExtractedText(text string) string {
	// Clean up any remaining XML-like patterns first
	text = xmlPattern.ReplaceAllString(text, "")

	// Remove any remaining control characters
	text = controlPattern.ReplaceAllString(text, "")

	// Remove excessive whitespace (after removing XML tags)
	text = whitespacePattern.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

func (m *Manager) isDocFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	signature := make([]byte, 8)
	n, err := file.Read(signature)
	if err != nil || n < 8 {
		return false
	}

	return signature[0] == 0xD0 && signature[1] == 0xCF &&
		signature[2] == 0x11 && signature[3] == 0xE0
}
