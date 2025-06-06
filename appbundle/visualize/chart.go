package visualize

import (
	"bitrise-plugins-analyze/appbundle/core"
	"fmt"
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
	ExpandDexFiles(bundle)
	ExpandBinaryFiles(bundle)
	ExpandCarFiles(bundle)

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
		} else {
			chart.Labels = append(chart.Labels, filepath.Base(files.RelativePath))
		}
		chart.Values = append(chart.Values, files.Size)
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

func ExpandBinaryFiles(bundle *core.AppBundle) {
	for _, file := range bundle.BinaryFiles {
		var traverseFiles func(files *core.FileInfo)
		traverseFiles = func(files *core.FileInfo) {
			if files.RelativePath == file.Path {
				var mappedSectionSizes int64
				for _, section := range file.Sections {
					mappedSectionSizes += section.Size

					sectionFile := core.FileInfo{
						RelativePath: filepath.Join(file.Path, section.Name),
						Type:         core.FileTypeBinary,
						Size:         section.Size,
						Children:     make([]core.FileInfo, 0),
					}

					// for _, symbol := range section.Symbols {
					// 	sectionFile.Children = append(sectionFile.Children, core.FileInfo{
					// 		RelativePath: filepath.Join(file.Path, section.Name, symbol.Name),
					// 		Type:         core.FileTypeBinary,
					// 		Size:         symbol.Size,
					// 		Children:     make([]core.FileInfo, 0),
					// 	})
					// }

					files.Children = append(files.Children, sectionFile)
				}

				unmappedSize := files.Size - mappedSectionSizes
				if unmappedSize > 0 {
					files.Children = append(files.Children, core.FileInfo{
						RelativePath: filepath.Join(file.Path, "Unmapped"),
						Type:         core.FileTypeBinary,
						Size:         unmappedSize,
						Children:     make([]core.FileInfo, 0),
					})
				}
				return
			}

			for i := range files.Children {
				traverseFiles(&files.Children[i])
			}
		}
		traverseFiles(&bundle.Files)
	}
}

func ExpandCarFiles(bundle *core.AppBundle) {
	for _, car := range bundle.CarFiles {
		var traverseFiles func(files *core.FileInfo)
		traverseFiles = func(files *core.FileInfo) {
			if files.RelativePath == car.Path {
				var totalCarSize int64
				for _, asset := range car.Assets {
					var totalAssetSize int64

					assetFile := core.FileInfo{
						RelativePath: filepath.Join(car.Path, asset.Name),
						Type:         core.FileTypeDirectory,
						Size:         0,
						Children:     make([]core.FileInfo, 0),
					}
					for _, rendition := range asset.RenditionInfo {
						totalAssetSize += rendition.Size

						assetFile.Children = append(assetFile.Children, core.FileInfo{
							RelativePath: filepath.Join(car.Path, asset.Name, fmt.Sprintf("%s@%dx~%s", rendition.RenditionName, rendition.Scale, rendition.Idiom)),
							Type:         core.FileTypeImage,
							Size:         rendition.Size,
							Children:     make([]core.FileInfo, 0),
						})
					}
					assetFile.Size = totalAssetSize
					totalCarSize += totalAssetSize

					files.Children = append(files.Children, assetFile)
				}

				unmappedSize := files.Size - totalCarSize
				if unmappedSize > 0 {
					files.Children = append(files.Children, core.FileInfo{
						RelativePath: filepath.Join(car.Path, "Unmapped"),
						Type:         core.FileTypeBinary,
						Size:         unmappedSize,
						Children:     make([]core.FileInfo, 0),
					})
				}
				return
			}

			for i := range files.Children {
				traverseFiles(&files.Children[i])
			}
		}
		traverseFiles(&bundle.Files)
	}
}
