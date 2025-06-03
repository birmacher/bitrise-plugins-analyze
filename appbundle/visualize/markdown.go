package visualize

import (
	"bitrise-plugins-analyze/appbundle/core"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateMarkdown generates a Markdown file containing the bundle analysis data
func GenerateMarkdown(bundle *core.AppBundle, outputDir string) error {
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
	content.WriteString("</details>\n\n")

	// Collect all duplicates
	duplicates := FindDuplicates(bundle.Files)

	// Write combined duplicates table
	if len(duplicates) > 0 {
		content.WriteString("<details>\n")
		content.WriteString(fmt.Sprintf("<summary>🔄 %d duplicate items found</summary>\n\n", len(duplicates)))

		// Calculate total size of duplicates
		totalDuplicateSize := int64(0)
		for _, group := range duplicates {
			totalDuplicateSize += group.Size * int64(len(group.Files)-1) // Only count wasted space
		}

		content.WriteString("| Name | Type | Size | Occurrences | Locations |\n")
		content.WriteString("|------|------|------|-------------|------------|\n")

		for _, group := range duplicates {
			locations := make([]string, len(group.Files))
			for i, f := range group.Files {
				locations[i] = f.RelativePath
			}
			content.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n",
				filepath.Base(group.Files[0].RelativePath),
				group.Files[0].Type,
				formatSize(group.Size),
				len(group.Files),
				strings.Join(locations, "<br>")))
		}
		content.WriteString("\n</details>\n\n")
	}

	// Add missing resources section
	missingResources := FindMissingResources(*bundle)
	if len(missingResources) > 0 {
		content.WriteString("## ❓ Missing Resources\n\n")
		content.WriteString("The following resources were expected but not found in the bundle:\n\n")
		content.WriteString("| Resource | Type | Size |\n")
		content.WriteString("|----------|------|------|\n")
		for _, res := range missingResources {
			content.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
				res.RelativePath,
				res.Type,
				formatSize(res.Size)))
		}
		content.WriteString("\n")
	}

	// Write the markdown file
	if err := os.WriteFile(mdPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %v", err)
	}

	return nil
}

// findDuplicateFiles returns a map of SHA256 hashes to files with that hash
func findDuplicateFiles(root core.FileInfo) map[string][]core.FileInfo {
	duplicates := make(map[string][]core.FileInfo)

	var traverse func(file core.FileInfo)
	traverse = func(file core.FileInfo) {
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
func getRelativePaths(files []core.FileInfo) []string {
	paths := make([]string, len(files))
	for i, file := range files {
		paths[i] = file.RelativePath
	}
	return paths
}
