package platform

import (
	"bitrise-app-analyze/internal/analyzer"
	"bitrise-app-analyze/internal/platform/android"
	"bitrise-app-analyze/internal/platform/ios"
)

// NewIOSAnalyzer creates a new iOS analyzer
func NewIOSAnalyzer() analyzer.Analyzer {
	return ios.NewIOSAnalyzer()
}

// NewAndroidAnalyzer creates a new Android analyzer
func NewAndroidAnalyzer() analyzer.Analyzer {
	return android.NewAndroidAnalyzer()
}
