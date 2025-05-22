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
	for i, module := range modules {
		if i >= 10 {
			break
		}
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

	// Duplicate Files
	content.WriteString("## Duplicate Files\n\n")
	duplicates := findDuplicateFiles(bundle.Files)
	if len(duplicates) > 0 {
		content.WriteString("### File System Duplicates\n\n")
		content.WriteString("| SHA256 | Size | Occurrences | Paths |\n")
		content.WriteString("|--------|------|-------------|--------|\n")

		for shasum, files := range duplicates {
			if len(files) > 1 { // Only show actual duplicates
				content.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n",
					shasum[:8], // Show only first 8 chars of hash
					formatSize(files[0].Size),
					len(files),
					strings.Join(getRelativePaths(files), "<br>")))
			}
		}
		content.WriteString("\n")
	}

	// CAR File Duplicates
	if len(bundle.CarFiles) > 0 {
		content.WriteString("### Asset Catalog Duplicates\n\n")
		content.WriteString("| Asset Name | Size | Occurrences | Locations |\n")
		content.WriteString("|------------|------|-------------|------------|\n")

		// Map to track duplicates across all CAR files
		assetDuplicates := make(map[string][]string)
		assetSizes := make(map[string]int64)

		for _, car := range bundle.CarFiles {
			for _, asset := range car.Assets {
				for _, rendition := range asset.RenditionInfo {
					if rendition.Shasum != "" {
						key := fmt.Sprintf("%s:%s", asset.Name, rendition.Shasum)
						assetDuplicates[key] = append(assetDuplicates[key],
							fmt.Sprintf("%s (%s)", car.Path, rendition.RenditionName))
						assetSizes[key] = rendition.Size
					}
				}
			}
		}

		// Filter and sort duplicates
		for key, locations := range assetDuplicates {
			if len(locations) > 1 {
				assetName := strings.Split(key, ":")[0]
				content.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n",
					assetName,
					formatSize(assetSizes[key]),
					len(locations),
					strings.Join(locations, "<br>")))
			}
		}
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
