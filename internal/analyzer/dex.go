package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DexClass struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Shasum string `json:"shasum"`
}

type DexPackage struct {
	DexFilePath string     `json:"dex_file_path"`
	Name        string     `json:"name"`
	Size        int64      `json:"size"`
	Classes     []DexClass `json:"classes"`
}

type AsrcFile struct {
	ResourceId string `json:"resource_id"`
	Type       string `json:"type"`
	Name       string `json:"name"`
}

func analyzeDexFiles(unzipedApkDir string) ([]DexPackage, error) {
	allPackages := []DexPackage{}

	// Walk through the APK path directory
	err := filepath.Walk(unzipedApkDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Find any *.dex file
		if filepath.Ext(path) == ".dex" {
			// check dex file with dex_py
			dexFiles, err := AnalyzeDex(path)
			if err != nil {
				// Log the error but continue with other dex files
				fmt.Printf("Failed to analyze DEX file %s: %v\n", path, err)
				return nil
			}

			// relative dex file path
			relDexFilePath, err := filepath.Rel(unzipedApkDir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative dex file path: %v", err)
			}

			// Go through the analyzed dex files and create DexPackage objects
			for classes, size := range dexFiles {
				packageName := strings.TrimPrefix(filepath.Dir(classes), "L")
				className := strings.TrimSuffix(filepath.Base(classes), ";")

				// Skip classnames that include "$" as they are inner classes
				if strings.Contains(className, "$") {
					continue
				}

				// add class to the package
				addClassToPackage(&allPackages, relDexFilePath, packageName, className, int64(size))
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through extracted APK directory: %v", err)
	}

	return allPackages, nil
}

func addClassToPackage(dexPackages *[]DexPackage, dexFilePath string, packageName, className string, size int64) {
	packagePath := filepath.Join(dexFilePath, strings.ReplaceAll(packageName, "/", "."))

	var pkgIdx int = -1
	for i := range *dexPackages {
		if (*dexPackages)[i].Name == packagePath {
			pkgIdx = i
			break
		}
	}
	if pkgIdx == -1 {
		// Not found, create a new one and append
		*dexPackages = append(*dexPackages, DexPackage{
			DexFilePath: dexFilePath,
			Name:        packagePath,
			Size:        0,
			Classes:     make([]DexClass, 0),
		})
		pkgIdx = len(*dexPackages) - 1
	}
	// Add the class to the package via slice index
	(*dexPackages)[pkgIdx].Classes = append((*dexPackages)[pkgIdx].Classes, DexClass{
		Name:   className,
		Size:   size,
		Shasum: "", // SHA256 will be calculated later
	})
	(*dexPackages)[pkgIdx].Size += size
}

func (dexPackage DexPackage) GetPath() string {
	return filepath.Join(dexPackage.DexFilePath, dexPackage.Name)

}

func (dexClass DexClass) GetPath(dexPackage DexPackage) string {
	return filepath.Join(dexPackage.GetPath(), dexClass.Name)

}
