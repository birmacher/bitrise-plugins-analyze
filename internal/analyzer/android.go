package analyzer

import (
	"fmt"
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
	// TODO: Remoe readableAPK and just use apkanalyzer on the bynaryXML
	// Read APK manifest and create bundle info
	bundle := &AppBundle{}

	readableApkPath, err := generateReadableApk(apkPath, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate APK for reading AndroidManifest: %v", err)
	}

	err = readAPKManifestFile(readableApkPath, bundle)
	if err != nil {
		return nil, err
	}

	// Unzip the APK to analyze its contents
	apkUnzipDir, err := unzipApk(apkPath, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to unzip APK: %v", err)
	}

	// Analyze the APK files
	files, err := AnalyzeFile(apkUnzipDir, apkUnzipDir)
	if err != nil {
		return nil, err
	}
	bundle.Files = files

	// Analyze DEX files
	// TODO: Export DEX files to file structure under the unzipped APK
	// Only after run analyzeFile as it will correctly setup the file structure
	dexPackages, err := analyzeDexFiles(apkPath, tempDir)
	if err != nil {
		// Log the error but don't fail the analysis
		fmt.Printf("Warning: failed to analyze DEX files: %v\n", err)
	} else {
		bundle.DexPackages = dexPackages
	}

	// Calculate sizes
	bundle.InstallSize = files.Size

	// Get the original APK file size for download size
	apkInfo, err := os.Stat(apkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get APK file size: %v", err)
	}
	bundle.DownloadSize = apkInfo.Size()

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

func readAPKManifestFile(apkContainerDir string, bundle *AppBundle) error {
	// Parse AndroidManifest.xml
	manifestPath := filepath.Join(apkContainerDir, "AndroidManifest.xml")
	manifest, err := parseAndroidManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to parse AndroidManifest.xml: %v", err)
	}

	fmt.Println("Manifest info:", manifest)

	// Update bundle with manifest info
	bundle.AppName = manifest.Application.Label
	bundle.BundleID = manifest.Package
	bundle.Version = manifest.VersionName
	bundle.SupportedPlatforms = []string{"Android"}

	// If app name wasn't in manifest, use file name
	if bundle.AppName == "" {
		bundle.AppName = filepath.Base(apkContainerDir)
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

func generateReadableApk(apkPath, tempDir string) (string, error) {
	// export APK with apktool to tempDir
	apktoolPath, err := exec.LookPath("apktool")
	if err != nil {
		return "", fmt.Errorf("apktool not found in PATH: %v", err)
	}

	// Create a container directory for the APK
	// Get the ApKPath file name without extension
	apkFileName := filepath.Base(apkPath)
	apkFileNameWithoutExt := apkFileName[:len(apkFileName)-len(filepath.Ext(apkFileName))]
	apkContainerDir := filepath.Join(tempDir, "apktool-"+apkFileNameWithoutExt)
	if err := os.MkdirAll(apkContainerDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create APK container directory: %v", err)
	}

	cmd := exec.Command(apktoolPath, "d", apkPath, "-o", apkContainerDir, "--force")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run apktool: %v", err)
	}

	return apkContainerDir, nil
}

func unzipApk(apkPath, tempDir string) (string, error) {
	// Create a container directory for the APK
	apkFileName := filepath.Base(apkPath)
	apkFileNameWithoutExt := apkFileName[:len(apkFileName)-len(filepath.Ext(apkFileName))]
	apkContainerDir := filepath.Join(tempDir, "unzip-"+apkFileNameWithoutExt)
	if err := os.MkdirAll(apkContainerDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create APK container directory: %v", err)
	}

	// Check if unzip is available
	unzipPath, err := exec.LookPath("unzip")
	if err != nil {
		return "", fmt.Errorf("unzip not found in PATH: %v", err)
	}

	// Unzip the APK
	cmd := exec.Command(unzipPath, apkPath, "-d", apkContainerDir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to unzip APK: %v", err)
	}

	return apkContainerDir, nil
}
