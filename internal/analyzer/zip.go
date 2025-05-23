package analyzer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func unzip(zip_path string) (string, error) {
	tempDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Open the ZIP file
	reader, err := zip.OpenReader(zip_path)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %v", err)
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

	return tempDir, nil
}
