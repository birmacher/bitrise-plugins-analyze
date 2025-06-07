package appbundle

import (
	"bitrise-plugins-analyze/appbundle/core"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func AnalyzeFile(filePath string, basePath string) (core.FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return core.FileInfo{}, fmt.Errorf("failed to get file info: %v", err)
	}

	relativePath, err := filepath.Rel(basePath, filePath)
	if err != nil {
		return core.FileInfo{}, fmt.Errorf("failed to get relative path: %v", err)
	}

	fileInfo := core.FileInfo{
		RelativePath: relativePath,
		Type:         getFileType(info),
	}

	// Recursively process directory contents
	if info.IsDir() {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return core.FileInfo{}, fmt.Errorf("failed to read directory: %v", err)
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name() < entries[j].Name()
		})

		var totalSize int64
		var childChecksums []string
		for _, entry := range entries {
			childPath := filepath.Join(filePath, entry.Name())
			childInfo, err := AnalyzeFile(childPath, basePath)
			if err != nil {
				return core.FileInfo{}, err
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
			return core.FileInfo{}, fmt.Errorf("failed to calculate SHA256: %v", err)
		}
		fileInfo.Shasum = shasum
	}

	return fileInfo, nil
}

func getFileType(info os.FileInfo) string {
	if info.IsDir() {
		return core.FileTypeDirectory
	}

	name := strings.ToLower(info.Name())
	ext := strings.ToLower(filepath.Ext(name))

	switch ext {
	// Fonts
	case ".otf", ".ttc", ".ttf", ".woff":
		return core.FileTypeFont

	// Localizations
	case ".strings", ".xcstrings", ".stringsdict":
		return core.FileTypeLocalization

	// Asset Catalogs
	case ".car", ".xcassets":
		return core.FileTypeAssetCatalog

	// Images
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".heic", ".heif":
		return core.FileTypeImage

	// Videos
	case ".mp4", ".mov", ".m4v":
		return core.FileTypeVideo

	// CoreML Models
	case ".mlmodel", ".mlmodelc":
		return core.FileTypeCoreMLModel

	default:
		return core.FileTypeBinary
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
