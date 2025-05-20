package analyzer

// AnalysisResult contains the results of analyzing an app
type AnalysisResult struct {
	// Common fields
	AppName            string      `json:"app_name"`
	BundleID           string      `json:"bundle_id"`
	Version            string      `json:"version"`
	BuildNumber        string      `json:"build_number"`
	Platform           string      `json:"platform"`
	FileSize           int64       `json:"file_size"`
	SigningInfo        SigningInfo `json:"signing_info"`
	EmbeddedFrameworks []Framework `json:"embedded_frameworks,omitempty"`
	Resources          []Resource  `json:"resources,omitempty"`
}

// SigningInfo contains information about the app's signing
type SigningInfo struct {
	TeamID              string                 `json:"team_id"`
	TeamName            string                 `json:"team_name"`
	ProvisioningProfile string                 `json:"provisioning_profile,omitempty"`
	Entitlements        map[string]interface{} `json:"entitlements,omitempty"`
}

// Framework represents an embedded framework
type Framework struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

// Resource represents an embedded resource
type Resource struct {
	Type string `json:"type"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// Analyzer defines the interface for platform-specific analyzers
type Analyzer interface {
	// Analyze takes a path to an app file and returns analysis results
	Analyze(path string) (*AnalysisResult, error)
}
