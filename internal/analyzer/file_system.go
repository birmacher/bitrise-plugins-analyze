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
		Type:         getFileType(info),
	}

	// Recursively process directory contents
	if info.IsDir() {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to read directory: %v", err)
		}

		var totalSize int64
		var childChecksums []string
		for _, entry := range entries {
			childPath := filepath.Join(filePath, entry.Name())
			childInfo, err := AnalyzeFile(childPath, basePath)
			if err != nil {
				return FileInfo{}, err
			}
			fileInfo.Children = append(fileInfo.Children, childInfo)
			totalSize += childInfo.Size
			childChecksums = append(childChecksums, childInfo.Shasum)
		}
		fileInfo.Size = totalSize

		// Calculate directory checksum by combining children's checksums
		if len(childChecksums) > 0 {
			hash := sha256.New()
			for _, checksum := range childChecksums {
				hash.Write([]byte(checksum))
			}
			fileInfo.Shasum = hex.EncodeToString(hash.Sum(nil))
		}
	} else {
		fileInfo.Size = info.Size()
		// Calculate SHA256 for files
		shasum, err := calculateSHA256(filePath)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to calculate SHA256: %v", err)
		}
		fileInfo.Shasum = shasum
	}

	return fileInfo, nil
}

func getFileType(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}

	name := strings.ToLower(info.Name())
	ext := strings.ToLower(filepath.Ext(name))

	switch ext {
	// Fonts
	case ".otf", ".ttc", ".ttf", ".woff":
		return "font"

	// Localizations
	case ".strings", ".xcstrings", ".stringsdict":
		return "localization"

	// Asset Catalogs
	case ".car", ".xcassets":
		return "asset_catalog"

	// Videos
	case ".mp4", ".mov", ".m4v":
		return "video"

	// CoreML Models
	case ".mlmodel", ".mlmodelc":
		return "coreml_model"

	default:
		return "binary"
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
