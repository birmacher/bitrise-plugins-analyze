package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
)

func analyzeAppBundle(app_path string) error {
	ext := strings.ToLower(filepath.Ext(app_path))

	switch ext {
	case AppExtension:
		return analyzeApp(app_path)
	case IpaExtension:
		return analyzeIpa(app_path)
	case XcarchiveExtension:
		return analyzeXcarchive(app_path)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func analyzeApp(app_path string) error {
	// Todo: Implement .app analysis
	return nil
}

func analyzeIpa(app_path string) error {
	// Todo: Implement .ipa analysis
	return nil
}

func analyzeXcarchive(app_path string) error {
	// Todo: Implement .xcarchive analysis
	return nil
}
