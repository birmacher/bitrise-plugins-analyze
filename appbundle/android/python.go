package android

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed py/dex_size.py
var pyScript []byte

//go:embed py/requirements.txt
var requirements []byte

// setupPythonEnvironment creates a Python virtual environment and installs required packages.
// It returns paths to the Python and pip binaries, as well as any error encountered.
func setupPythonEnvironment(venvName string, requirementsData []byte) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", venvName+"-tmp")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write requirements.txt if provided
	var reqFile string
	if len(requirementsData) > 0 {
		reqFile = filepath.Join(tmpDir, "requirements.txt")
		if err := os.WriteFile(reqFile, requirementsData, 0644); err != nil {
			return "", "", fmt.Errorf("failed to write requirements: %w", err)
		}
	}

	// Create or reuse venv (put in parent dir for caching)
	venvDir := filepath.Join(os.TempDir(), venvName+"-venv")
	pythonBin := filepath.Join(venvDir, "bin", "python")
	pipBin := filepath.Join(venvDir, "bin", "pip")

	// Create venv if it doesn't exist
	if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
		cmd := exec.Command("python3", "-m", "venv", venvDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", "", fmt.Errorf("venv creation failed: %w\nOutput: %s", err, out)
		}
	}

	// Upgrade pip, setuptools, wheel for best compatibility
	cmd := exec.Command(pipBin, "install", "--upgrade", "pip", "setuptools", "wheel")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("pip upgrade failed: %w\nOutput: %s", err, out)
	}

	// Install requirements if provided
	if reqFile != "" {
		cmd = exec.Command(pipBin, "install", "-r", reqFile)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", "", fmt.Errorf("pip install failed: %w\nOutput: %s", err, out)
		}
	}

	return pythonBin, pipBin, nil
}

func AnalyzeDex(dexPath string) (map[string]int, error) {
	tmpDir, err := os.MkdirTemp("", "dexpy")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write embedded Python script
	pyFile := filepath.Join(tmpDir, "dex_size.py")
	if err := os.WriteFile(pyFile, pyScript, 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Setup Python environment
	pythonBin, _, err := setupPythonEnvironment("bitrise-plugin-analyze", requirements)
	if err != nil {
		return nil, fmt.Errorf("failed to setup Python environment: %w", err)
	}

	// Run the script using the venv Python
	cmd := exec.Command(pythonBin, pyFile, dexPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python failed: %w\nOutput: %s", err, out)
	}

	// Parse output (class name and size per line)
	result := map[string]int{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		var size int
		fmt.Sscanf(parts[1], "%d", &size)
		result[name] = size
	}
	return result, nil
}

func AnalyzeArsc(arscPath string) (map[string]map[string]string, error) {
	// Setup Python environment
	_, pipBin, err := setupPythonEnvironment("bitrise-plugin-analyze", requirements)
	if err != nil {
		return nil, fmt.Errorf("failed to setup Python environment: %w", err)
	}

	// Ensure androguard is installed
	cmd := exec.Command(pipBin, "install", "androguard")
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pip install androguard failed: %w\nOutput: %s", err, out)
	}

	// Get path to androguard CLI
	androguardBin := filepath.Join(filepath.Dir(pipBin), "androguard")

	// Call androguard CLI from venv
	cmd = exec.Command(androguardBin, "arsc", "-i", arscPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("androguard failed: %w\nOutput: %s", err, out)
	}

	// Parse XML output for <public .../> tags
	result := map[string]map[string]string{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "<public ") && strings.HasSuffix(line, "/>") {
			attrs := map[string]string{}
			// Extract attributes: type, name, id
			for _, attr := range []string{"type", "name", "id"} {
				prefix := attr + "=\""
				start := strings.Index(line, prefix)
				if start == -1 {
					continue
				}
				start += len(prefix)
				end := strings.Index(line[start:], "\"")
				if end == -1 {
					continue
				}
				attrs[attr] = line[start : start+end]
			}
			if id, ok := attrs["id"]; ok {
				result[id] = map[string]string{
					"type": attrs["type"],
					"name": attrs["name"],
				}
			}
		}
	}
	return result, nil
}
