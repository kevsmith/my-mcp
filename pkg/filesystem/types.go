package filesystem

import "time"

// Tool argument types
type ChangeDirectoryArgs struct {
	Path string `json:"path"`
}

type ListDirectoryArgs struct {
	Path *string `json:"path,omitempty"` // Optional, defaults to CWD
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
