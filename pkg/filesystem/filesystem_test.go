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

func TestNewHandler(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Test successful creation
	handler, err := NewHandler([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Should start in first allowed root
	if handler.GetCurrentDirectory() != tmpDir {
		t.Errorf("Expected CWD %s, got %s", tmpDir, handler.GetCurrentDirectory())
	}

	// Test with non-existent directory
	_, err = NewHandler([]string{"/nonexistent/path"})
	if err == nil {
		t.Error("Expected error for non-existent root directory")
	}

	// Test with empty roots
	_, err = NewHandler([]string{})
	if err == nil {
		t.Error("Expected error for empty roots")
	}
}

func TestListDirectory(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler, err := NewHandler([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// List current directory (no path argument)
	files, err := handler.ListDirectory(nil)
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

	// Test with specific path
	subdirPath := "subdir"
	files, err = handler.ListDirectory(&subdirPath)
	if err != nil {
		t.Fatalf("Failed to list subdirectory: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file in subdir, got %d", len(files))
	}

	if files[0].Name != "sub.txt" {
		t.Errorf("Expected sub.txt, got %s", files[0].Name)
	}
}

func TestPathTraversalPrevention(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler, err := NewHandler([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test various path traversal attempts
	attackPaths := []string{
		"../",
		"../../etc/passwd",
		"../../../root/.ssh/id_rsa",
		"subdir/../../../etc/passwd",
	}

	for _, attackPath := range attackPaths {
		_, err := handler.ReadFile(attackPath)
		if err == nil {
			t.Errorf("Expected path traversal to be blocked for: %s", attackPath)
		}

		_, err = handler.GetFileInfo(attackPath)
		if err == nil {
			t.Errorf("Expected path traversal to be blocked for: %s", attackPath)
		}

		path := attackPath
		_, err = handler.ListDirectory(&path)
		if err == nil {
			t.Errorf("Expected path traversal to be blocked for: %s", attackPath)
		}
	}
}

func TestChangeDirectory(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler, err := NewHandler([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Change to subdirectory
	err = handler.ChangeDirectory("subdir")
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "subdir")
	if handler.GetCurrentDirectory() != expectedPath {
		t.Errorf("Expected CWD %s, got %s", expectedPath, handler.GetCurrentDirectory())
	}

	// Change back to parent
	err = handler.ChangeDirectory("..")
	if err != nil {
		t.Fatalf("Failed to change to parent directory: %v", err)
	}

	if handler.GetCurrentDirectory() != tmpDir {
		t.Errorf("Expected CWD %s, got %s", tmpDir, handler.GetCurrentDirectory())
	}

	// Try to escape - should fail
	err = handler.ChangeDirectory("../../")
	if err == nil {
		t.Error("Expected path traversal to be blocked")
	}
}

func TestReadFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler, err := NewHandler([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Read file with relative path
	content, err := handler.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "test content"
	if content != expected {
		t.Errorf("Expected %q, got %q", expected, content)
	}

	// Read file in subdirectory
	content, err = handler.ReadFile("subdir/sub.txt")
	if err != nil {
		t.Fatalf("Failed to read subdirectory file: %v", err)
	}

	expected = "sub content"
	if content != expected {
		t.Errorf("Expected %q, got %q", expected, content)
	}

	// Try to read directory - should fail
	_, err = handler.ReadFile("subdir")
	if err == nil {
		t.Error("Expected error when reading directory as file")
	}
}

func TestGlob(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler, err := NewHandler([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test simple glob
	result, err := handler.Glob("*.txt")
	if err != nil {
		t.Fatalf("Failed to glob: %v", err)
	}

	if len(result.Matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(result.Matches))
	}

	if result.Matches[0].Name != "test.txt" {
		t.Errorf("Expected test.txt, got %s", result.Matches[0].Name)
	}

	// Test recursive glob
	result, err = handler.Glob("**/*.txt")
	if err != nil {
		t.Fatalf("Failed to glob recursively: %v", err)
	}

	if len(result.Matches) < 1 {
		t.Errorf("Expected at least 1 match, got %d", len(result.Matches))
	}
}

func TestMultipleRoots(t *testing.T) {
	tmpDir1, cleanup1 := setupTestDir(t)
	defer cleanup1()

	tmpDir2, cleanup2 := setupTestDir(t)
	defer cleanup2()

	handler, err := NewHandler([]string{tmpDir1, tmpDir2})
	if err != nil {
		t.Fatalf("Failed to create handler with multiple roots: %v", err)
	}

	// Should start in first root
	if handler.GetCurrentDirectory() != tmpDir1 {
		t.Errorf("Expected CWD %s, got %s", tmpDir1, handler.GetCurrentDirectory())
	}

	// Should be able to access second root with absolute path
	absPath := filepath.Join(tmpDir2, "test.txt")
	content, err := handler.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Failed to read file from second root: %v", err)
	}

	if content != "test content" {
		t.Errorf("Expected 'test content', got %q", content)
	}

	// Should be able to change to second root
	err = handler.ChangeDirectory(tmpDir2)
	if err != nil {
		t.Fatalf("Failed to change to second root: %v", err)
	}

	if handler.GetCurrentDirectory() != tmpDir2 {
		t.Errorf("Expected CWD %s, got %s", tmpDir2, handler.GetCurrentDirectory())
	}
}

func TestGetDirectoryInfo(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	handler, err := NewHandler([]string{tmpDir})
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	info := handler.GetDirectoryInfo()

	if info.CurrentDirectory != tmpDir {
		t.Errorf("Expected current directory %s, got %s", tmpDir, info.CurrentDirectory)
	}

	if len(info.AllowedRoots) != 1 {
		t.Errorf("Expected 1 allowed root, got %d", len(info.AllowedRoots))
	}

	if info.AllowedRoots[0] != tmpDir {
		t.Errorf("Expected allowed root %s, got %s", tmpDir, info.AllowedRoots[0])
	}
}
