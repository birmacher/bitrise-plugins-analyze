package visualize

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"bitrise-plugins-analyze/internal/analyzer"
)

// FindDuplicates analyzes the bundle for duplicate files based on size and checksum
func FindDuplicates(bundle *analyzer.AppBundle) (map[string][]analyzer.FileInfo, error) {
	// First pass: group files by size
	sizeGroups := make(map[int64][]analyzer.FileInfo)

	// Second pass: group files of the same size by their checksum
	duplicates := make(map[string][]analyzer.FileInfo)
	for _, sizeGroup := range sizeGroups {
		if len(sizeGroup) < 2 {
			continue
		}

		// Group files by their checksum
		for _, file := range sizeGroup {
			duplicates[file.Shasum] = append(duplicates[file.Shasum], file)
		}
	}

	// Remove unique files (groups with only one file)
	for checksum, group := range duplicates {
		if len(group) < 2 {
			delete(duplicates, checksum)
		}
	}

	return duplicates, nil
}

// calculateChecksum calculates SHA-256 checksum of a file
func calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
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
