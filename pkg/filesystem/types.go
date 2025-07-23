package filesystem

import "time"

// Tool argument types
type ChangeDirectoryArgs struct {
	Path string `json:"path"`
}

type ListDirectoryArgs struct {
	Path  *string `json:"path,omitempty"`  // Optional, defaults to CWD
	Limit *int    `json:"limit,omitempty"` // Optional, limits number of entries returned
	Skip  *int    `json:"skip,omitempty"`  // Optional, number of entries to skip for pagination
}

type GlobArgs struct {
	Pattern string `json:"pattern"`
}

type GetFileInfoArgs struct {
	Path string `json:"path"`
}

type ReadFileArgs struct {
	Path string `json:"path"`
}

// Response types
type FileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`          // Always absolute
	RelativePath string    `json:"relative_path"` // Relative to CWD for display
	IsDir        bool      `json:"is_dir"`
	Size         int64     `json:"size"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

type DirectoryInfo struct {
	CurrentDirectory string   `json:"current_directory"`
	AllowedRoots     []string `json:"allowed_roots"`
}

type GlobResult struct {
	Pattern string     `json:"pattern"`
	Matches []FileInfo `json:"matches"`
}

// DirectoryListResult represents paginated directory listing results
type DirectoryListResult struct {
	Files         []FileInfo `json:"files"`
	TotalCount    int        `json:"total_count"`    // Total number of entries in directory
	ReturnedCount int        `json:"returned_count"` // Number of entries actually returned
	Skipped       int        `json:"skipped"`        // Number of entries skipped
	HasMore       bool       `json:"has_more"`       // Whether there are more entries available
}
