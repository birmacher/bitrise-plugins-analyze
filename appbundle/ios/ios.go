package ios

import (
	"bitrise-plugins-analyze/appbundle/core"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func GetTempIPAPath(bundle_path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(bundle_path))

	var app_path string

	var err error
	switch ext {
	case core.AppExtension:
		app_path = bundle_path
	case core.IpaExtension:
		app_path, _, err = analyzeIpa(bundle_path)
	case core.XcarchiveExtension:
		app_path, err = analyzeXcarchive(bundle_path)
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return "", err
	}

	return app_path, nil
}

func analyzeXcarchive(app_path string) (string, error) {
	productsPath := filepath.Join(app_path, "Products", "Applications")
	return findAppPath(productsPath)
}

func analyzeIpa(app_path string) (string, string, error) {
	tempDir, err := core.Unzip(app_path)
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

func CalculateDownloadSize(bundlePath string) (int64, error) {
	tempDir, err := os.MkdirTemp("", "app-*")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(tempDir)

	// Create zip file path
	zipPath := filepath.Join(tempDir, "app.zip")

	// Run ditto command to create zip
	cmd := exec.Command("ditto", "-c", "-k", "--sequesterRsrc", "--keepParent", bundlePath, zipPath)
	if err := cmd.Run(); err != nil {
		return 0, err
	}

	// Get zip file size using stat
	cmd = exec.Command("stat", "-f%z", zipPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse size from stat output
	return strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
}

func CalculateInstallSize(bundlePath string) (int64, error) {
	cmd := exec.Command("du", "-sk", bundlePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return 0, fmt.Errorf("failed to parse du output")
	}

	sizeKB, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, err
	}

	return sizeKB * 1024, nil
}
