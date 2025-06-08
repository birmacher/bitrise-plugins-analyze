package core

import "strings"

const (
	FileTypeDirectory     = "directory"
	FileTypeBinary        = "binary"
	FileTypeFont          = "font"
	FileTypeLocalization  = "localization"
	FileTypeImage         = "image"
	FileTypeVideo         = "video"
	FileTypeAudio         = "audio"
	FileTypeAssetCatalog  = "asset_catalog"
	FileTypeCoreMLModel   = "coreml_model"
	FileTypeDex           = "dex"
	FileTypeMetadata      = "metadata"
	FileTypeAsset         = "asset"
	FileTypeResource      = "resource"
	FileTypeNativeLibrary = "native_library"
	// Inner Types
	FileTypeDuplicate = "duplicate"
	FileTypeUnmapped  = "unmapped"
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
	FileTypeDirectory:    {Label: "Directory", Color: "#bdbdbd"},    // Gray
	FileTypeBinary:       {Label: "Binary", Color: "#1976d2"},       // Blue
	FileTypeFont:         {Label: "Font", Color: "#8d6e63"},         // Brown
	FileTypeLocalization: {Label: "Localization", Color: "#388e3c"}, // Green
	FileTypeImage:        {Label: "Image", Color: "#fbc02d"},        // Yellow
	FileTypeVideo:        {Label: "Video", Color: "#d32f2f"},        // Red
	FileTypeAudio:        {Label: "Audio", Color: "#7b1fa2"},        // Purple
	FileTypeDuplicate:    {Label: "Duplicate", Color: "#c62828"},    // Dark Red
	FileTypeUnmapped:     {Label: "Unmapped", Color: "#bdbdbd"},     // Light Gray
}

// IOSFileTypes lists iOS specific file types
var IOSFileTypes = map[string]FileTypeInfo{
	FileTypeAssetCatalog: {Label: "Asset Catalog", Color: "#ffa000"}, // Amber
	FileTypeCoreMLModel:  {Label: "CoreML Model", Color: "#0288d1"},  // Light Blue
}

// AndroidFileTypes lists Android specific file types
var AndroidFileTypes = map[string]FileTypeInfo{
	FileTypeDex:           {Label: "Dex", Color: "#4e747b"},            // Green
	FileTypeMetadata:      {Label: "Metadata", Color: "#455a64"},       // Blue Gray
	FileTypeAsset:         {Label: "Asset", Color: "#fbc02d"},          // Yellow
	FileTypeResource:      {Label: "Resource", Color: "#0288d1"},       // Light Blue
	FileTypeNativeLibrary: {Label: "Native Library", Color: "#6d4c41"}, // Dark Brown
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
