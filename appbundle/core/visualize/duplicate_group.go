package visualize

import "bitrise-plugins-analyze/appbundle/core"

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Files         []core.FileInfo `json:"files"`
	Size          int64           `json:"size"`
	WastedSpace   int64           `json:"wasted_space"`
	TotalWasted   int64           `json:"total_wasted"`
	WastedPercent float64         `json:"wasted_percent"`
}
