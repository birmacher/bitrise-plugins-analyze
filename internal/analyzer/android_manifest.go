package analyzer

import (
	"encoding/xml"
	"fmt"
	"os"
)

type AndroidManifest struct {
	XMLName     xml.Name `xml:"manifest"`
	Package     string   `xml:"package,attr"`
	VersionCode string   `xml:"versionCode,attr"`
	VersionName string   `xml:"versionName,attr"`
	Application struct {
		Label string `xml:"label,attr"`
	} `xml:"application"`
}

func parseAndroidManifest(path string) (*AndroidManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %v", err)
	}

	var manifest AndroidManifest
	if err := xml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %v", err)
	}

	return &manifest, nil
}
