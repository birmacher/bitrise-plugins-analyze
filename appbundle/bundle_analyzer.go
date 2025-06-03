package appbundle

import (
	"bitrise-plugins-analyze/appbundle/core"
	"fmt"
	"path/filepath"
	"strings"
)

func AnalyzeBundlePath(bundle_path string) (*core.AppBundle, error) {
	ext := strings.ToLower(filepath.Ext(bundle_path))

	switch ext {
	case core.AppExtension, core.IpaExtension, core.XcarchiveExtension:
		return analyzeIOSBundle(bundle_path)
	case core.ApkExtension, core.AabExtension:
		return analyzeAndroidBundle(bundle_path)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
