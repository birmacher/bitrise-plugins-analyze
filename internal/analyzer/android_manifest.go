package analyzer

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type AndroidManifest struct {
	XMLName     xml.Name `xml:"manifest"`
	Package     string   `xml:"package,attr"`
	VersionCode string   `xml:"versionCode,attr"`
	VersionName string   `xml:"versionName,attr"`
	Application struct {
		Label string `xml:"label,attr"`
	} `xml:"application"`
}

func parseAndroidManifest(apkPath string) (*AndroidManifest, error) {
	// Path to the apkanalyzer tool
	apkanalyzerPath := filepath.Join(os.Getenv("HOME"), "Library/Android/sdk/cmdline-tools/latest/bin/apkanalyzer")

	// Check if apkanalyzer exists
	if _, err := os.Stat(apkanalyzerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("apkanalyzer not found at %s", apkanalyzerPath)
	}

	// Prepare the command to extract AndroidManifest.xml
	cmd := exec.Command(apkanalyzerPath, "manifest", "print", apkPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute apkanalyzer: %v, output: %s", err, string(output))
	}

	// Parse the XML output into the AndroidManifest struct
	var manifest AndroidManifest
	if err := xml.Unmarshal(output, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse AndroidManifest.xml: %v", err)
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

	return &manifest, nil
}
