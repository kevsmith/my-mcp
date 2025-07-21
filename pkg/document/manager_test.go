package document

import (
	"os"
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

func TestExtractText_PptFormat(t *testing.T) {
	manager := NewManager()
	_, err := manager.ExtractText("test.ppt")
	if err == nil {
		t.Fatal("Expected error for PPT format")
	}
	expectedMsg := "PPT files are not yet supported, please convert to PPTX format"
	if err.Error() != expectedMsg {
		t.Fatalf("Unexpected error message: %s", err.Error())
	}
}

func TestGetDocumentInfo_PptxSupported(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test*.pptx")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	manager := NewManager()
	info, err := manager.GetDocumentInfo(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to get document info: %v", err)
	}

	if info.Extension != ".pptx" {
		t.Errorf("Expected .pptx extension, got %s", info.Extension)
	}

	if !info.IsSupported {
		t.Error("Expected PPTX to be supported")
	}
}

func TestExtractCleanTextFromXML(t *testing.T) {
	manager := NewManager()

	testCases := []struct {
		name     string
		xmlInput string
		expected string
	}{
		{
			name:     "Simple text with XML tags",
			xmlInput: `<w:p><w:r><w:t>Hello World</w:t></w:r></w:p>`,
			expected: "Hello World",
		},
		{
			name:     "Multiple paragraphs",
			xmlInput: `<w:p><w:r><w:t>First paragraph.</w:t></w:r></w:p><w:p><w:r><w:t>Second paragraph.</w:t></w:r></w:p>`,
			expected: "First paragraph. Second paragraph.",
		},
		{
			name:     "Text with attributes and complex XML",
			xmlInput: `<w:p w:rsidR="00123456"><w:r><w:rPr><w:b/></w:rPr><w:t>Bold text</w:t></w:r><w:r><w:t> and normal text.</w:t></w:r></w:p>`,
			expected: "Bold text and normal text.",
		},
		{
			name:     "Empty and whitespace-only elements",
			xmlInput: `<w:p><w:r><w:t>   </w:t></w:r></w:p><w:p><w:r><w:t>Real text</w:t></w:r></w:p><w:p><w:r><w:t>  </w:t></w:r></w:p>`,
			expected: "Real text",
		},
		{
			name:     "Text with excessive whitespace",
			xmlInput: `<w:p><w:r><w:t>Text    with     multiple     spaces</w:t></w:r></w:p>`,
			expected: "Text with multiple spaces",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := manager.extractCleanTextFromXML(tc.xmlInput)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Expected: %q, Got: %q", tc.expected, result)
			}
		})
	}
}

func TestCleanExtractedText(t *testing.T) {
	manager := NewManager()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Multiple spaces",
			input:    "Text   with    multiple   spaces",
			expected: "Text with multiple spaces",
		},
		{
			name:     "Remaining XML tags",
			input:    "Text with <remaining>tags</remaining> here",
			expected: "Text with tags here",
		},
		{
			name:     "Control characters",
			input:    "Text\x00with\x08control\x1Fcharacters",
			expected: "Textwithcontrolcharacters",
		},
		{
			name:     "Mixed whitespace and tags",
			input:    "  Text  <tag>  with  </tag>  mixed   issues  ",
			expected: "Text with mixed issues",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := manager.cleanExtractedText(tc.input)
			if result != tc.expected {
				t.Errorf("Expected: %q, Got: %q", tc.expected, result)
			}
		})
	}
}
