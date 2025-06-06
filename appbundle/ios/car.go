package ios

import (
	"bitrise-plugins-analyze/appbundle/core/ios"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// AssetsutilCatalog represents the JSON structure returned by assetutil
type AssetsutilCatalog struct {
	Type          string `json:"AssetType"`
	Name          string `json:"Name"`
	RenditionName string `json:"RenditionName"`
	Scale         int64  `json:"Scale"`
	Idiom         string `json:"Idiom"`
	SizeOnDisk    int64  `json:"SizeOnDisk"`
	Compression   string `json:"Compression"`
	SHA1Digest    string `json:"SHA1Digest"`
}

// ParseCARFile uses assetutil to analyze the .car file and returns structured information
func ParseCARFile(path string, basePath string) (*ios.CarFileInfo, error) {
	// Check if assetutil exists
	if _, err := exec.LookPath("assetutil"); err != nil {
		return nil, fmt.Errorf("assetutil not found: this tool requires macOS")
	}

	// Run assetutil to get JSON output
	cmd := exec.Command("assetutil", "--info", path)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run assetutil: %v", err)
	}

	// Parse the JSON output into AssetsutilCatalog slice
	var catalogs []AssetsutilCatalog
	if err := json.Unmarshal(output, &catalogs); err != nil {
		return nil, fmt.Errorf("failed to parse assetutil output: %v", err)
	}

	// Group renditions by name
	assetMap := make(map[string]*ios.AssetInfo)
	for _, catalog := range catalogs {
		if catalog.SizeOnDisk == 0 {
			continue
		}

		// Skip renditions with empty names or system-generated packed assets
		if catalog.Name == "" || strings.HasPrefix(catalog.Name, "ZZZZPackedAsset-") {
			continue
		}

		// Get or create the AssetInfo for this name
		name := catalog.Name
		if name == "" {
			name = catalog.SHA1Digest
		}
		asset, exists := assetMap[name]
		if !exists {
			asset = &ios.AssetInfo{
				Name:          catalog.Name,
				RenditionInfo: make([]ios.RenditionInfo, 0),
			}
			assetMap[catalog.Name] = asset
		}

		// Skip renditions with empty names or system-generated packed assets
		if catalog.RenditionName == "" || strings.HasPrefix(catalog.RenditionName, "ZZZZPackedAsset-") {
			continue
		}

		// Add the rendition info
		renditionName := catalog.RenditionName
		if renditionName == "" {
			renditionName = catalog.SHA1Digest
		}
		rendition := ios.RenditionInfo{
			RenditionName: renditionName,
			Size:          catalog.SizeOnDisk,
			Idiom:         catalog.Idiom,
			Scale:         catalog.Scale,
			Compression:   catalog.Compression,
			Shasum:        catalog.SHA1Digest,
		}
		asset.RenditionInfo = append(asset.RenditionInfo, rendition)
	}

	// Convert map to slice
	assets := make([]ios.AssetInfo, 0, len(assetMap))
	for _, asset := range assetMap {
		assets = append(assets, *asset)
	}

	relativePath, err := filepath.Rel(basePath, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %v", err)
	}

	return &ios.CarFileInfo{
		Path:   relativePath,
		Assets: assets,
	}, nil
}
