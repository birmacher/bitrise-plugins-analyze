package core

import "strings"

const (
	FileTypeDirectory    = "directory"
	FileTypeBinary       = "binary"
	FileTypeFont         = "font"
	FileTypeLocalization = "localization"
	FileTypeImage        = "image"
	FileTypeVideo        = "video"
	FileTypeAssetCatalog = "asset_catalog"
	FileTypeCoreMLModel  = "coreml_model"
	FileTypeDex          = "dex"
	FileTypeDuplicate    = "duplicate"
)

// FileTypeInfo holds display name and color for a file type
// Color should be a hex color code used for visualization
// Label is a human friendly label shown in the legend
// Key refers to the FileInfo.Type value

type FileTypeInfo struct {
	Label string `json:"label"`
	Color string `json:"color"`
}

// CoreFileTypes contains file types that are common for all platforms
var CoreFileTypes = map[string]FileTypeInfo{
	FileTypeDirectory:    {Label: "Directory", Color: "#b0b4ff"},
	FileTypeBinary:       {Label: "Binary", Color: "#a5d8ff"},
	FileTypeFont:         {Label: "Font", Color: "#ff9f0a"},
	FileTypeLocalization: {Label: "Localization", Color: "#30d158"},
	FileTypeImage:        {Label: "Image", Color: "#64d2ff"},
	FileTypeVideo:        {Label: "Video", Color: "#bf5af2"},
	FileTypeDuplicate:    {Label: "Duplicate", Color: "#ff3b30"},
}

// IOSFileTypes lists iOS specific file types
var IOSFileTypes = map[string]FileTypeInfo{
	FileTypeAssetCatalog: {Label: "Asset Catalog", Color: "#ffe066"},
	FileTypeCoreMLModel:  {Label: "CoreML Model", Color: "#ff453a"},
}

// AndroidFileTypes lists Android specific file types
var AndroidFileTypes = map[string]FileTypeInfo{
	FileTypeDex: {Label: "Dex", Color: "#8e8d8a"},
}

// FileTypesForPlatforms returns a combined set of file types
// based on the supported platforms of the analyzed bundle.
func FileTypesForPlatforms(platforms []string) map[string]FileTypeInfo {
	types := make(map[string]FileTypeInfo)
	for k, v := range CoreFileTypes {
		types[k] = v
	}
	for _, p := range platforms {
		switch strings.ToLower(p) {
		case "ios", "iphoneos", "iphonesimulator", "ipados", "macos":
			for k, v := range IOSFileTypes {
				types[k] = v
			}
		case "android":
			for k, v := range AndroidFileTypes {
				types[k] = v
			}
		}
	}
	return types
}
