package visualize

import (
	"bitrise-plugins-analyze/internal/analyzer"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateJSON generates a JSON file containing the bundle analysis data
func GenerateJSON(bundle *analyzer.AppBundle, outputDir string) error {
	// Create JSON file named after bundle ID
	jsonFileName := fmt.Sprintf("%s.json", bundle.BundleID)
	jsonPath := filepath.Join(outputDir, jsonFileName)

	// Marshal the bundle with indentation for better readability
	jsonData, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bundle data: %v", err)
	}

	// Write JSON file
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}

	return nil
}
