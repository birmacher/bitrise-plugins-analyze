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
		Type:         getFileType(relativePath, info),
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

func getFileType(relativePath string, info os.FileInfo) string {
	name := strings.ToLower(info.Name())
	ext := strings.ToLower(filepath.Ext(name))

	directory_type := ""

	if relativePath == "META-INF" || strings.HasPrefix(relativePath, "META-INF/") {
		directory_type = core.FileTypeMetadata
	}

	if relativePath == "assets" || strings.HasPrefix(relativePath, "assets/") {
		directory_type = core.FileTypeAsset
	}

	if relativePath == "res" || strings.HasPrefix(relativePath, "res/") {
		directory_type = core.FileTypeResource
	}

	if relativePath == "lib" || strings.HasPrefix(relativePath, "lib/") {
		directory_type = core.FileTypeNativeLibrary
	}

	if info.IsDir() {
		if directory_type != "" {
			return directory_type
		}

		return core.FileTypeDirectory
	}

	if strings.Contains(relativePath, "AndroidManifest.xml") ||
		strings.Contains(relativePath, "package.xml") {
		return core.FileTypeMetadata
	}

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

	// Audio
	case ".mp3", ".wav", ".aac", ".m4a", ".flac":
		return core.FileTypeAudio

	// CoreML Models
	case ".mlmodel", ".mlmodelc":
		return core.FileTypeCoreMLModel

	case ".json", ".plist", ".xml":
		return core.FileTypeResource

	// Native Libraries
	case ".so", ".dll", ".dylib":
		return core.FileTypeNativeLibrary

	default:
		if directory_type != "" {
			return directory_type
		}
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
