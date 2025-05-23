package analyzer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	// Create a debug keystore if it doesn't exist
	keystorePath := filepath.Join(tempDir, "debug.keystore")
	if err := createDebugKeystore(keystorePath); err != nil {
		return nil, fmt.Errorf("failed to create debug keystore: %v", err)
	}

	// Convert AAB to universal APK
	universalApkPath := filepath.Join(tempDir, "universal.apk")
	if err := generateUniversalApk(aabPath, universalApkPath, keystorePath); err != nil {
		return nil, fmt.Errorf("failed to generate universal APK: %v", err)
	}

	// Now analyze the universal APK
	return analyzeApk(universalApkPath, tempDir)
}

func createDebugKeystore(keystorePath string) error {
	// Create a debug keystore for signing
	cmd := exec.Command("keytool", "-genkeypair",
		"-keystore", keystorePath,
		"-alias", "debug",
		"-keyalg", "RSA",
		"-keysize", "2048",
		"-validity", "10000",
		"-dname", "CN=Debug,OU=Development,O=Bitrise,L=Debug,S=Debug,C=US",
		"-storepass", "android",
		"-keypass", "android")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create debug keystore: %v", err)
	}

	return nil
}

func generateUniversalApk(aabPath, outputPath, keystorePath string) error {
	// Generate universal APK from AAB
	cmd := exec.Command("bundletool",
		"build-apks",
		"--bundle="+aabPath,
		"--output="+outputPath+".apks",
		"--mode=universal",
		"--ks="+keystorePath,
		"--ks-pass=pass:android",
		"--ks-key-alias=debug",
		"--key-pass=pass:android")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate universal APK: %v", err)
	}

	// Extract the universal.apk from the .apks file (it's just a zip)
	cmd = exec.Command("unzip", "-p", outputPath+".apks", "universal.apk")
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract universal APK: %v", err)
	}

	return nil
}
