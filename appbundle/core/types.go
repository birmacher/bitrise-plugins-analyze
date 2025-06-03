package core

import "path/filepath"

const (
	AppExtension       = ".app"
	IpaExtension       = ".ipa"
	XcarchiveExtension = ".xcarchive"
	ApkExtension       = ".apk"
	AabExtension       = ".aab"
)

func IsIos(path string) bool {
	extension := filepath.Ext(path)
	return extension == AppExtension || extension == IpaExtension || extension == XcarchiveExtension
}

func IsAndroid(path string) bool {
	extension := filepath.Ext(path)
	return extension == ApkExtension || extension == AabExtension
}
