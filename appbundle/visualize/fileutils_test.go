package visualize

import (
	corepkg "bitrise-plugins-analyze/appbundle/core"
	androidcore "bitrise-plugins-analyze/appbundle/core/android"
	"testing"
)

func testTree() corepkg.FileInfo {
	return corepkg.FileInfo{
		RelativePath: ".",
		Type:         "directory",
		Size:         108,
		Children: []corepkg.FileInfo{
			{RelativePath: "a.txt", Size: 10, Type: "binary", Shasum: "sha"},
			{RelativePath: "b.txt", Size: 10, Type: "binary", Shasum: "sha"},
			{
				RelativePath: "dir",
				Type:         "directory",
				Size:         80,
				Children: []corepkg.FileInfo{
					{RelativePath: "dir/c.png", Size: 30, Type: "image", Shasum: "shaB"},
					{
						RelativePath: "dir/sub",
						Type:         "directory",
						Size:         50,
						Children: []corepkg.FileInfo{
							{RelativePath: "dir/sub/d.ttc", Size: 50, Type: "font", Shasum: "shaC"},
						},
					},
				},
			},
			{
				RelativePath: "res",
				Type:         "directory",
				Size:         8,
				Children: []corepkg.FileInfo{
					{RelativePath: "res/drawable/icon.png", Size: 5, Type: "image", Shasum: "shaD"},
					{RelativePath: "res/layout/main.xml", Size: 3, Type: "binary", Shasum: "shaE"},
				},
			},
		},
	}
}

func TestCountFiles(t *testing.T) {
	root := testTree()
	if n := CountFiles(root); n != 6 {
		t.Errorf("CountFiles = %d, want 6", n)
	}
}

func TestFindLargestFiles(t *testing.T) {
	root := testTree()
	files := FindLargestFiles(root)
	if len(files) == 0 || files[0].RelativePath != "dir/sub/d.ttc" {
		t.Fatalf("unexpected largest file: %+v", files)
	}
	if files[0].Size != 50 {
		t.Errorf("largest size = %d, want 50", files[0].Size)
	}
}

func TestFindLargestModules(t *testing.T) {
	root := testTree()
	mods := FindLargestModules(root)
	if len(mods) < 2 {
		t.Fatalf("expected modules")
	}
	if mods[0].RelativePath != "dir" || mods[0].Size != 80 {
		t.Errorf("unexpected first module %+v", mods[0])
	}
	if mods[1].RelativePath != "dir/sub" {
		t.Errorf("second module path = %s", mods[1].RelativePath)
	}
}

func TestCalculateTypeBreakdown(t *testing.T) {
	root := testTree()
	bd := CalculateTypeBreakdown(root)
	if len(bd) == 0 || bd[0].Type != "font" {
		t.Fatalf("unexpected breakdown: %+v", bd)
	}
	if bd[0].Percentage < 46 || bd[0].Percentage > 47 {
		t.Errorf("font percentage = %.1f", bd[0].Percentage)
	}
}

func TestFindDuplicates(t *testing.T) {
	root := testTree()
	dups := FindDuplicates(root)
	if len(dups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(dups))
	}
	if dups[0].WastedSpace != 10 {
		t.Errorf("wasted space = %d, want 10", dups[0].WastedSpace)
	}
}

func TestFindMissingResources(t *testing.T) {
	root := testTree()
	bundle := corepkg.AppBundle{Files: root,
		AsrcFiles: []androidcore.AsrcFile{{Type: "drawable", Name: "icon"}},
	}
	miss := FindMissingResources(bundle)
	if len(miss) != 1 || miss[0].RelativePath != "res/layout/main.xml" {
		t.Fatalf("unexpected missing resources: %+v", miss)
	}
}

func TestFormatSize(t *testing.T) {
	if formatSize(500) != "500 B" {
		t.Errorf("formatSize small failed")
	}
	if formatSize(1024) != "1.0 KB" {
		t.Errorf("formatSize KB failed: %s", formatSize(1024))
	}
	if formatSize(1024*1024) != "1.0 MB" {
		t.Errorf("formatSize MB failed: %s", formatSize(1024*1024))
	}
}
