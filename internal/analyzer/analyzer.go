package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	RelativePath string     `json:"relative_path"`
	Size         int64      `json:"size"`
	Shasum       string     `json:"shasum"`
	Type         string     `json:"type"`
	Children     []FileInfo `json:"children,omitempty"`
}

func AnalyzeBaseDirectory(directoryPath string) (FileInfo, error) {
	return AnalyzeFile(directoryPath, directoryPath)
}

func AnalyzeFile(filePath string, basePath string) (FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to get file info: %v", err)
	}

	relativePath, err := filepath.Rel(basePath, filePath)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to get relative path: %v", err)
	}

	fileInfo := FileInfo{
		RelativePath: relativePath,
		Size:         info.Size(),
		Type:         getFileType(info),
	}

	// Calculate SHA256 for files
	if !info.IsDir() {
		shasum, err := calculateSHA256(filePath)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to calculate SHA256: %v", err)
		}
		fileInfo.Shasum = shasum
	}

	// Recursively process directory contents
	if info.IsDir() {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to read directory: %v", err)
		}

		for _, entry := range entries {
			childPath := filepath.Join(filePath, entry.Name())
			childInfo, err := AnalyzeFile(childPath, basePath)
			if err != nil {
				return FileInfo{}, err
			}
			fileInfo.Children = append(fileInfo.Children, childInfo)
		}
	}

	return fileInfo, nil
}

func getFileType(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	ext := strings.ToLower(filepath.Ext(info.Name()))
	switch ext {
	case ".otf", ".ttc", ".ttf", ".woff":
		return "font"
	case ".strings", ".xcstrings":
		return "localization"
	case "car":
		return "asset"
	default:
		return "file"
	}
}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
