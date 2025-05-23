package analyzer

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	AppExtension       = ".app"
	IpaExtension       = ".ipa"
	XcarchiveExtension = ".xcarchive"
	ApkExtension       = ".apk"
	AabExtension       = ".aab"
)

func AnalyzeBundlePath(bundle_path string) (*AppBundle, error) {
	ext := strings.ToLower(filepath.Ext(bundle_path))

	switch ext {
	case AppExtension, IpaExtension, XcarchiveExtension:
		return analyzeIOSBundle(bundle_path)
	case ApkExtension, AabExtension:
		return analyzeAndroidBundle(bundle_path)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
