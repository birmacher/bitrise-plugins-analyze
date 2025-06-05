package core

import "strings"

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
	"directory":    {Label: "Directory", Color: "#b0b4ff"},
	"binary":       {Label: "Binary", Color: "#a5d8ff"},
	"font":         {Label: "Font", Color: "#ff9f0a"},
	"localization": {Label: "Localization", Color: "#30d158"},
	"image":        {Label: "Image", Color: "#64d2ff"},
	"video":        {Label: "Video", Color: "#bf5af2"},
}

// IOSFileTypes lists iOS specific file types
var IOSFileTypes = map[string]FileTypeInfo{
	"asset_catalog": {Label: "Asset Catalog", Color: "#ffe066"},
	"coreml_model":  {Label: "CoreML Model", Color: "#ff453a"},
}

// AndroidFileTypes lists Android specific file types
var AndroidFileTypes = map[string]FileTypeInfo{
	"dex": {Label: "Dex", Color: "#8e8d8a"},
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
