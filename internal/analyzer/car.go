package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
func ParseCARFile(path string, basePath string) (*CarFileInfo, error) {
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
	assetMap := make(map[string]*AssetInfo)
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
			asset = &AssetInfo{
				Name:          catalog.Name,
				RenditionInfo: make([]RenditionInfo, 0),
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
		rendition := RenditionInfo{
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
	assets := make([]AssetInfo, 0, len(assetMap))
	for _, asset := range assetMap {
		assets = append(assets, *asset)
	}

	relativePath, err := filepath.Rel(basePath, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %v", err)
	}

	return &CarFileInfo{
		Path:   relativePath,
		Assets: assets,
	}, nil
}

// FindAndAnalyzeCarFiles searches for and analyzes all .car files in the bundle
func FindAndAnalyzeCarFiles(bundlePath string, bundle *AppBundle) error {
	err := filepath.Walk(bundlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".car" {
			carInfo, err := ParseCARFile(path, bundlePath)
			if err != nil {
				return fmt.Errorf("failed to analyze %s: %v", path, err)
			}
			bundle.CarFiles = append(bundle.CarFiles, *carInfo)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk bundle directory: %v", err)
	}

	return nil
}
