package visualize

import (
	"bitrise-plugins-analyze/internal/analyzer"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// fileWithSize represents a file with its size for sorting
type fileWithSize struct {
	path string
	size int64
}

// moduleWithSize represents a directory module with its total size
type moduleWithSize struct {
	path      string
	size      int64
	fileCount int
}

// duplicateInfo represents information about duplicate content
type duplicateInfo struct {
	name        string   // file name or asset name
	size        int64    // size of the duplicate
	occurrences int      // number of occurrences
	locations   []string // where the duplicates are found
	isAsset     bool     // whether this is an asset catalog duplicate
}

// GenerateMarkdown generates a Markdown file containing the bundle analysis data
func GenerateMarkdown(bundle *analyzer.AppBundle, outputDir string) error {
	// Create Markdown file named after bundle ID
	mdFileName := fmt.Sprintf("%s.md", bundle.BundleID)
	mdPath := filepath.Join(outputDir, mdFileName)

	// Build markdown content
	var content strings.Builder

	// Header
	content.WriteString(fmt.Sprintf("# App Bundle Analysis: %s\n\n", bundle.AppName))

	// Basic Information
	content.WriteString("## Basic Information\n\n")
	content.WriteString("| Property | Value |\n")
	content.WriteString("|----------|-------|\n")
	content.WriteString(fmt.Sprintf("| Bundle ID | `%s` |\n", bundle.BundleID))
	content.WriteString(fmt.Sprintf("| Version | %s |\n", bundle.Version))
	content.WriteString(fmt.Sprintf("| Minimum OS Version | %s |\n", bundle.MinimumOSVersion))
	content.WriteString(fmt.Sprintf("| Download Size | %s |\n", formatSize(bundle.DownloadSize)))
	content.WriteString(fmt.Sprintf("| Install Size | %s |\n", formatSize(bundle.InstallSize)))
	content.WriteString(fmt.Sprintf("| Supported Platforms | %s |\n\n", strings.Join(bundle.SupportedPlatforms, ", ")))

	// Top 10 Largest Modules
	content.WriteString("## Top 10 Largest Modules\n\n")
	content.WriteString("| Module | Size | File Count | % of Total |\n")
	content.WriteString("|--------|------|------------|------------|\n")

	modules := findLargestModules(bundle.Files)
	for _, module := range modules[1:11] {
		percentage := float64(module.size) / float64(bundle.InstallSize) * 100
		content.WriteString(fmt.Sprintf("| %s | %s | %d | %.1f%% |\n",
			module.path,
			formatSize(module.size),
			module.fileCount,
			percentage))
	}
	content.WriteString("\n")

	// Top 10 Largest Files
	content.WriteString("## Top 10 Largest Files\n\n")
	content.WriteString("| File | Size | % of Total |\n")
	content.WriteString("|------|------|------------|\n")

	files := findLargestFiles(bundle.Files)
	for i, file := range files {
		if i >= 10 {
			break
		}
		percentage := float64(file.size) / float64(bundle.InstallSize) * 100
		content.WriteString(fmt.Sprintf("| %s | %s | %.1f%% |\n",
			file.path,
			formatSize(file.size),
			percentage))
	}
	content.WriteString("\n")

	// Collect all duplicates
	var allDuplicates []duplicateInfo

	// Add filesystem duplicates
	fsDuplicates := findDuplicateFiles(bundle.Files)
	for _, files := range fsDuplicates {
		if len(files) > 1 {
			allDuplicates = append(allDuplicates, duplicateInfo{
				name:        filepath.Base(files[0].RelativePath),
				size:        files[0].Size,
				occurrences: len(files),
				locations:   getRelativePaths(files),
				isAsset:     false,
			})
		}
	}

	// Add CAR file duplicates
	if len(bundle.CarFiles) > 0 {
		assetDuplicates := make(map[string]*duplicateInfo)

		for _, car := range bundle.CarFiles {
			for _, asset := range car.Assets {
				for _, rendition := range asset.RenditionInfo {
					if rendition.Shasum != "" {
						key := fmt.Sprintf("%s:%s", asset.Name, rendition.Shasum)
						info, exists := assetDuplicates[key]
						if !exists {
							info = &duplicateInfo{
								name:        asset.Name,
								size:        rendition.Size,
								occurrences: 0,
								locations:   make([]string, 0),
								isAsset:     true,
							}
							assetDuplicates[key] = info
						}
						info.occurrences++
						info.locations = append(info.locations,
							fmt.Sprintf("%s (%s)", car.Path, rendition.RenditionName))
					}
				}
			}
		}

		// Add asset duplicates to the main list
		for _, info := range assetDuplicates {
			if info.occurrences > 1 {
				allDuplicates = append(allDuplicates, *info)
			}
		}
	}

	// Sort duplicates by size
	sort.Slice(allDuplicates, func(i, j int) bool {
		return allDuplicates[i].size > allDuplicates[j].size
	})

	// Write combined duplicates table
	if len(allDuplicates) > 0 {
		content.WriteString("## Duplicate Content\n\n")
		content.WriteString("| Name | Type | Size | Occurrences | Locations |\n")
		content.WriteString("|------|------|------|-------------|------------|\n")

		for _, dup := range allDuplicates {
			contentType := "File"
			if dup.isAsset {
				contentType = "Asset"
			}

			content.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n",
				dup.name,
				contentType,
				formatSize(dup.size),
				dup.occurrences,
				strings.Join(dup.locations, "<br>")))
		}
		content.WriteString("\n")
	}

	// Write the markdown file
	if err := os.WriteFile(mdPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %v", err)
	}

	return nil
}

