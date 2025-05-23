package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func analyzeAndroidBundle(bundle_path string) (*AppBundle, error) {
	ext := filepath.Ext(bundle_path)

	if ext == ApkExtension {
		return analyzeApk(bundle_path)
	} else if ext == AabExtension {
		bundle_path, err := analyzeAab(bundle_path)
		defer os.RemoveAll(bundle_path)

		if err != nil {
			return nil, fmt.Errorf("failed to analyze AAB: %v", err)
		}

		return analyzeApk(bundle_path)
	}

	return nil, fmt.Errorf("unsupported Android file type: %s", ext)
}

func analyzeApk(apkPath string) (*AppBundle, error) {
	// Create bundle info
	bundle := &AppBundle{}

	// Parse AndroidManifest.xml using apkanalyzer
	manifest, err := parseAndroidManifest(apkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AndroidManifest.xml: %v", err)
	}

	// Set bundle metadata from manifest
	bundle.AppName = "Todo"
	bundle.BundleID = manifest.Package
	bundle.Version = manifest.VersionName + " (" + manifest.VersionCode + ")"

	unzipedApkDir, err := unzip(apkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unzip APK: %v", err)
	}

	// Analyze the APK files
	files, err := AnalyzeFile(unzipedApkDir, unzipedApkDir)
	if err != nil {
		return nil, err
	}
	bundle.Files = files

	// Analyze DEX files
	// TODO: Export DEX files to file structure under the unzipped APK
	// Only after run analyzeFile as it will correctly setup the file structure
	dexPackages, err := analyzeDexFiles(unzipedApkDir)
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

func analyzeAab(aabPath string) (string, error) {
	tempDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a debug keystore if it doesn't exist
	keystorePath := filepath.Join(tempDir, "debug.keystore")
	if err := createDebugKeystore(keystorePath); err != nil {
		return "", fmt.Errorf("failed to create debug keystore: %v", err)
	}

	// Convert AAB to universal APK
	universalApkPath := filepath.Join(tempDir, "universal.apk")
	apkPath, err := generateUniversalApk(aabPath, universalApkPath, keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to generate universal APK: %v", err)
	}

	return apkPath, nil
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

func generateUniversalApk(aabPath, outputPath, keystorePath string) (string, error) {
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
		return "", fmt.Errorf("failed to generate universal APK: %v", err)
	}

	apkPath, err := unzip(outputPath + ".apks")
	if err != nil {
		return "", fmt.Errorf("failed to unzip .apks file: %v", err)
	}
	return filepath.Join(apkPath, "universal.apk"), nil
}
