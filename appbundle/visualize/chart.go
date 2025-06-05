package visualize

import (
	"bitrise-plugins-analyze/appbundle/core"
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

func GeneratePlotlyChart(bundle *core.AppBundle) Chart {
	ExpandAndroidFiles(bundle)

	chart := Chart{
		Labels:  []string{},
		Parents: []string{},
		Values:  []int64{},
		Types:   []string{},
		Ids:     []string{},
	}

	var traverseFiles func(files core.FileInfo, parent string)
	traverseFiles = func(files core.FileInfo, parent string) {
		if parent == "" {
			chart.Labels = append(chart.Labels, bundle.AppName)
			chart.Values = append(chart.Values, bundle.InstallSize)
		} else {
			chart.Labels = append(chart.Labels, filepath.Base(files.RelativePath))
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

func ExpandAndroidFiles(bundle *core.AppBundle) {
	ExpandDexFiles(bundle)
}

func ExpandDexFiles(bundle *core.AppBundle) {
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
	dexFileInfo := core.FileInfo{
		RelativePath: "dex",
		Type:         core.FileTypeDirectory,
		Size:         totalDexSize,
		Children:     make([]core.FileInfo, 0),
	}

	dexPackageSizes := int64(0)
	for _, dexPackage := range bundle.DexPackages {
		parent := &dexFileInfo

		if dexPackage.GetPath() == "" || dexPackage.GetPath() == "." {
			continue
		}

		dexPackagePath := filepath.Join("dex", dexPackage.GetPath())
		dexPackagePathParts := strings.Split(dexPackagePath, "/")

		for i := 1; i < len(dexPackagePathParts)-1; i++ {
			intermediatePath := filepath.Join("dex", strings.Join(dexPackagePathParts[1:i+1], "/"))

			if dexPackage.Size == 0 {
				continue
			}

			exists := false
			for j := range parent.Children {
				if parent.Children[j].RelativePath == intermediatePath {
					parent.Children[j].Size += dexPackage.Size
					parent = &parent.Children[j]
					exists = true
					break
				}
			}
			if !exists {
				parent.Children = append(parent.Children, core.FileInfo{
					RelativePath: intermediatePath,
					Type:         core.FileTypeDirectory,
					Size:         dexPackage.Size,
					Children:     make([]core.FileInfo, 0),
				})
				parent = &parent.Children[len(parent.Children)-1]
			}
		}

		parent.Children = append(parent.Children, core.FileInfo{
			RelativePath: filepath.Join(dexPackagePath, strings.ReplaceAll(dexPackage.GetPath(), "/", ".")),
			Type:         core.FileTypeDex,
			Size:         dexPackage.Size,
			Children:     make([]core.FileInfo, 0),
		})
		dexPackageSizes += dexPackage.Size
	}

	// Add Unmapped dex file size if there are any
	if totalDexSize-dexPackageSizes > 0 {
		dexFileInfo.Children = append(dexFileInfo.Children, core.FileInfo{
			RelativePath: "dex/Unmapped dex",
			Type:         core.FileTypeDex,
			Size:         totalDexSize - dexPackageSizes,
			Children:     make([]core.FileInfo, 0),
		})
	}

	dexFileInfo.Children = simplifyDirectory(dexFileInfo.Children)

	// Append Dex Packages to the bundle
	bundle.Files.Children = append(bundle.Files.Children, dexFileInfo)
}

func simplifyDirectory(children []core.FileInfo) []core.FileInfo {
	for i, child := range children {
		children[i].Children = simplifyDirectory(child.Children)
	}

	for i, child := range children {
		if child.Type == core.FileTypeDirectory && len(child.Children) == 1 && child.Children[0].Type == core.FileTypeDirectory {
			relativePath := strings.Join([]string{child.RelativePath, filepath.Base(child.Children[0].RelativePath)}, ".")
			children[i] = child.Children[0]
			children[i].RelativePath = relativePath
		}
	}

	return children
}
