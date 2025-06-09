package ios

import (
	"bitrise-plugins-analyze/appbundle/core"
	coreios "bitrise-plugins-analyze/appbundle/core/ios"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// convertToHEICFunc is used to convert an image to HEIC format.
// It can be overridden in tests.
var convertToHEICFunc = convertToHEIC

// extractCarAssetFunc is used to extract a rendition from a .car file.
// It can be overridden in tests.
var extractCarAssetFunc = extractCarAsset

// AnalyzeImages walks the provided bundle directory, converts supported images
// to HEIC format and reports potential savings.
// AnalyzeImages walks the provided bundle directory and the parsed .car files,
// converts supported images to HEIC format and reports potential savings.
func AnalyzeImages(bundleDir string, carFiles []coreios.CarFileInfo) ([]core.OversizedImage, error) {
	oversizedImages := []core.OversizedImage{}

	err := filepath.Walk(bundleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		lowerPath := strings.ToLower(path)
		ext := filepath.Ext(lowerPath)
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
			originalSize, convertedSize, err := convertToHEICFunc(path)
			if err != nil {
				return err
			}
			if convertedSize < originalSize {
				relPath, err := filepath.Rel(bundleDir, path)
				if err != nil {
					return err
				}
				oversizedImages = append(oversizedImages, core.OversizedImage{
					RelativePath: relPath,
					OriginalSize: originalSize,
					Saving:       originalSize - convertedSize,
					ConvertedTo:  "heic",
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Analyze images inside .car files
	for _, car := range carFiles {
		carPath := filepath.Join(bundleDir, car.Path)
		for _, asset := range car.Assets {
			for _, rendition := range asset.RenditionInfo {
				extracted, err := extractCarAssetFunc(carPath, rendition.RenditionName)
				if err != nil {
					continue
				}
				origSize, convSize, err := convertToHEICFunc(extracted)
				os.RemoveAll(filepath.Dir(extracted))
				if err != nil {
					continue
				}
				if convSize < origSize {
					rel := fmt.Sprintf("%s:%s", car.Path, rendition.RenditionName)
					oversizedImages = append(oversizedImages, core.OversizedImage{
						RelativePath: rel,
						OriginalSize: origSize,
						Saving:       origSize - convSize,
						ConvertedTo:  "heic",
					})
				}
			}
		}
	}

	sort.Slice(oversizedImages, func(i, j int) bool { return oversizedImages[i].Saving > oversizedImages[j].Saving })
	return oversizedImages, nil
}

// convertToHEIC converts the given image to HEIC format using either `sips` or
// `heif-convert` tools and returns the original and converted file sizes.
func convertToHEIC(imagePath string) (int64, int64, error) {
	sipsPath, errSips := exec.LookPath("sips")
	heifPath, errHeif := exec.LookPath("heif-convert")
	if errSips != nil && errHeif != nil {
		return 0, 0, fmt.Errorf("neither sips nor heif-convert found: please install one of them to analyze images")
	}

	tmpDir, err := os.MkdirTemp("", "heic-conversion")
	if err != nil {
		return 0, 0, err
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "converted.heic")
	var cmd *exec.Cmd
	if errSips == nil {
		cmd = exec.Command(sipsPath, "-s", "format", "heic", imagePath, "--out", tmpFile)
	} else {
		cmd = exec.Command(heifPath, imagePath, tmpFile)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("conversion failed: %v\nOutput: %s", err, string(output))
	}

	origInfo, err := os.Stat(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to stat original file: %v", err)
	}
	newInfo, err := os.Stat(tmpFile)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to stat heic file: %v", err)
	}
	return origInfo.Size(), newInfo.Size(), nil
}

// extractCarAsset extracts the specified rendition from the given .car file and
// returns the path to the extracted image. The caller is responsible for
// deleting the returned file's directory when done.
func extractCarAsset(carPath, rendition string) (string, error) {
	if _, err := exec.LookPath("assetutil"); err != nil {
		return "", fmt.Errorf("assetutil not found: this tool requires macOS")
	}

	tmpDir, err := os.MkdirTemp("", "car-asset")
	if err != nil {
		return "", err
	}
	tmpFile := filepath.Join(tmpDir, rendition+".png")
	cmd := exec.Command("assetutil", "--extract", rendition, "--output", tmpFile, carPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to extract asset: %v\nOutput: %s", err, string(output))
	}
	return tmpFile, nil
}
