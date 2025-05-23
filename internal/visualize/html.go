package visualize

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"bitrise-plugins-analyze/internal/analyzer"
)

//go:embed templates/template.html
var tmplFS embed.FS

// templateData represents the data structure for the HTML template
type templateData struct {
	Title          string
	AppName        string
	BundleID       string
	Platform       string
	Version        string
	DownloadSize   string
	InstallSize    string
	FileTree       template.JS
	LargestFiles   []analyzer.FileInfo
	LargestModules []analyzer.FileInfo
	TypeBreakdown  []TypeBreakdown
	Duplicates     []DuplicateGroup
}

// formatSize converts bytes to a human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GenerateHTML generates an HTML visualization of the bundle analysis
func GenerateHTML(bundle *analyzer.AppBundle, outputDir string) error {
	// Parse the template from the embedded file
	tmpl, err := template.New("template.html").Funcs(template.FuncMap{
		"formatSize": formatSize,
	}).ParseFS(tmplFS, "templates/template.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Extract app name from the bundle path
	appName := bundle.AppName
	if filepath.Ext(appName) == ".app" {
		appName = appName[:len(appName)-4] // Remove .app extension
	}

	// Convert FileTree to JSON string to make it safe for JavaScript
	fileInfo, err := analyzer.FilesIncludingMetaInformation(bundle)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fileTreeJSON, err := json.Marshal(fileInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal file tree: %v", err)
	}

	// Pre-calculate largest files and modules
	largestFiles := FindLargestFiles(fileInfo)
	if len(largestFiles) > 10 {
		largestFiles = largestFiles[:10]
	}

	largestModules := FindLargestModules(fileInfo)
	if len(largestModules) > 10 {
		largestModules = largestModules[:10]
	}

	// Calculate type breakdown
	typeBreakdown := CalculateTypeBreakdown(fileInfo)

	// Find duplicate files
	duplicates := FindDuplicates(fileInfo)

	// Create template data
	data := templateData{
		Title:          "App Bundle Analysis",
		AppName:        appName,
		BundleID:       bundle.BundleID,
		Platform:       strings.Join(bundle.SupportedPlatforms, ", "),
		Version:        bundle.Version,
		DownloadSize:   formatSize(bundle.DownloadSize),
		InstallSize:    formatSize(bundle.InstallSize),
		FileTree:       template.JS(fileTreeJSON),
		LargestFiles:   largestFiles,
		LargestModules: largestModules,
		TypeBreakdown:  typeBreakdown,
		Duplicates:     duplicates,
	}

	// Create a buffer to store the rendered template
	var buf bytes.Buffer

	// Execute the template with the data
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Create HTML file named after bundle ID
	htmlFileName := fmt.Sprintf("%s.html", bundle.BundleID)
	htmlPath := filepath.Join(outputDir, htmlFileName)

	// Write the rendered template to the output file
	if err := os.WriteFile(htmlPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %v", err)
	}

	return nil
}
