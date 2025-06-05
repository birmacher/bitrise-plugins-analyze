package core

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func createTestZip(path string, files map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			return err
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			return err
		}
	}
	return w.Close()
}

func TestUnzip(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "sample.zip")
	files := map[string]string{
		"file1.txt":     "hello",
		"dir/file2.txt": "world",
	}
	if err := createTestZip(zipPath, files); err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}

	outDir, err := Unzip(zipPath)
	if err != nil {
		t.Fatalf("Unzip returned error: %v", err)
	}
	for name, content := range files {
		data, err := os.ReadFile(filepath.Join(outDir, name))
		if err != nil {
			t.Fatalf("missing file %s: %v", name, err)
		}
		if string(data) != content {
			t.Errorf("file %s content mismatch: got %q want %q", name, data, content)
		}
	}
}
