package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func analyzeIOSBundle(bundle_path string) (*AppBundle, error) {
	ext := strings.ToLower(filepath.Ext(bundle_path))

	var app_path string
	var tempDir string
	defer func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	}()
	var err error
	switch ext {
	case AppExtension:
		app_path = bundle_path
	case IpaExtension:
		app_path, tempDir, err = analyzeIpa(bundle_path)
	case XcarchiveExtension:
		app_path, err = analyzeXcarchive(bundle_path)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	return AnalyzeAppBundle(app_path)
}

func analyzeXcarchive(app_path string) (string, error) {
	productsPath := filepath.Join(app_path, "Products", "Applications")
	return findAppPath(productsPath)
}

func analyzeIpa(app_path string) (string, string, error) {
	tempDir, err := unzip(app_path)
	if err != nil {
		return "", "", fmt.Errorf("failed to unzip IPA: %v", err)
	}

	// Find the .app file in Payload directory
	payloadPath := filepath.Join(tempDir, "Payload")
	appPath, err := findAppPath(payloadPath)
	return appPath, tempDir, err
}

func findAppPath(directory string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(directory, "*.app"))
	if err != nil {
		return "", fmt.Errorf("error searching for .app file: %v", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no .app file found in Payload directory")
	}

	// Get absolute path of the first match
	absPath, err := filepath.Abs(matches[0])
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %v", err)
	}

	return absPath, nil
}
