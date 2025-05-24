package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DexClass struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type DexPackage struct {
	Name    string     `json:"name"`
	Size    int64      `json:"size"`
	Classes []DexClass `json:"classes"`
}

func generateDecompiledCode(dexFilePath string) (string, error) {
	// Check if jadx is available
	tempDir, err := os.MkdirTemp("", "*")

	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}

	// Run jadx to decompile the APK
	cmd := exec.Command("jadx",
		"--no-res", // Skip resources
		"--output-dir", tempDir,
		dexFilePath)

	err = cmd.Run()
	if err != nil {
		// Skip this error as most of the time jadx will fail to decompile some classes
		// but we still want to analyze the rest of the classes
		fmt.Println("Some error returned, but continuing for:", dexFilePath)
	}

	return tempDir, nil
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
			// Call generateDecompiledCode on each dex file
			decompiledCodeDir, err := generateDecompiledCode(path)
			if err != nil {
				// Log the error but continue with other dex files
				fmt.Printf("Failed to decompile DEX file %s: %v\n", path, err)
				return nil
			}

			// // Analyze the decompiled code and get packages
			packages, err := analyzeDecompiledCode(decompiledCodeDir)
			if err != nil {
				fmt.Printf("Warning: failed to analyze decompiled code for %s: %v\n", path, err)
				return nil
			}

			// Merge packages with existing results
			allPackages = append(allPackages, packages...)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through extracted APK directory: %v", err)
	}

	return allPackages, nil
}

func analyzeDecompiledCode(decompiledCodeDir string) ([]DexPackage, error) {
	dexPackages := []DexPackage{}

	// Check if the sources directory exists
	sourcesDir := filepath.Join(decompiledCodeDir, "sources")
	if _, err := os.Stat(sourcesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("sources directory not found in decompiled code")
	}

	// Analyze the decompiled sources
	err := filepath.Walk(sourcesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		classPath, err := filepath.Rel(sourcesDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		packagePath := filepath.Dir(classPath)
		className := strings.TrimSuffix(filepath.Base(classPath), ".java")
		size := info.Size()
		fmt.Println("Package:", packagePath, "class:", className, "size:", size)

		// Try to find the package in the slice
		var pkgIdx int = -1
		for i := range dexPackages {
			if dexPackages[i].Name == packagePath {
				pkgIdx = i
				break
			}
		}
		if pkgIdx == -1 {
			// Not found, create a new one and append
			dexPackages = append(dexPackages, DexPackage{
				Name:    packagePath,
				Size:    0,
				Classes: make([]DexClass, 0),
			})
			pkgIdx = len(dexPackages) - 1
		}

		// Add the class to the package via slice index
		dexPackages[pkgIdx].Classes = append(dexPackages[pkgIdx].Classes, DexClass{
			Name: className,
			Size: size,
		})
		dexPackages[pkgIdx].Size += size

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through dex file directory: %v", err)
	}

	return dexPackages, nil
}

// // Helper function to add a class to the package tree
// func addClassToPackageTree(packages *[]DexPackage, packagePath, className string, size int64) {
// 	// Handle empty package (default package)
// 	if packagePath == "." {
// 		packagePath = "default"
// 	}

// 	// Split the package path into components
// 	components := strings.Split(packagePath, ".")

// 	// Start at the root level
// 	currentLevel := packages
// 	var currentPath string

// 	// Navigate through the package hierarchy
// 	for i, component := range components {
// 		if currentPath == "" {
// 			currentPath = component
// 		} else {
// 			currentPath = currentPath + "." + component
// 		}

// 		// Find or create the package at this level
// 		var pkg *DexPackage
// 		for i := range *currentLevel {
// 			if (*currentLevel)[i].Name == currentPath {
// 				pkg = &(*currentLevel)[i]
// 				break
// 			}
// 		}

// 		// Create new package if not found
// 		if pkg == nil {
// 			*currentLevel = append(*currentLevel, DexPackage{
// 				Name:     currentPath,
// 				Size:     0,
// 				Classes:  make([]DexClass, 0),
// 				Children: make([]DexPackage, 0),
// 			})
// 			pkg = &(*currentLevel)[len(*currentLevel)-1]
// 		}

// 		// Add the class size to this package level
// 		pkg.Size += size

// 		// If this is the last component, add the class to this package
// 		if i == len(components)-1 {
// 			pkg.Classes = append(pkg.Classes, DexClass{
// 				Name: className,
// 				Size: size,
// 			})
// 		}

// 		// Move to the next level in the hierarchy
// 		currentLevel = &pkg.Children
// 	}
// }
