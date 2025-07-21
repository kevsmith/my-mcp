package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	case ".doc":
		return "", fmt.Errorf("DOC files are not yet supported, please convert to DOCX format")
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
	isSupported := ext == ".pdf" || ext == ".docx" || ext == ".doc"

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
	text := document.GetContent()

	return strings.TrimSpace(text), nil
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
