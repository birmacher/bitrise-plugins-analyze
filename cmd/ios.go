package cmd

import (
	"archive/zip"
	"bitrise-plugins-analyze/internal/analyzer"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func analyzeAppBundle(bundle_path string) error {
	ext := strings.ToLower(filepath.Ext(bundle_path))

	var app_path string
	var err error
	switch ext {
	case AppExtension:
		app_path = bundle_path
	case IpaExtension:
		app_path, err = analyzeIpa(bundle_path)
	case XcarchiveExtension:
		app_path, err = analyzeXcarchive(bundle_path)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return err
	}

	fileInfo, err := analyzer.AnalyzeFile(app_path, "")
	if err != nil {
		return fmt.Errorf("failed to analyze file: %v", err)
	}

	fmt.Println(fileInfo)

	return nil
}

func analyzeXcarchive(app_path string) (string, error) {
	productsPath := filepath.Join(app_path, "Products", "Applications")
	return findAppPath(productsPath)
}

func analyzeIpa(app_path string) (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "ipa-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory when done

	// Open the IPA file
	reader, err := zip.OpenReader(app_path)
	if err != nil {
		return "", fmt.Errorf("failed to open IPA file: %v", err)
	}
	defer reader.Close()

	// Extract all files to temp directory
	for _, file := range reader.File {
		filePath := filepath.Join(tempDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", fmt.Errorf("failed to create file: %v", err)
		}

		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return "", fmt.Errorf("failed to open zip file: %v", err)
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			srcFile.Close()
			dstFile.Close()
			return "", fmt.Errorf("failed to extract file: %v", err)
		}

		srcFile.Close()
		dstFile.Close()
	}

	// Find the .app file in Payload directory
	payloadPath := filepath.Join(tempDir, "Payload")
	return findAppPath(payloadPath)
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
