package python

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed py/dex_size.py
var dexScript []byte

//go:embed py/lief_parser.py
var liefScript []byte

//go:embed py/requirements.txt
var requirements []byte

type PythonEnvironment struct {
	DirPath string
}

func (env *PythonEnvironment) PythonBin() string {
	return filepath.Join(env.VenvDir(), "bin", "python")
}

func (env *PythonEnvironment) PipBin() string {
	return filepath.Join(env.VenvDir(), "bin", "pip")
}

func (env *PythonEnvironment) VenvDir() string {
	return filepath.Join(env.DirPath, "venv")
}

func (env *PythonEnvironment) isSetup() bool {
	return env.DirPath != ""
}

// pythonInterpreter returns the Python binary to use when creating the
// virtual environment. It reads the path from the PYTHON_BIN environment
// variable and falls back to "python3" if not set.
func pythonInterpreter() string {
	if bin := os.Getenv("PYTHON_BIN"); bin != "" {
		return bin
	}
	return "python3"
}

// Cleanup removes the temporary virtual environment directory
func (env *PythonEnvironment) Cleanup() {
	if env.DirPath != "" {
		os.RemoveAll(env.DirPath)
		env.DirPath = ""
	}
}

// setupPythonEnvironment creates a Python virtual environment and installs required packages.
// It returns paths to the Python and pip binaries, as well as any error encountered.
func (env *PythonEnvironment) SetupPythonEnvironment() error {
	var err error
	env.DirPath, err = os.MkdirTemp("", "*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Write requirements.txt if provided
	var reqFile string
	if len(requirements) > 0 {
		reqFile = filepath.Join(env.DirPath, "requirements.txt")
		if err := os.WriteFile(reqFile, requirements, 0644); err != nil {
			return fmt.Errorf("failed to write requirements: %w", err)
		}
	}

	fmt.Println("Using Python virtual environment at:", env.DirPath)

	// Create venv if it doesn't exist
	if _, err := os.Stat(env.PythonBin()); os.IsNotExist(err) {
		fmt.Println("  Creating virtual environment...")
		cmd := exec.Command(pythonInterpreter(), "-m", "venv", env.VenvDir())
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("venv creation failed: %w\nOutput: %s", err, out)
		}

		fmt.Println("  Upgradeing pip")

		// Upgrade pip, setuptools, wheel for best compatibility
		cmd = exec.Command(env.PipBin(), "install", "--upgrade", "pip", "setuptools", "wheel")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("pip upgrade failed: %w\nOutput: %s", err, out)
		}

		fmt.Println("  Installing requirements")

		// Install requirements if provided
		if reqFile != "" {
			cmd := exec.Command(env.PipBin(), "install", "-r", reqFile)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("pip install failed: %w\nOutput: %s", err, out)
			}
		}

		fmt.Println("  ✅ Environment setup complete")
	}

	return nil
}

func (env *PythonEnvironment) AnalyzeDex(dexPath string) (map[string]int64, error) {
	if !env.isSetup() {
		return nil, fmt.Errorf("python environment not set up")
	}

	// Write embedded Python script
	pyFile := filepath.Join(env.DirPath, "dex_size.py")
	if err := os.WriteFile(pyFile, dexScript, 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Run the script using the venv Python
	cmd := exec.Command(env.PythonBin(), pyFile, dexPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python failed: %w\nOutput: %s", err, out)
	}

	// Parse output (class name and size per line)
	result := map[string]int64{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		var size int64
		fmt.Sscanf(parts[1], "%d", &size)
		result[name] = size
	}
	return result, nil
}

func (env *PythonEnvironment) AnalyzeLief(liefPath string) (map[string]map[string]int64, error) {
	if !env.isSetup() {
		return nil, fmt.Errorf("python environment not set up")
	}

	// Write embedded Python script
	pyFile := filepath.Join(env.DirPath, "lief_parser.py")
	if err := os.WriteFile(pyFile, liefScript, 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Run the script using the venv Python
	cmd := exec.Command(env.PythonBin(), pyFile, liefPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python failed: %w\nOutput: %s", err, out)
	}

	result := map[string]map[string]int64{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		group := parts[0]
		name := parts[1]
		var size int64
		fmt.Sscanf(parts[2], "%d", &size)
		if _, ok := result[group]; !ok {
			result[group] = map[string]int64{}
		}
		result[group][name] = size
	}
	return result, nil
}

func (env *PythonEnvironment) AnalyzeArsc(arscPath string) (map[string]map[string]string, error) {
	if !env.isSetup() {
		return nil, fmt.Errorf("python environment not set up")
	}

	// Get path to androguard CLI
	androguardBin := filepath.Join(env.VenvDir(), "bin", "androguard")

	// Call androguard CLI from venv
	cmd := exec.Command(androguardBin, "arsc", "-i", arscPath)
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
