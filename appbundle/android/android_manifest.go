package android

import (
	"bitrise-plugins-analyze/appbundle/core/android"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ParseAndroidManifest(apkPath string) (android.AndroidManifest, error) {
	var manifest android.AndroidManifest

	var searchPaths []string
	var apkanalyzerPath string

	if sdkRoot := os.Getenv("ANDROID_SDK_ROOT"); sdkRoot != "" {
		candidate := filepath.Join(sdkRoot, "cmdline-tools/latest/bin/apkanalyzer")
		searchPaths = append(searchPaths, candidate)
		if _, err := os.Stat(candidate); err == nil {
			apkanalyzerPath = candidate
		}
	}

	if apkanalyzerPath == "" {
		if androidHome := os.Getenv("ANDROID_HOME"); androidHome != "" {
			candidate := filepath.Join(androidHome, "cmdline-tools/latest/bin/apkanalyzer")
			searchPaths = append(searchPaths, candidate)
			if _, err := os.Stat(candidate); err == nil {
				apkanalyzerPath = candidate
			}
		}
	}

	if apkanalyzerPath == "" {
		candidate := filepath.Join(os.Getenv("HOME"), "Library/Android/sdk/cmdline-tools/latest/bin/apkanalyzer")
		searchPaths = append(searchPaths, candidate)
		if _, err := os.Stat(candidate); err == nil {
			apkanalyzerPath = candidate
		}
	}

	if apkanalyzerPath == "" {
		return manifest, fmt.Errorf("apkanalyzer not found. searched: %s", strings.Join(searchPaths, ", "))
	}

	// Prepare the command to extract AndroidManifest.xml
	cmd := exec.Command(apkanalyzerPath, "manifest", "print", apkPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return manifest, fmt.Errorf("failed to execute apkanalyzer: %v, output: %s", err, string(output))
	}

	// Parse the XML output into the AndroidManifest struct
	if err := xml.Unmarshal(output, &manifest); err != nil {
		return manifest, fmt.Errorf("failed to parse AndroidManifest.xml: %v", err)
	}

	// If version code or name are empty, try to extract them with more specific commands
	if manifest.VersionCode == "" || manifest.VersionName == "" {
		// Get version code
		cmdVersionCode := exec.Command(apkanalyzerPath, "manifest", "get-attr", "--xpath", "/manifest", "versionCode", apkPath)
		versionCodeOutput, err := cmdVersionCode.CombinedOutput()
		if err == nil {
			manifest.VersionCode = strings.TrimSpace(string(versionCodeOutput))
		}

		// Get version name
		cmdVersionName := exec.Command(apkanalyzerPath, "manifest", "get-attr", "--xpath", "/manifest", "versionName", apkPath)
		versionNameOutput, err := cmdVersionName.CombinedOutput()
		if err == nil {
			manifest.VersionName = strings.TrimSpace(string(versionNameOutput))
		}
	}

	// Get application label if empty
	if manifest.Application.Label == "" {
		cmdAppLabel := exec.Command(apkanalyzerPath, "manifest", "get-attr", "--xpath", "/manifest/application", "android:label", apkPath)
		appLabelOutput, err := cmdAppLabel.CombinedOutput()
		if err == nil {
			manifest.Application.Label = strings.TrimSpace(string(appLabelOutput))
		}
	}

	return manifest, nil
}
