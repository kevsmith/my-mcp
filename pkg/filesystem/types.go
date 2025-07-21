package filesystem

import "time"

type ListDirectoryArgs struct {
	Path string `json:"path"`
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

type GetAbsolutePathArgs struct {
	Path string `json:"path"`
}

type FileInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsDir    bool      `json:"is_dir"`
	Size     int64     `json:"size"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}
