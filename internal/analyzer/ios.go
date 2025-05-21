package analyzer

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"howett.net/plist"
)

// plistData represents the structure of an Info.plist file
type plistData struct {
	XMLName xml.Name `xml:"plist"`
	Dict    struct {
		Key    []string `xml:"key"`
		String []string `xml:"string"`
		Array  []struct {
			String []string `xml:"string"`
		} `xml:"array"`
	} `xml:"dict"`
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