// findLargestModules returns a sorted list of largest modules (directories)
func findLargestModules(root analyzer.FileInfo) []moduleWithSize {
	modules := make([]moduleWithSize, 0)

	var traverse func(file analyzer.FileInfo)
	traverse = func(file analyzer.FileInfo) {
		if len(file.Children) > 0 {
			fileCount := 0
			var totalSize int64
			for _, child := range file.Children {
				if len(child.Children) == 0 {
					fileCount++
					totalSize += child.Size
				}
				traverse(child)
			}
			if totalSize > 0 {
				modules = append(modules, moduleWithSize{
					path:      file.RelativePath,
					size:      totalSize,
					fileCount: fileCount,
				})
			}
		}
	}

	traverse(root)

	// Sort by size in descending order
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].size > modules[j].size
	})

	return modules
}

// findLargestFiles returns a sorted list of largest individual files
func findLargestFiles(root analyzer.FileInfo) []fileWithSize {
	files := make([]fileWithSize, 0)

	var traverse func(file analyzer.FileInfo)
	traverse = func(file analyzer.FileInfo) {
		if len(file.Children) == 0 && file.Size > 0 {
			files = append(files, fileWithSize{
				path: file.RelativePath,
				size: file.Size,
			})
		}
		for _, child := range file.Children {
			traverse(child)
		}
	}

	traverse(root)

	// Sort by size in descending order
	sort.Slice(files, func(i, j int) bool {
		return files[i].size > files[j].size
	})

	return files
}

// findDuplicateFiles returns a map of SHA256 hashes to files with that hash
func findDuplicateFiles(root analyzer.FileInfo) map[string][]analyzer.FileInfo {
	duplicates := make(map[string][]analyzer.FileInfo)

	var traverse func(file analyzer.FileInfo)
	traverse = func(file analyzer.FileInfo) {
		if len(file.Children) == 0 && file.Shasum != "" {
			duplicates[file.Shasum] = append(duplicates[file.Shasum], file)
		}
		for _, child := range file.Children {
			traverse(child)
		}
	}

	traverse(root)

	return duplicates
}

// getRelativePaths returns a list of relative paths for the given files
func getRelativePaths(files []analyzer.FileInfo) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.RelativePath
	}
	return paths
}
