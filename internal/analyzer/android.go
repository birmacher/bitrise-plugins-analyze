package analyzer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func analyzeAndroidBundle(bundle_path string) (*AppBundle, error) {
	ext := filepath.Ext(bundle_path)

	tempDir, err := os.MkdirTemp("", "android-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if ext == ApkExtension {
		return analyzeApk(bundle_path, tempDir)
	} else if ext == AabExtension {
		return analyzeAab(bundle_path, tempDir)
	}

	return nil, fmt.Errorf("unsupported Android file type: %s", ext)
}

func analyzeApk(apkPath string, tempDir string) (*AppBundle, error) {
	// Open and extract APK (it's a ZIP file)
	reader, err := zip.OpenReader(apkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open APK: %v", err)
	}
	defer reader.Close()

	// Extract APK contents
	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Create containing directory if it doesn't exist
		outPath := filepath.Join(tempDir, file.Name)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return nil, err
		}

		// Extract file
		outFile, err := os.Create(outPath)
		if err != nil {
			return nil, err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return nil, err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return nil, err
		}
	}

	// Create AppBundle with basic info
	bundle := &AppBundle{
		AppName:            filepath.Base(apkPath),
		SupportedPlatforms: []string{"Android"},
	}

	// Analyze the files in the bundle
	files, err := AnalyzeFile(tempDir, tempDir)
	if err != nil {
		return nil, err
	}
	bundle.Files = files

	// Calculate sizes
	bundle.InstallSize = files.Size
	bundle.DownloadSize = files.Size // For APK, download size is the same as file size

	return bundle, nil
}

func analyzeAab(aabPath string, tempDir string) (*AppBundle, error) {
	// Similar to APK but handling AAB specific structure
	// This would require bundletool to properly analyze
	return nil, fmt.Errorf("AAB analysis not yet implemented")
}
