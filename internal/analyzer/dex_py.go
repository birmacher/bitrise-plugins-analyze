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

// //go:embed py/arsc_parsing.py
// var arscScript []byte

//go:embed py/requirements.txt
var requirements []byte

// AnalyzeDex returns a map of class names to code sizes for the given dex file.
func runPythonScript(pyScript []byte, requirements []byte, scriptArgs ...string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "dexpy")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write embedded Python script
	pyFile := filepath.Join(tmpDir, "script.py")
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
	args := append([]string{pyFile}, scriptArgs...)
	cmd = exec.Command(pythonBin, args...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python failed: %w\nOutput: %s", err, out)
	}
	return out, nil
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

	// Write embedded requirements.txt
	reqFile := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqFile, requirements, 0644); err != nil {
		return nil, fmt.Errorf("failed to write requirements: %w", err)
	}

	// Create or reuse venv (put in parent dir for caching)
	venvDir := filepath.Join(os.TempDir(), "bitrise-plugin-analyze-venv")
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

// 	// Use the same venv as runPythonScript
// 	venvDir := filepath.Join(os.TempDir(), "bitrise-pluglin-analyze-venv")
// 	pythonBin := filepath.Join(venvDir, "bin", "python")
// 	pipBin := filepath.Join(venvDir, "bin", "pip")
// 	androguardBin := filepath.Join(venvDir, "bin", "androguard")

// 	// Ensure venv exists
// 	if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
// 		cmd := exec.Command("python3", "-m", "venv", venvDir)
// 		if out, err := cmd.CombinedOutput(); err != nil {
// 			return nil, fmt.Errorf("venv failed: %w\nOutput: %s", err, out)
// 		}
// 	}

// 	// Upgrade pip, setuptools, wheel
// 	cmd := exec.Command(pipBin, "install", "--upgrade", "pip", "setuptools", "wheel")
// 	if out, err := cmd.CombinedOutput(); err != nil {
// 		return nil, fmt.Errorf("pip upgrade failed: %w\nOutput: %s", err, out)
// 	}

// 	// Ensure androguard is installed
// 	cmd = exec.Command(pipBin, "install", "androguard")
// 	if out, err := cmd.CombinedOutput(); err != nil {
// 		return nil, fmt.Errorf("pip install androguard failed: %w\nOutput: %s", err, out)
// 	}

// 	// Call androguard CLI from venv
// 	cmd = exec.Command(androguardBin, "disassemble", dexPath)
// 	out, err := cmd.CombinedOutput()
// 	fmt.Println(string(out))
// 	if err != nil {
// 		fmt.Println(out)
// 		return nil, fmt.Errorf("androguard failed: %w\nOutput: %s", err, out)
// 	}

// 	// Parse output for class name and size per line (assuming similar output as before)
// 	result := map[string]int{}
// 	lines := strings.Split(string(out), "\n")
// 	for _, line := range lines {
// 		parts := strings.Fields(line)
// 		if len(parts) != 2 {
// 			continue
// 		}
// 		name := parts[0]
// 		var size int
// 		fmt.Sscanf(parts[1], "%d", &size)
// 		result[name] = size
// 	}
// 	return result, nil
// }

func AnalyzeArsc(arscPath string) (map[string]map[string]string, error) {
	// Use the same venv as runPythonScript
	venvDir := filepath.Join(os.TempDir(), "dexpy-venv")
	pythonBin := filepath.Join(venvDir, "bin", "python")
	pipBin := filepath.Join(venvDir, "bin", "pip")
	androguardBin := filepath.Join(venvDir, "bin", "androguard")

	// Ensure venv exists
	if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
		cmd := exec.Command("python3", "-m", "venv", venvDir)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("venv failed: %w\nOutput: %s", err, out)
		}
	}

	// Upgrade pip, setuptools, wheel
	cmd := exec.Command(pipBin, "install", "--upgrade", "pip", "setuptools", "wheel")
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pip upgrade failed: %w\nOutput: %s", err, out)
	}

	// Ensure androguard is installed
	cmd = exec.Command(pipBin, "install", "androguard")
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pip install androguard failed: %w\nOutput: %s", err, out)
	}

	// Call androguard CLI from venv
	cmd = exec.Command(androguardBin, "arsc", "-i", arscPath)
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
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
