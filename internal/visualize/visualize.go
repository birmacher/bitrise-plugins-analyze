package visualize

import (
	"fmt"
	"strings"

	"bitrise-plugins-analyze/internal/analyzer"
)

// Visualize generates a visualization of the app bundle analysis
func Visualize(bundle *analyzer.AppBundle) error {
	fmt.Println("📱 App Bundle Analysis")
	fmt.Printf("Bundle ID: %s\n", bundle.BundleID)
	fmt.Printf("Version: %s\n", bundle.Version)
	fmt.Printf("Minimum OS Version: %s\n", bundle.MinimumOSVersion)
	fmt.Printf("Supported Platforms: %s\n", strings.Join(bundle.SupportedPlatforms, ", "))
	fmt.Printf("Download Size: %.2f MB\n", float64(bundle.DownloadSize)/(1024*1024))
	fmt.Printf("Install Size: %.2f MB\n\n", float64(bundle.InstallSize)/(1024*1024))

	// Show duplicates if requested
	duplicates, err := FindDuplicates(bundle)
	if err != nil {
		return fmt.Errorf("failed to analyze duplicates: %v", err)
	}
	if len(duplicates) > 0 {
		fmt.Println("👥 Duplicate Files:")
		var totalWasted int64
		for checksum, group := range duplicates {
			wastedSpace := int64(len(group)-1) * group[0].Size
			totalWasted += wastedSpace

			fmt.Printf("  Checksum: %s\n", checksum[:8])
			fmt.Printf("  Size: %.2f MB (%.2f MB wasted)\n", float64(group[0].Size)/(1024*1024), float64(wastedSpace)/(1024*1024))
			fmt.Printf("  Occurrences: %d\n", len(group))
			for _, file := range group {
				fmt.Printf("    - %s\n", file.RelativePath)
			}
			fmt.Println()
		}
		fmt.Printf("Total Space Wasted by Duplicates: %.2f MB\n\n", float64(totalWasted)/(1024*1024))
	}

	return nil
}
