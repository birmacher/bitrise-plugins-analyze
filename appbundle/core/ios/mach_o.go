package ios

// MachOInfo represents information about a Mach-O binary
type MachOInfo struct {
	Path         string   `json:"path"`
	Architecture []string `json:"architecture"`
	LoadCommands []string `json:"load_commands,omitempty"`
	MinOSVersion string   `json:"min_os_version,omitempty"`
	LinkedLibs   []string `json:"linked_libraries,omitempty"`
	RPaths       []string `json:"rpaths,omitempty"`
	Size         int64    `json:"size"`
}
