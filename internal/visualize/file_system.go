package visualize

import (
	"sort"

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

// FindLargestFiles returns the top N largest files in the bundle
func FindLargestFiles(bundle *analyzer.AppBundle, n int) []analyzer.FileInfo {
	// Create a slice to store all files
	var files []analyzer.FileInfo

	if bundle.Files.Type != "directory" {
		files = append(files, bundle.Files)
	}

	// Add all child files recursively
	addChildFiles(&bundle.Files, &files)

	// Sort files by size in descending order
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	// Return top N files (or all if less than N)
	if len(files) < n {
		return files
	}
	return files[:n]
}

// addChildFiles recursively adds all child files to the files slice
func addChildFiles(info *analyzer.FileInfo, files *[]analyzer.FileInfo) {
	for _, child := range info.Children {
		// Only add regular files, not directories
		if child.Type != "directory" {
			*files = append(*files, child)
		}
		// Recursively process children
		addChildFiles(&child, files)
	}
}
