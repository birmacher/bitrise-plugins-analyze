package platform

import (
	"bitrise-plugins-analyze/internal/analyzer"
	"bitrise-plugins-analyze/internal/platform/android"
	"bitrise-plugins-analyze/internal/platform/ios"
)

// NewIOSAnalyzer creates a new iOS analyzer
func NewIOSAnalyzer() analyzer.Analyzer {
	return ios.NewIOSAnalyzer()
}

// NewAndroidAnalyzer creates a new Android analyzer
func NewAndroidAnalyzer() analyzer.Analyzer {
	return android.NewAndroidAnalyzer()
}
