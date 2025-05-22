package visualize

import (
	"bitrise-plugins-analyze/internal/analyzer"
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
