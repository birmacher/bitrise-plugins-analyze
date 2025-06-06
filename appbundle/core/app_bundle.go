package core

import (
	"bitrise-plugins-analyze/appbundle/core/android"
	"bitrise-plugins-analyze/appbundle/core/ios"
)

// AppBundle represents an analyzed application bundle
type AppBundle struct {
	DownloadSize       int64                `json:"download_size"`
	InstallSize        int64                `json:"install_size"`
	BundleID           string               `json:"bundle_id"`
	SupportedPlatforms []string             `json:"supported_platforms"`
	Version            string               `json:"version"`
	MinimumOSVersion   string               `json:"minimum_os_version"`
	AppName            string               `json:"app_name"`
	Files              FileInfo             `json:"files"`
	CarFiles           []ios.CarFileInfo    `json:"car_files,omitempty"`
	MachOFiles         []any                `json:"mach_o_files,omitempty"`
	BinaryFiles        []LiefBinary         `json:"binary_files,omitempty"`
	DexPackages        []android.DexPackage `json:"dex_files,omitempty"`
	AsrcFiles          []android.AsrcFile   `json:"arsc_files,omitempty"`
	OversizedImages    []OversizedImage     `json:"oversized_images,omitempty"`
}
