package ios

type RenditionInfo struct {
	RenditionName string `json:"rendition_name"`
	Size          int64  `json:"size"`
	Idiom         string `json:"idiom"`
	Scale         int64  `json:"scale"`
	Compression   string `json:"compression"`
	Shasum        string `json:"shasum"`
}

// AssetInfo represents information about an asset in the .car file
type AssetInfo struct {
	Name          string          `json:"name"`
	RenditionInfo []RenditionInfo `json:"rendition_info"`
}

// CarFileInfo represents the analyzed contents of a .car file
type CarFileInfo struct {
	Path   string      `json:"path"`
	Assets []AssetInfo `json:"assets"`
}
