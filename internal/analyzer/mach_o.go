package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MachOInfo represents information about a Mach-O binary
type MachOInfo struct {
	Path         string   `json:"path"`
	Architecture []string `json:"architecture"`
	LoadCommands []string `json:"load_commands,omitempty"`
	MinOSVersion string   `json:"min_os_version,omitempty"`
	LinkedLibs   []string `json:"linked_libraries,omitempty"`
	RPaths       []string `json:"rpaths,omitempty"`
	Size         int64    `json:"size"`
}

// FindAndAnalyzeMachO searches for and analyzes Mach-O binaries in the bundle
func FindAndAnalyzeMachO(bundlePath string, bundle *AppBundle) error {
	// Check if otool exists
	if _, err := exec.LookPath("otool"); err != nil {
		return fmt.Errorf("otool not found: this tool requires macOS")
	}

	err := filepath.Walk(bundlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-regular files
		if info.IsDir() || !info.Mode().IsRegular() {
			return nil
		}

		// Run file command to check if it's a Mach-O binary
		cmd := exec.Command("file", path)
		output, err := cmd.Output()
		if err != nil {
			return nil // Skip if file command fails
		}

		// Check if the file is a Mach-O binary
		if !strings.Contains(string(output), "Mach-O") {
			return nil
		}

		// Analyze the Mach-O binary
		machO, err := analyzeMachO(path)
		if err != nil {
			return fmt.Errorf("failed to analyze Mach-O binary %s: %v", path, err)
		}

		// Add to bundle's Mach-O information
		bundle.MachOFiles = append(bundle.MachOFiles, *machO)
		return nil
	})

	return err
}

// analyzeMachO analyzes a single Mach-O binary using otool
func analyzeMachO(path string) (*MachOInfo, error) {
	info := &MachOInfo{
		Path: path,
	}

	// Get file size
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	info.Size = fileInfo.Size()

	// Get architectures
	cmd := exec.Command("lipo", "-info", path)
	output, err := cmd.Output()
	if err == nil {
		// Parse architectures from lipo output
		parts := strings.Split(string(output), ":")
		if len(parts) > 1 {
			archs := strings.Split(strings.TrimSpace(parts[len(parts)-1]), " ")
			info.Architecture = archs
		}
	}

	// Get load commands and linked libraries
	cmd = exec.Command("otool", "-l", "-L", path)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("otool failed: %v", err)
	}

	// Parse otool output
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Extract minimum OS version
		if strings.Contains(line, "LC_VERSION_MIN_IPHONEOS") && i+2 < len(lines) {
			versionLine := strings.TrimSpace(lines[i+2])
			if strings.HasPrefix(versionLine, "version") {
				parts := strings.Fields(versionLine)
				if len(parts) > 1 {
					info.MinOSVersion = parts[1]
				}
			}
		}

		// Extract linked libraries
		if strings.HasPrefix(line, "/") && strings.Contains(line, ".dylib") {
			lib := strings.Fields(line)[0]
			info.LinkedLibs = append(info.LinkedLibs, lib)
		}

		// Extract RPaths
		if strings.Contains(line, "LC_RPATH") && i+2 < len(lines) {
			pathLine := strings.TrimSpace(lines[i+2])
			if strings.HasPrefix(pathLine, "path") {
				parts := strings.Fields(pathLine)
				if len(parts) > 1 {
					info.RPaths = append(info.RPaths, parts[1])
				}
			}
		}
	}

	return info, nil
}
