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
	allowedRoots []string // Pre-cleaned absolute paths (stored without trailing separators)
	rootPrefixes []string // Pre-computed roots with trailing separators for efficient matching
	currentWD    string   // Current working directory (absolute)
}

func NewHandler(allowedRoots []string) (*Handler, error) {
	if len(allowedRoots) == 0 {
		return nil, fmt.Errorf("at least one allowed root directory is required")
	}

	// Clean and validate all allowed roots, pre-compute prefixes
	var cleanRoots []string
	var rootPrefixes []string
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
		
		// Pre-compute prefix with trailing separator for efficient matching
		rootPrefix := absRoot
		if !strings.HasSuffix(rootPrefix, string(filepath.Separator)) {
			rootPrefix += string(filepath.Separator)
		}
		rootPrefixes = append(rootPrefixes, rootPrefix)
	}

	// Start in the first allowed root
	initialWD := cleanRoots[0]

	return &Handler{
		allowedRoots: cleanRoots,
		rootPrefixes: rootPrefixes,
		currentWD:    initialWD,
	}, nil
}

// Core security function - resolves and validates any path
// Optimized with pre-cleaning and efficient validation
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

	// Optimized validation against allowed roots
	if !h.isPathAllowedOptimized(absPath) {
		return "", fmt.Errorf("access denied: path outside allowed roots")
	}

	return absPath, nil
}

// Legacy method for backward compatibility
func (h *Handler) isPathAllowed(path string) bool {
	return h.isPathAllowedOptimized(path)
}

// Optimized path validation with pre-computed prefixes
func (h *Handler) isPathAllowedOptimized(path string) bool {
	// Pre-clean path once
	cleanPath := filepath.Clean(path)
	
	// First check for exact root matches (most common case)
	for _, root := range h.allowedRoots {
		if cleanPath == root {
			return true
		}
	}
	
	// Check if path is under any allowed root using pre-computed prefixes
	for _, rootPrefix := range h.rootPrefixes {
		if strings.HasPrefix(cleanPath, rootPrefix) {
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
	// For backward compatibility, call the optimized version with no limits
	result, err := h.ListDirectoryOptimized(path, nil, nil)
	if err != nil {
		return nil, err
	}
	return result.Files, nil
}

// ListDirectoryOptimized provides streaming directory listing with limits and pagination
func (h *Handler) ListDirectoryOptimized(path *string, limit *int, skip *int) (*DirectoryListResult, error) {
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
	
	// Set default values for pagination
	var skipCount, limitCount int
	if skip != nil {
		skipCount = *skip
	}
	if limit != nil {
		limitCount = *limit
	} else {
		limitCount = -1 // No limit
	}

	// Open directory for streaming read
	dir, err := os.Open(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %w", err)
	}
	defer dir.Close()
	
	// Use streaming read for better performance with large directories
	var files []FileInfo
	var totalCount int
	var processedCount int
	
	// Read directory entries in batches for memory efficiency
	batchSize := 1000
	if limitCount > 0 && limitCount < batchSize {
		batchSize = limitCount * 2 // Read a bit more than needed for sorting
	}
	
	for {
		entries, err := dir.ReadDir(batchSize)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read directory entries: %w", err)
		}
		
		if len(entries) == 0 {
			break
		}
		
		// Process entries in this batch
		for _, entry := range entries {
			totalCount++
			
			// Skip entries if needed for pagination
			if processedCount < skipCount {
				processedCount++
				continue
			}
			
			// Check limit after skipping
			if limitCount > 0 && len(files) >= limitCount {
				break
			}
			
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
			processedCount++
		}
		
		// Break if we've reached our limit
		if limitCount > 0 && len(files) >= limitCount {
			break
		}
		
		// If we got fewer entries than batch size, we're at EOF
		if len(entries) < batchSize {
			break
		}
	}

	// Sort the collected files
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})
	
	// Determine if there are more entries available
	hasMore := false
	if limitCount > 0 {
		// Check if total count is greater than what we've returned + skipped
		hasMore = totalCount > (len(files) + skipCount)
	}

	return &DirectoryListResult{
		Files:         files,
		TotalCount:    totalCount,
		ReturnedCount: len(files),
		Skipped:       skipCount,
		HasMore:       hasMore,
	}, nil
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
		if !h.isPathAllowedOptimized(match) {
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
