package visualize

import (
	"bitrise-plugins-analyze/internal/analyzer"
	"path/filepath"
	"slices"
	"strings"
)

type Chart struct {
	Labels  []string `json:"labels"`
	Parents []string `json:"parents"`
	Values  []int64  `json:"values"`
	Types   []string `json:"types"`
	Ids     []string `json:"ids"`
}

func GeneratePlotlyChart(bundle *analyzer.AppBundle) Chart {
	ExpandAndroidFiles(bundle)

	chart := Chart{
		Labels:  []string{},
		Parents: []string{},
		Values:  []int64{},
		Types:   []string{},
		Ids:     []string{},
	}

	var traverseFiles func(files analyzer.FileInfo, parent string)
	traverseFiles = func(files analyzer.FileInfo, parent string) {
		if parent == "" {
			chart.Labels = append(chart.Labels, bundle.AppName)
			chart.Values = append(chart.Values, bundle.InstallSize)
		} else {
			chart.Labels = append(chart.Labels, files.RelativePath)
			chart.Values = append(chart.Values, files.Size)
		}
		chart.Parents = append(chart.Parents, parent)
		chart.Types = append(chart.Types, files.Type)
		chart.Ids = append(chart.Ids, files.RelativePath)

		for _, child := range files.Children {
			traverseFiles(child, files.RelativePath)
		}
	}
	traverseFiles(bundle.Files, "")

	return chart
}

func ExpandAndroidFiles(bundle *analyzer.AppBundle) {
	ExpandDexFiles(bundle)
}

func ExpandDexFiles(bundle *analyzer.AppBundle) {
	// Calculate total .dex file sizes
	// Dex files are located at the root level
	totalDexSize := int64(0)
	for i := len(bundle.Files.Children) - 1; i >= 0; i-- {
		file := bundle.Files.Children[i]
		if strings.HasSuffix(strings.ToLower(file.RelativePath), ".dex") {
			totalDexSize += file.Size
			bundle.Files.Children = slices.Delete(bundle.Files.Children, i, i+1)
		}
	}
	// add a new root level file for dex files
	dexFileInfo := analyzer.FileInfo{
		RelativePath: "dex",
		Type:         "directory",
		Size:         totalDexSize,
		Children:     make([]analyzer.FileInfo, 0),
	}

	dexPackageSizes := int64(0)
	for _, dexPackage := range bundle.DexPackages {
		dexFileInfo.Children = append(dexFileInfo.Children, analyzer.FileInfo{
			RelativePath: filepath.Join("dex", dexPackage.GetPath()),
			Type:         "Dex",
			Size:         dexPackage.Size,
			Children:     make([]analyzer.FileInfo, 0),
		})
		dexPackageSizes += dexPackage.Size
	}

	// Add Unmapped dex file size if there are any
	if totalDexSize-dexPackageSizes > 0 {
		dexFileInfo.Children = append(dexFileInfo.Children, analyzer.FileInfo{
			RelativePath: "dex/Unmapped dex",
			Type:         "Dex",
			Size:         totalDexSize - dexPackageSizes,
			Children:     make([]analyzer.FileInfo, 0),
		})
	}

	// Append Dex Packages to the bundle
	bundle.Files.Children = append(bundle.Files.Children, dexFileInfo)
}
