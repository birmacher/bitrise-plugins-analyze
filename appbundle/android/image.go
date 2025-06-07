package android

import (
	"bitrise-plugins-analyze/appbundle/core"
	"bitrise-plugins-analyze/appbundle/core/android"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func AnalyzeImages(bundleDir string, manifest android.AndroidManifest) ([]core.OversizedImage, error) {
	oversizedImages := []core.OversizedImage{}
	err := filepath.Walk(bundleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		lowerPath := strings.ToLower(path)
		extension := filepath.Ext(lowerPath)
		if extension == ".png" || extension == ".jpg" || extension == ".bmp" {
			// .9.pngs not supported with WebP
			if strings.HasSuffix(lowerPath, ".9.png") {
				return nil
			}

			// Transparent WebP not supported under SDK level 18
			if extension == ".png" && manifest.UsesSdk.MinSdk < 18 {
				transparent, err := hasTransparency(path)
				if err != nil {
					return err
				}
				if transparent {
					return nil
				}
			}

			originalSize, convertedSize, err := convertToWebP(path)
			if err != nil {
				return err
			}

			if convertedSize < originalSize {
				Relpath, err := filepath.Rel(bundleDir, path)
				if err != nil {
					return err
				}

				oversizedImages = append(oversizedImages, core.OversizedImage{
					RelativePath: Relpath,
					OriginalSize: originalSize,
					Saving:       originalSize - convertedSize,
					ConvertedTo:  "webp",
				})
			}
		}

		return nil
	})

	return oversizedImages, err
}

func convertToWebP(imagePath string) (int64, int64, error) {
	if _, err := exec.LookPath("cwebp"); err != nil {
		return 0, 0, fmt.Errorf("cwebp not found: please install it to analyze images")
	}

	tmpDir, err := os.MkdirTemp("", "webp-conversion")
	if err != nil {
		return 0, 0, err
	}
	defer (func() { os.RemoveAll(tmpDir) })()

	tmpFile := filepath.Join(tmpDir, "converted.webp")

	cmd := exec.Command("cwebp", imagePath, "-o", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("conversion failed: %v\nOutput: %s", err, string(output))
	}

	// Get Original image size
	originalInfo, err := os.Stat(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to stat original file: %v", err)
	}

	// Get WebP file size
	webpInfo, err := os.Stat(tmpFile)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to stat webp file: %v", err)
	}

	return originalInfo.Size(), webpInfo.Size(), nil
}

func hasTransparency(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return false, fmt.Errorf("not a PNG or failed to decode: %v", err)
	}

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 0xffff { // 0xffff = 65535 (fully opaque in 16-bit)
				return true, nil
			}
		}
	}
	return false, nil
}
