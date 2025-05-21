package analyzer

import "strings"

// AppBundle represents an analyzed application bundle
type AppBundle struct {
	DownloadSize       int64         `json:"download_size"`
	InstallSize        int64         `json:"install_size"`
	BundleID           string        `json:"bundle_id"`
	SupportedPlatforms []string      `json:"supported_platforms"`
	Version            string        `json:"version"`
	MinimumOSVersion   string        `json:"minimum_os_version"`
	Files              FileInfo      `json:"files"`
	CarFiles           []CarFileInfo `json:"car_files,omitempty"`
}

// AnalyzeAppBundle analyzes the provided app bundle directory and returns the analysis results
func AnalyzeAppBundle(bundlePath string) (*AppBundle, error) {
	bundle := &AppBundle{}

	// Analyze the files in the bundle
	files, err := AnalyzeFile(bundlePath, bundlePath)
	if err != nil {
		return nil, err
	}

	bundle.Files = files
	// Todo: Correctly calculate download and install size
	bundle.DownloadSize = files.Size

	// iOS app bundle
	if strings.HasSuffix(bundlePath, ".app") {
		// Analyze Info.plist
		err = AnalyzeInfoPlist(bundlePath, bundle)
		if err != nil {
			return nil, err
		}

		// Analyze .car files if present
		carFiles, err := FindAndAnalyzeCarFiles(bundlePath)
		if err != nil {
			// Log the error but don't fail the analysis
			// Some bundles might not have .car files
			bundle.CarFiles = []CarFileInfo{}
		} else {
			bundle.CarFiles = carFiles
		}
	}

	return bundle, nil
}
