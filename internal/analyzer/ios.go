package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"howett.net/plist"
)

// AssetInfo represents information about an asset in the .car file
type AssetInfo struct {
	Name          string `json:"name"`
	RenditionName string `json:"rendition_name"`
	Size          int64  `json:"size"`
	Idiom         string `json:"idiom"`
	Scale         int64  `json:"scale"`
	Compression   string `json:"compression"`
	Shasum        string `json:"shasum"`
}

// CarFileInfo represents the analyzed contents of a .car file
type CarFileInfo struct {
	Path   string      `json:"path"`
	Assets []AssetInfo `json:"assets"`
	Type   string      `json:"type"`
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

// AnalyzeInfoPlist reads and parses the Info.plist file from the provided path
// and updates the AppBundle with the extracted information
func AnalyzeInfoPlist(bundlePath string, bundle *AppBundle) error {
	infoPlistPath := filepath.Join(bundlePath, "Info.plist")

	f, err := os.Open(infoPlistPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var data map[string]interface{}
	decoder := plist.NewDecoder(f)
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	// Handle supported platforms array
	if platforms, ok := data["CFBundleSupportedPlatforms"].([]interface{}); ok {
		bundle.SupportedPlatforms = make([]string, len(platforms))
		for i, platform := range platforms {
			if str, ok := platform.(string); ok {
				bundle.SupportedPlatforms[i] = str
			}
		}
	}

	// Safely extract string values with type checking
	if str, ok := data["CFBundleIdentifier"].(string); ok {
		bundle.BundleID = str
	} else {
		return fmt.Errorf("CFBundleIdentifier not found or invalid type")
	}

	if str, ok := data["CFBundleShortVersionString"].(string); ok {
		bundle.Version = str
	} else {
		return fmt.Errorf("CFBundleShortVersionString not found or invalid type")
	}

	if str, ok := data["MinimumOSVersion"].(string); ok {
		bundle.MinimumOSVersion = str
	} else {
		return fmt.Errorf("MinimumOSVersion not found or invalid type")
	}

	return nil
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

	// Convert the assetutil output to our structure
	assets := make([]AssetInfo, 0)
	for _, catalog := range catalogs {
		fmt.Println("Catalog:", catalog)
		asset := AssetInfo{
			Name:          catalog.Name,
			RenditionName: catalog.RenditionName,
			Size:          catalog.SizeOnDisk,
			Idiom:         catalog.Idiom,
			Scale:         catalog.Scale,
			Compression:   catalog.Compression,
			Shasum:        catalog.SHA1Digest,
		}
		assets = append(assets, asset)
	}

	return &CarFileInfo{
		Path:   path,
		Type:   "asset_catalog",
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
