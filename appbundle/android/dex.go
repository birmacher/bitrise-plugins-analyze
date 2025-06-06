package android

import (
	"bitrise-plugins-analyze/appbundle/core"
	"bitrise-plugins-analyze/appbundle/core/android"
	"path/filepath"
	"strings"
)

func ParseDexOutput(dexOutput map[string]int64, bundle *core.AppBundle) error {
	// Go through the dex output and create DexPackage objects
	for classes, size := range dexOutput {
		packageName := strings.TrimPrefix(filepath.Dir(classes), "L")
		className := strings.TrimSuffix(filepath.Base(classes), ";")

		// Skip classnames that include "$" as they are inner classes
		if strings.Contains(className, "$") {
			continue
		}

		// add class to the package
		addClassToPackage(&bundle.DexPackages, packageName, className, size)
	}

	return nil
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
		Name: className,
		Size: size,
	})
	(*dexPackages)[pkgIdx].Size += size
}
