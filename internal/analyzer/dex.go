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

func analyzeDexFile(dexPath string) ([]DexPackage, error) {
	dexDir, err := generateDecompiledCode(dexPath)
	defer os.RemoveAll(dexDir)

	if err != nil {
		return nil, fmt.Errorf("failed to generate decompiled code: %v", err)
	}

	dexPackages := make([]DexPackage, 0)

	// Check if the sources directory exists
	sourcesDir := filepath.Join(dexDir, "sources")
	if _, err := os.Stat(sourcesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("sources directory not found in decompiled code")
	}

	// Walk through the decompiled sources
	err = filepath.Walk(sourcesDir, func(path string, info os.FileInfo, err error) error {
		fmt.Println("Analyzing file:", path)
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-Java files
		if !strings.HasSuffix(info.Name(), ".java") {
			return nil
		}

		fmt.Println(sourcesDir + ": " + path)

		// Get file size
		// size := info.Size()

		// // Get package path by calculating relative path from sources dir
		// relPath, err := filepath.Rel(sourcesDir, path)
		// if err != nil {
		// 	return fmt.Errorf("failed to get relative path: %v", err)
		// }

		// // Get package components from directory structure
		// dirPath := filepath.Dir(relPath)
		// packagePath := strings.ReplaceAll(dirPath, string(os.PathSeparator), ".")

		// // Find or create package in the hierarchy
		// className := strings.TrimSuffix(info.Name(), ".java")

		// Add or update package in the tree
		// addClassToPackageTree(&dexPackages, packagePath, className, size)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze decompiled sources: %v", err)
	}

	return dexPackages, nil
}

func generateDecompiledCode(dexFilePath string) (string, error) {
	fmt.Println("Generating decompiled code for DEX file:", dexFilePath)
	// Check if jadx is available
	jadxPath, err := exec.LookPath("jadx")
	if err != nil {
		return "", fmt.Errorf("jadx not found in PATH: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "*")

	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}

	// Run jadx to decompile the APK
	cmd := exec.Command(jadxPath,
		"--no-res", // Skip resources
		"--output-dir", tempDir,
		dexFilePath)

	output, err := cmd.CombinedOutput()
	fmt.Printf("jadx output: %s\n", string(output))
	if err != nil {
		return "", fmt.Errorf("failed to run jadx: %v", err)
	}

	return tempDir, nil
}

func analyzeDexFiles(unzipedApkDir string) ([]DexPackage, error) {
	fmt.Println("Analyzing DEX files in directory:", unzipedApkDir)
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
			fmt.Println("Found DEX file:", path)
			// Call generateDecompiledCode on each dex file
			_, err := generateDecompiledCode(path)
			if err != nil {
				// Log the error but continue with other dex files
				fmt.Printf("Warning: failed to decompile DEX file %s: %v\n", path, err)
				return nil
			}

			// // Analyze the decompiled code and get packages
			// packages, err := analyzeDecompiledCode(decompiledCodeDir)
			// if err != nil {
			// 	fmt.Printf("Warning: failed to analyze decompiled code for %s: %v\n", path, err)
			// 	return nil
			// }

			// // Merge packages with existing results
			// allPackages = append(allPackages, packages...)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through extracted APK directory: %v", err)
	}

	return allPackages, nil
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
