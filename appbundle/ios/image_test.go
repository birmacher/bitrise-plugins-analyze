package ios

import (
	coreios "bitrise-plugins-analyze/appbundle/core/ios"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeImages(t *testing.T) {
	dir := t.TempDir()

	// create dummy files
	if err := os.WriteFile(filepath.Join(dir, "smaller.png"), []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bigger.jpg"), []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// stub conversion
	origFunc := convertToHEICFunc
	defer func() { convertToHEICFunc = origFunc }()
	convertToHEICFunc = func(p string) (int64, int64, error) {
		if strings.Contains(p, "smaller") {
			return 1000, 600, nil
		}
		return 1000, 1200, nil
	}

	imgs, err := AnalyzeImages(dir, nil)
	if err != nil {
		t.Fatalf("AnalyzeImages returned error: %v", err)
	}

	if len(imgs) != 1 {
		t.Fatalf("expected 1 oversized image, got %d", len(imgs))
	}
	if imgs[0].RelativePath != "smaller.png" {
		t.Errorf("unexpected relative path: %s", imgs[0].RelativePath)
	}
	if imgs[0].Saving != 400 {
		t.Errorf("unexpected saving: %d", imgs[0].Saving)
	}
}

func TestAnalyzeImagesFromCAR(t *testing.T) {
	dir := t.TempDir()

	carFile := coreios.CarFileInfo{
		Path: "dummy.car",
		Assets: []coreios.AssetInfo{{
			Name:          "icon",
			RenditionInfo: []coreios.RenditionInfo{{RenditionName: "icon"}},
		}},
	}

	origConv := convertToHEICFunc
	origExtract := extractCarAssetFunc
	defer func() {
		convertToHEICFunc = origConv
		extractCarAssetFunc = origExtract
	}()

	convertToHEICFunc = func(p string) (int64, int64, error) { return 1000, 600, nil }
	extractCarAssetFunc = func(carPath, rendition string) (string, error) { return filepath.Join(dir, "icon.png"), nil }

	imgs, err := AnalyzeImages(dir, []coreios.CarFileInfo{carFile})
	if err != nil {
		t.Fatalf("AnalyzeImages returned error: %v", err)
	}
	if len(imgs) != 1 {
		t.Fatalf("expected 1 oversized image, got %d", len(imgs))
	}
	if imgs[0].RelativePath != "dummy.car:icon" {
		t.Errorf("unexpected relative path: %s", imgs[0].RelativePath)
	}
	if imgs[0].Saving != 400 {
		t.Errorf("unexpected saving: %d", imgs[0].Saving)
	}
}
