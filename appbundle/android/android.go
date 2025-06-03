package android

import (
	"bitrise-plugins-analyze/appbundle/core"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func GetTempAPKPath(bundle_path string) (string, error) {
	ext := filepath.Ext(bundle_path)

	if ext == core.ApkExtension {
		return bundle_path, nil
	} else if ext == core.AabExtension {
		bundle_path, err := analyzeAab(bundle_path)

		if err != nil {
			return "", fmt.Errorf("failed to analyze AAB: %v", err)
		}

		return bundle_path, nil
	}

	return "", fmt.Errorf("unsupported Android file type: %s", ext)
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

	apkPath, err := core.Unzip(outputPath + ".apks")
	if err != nil {
		return "", fmt.Errorf("failed to unzip .apks file: %v", err)
	}
	return filepath.Join(apkPath, "universal.apk"), nil
}
