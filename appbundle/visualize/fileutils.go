package visualize

import (
	"bitrise-plugins-analyze/appbundle/core"
	"bitrise-plugins-analyze/appbundle/core/visualize"
	"fmt"
	"sort"
	"strings"
)

// FindLargestFiles returns a sorted list of largest individual files
func FindLargestFiles(root core.FileInfo) []core.FileInfo {
	files := make([]core.FileInfo, 0)

	var traverse func(file core.FileInfo)
	traverse = func(file core.FileInfo) {
		if len(file.Children) == 0 && file.Size > 0 && file.Type != core.FileTypeUnmapped {
			files = append(files, file)
		}
		for _, child := range file.Children {
			traverse(child)
		}
	}

	traverse(root)

	// Sort by size in descending order
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	return files
}

// CountFiles returns the number of files (non-directory nodes) in a FileInfo tree
func CountFiles(root core.FileInfo) int {
	count := 0
	if len(root.Children) == 0 {
		return 1
	}
	for _, child := range root.Children {
		if len(child.Children) == 0 {
			count++
		} else {
			count += CountFiles(child)
		}
	}
	return count
}

// FindLargestModules returns a sorted list of largest modules (directories)
func FindLargestModules(root core.FileInfo) []core.FileInfo {
	modules := make([]core.FileInfo, 0)

	// Process only children of root to skip the root directory itself
	for _, child := range root.Children {
		var traverse func(file core.FileInfo)
		traverse = func(file core.FileInfo) {
			if len(file.Children) > 0 && file.Type != core.FileTypeUnmapped {
				modules = append(modules, file)
				for _, child := range file.Children {
					traverse(child)
				}
			}
		}
		traverse(child)
	}

	// Sort by size in descending order
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Size > modules[j].Size
	})

	return modules
}

// CalculateTypeBreakdown returns a sorted list of size breakdowns by file type
func CalculateTypeBreakdown(root core.FileInfo) []visualize.TypeBreakdown {
	breakdown := make(map[string]int64)
	totalSize := root.Size

	var traverse func(file core.FileInfo)
	traverse = func(file core.FileInfo) {
		if len(file.Children) == 0 && file.Type != core.FileTypeDirectory {
			fileType := file.Type
			if fileType == "" {
				fileType = "Other"
			}
			breakdown[fileType] += file.Size
		}
		for _, child := range file.Children {
			traverse(child)
		}
	}

	traverse(root)

	// Convert map to slice and calculate percentages
	result := make([]visualize.TypeBreakdown, 0, len(breakdown))
	for fileType, size := range breakdown {
		percentage := float64(size) / float64(totalSize) * 100
		result = append(result, visualize.TypeBreakdown{
			Type:       fileType,
			Size:       size,
			Percentage: percentage,
		})
	}

	// Sort by size in descending order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Size > result[j].Size
	})

	return result
}

// FindDuplicates returns groups of duplicate files sorted by size
func FindDuplicates(root core.FileInfo) []visualize.DuplicateGroup {
	fileMap := make(map[string][]core.FileInfo)
	totalSize := root.Size

	var traverse func(file core.FileInfo)
	traverse = func(file core.FileInfo) {
		if len(file.Children) == 0 && file.Shasum != "" {
			// Create a key combining size and shasum to identify duplicates
			key := fmt.Sprintf("%d-%s", file.Size, file.Shasum)
			fileMap[key] = append(fileMap[key], file)
		}
		for _, child := range file.Children {
			traverse(child)
		}
	}

	traverse(root)

	// Convert map to slice of DuplicateGroup
	duplicates := make([]visualize.DuplicateGroup, 0)
	var totalWastedSpace int64

	for _, files := range fileMap {
		if len(files) > 1 {
			wastedSpace := files[0].Size * int64(len(files)-1)
			totalWastedSpace += wastedSpace

			duplicates = append(duplicates, visualize.DuplicateGroup{
				Files:       files,
				Size:        files[0].Size,
				WastedSpace: wastedSpace,
			})
		}
	}

	// Sort by wasted space in descending order
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].WastedSpace > duplicates[j].WastedSpace
	})

	// Calculate total wasted percentage for each group
	for i := range duplicates {
		duplicates[i].TotalWasted = totalWastedSpace
		duplicates[i].WastedPercent = float64(totalWastedSpace) / float64(totalSize) * 100
	}

	return duplicates
}

// Find those resources that are not in the arsc file
func FindMissingResources(bundle core.AppBundle) []core.FileInfo {
	var missingResources []core.FileInfo
	// Copy bundle.Files to a new var

	var traverse func(file core.FileInfo)
	traverse = func(file core.FileInfo) {
		if strings.HasPrefix(file.RelativePath, "res/") {
			// get res/{type}/{name}.{extension}
			parts := strings.Split(file.RelativePath, "/")
			if len(parts) == 3 {
				resourceType := parts[1]
				resourceName := strings.Split(parts[2], ".")[0]

				// Find in arscResources
				for _, asrc := range bundle.AsrcFiles {
					if strings.HasPrefix(resourceType, asrc.Type) && asrc.Name == resourceName {
						return
					}
				}
				// Not found in arsc, add to missing resources
				missingResources = append(missingResources, file)
			}
		}

		for _, child := range file.Children {
			traverse(child)
		}
	}

	traverse(bundle.Files)

	return missingResources
}
