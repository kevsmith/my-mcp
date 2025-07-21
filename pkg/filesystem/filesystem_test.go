package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "fs-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	subFile := filepath.Join(subDir, "sub.txt")
	if err := os.WriteFile(subFile, []byte("sub content"), 0644); err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestFilesystemHandler_ListDirectory(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	files, err := handler.ListDirectory(".")
	if err != nil {
		t.Fatalf("Failed to list directory: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}

	var foundDir, foundFile bool
	for _, file := range files {
		if file.Name == "subdir" && file.IsDir {
			foundDir = true
		}
		if file.Name == "test.txt" && !file.IsDir && file.Size == 12 {
			foundFile = true
		}
	}

	if !foundDir {
		t.Error("Did not find expected directory")
	}
	if !foundFile {
		t.Error("Did not find expected file")
	}
}

func TestFilesystemHandler_ListDirectory_PathTraversal(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	_, err := handler.ListDirectory("../")
	if err == nil {
		t.Error("Expected path traversal to be blocked")
	}
}

func TestFilesystemHandler_Glob(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	files, err := handler.Glob("*.txt")
	if err != nil {
		t.Fatalf("Failed to glob: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if files[0].Name != "test.txt" {
		t.Errorf("Expected test.txt, got %s", files[0].Name)
	}
}

func TestFilesystemHandler_Glob_Recursive(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	files, err := handler.Glob("**/*.txt")
	if err != nil {
		t.Fatalf("Failed to glob recursively: %v", err)
	}

	if len(files) < 1 {
		t.Errorf("Expected at least 1 file, got %d", len(files))
	}
}

func TestFilesystemHandler_GetFileInfo(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	fileInfo, err := handler.GetFileInfo("test.txt")
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if fileInfo.Name != "test.txt" {
		t.Errorf("Expected test.txt, got %s", fileInfo.Name)
	}

	if fileInfo.Size != 12 {
		t.Errorf("Expected size 12, got %d", fileInfo.Size)
	}

	if fileInfo.IsDir {
		t.Error("Expected file, got directory")
	}

	if fileInfo.Modified.IsZero() {
		t.Error("Modified time should not be zero")
	}
}

func TestFilesystemHandler_GetFileInfo_Directory(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	fileInfo, err := handler.GetFileInfo("subdir")
	if err != nil {
		t.Fatalf("Failed to get directory info: %v", err)
	}

	if !fileInfo.IsDir {
		t.Error("Expected directory, got file")
	}
}

func TestFilesystemHandler_ReadFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	content, err := handler.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "test content"
	if content != expected {
		t.Errorf("Expected %q, got %q", expected, content)
	}
}

func TestFilesystemHandler_ReadFile_Directory(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	_, err := handler.ReadFile("subdir")
	if err == nil {
		t.Error("Expected error when reading directory as file")
	}
}

func TestFilesystemHandler_ReadFile_PathTraversal(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	_, err := handler.ReadFile("../../../etc/passwd")
	if err == nil {
		t.Error("Expected path traversal to be blocked")
	}
}

func TestFilesystemHandler_IsPathSafe(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	testCases := []struct {
		path string
		safe bool
	}{
		{filepath.Join(tmpDir, "test.txt"), true},
		{filepath.Join(tmpDir, "subdir/sub.txt"), true},
		{tmpDir, true},
		{filepath.Join(tmpDir, "../"), false},
		{"/etc/passwd", false},
		{filepath.Join(tmpDir, "../../../etc/passwd"), false},
	}

	for _, tc := range testCases {
		result := handler.isPathSafe(tc.path)
		if result != tc.safe {
			t.Errorf("Path %q: expected safe=%v, got %v", tc.path, tc.safe, result)
		}
	}
}

func TestFilesystemHandler_NonExistentPath(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	_, err := handler.GetFileInfo("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	_, err = handler.ReadFile("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	_, err = handler.ListDirectory("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	_, err = handler.GetAbsolutePath("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestFilesystemHandler_GetAbsolutePath(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	absPath, err := handler.GetAbsolutePath("test.txt")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test.txt")
	if absPath != expectedPath {
		t.Errorf("Expected %q, got %q", expectedPath, absPath)
	}

	absPath, err = handler.GetAbsolutePath("subdir")
	if err != nil {
		t.Fatalf("Failed to get absolute path for directory: %v", err)
	}

	expectedPath = filepath.Join(tmpDir, "subdir")
	if absPath != expectedPath {
		t.Errorf("Expected %q, got %q", expectedPath, absPath)
	}

	absPath, err = handler.GetAbsolutePath("subdir/sub.txt")
	if err != nil {
		t.Fatalf("Failed to get absolute path for nested file: %v", err)
	}

	expectedPath = filepath.Join(tmpDir, "subdir", "sub.txt")
	if absPath != expectedPath {
		t.Errorf("Expected %q, got %q", expectedPath, absPath)
	}
}

func TestFilesystemHandler_GetAbsolutePath_PathTraversal(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler := NewHandler(tmpDir)

	_, err := handler.GetAbsolutePath("../")
	if err == nil {
		t.Error("Expected path traversal to be blocked")
	}

	_, err = handler.GetAbsolutePath("../../../etc/passwd")
	if err == nil {
		t.Error("Expected path traversal to be blocked")
	}
}
