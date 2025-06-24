package appbundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeFile(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "img.png"), []byte("1234567890"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	fi, err := AnalyzeFile(tmp, tmp)
	if err != nil {
		t.Fatalf("AnalyzeFile returned error: %v", err)
	}

	if fi.Type != "directory" {
		t.Errorf("root type = %s, want directory", fi.Type)
	}
	if fi.Size != 5+10 {
		t.Errorf("root size = %d, want %d", fi.Size, 15)
	}
	if len(fi.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(fi.Children))
	}
	var foundTxt, foundImg bool
	for _, child := range fi.Children {
		switch child.RelativePath {
		case "sub":
			if child.Size != 5 {
				t.Errorf("sub size = %d, want 5", child.Size)
			}
			if child.Type != "directory" {
				t.Errorf("sub type = %s", child.Type)
			}
		case "img.png":
			foundImg = true
			if child.Type != "image" {
				t.Errorf("img.png type = %s", child.Type)
			}
			if child.Size != 10 {
				t.Errorf("img.png size = %d", child.Size)
			}
		}
		if child.RelativePath == "sub" {
			if len(child.Children) == 1 && child.Children[0].RelativePath == filepath.Join("sub", "a.txt") {
				foundTxt = true
			}
		}
	}
	if !foundImg {
		t.Error("img.png child not found")
	}
	if !foundTxt {
		t.Error("a.txt child not found in sub directory")
	}
}
