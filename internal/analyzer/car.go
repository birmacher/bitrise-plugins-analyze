package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
func ParseCARFile(path string) (*CarFileInfo, error) {
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
		// Get or create the AssetInfo for this name
		asset, exists := assetMap[catalog.Name]
		if !exists {
			asset = &AssetInfo{
				Name:          catalog.Name,
				RenditionInfo: make([]RenditionInfo, 0),
			}
			assetMap[catalog.Name] = asset
		}

		// Add the rendition info
		rendition := RenditionInfo{
			RenditionName: catalog.RenditionName,
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

	return &CarFileInfo{
		Path:   path,
		Assets: assets,
	}, nil
}

// FindAndAnalyzeCarFiles searches for and analyzes all .car files in the bundle
func FindAndAnalyzeCarFiles(bundlePath string) ([]CarFileInfo, error) {
	var carFiles []CarFileInfo

	err := filepath.Walk(bundlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".car" {
			carInfo, err := ParseCARFile(path)
			if err != nil {
				return fmt.Errorf("failed to analyze %s: %v", path, err)
			}
			carFiles = append(carFiles, *carInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk bundle directory: %v", err)
	}

	return carFiles, nil
}
