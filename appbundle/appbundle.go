package appbundle

import (
	"bitrise-plugins-analyze/appbundle/android"
	"bitrise-plugins-analyze/appbundle/core"
	"bitrise-plugins-analyze/appbundle/ios"
	"fmt"
	"path/filepath"
	"strings"
)

func Analyze(bundle_path string) (*core.AppBundle, error) {
	ext := strings.ToLower(filepath.Ext(bundle_path))

	var err error
	var tmp_path string

	switch ext {
	case core.AppExtension, core.IpaExtension, core.XcarchiveExtension:
		tmp_path, err = ios.GetTempIPAPath(bundle_path)
	case core.ApkExtension, core.AabExtension:
		tmp_path, err = android.GetTempAPKPath(bundle_path)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return nil, err
	}
	return AnalyzeAppBundle(tmp_path)
}
