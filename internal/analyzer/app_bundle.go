package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// AppBundle represents an analyzed application bundle
type AppBundle struct {
	DownloadSize       int64         `json:"download_size"`
	InstallSize        int64         `json:"install_size"`
	BundleID           string        `json:"bundle_id"`
	SupportedPlatforms []string      `json:"supported_platforms"`
	Version            string        `json:"version"`
	MinimumOSVersion   string        `json:"minimum_os_version"`
	AppName            string        `json:"app_name"`
	Files              FileInfo      `json:"files"`
	CarFiles           []CarFileInfo `json:"car_files,omitempty"`
	MachOFiles         []MachOInfo   `json:"mach_o_files,omitempty"`
	DexPackages        []DexPackage  `json:"dex_files,omitempty"`
}

// AnalyzeAppBundle analyzes the provided app bundle directory and returns the analysis results
func AnalyzeAppBundle(bundlePath string) (*AppBundle, error) {
	bundle := &AppBundle{}

	// Analyze the files in the bundle
	files, err := AnalyzeFile(bundlePath, bundlePath)
	if err != nil {
		return nil, err
	}

	bundle.AppName = filepath.Base(bundlePath)
	bundle.Files = files

	// Calculate download size
	bundle.DownloadSize, err = calculateDownloadSize(bundlePath)
	if err != nil {
		return nil, err
	}

	// Calculate install size using du command
	bundle.InstallSize, err = calculateInstallSize(bundlePath)
	if err != nil {
		return nil, err
	}

	// iOS app bundle
	if strings.HasSuffix(bundlePath, ".app") {
		// Analyze Info.plist
		err = AnalyzeInfoPlist(bundlePath, bundle)
		if err != nil {
			return nil, err
		}

		// Analyze .car files if present
		err := FindAndAnalyzeCarFiles(bundlePath, bundle)
		if err != nil {
			return nil, err
		}

		// Analyze Mach-O binaries
		err = FindAndAnalyzeMachO(bundlePath, bundle)
		if err != nil {
			return nil, err
		}
	}

	return bundle, nil
}

func calculateDownloadSize(bundlePath string) (int64, error) {
	tempDir, err := os.MkdirTemp("", "app-*")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(tempDir)

	// Create zip file path
	zipPath := filepath.Join(tempDir, "app.zip")

	// Run ditto command to create zip
	cmd := exec.Command("ditto", "-c", "-k", "--sequesterRsrc", "--keepParent", bundlePath, zipPath)
	if err := cmd.Run(); err != nil {
		return 0, err
	}

	// Get zip file size using stat
	cmd = exec.Command("stat", "-f%z", zipPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse size from stat output
	return strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
}

func calculateInstallSize(bundlePath string) (int64, error) {
	cmd := exec.Command("sh", "-c", "du -sk "+bundlePath+" | awk '{print $1 * 1024}'")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
}

func FilesIncludingMetaInformation(bundle *AppBundle) (FileInfo, error) {
	extendedFiles := bundle.Files

	// For each CAR file, find matching file in extendedFiles and add asset info as children
	for _, carFile := range bundle.CarFiles {
		// Find matching file path in extendedFiles tree
		var findAndAddAssets func(files *FileInfo) bool
		findAndAddAssets = func(files *FileInfo) bool {
			if files.RelativePath == carFile.Path {

				// Add each asset as a child
				for _, asset := range carFile.Assets {
					assetPath := filepath.Join(carFile.Path, asset.Name)
					assetInfo := FileInfo{
						RelativePath: assetPath,
						Type:         "image",
						Children:     make([]FileInfo, 0),
					}

					// Add renditions as children of the asset
					for _, rendition := range asset.RenditionInfo {
						renditionInfo := FileInfo{
							RelativePath: filepath.Join(assetPath, fmt.Sprintf("%s @ %dx (%s)", rendition.RenditionName, rendition.Scale, rendition.Idiom)),
							Size:         rendition.Size,
							Shasum:       rendition.Shasum,
							Type:         "image",
						}
						assetInfo.Children = append(assetInfo.Children, renditionInfo)

						assetInfo.Size += rendition.Size
					}

					files.Children = append(files.Children, assetInfo)
				}
				return true
			}

			// Recursively search children
			for i := range files.Children {
				if findAndAddAssets(&files.Children[i]) {
					return true
				}
			}
			return false
		}

		findAndAddAssets(&extendedFiles)
	}

	return extendedFiles, nil
}
