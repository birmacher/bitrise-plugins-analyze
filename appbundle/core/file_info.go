package core

type FileInfo struct {
	RelativePath string     `json:"relative_path"`
	Size         int64      `json:"size"`
	Shasum       string     `json:"shasum"`
	Type         string     `json:"type"`
	Children     []FileInfo `json:"children,omitempty"`
}
