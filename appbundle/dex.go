package appbundle

import (
	"bitrise-plugins-analyze/appbundle/core/android"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func analyzeDexFiles(unzipedApkDir string) ([]android.DexPackage, error) {
	allPackages := []android.DexPackage{}

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

			// Go through the analyzed dex files and create DexPackage objects
			for classes, size := range dexFiles {
				packageName := strings.TrimPrefix(filepath.Dir(classes), "L")
				className := strings.TrimSuffix(filepath.Base(classes), ";")

				// Skip classnames that include "$" as they are inner classes
				if strings.Contains(className, "$") {
					continue
				}

				// add class to the package
				addClassToPackage(&allPackages, packageName, className, int64(size))
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through extracted APK directory: %v", err)
	}

	return allPackages, nil
}

func addClassToPackage(dexPackages *[]android.DexPackage, packageName, className string, size int64) {
	var pkgIdx int = -1
	for i := range *dexPackages {
		if (*dexPackages)[i].Name == packageName {
			pkgIdx = i
			break
		}
	}
	if pkgIdx == -1 {
		// Not found, create a new one and append
		*dexPackages = append(*dexPackages, android.DexPackage{
			Name:    packageName,
			Size:    0,
			Classes: make([]android.DexClass, 0),
		})
		pkgIdx = len(*dexPackages) - 1
	}
	// Add the class to the package via slice index
	(*dexPackages)[pkgIdx].Classes = append((*dexPackages)[pkgIdx].Classes, android.DexClass{
		Name:   className,
		Size:   size,
		Shasum: "", // SHA256 will be calculated later
	})
	(*dexPackages)[pkgIdx].Size += size
}
