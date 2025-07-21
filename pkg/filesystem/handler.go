package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Handler struct {
	allowedRoots []string // Pre-cleaned absolute paths
	currentWD    string   // Current working directory (absolute)
}

func NewHandler(allowedRoots []string) (*Handler, error) {
	if len(allowedRoots) == 0 {
		return nil, fmt.Errorf("at least one allowed root directory is required")
	}

	// Clean and validate all allowed roots
	var cleanRoots []string
	for _, root := range allowedRoots {
		absRoot, err := filepath.Abs(filepath.Clean(root))
		if err != nil {
			return nil, fmt.Errorf("invalid root path %s: %w", root, err)
		}

		// Verify root exists and is a directory
		info, err := os.Stat(absRoot)
		if err != nil {
			return nil, fmt.Errorf("root path %s does not exist: %w", absRoot, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("root path %s is not a directory", absRoot)
		}

		cleanRoots = append(cleanRoots, absRoot)
	}

	// Start in the first allowed root
	initialWD := cleanRoots[0]

	return &Handler{
		allowedRoots: cleanRoots,
		currentWD:    initialWD,
	}, nil
}

// Core security function - resolves and validates any path
func (h *Handler) resolvePath(inputPath string) (string, error) {
	var resolvedPath string

	if filepath.IsAbs(inputPath) {
		// Absolute path - use as-is but validate
		resolvedPath = inputPath
	} else {
		// Relative path - resolve from CWD
		resolvedPath = filepath.Join(h.currentWD, inputPath)
	}

	// Critical: Clean the path to resolve all ./ ../ shenanigans
	cleanPath := filepath.Clean(resolvedPath)

	// Make it absolute to handle any remaining edge cases
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Validate against allowed roots
	if !h.isPathAllowed(absPath) {
		return "", fmt.Errorf("access denied: path outside allowed roots")
	}

	return absPath, nil
}

func (h *Handler) isPathAllowed(path string) bool {
	cleanPath := filepath.Clean(path)

	for _, root := range h.allowedRoots {
		cleanRoot := filepath.Clean(root)

		// Path must be inside or equal to an allowed root
		if strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) ||
			cleanPath == cleanRoot {
			return true
		}
	}
	return false
}

// Get relative path for display purposes
func (h *Handler) getRelativePath(absPath string) string {
	relPath, err := filepath.Rel(h.currentWD, absPath)
	if err != nil {
		return absPath // Fallback to absolute if relative fails
	}
	return relPath
}

// Navigation functions
func (h *Handler) ChangeDirectory(path string) error {
	// Resolve the new directory path
	newWD, err := h.resolvePath(path)
	if err != nil {
		return err
	}

	// Verify it's actually a directory
	info, err := os.Stat(newWD)
	if err != nil {
		return fmt.Errorf("directory does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", newWD)
	}

	// Safe to change
	h.currentWD = newWD
	return nil
}

func (h *Handler) GetCurrentDirectory() string {
	return h.currentWD
}

func (h *Handler) GetDirectoryInfo() DirectoryInfo {
	return DirectoryInfo{
		CurrentDirectory: h.currentWD,
		AllowedRoots:     h.allowedRoots,
	}
}

// File operations with new logic
func (h *Handler) ListDirectory(path *string) ([]FileInfo, error) {
	var targetPath string
	if path != nil && *path != "" {
		resolvedPath, err := h.resolvePath(*path)
		if err != nil {
			return nil, err
		}
		targetPath = resolvedPath
	} else {
		// Default to current working directory
		targetPath = h.currentWD
	}

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		absPath := filepath.Join(targetPath, entry.Name())
		fileInfo := FileInfo{
			Name:         entry.Name(),
			Path:         absPath,
			RelativePath: h.getRelativePath(absPath),
			IsDir:        entry.IsDir(),
			Size:         info.Size(),
			Modified:     info.ModTime(),
		}

		if stat := info.Sys(); stat != nil {
			fileInfo.Created = extractCreationTime(stat)
		}

		files = append(files, fileInfo)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	return files, nil
}

func (h *Handler) Glob(pattern string) (*GlobResult, error) {
	// Resolve pattern from current working directory
	var fullPattern string
	if filepath.IsAbs(pattern) {
		fullPattern = pattern
	} else {
		fullPattern = filepath.Join(h.currentWD, pattern)
	}

	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("glob pattern error: %w", err)
	}

	var files []FileInfo
	for _, match := range matches {
		if !h.isPathAllowed(match) {
			continue // Skip matches outside allowed roots
		}

		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:         filepath.Base(match),
			Path:         match,
			RelativePath: h.getRelativePath(match),
			IsDir:        info.IsDir(),
			Size:         info.Size(),
			Modified:     info.ModTime(),
		}

		if stat := info.Sys(); stat != nil {
			fileInfo.Created = extractCreationTime(stat)
		}

		files = append(files, fileInfo)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return &GlobResult{
		Pattern: pattern,
		Matches: files,
	}, nil
}

func (h *Handler) GetFileInfo(path string) (*FileInfo, error) {
	fullPath, err := h.resolvePath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileInfo := &FileInfo{
		Name:         filepath.Base(fullPath),
		Path:         fullPath,
		RelativePath: h.getRelativePath(fullPath),
		IsDir:        info.IsDir(),
		Size:         info.Size(),
		Modified:     info.ModTime(),
	}

	if stat := info.Sys(); stat != nil {
		fileInfo.Created = extractCreationTime(stat)
	}

	return fileInfo, nil
}

func (h *Handler) ReadFile(path string) (string, error) {
	fullPath, err := h.resolvePath(path)
	if err != nil {
		return "", err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("cannot read directory as file")
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return string(content), nil
}
