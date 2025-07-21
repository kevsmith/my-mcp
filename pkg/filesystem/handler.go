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
	basePath string
}

func NewHandler(basePath string) *Handler {
	return &Handler{basePath: basePath}
}

func (fh *Handler) ListDirectory(path string) ([]FileInfo, error) {
	fullPath := filepath.Join(fh.basePath, path)

	if !fh.isPathSafe(fullPath) {
		return nil, fmt.Errorf("access denied: path outside allowed directory")
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:     entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			Modified: info.ModTime(),
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

func (fh *Handler) Glob(pattern string) ([]FileInfo, error) {
	fullPattern := filepath.Join(fh.basePath, pattern)

	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("glob pattern error: %w", err)
	}

	var files []FileInfo
	for _, match := range matches {
		if !fh.isPathSafe(match) {
			continue
		}

		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		relPath, err := filepath.Rel(fh.basePath, match)
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:     filepath.Base(match),
			Path:     relPath,
			IsDir:    info.IsDir(),
			Size:     info.Size(),
			Modified: info.ModTime(),
		}

		if stat := info.Sys(); stat != nil {
			fileInfo.Created = extractCreationTime(stat)
		}

		files = append(files, fileInfo)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files, nil
}

func (fh *Handler) GetFileInfo(path string) (*FileInfo, error) {
	fullPath := filepath.Join(fh.basePath, path)

	if !fh.isPathSafe(fullPath) {
		return nil, fmt.Errorf("access denied: path outside allowed directory")
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileInfo := &FileInfo{
		Name:     filepath.Base(path),
		Path:     path,
		IsDir:    info.IsDir(),
		Size:     info.Size(),
		Modified: info.ModTime(),
	}

	if stat := info.Sys(); stat != nil {
		fileInfo.Created = extractCreationTime(stat)
	}

	return fileInfo, nil
}

func (fh *Handler) ReadFile(path string) (string, error) {
	fullPath := filepath.Join(fh.basePath, path)

	if !fh.isPathSafe(fullPath) {
		return "", fmt.Errorf("access denied: path outside allowed directory")
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

func (fh *Handler) GetAbsolutePath(path string) (string, error) {
	fullPath := filepath.Join(fh.basePath, path)

	if !fh.isPathSafe(fullPath) {
		return "", fmt.Errorf("access denied: path outside allowed directory")
	}

	_, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to access path: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

func (fh *Handler) isPathSafe(path string) bool {
	cleanPath := filepath.Clean(path)
	cleanBase := filepath.Clean(fh.basePath)

	return strings.HasPrefix(cleanPath, cleanBase+string(filepath.Separator)) ||
		cleanPath == cleanBase
}
