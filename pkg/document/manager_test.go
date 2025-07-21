package document

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestGetDocumentInfo_NonExistentFile(t *testing.T) {
	manager := NewManager()
	_, err := manager.GetDocumentInfo("nonexistent.pdf")
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

func TestExtractText_UnsupportedFormat(t *testing.T) {
	manager := NewManager()
	_, err := manager.ExtractText("test.txt")
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if err.Error() != "unsupported file format: .txt" {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}
}

func TestExtractText_DocFormat(t *testing.T) {
	manager := NewManager()
	_, err := manager.ExtractText("test.doc")
	if err == nil {
		t.Fatal("Expected error for DOC format")
	}
	expectedMsg := "DOC files are not yet supported, please convert to DOCX format"
	if err.Error() != expectedMsg {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}
}
