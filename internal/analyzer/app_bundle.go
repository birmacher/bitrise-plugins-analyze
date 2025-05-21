package analyzer

// AppBundle represents an analyzed application bundle
type AppBundle struct {
	DownloadSize       int64    `json:"download_size"`
	InstallSize        int64    `json:"install_size"`
	BundleID           string   `json:"bundle_id"`
	SupportedPlatforms []string `json:"supported_platforms"`
	Version            string   `json:"version"`
	MinimumOSVersion   string   `json:"minimum_os_version"`
	Files              FileInfo `json:"files"`
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
	bundle.DownloadSize = files.Size

	return bundle, nil
}
