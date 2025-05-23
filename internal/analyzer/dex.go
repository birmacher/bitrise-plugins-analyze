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
	Name     string       `json:"name"`
	Size     int64        `json:"size"`
	Classes  []DexClass   `json:"classes"`
	Children []DexPackage `json:"children,omitempty"`
}

func generateDecompiledCode(apkPath, tempDir string) (string, error) {
	// Check if jadx is available
	jadxPath, err := exec.LookPath("jadx")
	if err != nil {
		return "", fmt.Errorf("jadx not found in PATH: %v", err)
	}

	// Create a container directory for the decompiled code
	apkFileName := filepath.Base(apkPath)
	apkFileNameWithoutExt := apkFileName[:len(apkFileName)-len(filepath.Ext(apkFileName))]
	decompiledDir := filepath.Join(tempDir, "jadx-"+apkFileNameWithoutExt)
	if err := os.MkdirAll(decompiledDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create decompiled code directory: %v", err)
	}

	// Run jadx to decompile the APK
	fmt.Println(jadxPath,
		// "--no-res", // Skip resources
		// "--output-dir", decompiledDir,
		"-d", decompiledDir,
		apkPath)
	cmd := exec.Command(jadxPath,
		// "--no-res", // Skip resources
		// "--output-dir", decompiledDir,
		"-d", decompiledDir,
		apkPath)

	output, err := cmd.CombinedOutput()
	fmt.Printf("jadx output: %s\n", string(output))
	if err != nil {
		return "", fmt.Errorf("failed to run jadx: %v", err)
	}

	return decompiledDir, nil
}

func analyzeDexFiles(apkPath, tempDir string) ([]DexPackage, error) {
	// First decompile the APK
	decompiledDir, err := generateDecompiledCode(apkPath, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to decompile APK: %v", err)
	}

	// Create package tree
	dexPackages := make([]DexPackage, 0)

	// Walk through the decompiled classes
	dexDir := filepath.Join(decompiledDir, "sources")
	err = filepath.Walk(dexDir, func(path string, info os.FileInfo, err error) error {
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

		// Get file size
		size := info.Size()

		// Get relative path to determine package
		relPath, err := filepath.Rel(dexDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Split the path into components to build package hierarchy
		components := strings.Split(filepath.Dir(relPath), string(os.PathSeparator))

		// Add to package tree
		currentLevel := &dexPackages
		var currentPath string

		for _, component := range components {
			if component == "." {
				continue
			}

			if currentPath == "" {
				currentPath = component
			} else {
				currentPath = filepath.Join(currentPath, component)
			}

			// Find or create package at current level
			var pkg *DexPackage
			for i := range *currentLevel {
				if (*currentLevel)[i].Name == currentPath {
					pkg = &(*currentLevel)[i]
					break
				}
			}

			if pkg == nil {
				*currentLevel = append(*currentLevel, DexPackage{
					Name:     currentPath,
					Size:     0,
					Classes:  make([]DexClass, 0),
					Children: make([]DexPackage, 0),
				})
				pkg = &(*currentLevel)[len(*currentLevel)-1]
			}

			pkg.Size += size
			currentLevel = &pkg.Children
		}

		// Add class info
		className := strings.TrimSuffix(info.Name(), ".java")
		if len(components) > 0 {
			pkg := findPackage(&dexPackages, filepath.Join(components...))
			if pkg != nil {
				pkg.Classes = append(pkg.Classes, DexClass{
					Name: className,
					Size: size,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze dex files: %v", err)
	}

	return dexPackages, nil
}

func findPackage(packages *[]DexPackage, path string) *DexPackage {
	for i := range *packages {
		if (*packages)[i].Name == path {
			return &(*packages)[i]
		}
		if strings.HasPrefix(path, (*packages)[i].Name+"/") {
			if pkg := findPackage(&(*packages)[i].Children, path); pkg != nil {
				return pkg
			}
		}
	}
	return nil
}
