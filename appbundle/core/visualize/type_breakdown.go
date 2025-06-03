package visualize

// TypeBreakdown represents size information for a specific file type
type TypeBreakdown struct {
	Type       string  `json:"type"`
	Size       int64   `json:"size"`
	Percentage float64 `json:"percentage"`
}
