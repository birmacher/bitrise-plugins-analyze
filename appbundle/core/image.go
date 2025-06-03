package core

type OversizedImage struct {
	RelativePath string `json:"relative_path"`
	OriginalSize int64  `json:"original_size"`
	Saving       int64  `json:"saving"`
	ConvertedTo  string `json:"converted_to"`
}
