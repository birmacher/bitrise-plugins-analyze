package core

import "testing"

func TestIsIos(t *testing.T) {
	cases := map[string]bool{
		"app.app":          true,
		"archive.ipa":      true,
		"bundle.xcarchive": true,
		"file.apk":         false,
		"":                 false,
	}
	for path, expected := range cases {
		if IsIos(path) != expected {
			t.Errorf("IsIos(%q) = %v, want %v", path, IsIos(path), expected)
		}
	}
}

func TestIsAndroid(t *testing.T) {
	cases := map[string]bool{
		"test.apk":   true,
		"bundle.aab": true,
		"foo.app":    false,
		"":           false,
	}
	for path, expected := range cases {
		if IsAndroid(path) != expected {
			t.Errorf("IsAndroid(%q) = %v, want %v", path, IsAndroid(path), expected)
		}
	}
}
