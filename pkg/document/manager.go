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

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

type DocumentInfo struct {
	FilePath    string
	FileSize    int64
	ModTime     time.Time
	Extension   string
	IsSupported bool
}

func (m *Manager) ExtractText(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return m.extractPDFText(filePath)
	case ".docx":
		return m.extractDocxText(filePath)
	case ".pptx":
		return m.extractPptxText(filePath)
	case ".doc":
		return "", fmt.Errorf("DOC files are not yet supported, please convert to DOCX format")
	case ".ppt":
		return "", fmt.Errorf("PPT files are not yet supported, please convert to PPTX format")
	default:
		return "", fmt.Errorf("unsupported file format: %s", ext)
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

	var text strings.Builder
	totalPages := reader.NumPage()

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
	xmlPattern := regexp.MustCompile(`<[^>]*>`)
	text = xmlPattern.ReplaceAllString(text, "")

	// Remove any remaining control characters
	controlPattern := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	text = controlPattern.ReplaceAllString(text, "")

	// Remove excessive whitespace (after removing XML tags)
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

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
