package visualize

import (
	"bitrise-plugins-analyze/internal/analyzer"
	"fmt"
	"sort"
)

// FindLargestFiles returns a sorted list of largest individual files
func FindLargestFiles(root analyzer.FileInfo) []analyzer.FileInfo {
	files := make([]analyzer.FileInfo, 0)

	var traverse func(file analyzer.FileInfo)
	traverse = func(file analyzer.FileInfo) {
		if len(file.Children) == 0 && file.Size > 0 {
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
func CountFiles(root analyzer.FileInfo) int {
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
func FindLargestModules(root analyzer.FileInfo) []analyzer.FileInfo {
	modules := make([]analyzer.FileInfo, 0)

	// Process only children of root to skip the root directory itself
	for _, child := range root.Children {
		var traverse func(file analyzer.FileInfo)
		traverse = func(file analyzer.FileInfo) {
			if len(file.Children) > 0 {
				var totalSize int64
				for _, child := range file.Children {
					if len(child.Children) == 0 {
						totalSize += child.Size
					}
					traverse(child)
				}
				if totalSize > 0 {
					// Create a new FileInfo for the module with calculated size
					moduleInfo := analyzer.FileInfo{
						RelativePath: file.RelativePath,
						Size:         totalSize,
						Children:     file.Children,
						Type:         "directory",
					}
					modules = append(modules, moduleInfo)
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

// TypeBreakdown represents size information for a specific file type
type TypeBreakdown struct {
	Type       string  `json:"type"`
	Size       int64   `json:"size"`
	Percentage float64 `json:"percentage"`
}

// CalculateTypeBreakdown returns a sorted list of size breakdowns by file type
func CalculateTypeBreakdown(root analyzer.FileInfo) []TypeBreakdown {
	breakdown := make(map[string]int64)
	totalSize := root.Size

	var traverse func(file analyzer.FileInfo)
	traverse = func(file analyzer.FileInfo) {
		if len(file.Children) == 0 {
			fileType := file.Type
			if fileType == "" {
				fileType = "unknown"
			}
			breakdown[fileType] += file.Size
		}
		for _, child := range file.Children {
			traverse(child)
		}
	}

	traverse(root)

	// Convert map to slice and calculate percentages
	result := make([]TypeBreakdown, 0, len(breakdown))
	for fileType, size := range breakdown {
		percentage := float64(size) / float64(totalSize) * 100
		result = append(result, TypeBreakdown{
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

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Files         []analyzer.FileInfo `json:"files"`
	Size          int64               `json:"size"`
	WastedSpace   int64               `json:"wasted_space"`
	TotalWasted   int64               `json:"total_wasted"`
	WastedPercent float64             `json:"wasted_percent"`
}

// FindDuplicates returns groups of duplicate files sorted by size
func FindDuplicates(root analyzer.FileInfo) []DuplicateGroup {
	fileMap := make(map[string][]analyzer.FileInfo)
	totalSize := root.Size

	var traverse func(file analyzer.FileInfo)
	traverse = func(file analyzer.FileInfo) {
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
	duplicates := make([]DuplicateGroup, 0)
	var totalWastedSpace int64

	for _, files := range fileMap {
		if len(files) > 1 {
			wastedSpace := files[0].Size * int64(len(files)-1)
			totalWastedSpace += wastedSpace

			duplicates = append(duplicates, DuplicateGroup{
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
