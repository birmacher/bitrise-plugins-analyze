package android

import "encoding/xml"

type AndroidManifest struct {
	XMLName     xml.Name `xml:"manifest"`
	Package     string   `xml:"package,attr"`
	VersionCode string   `xml:"versionCode,attr"`
	VersionName string   `xml:"versionName,attr"`
	Application struct {
		Label string `xml:"label,attr"`
	} `xml:"application"`
}
