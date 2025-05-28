package visualize

import (
	"bitrise-plugins-analyze/internal/analyzer"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
	content.WriteString(fmt.Sprintf("# 📱 App Bundle Analysis: %s\n\n", bundle.AppName))

	// Basic Information
	content.WriteString("| Bundle ID | Version | Download Size | Install Size |\n")
	content.WriteString("|------|---------|----------------|--------------|\n")
	content.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
		bundle.BundleID,
		bundle.Version,
		formatSize(bundle.DownloadSize),
		formatSize(bundle.InstallSize)))

	// Top Largest Modules
	modules := FindLargestModules(bundle.Files)
	topLargestModuleCount := min(len(modules), 10)
	content.WriteString("<details>\n")
	content.WriteString(fmt.Sprintf("<summary>📦 Top %d largest modules</summary>\n\n", topLargestModuleCount))
	content.WriteString("| Module | Size | % of Total |\n")
	content.WriteString("|------|------|------------|\n")
	for i, module := range modules {
		if i >= 10 {
			break
		}
		percentage := float64(module.Size) / float64(bundle.InstallSize) * 100
		content.WriteString(fmt.Sprintf("| %s | %s | %.1f%% |\n",
			module.RelativePath,
			formatSize(module.Size),
			percentage))
	}
	content.WriteString("\n</details>\n\n")

	// Top Largest Files
	files := FindLargestFiles(bundle.Files)
	topLargestFilesCount := min(len(files), 10)
	content.WriteString("<details>\n")
	content.WriteString(fmt.Sprintf("<summary>📄 Top %d largest files</summary>\n\n", topLargestFilesCount))
	content.WriteString("| File | Size | % of Total |\n")
	content.WriteString("|------|------|------------|\n")
	for i, file := range files {
		if i >= 10 {
			break
		}
		percentage := float64(file.Size) / float64(bundle.InstallSize) * 100
		content.WriteString(fmt.Sprintf("| %s | %s | %.1f%% |\n",
			file.RelativePath,
			formatSize(file.Size),
			percentage))
	}
	content.WriteString("</details>\n")

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
		content.WriteString("## 🔄 Duplicate Content\n\n")
		content.WriteString("<details>\n")

		// Calculate total size of duplicates
		totalDuplicateSize := int64(0)
		for _, dup := range allDuplicates {
			totalDuplicateSize += dup.size * int64(dup.occurrences-1) // Count only the duplicate space
		}

		content.WriteString(fmt.Sprintf("<summary>Found %d duplicated items wasting %s of space, click to expand</summary>\n\n",
			len(allDuplicates), formatSize(totalDuplicateSize)))
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
		content.WriteString("\n</details>\n\n")
	}

	// Write the markdown file
	if err := os.WriteFile(mdPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %v", err)
	}

	return nil
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
