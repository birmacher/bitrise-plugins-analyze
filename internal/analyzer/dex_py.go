package analyzer

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

// AnalyzeDex returns a map of class names to code sizes for the given dex file.
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

	// Write embedded requirements.txt
	reqFile := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqFile, requirements, 0644); err != nil {
		return nil, fmt.Errorf("failed to write requirements: %w", err)
	}

	// Create or reuse venv (put in parent dir for caching)
	venvDir := filepath.Join(os.TempDir(), "dexpy-venv")
	pythonBin := filepath.Join(venvDir, "bin", "python")
	pipBin := filepath.Join(venvDir, "bin", "pip")
	if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
		// venv does not exist, create it
		cmd := exec.Command("python3", "-m", "venv", venvDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("venv failed: %w\nOutput: %s", err, out)
		}
	}

	// Upgrade pip, setuptools, wheel for best compatibility
	cmd := exec.Command(pipBin, "install", "--upgrade", "pip", "setuptools", "wheel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pip upgrade failed: %w\nOutput: %s", err, out)
	}

	// Install requirements
	cmd = exec.Command(pipBin, "install", "-r", reqFile)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pip install failed: %w\nOutput: %s", err, out)
	}

	// Run the script using the venv Python
	cmd = exec.Command(pythonBin, pyFile, dexPath)
	out, err = cmd.CombinedOutput()
	fmt.Printf("Python script output:\n%s\n", out)
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
