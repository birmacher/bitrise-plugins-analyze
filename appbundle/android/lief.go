package android

import (
	"bitrise-plugins-analyze/appbundle/core"
)

func ParseLiefOutput(filePath string, liefOutput map[string]map[string]int64, bundle *core.AppBundle) error {
	binary := core.LiefBinary{
		Path:     filePath,
		Sections: []core.LiefSection{},
	}

	for section, symbols := range liefOutput {
		section := core.LiefSection{
			Name:    section,
			Size:    0,
			Symbols: []core.LiefSymbol{},
		}

		for symbol, size := range symbols {
			section.Size += size
			section.Symbols = append(section.Symbols, core.LiefSymbol{
				Name: symbol,
				Size: size,
			})
		}

		binary.Sections = append(binary.Sections, section)
	}

	bundle.BinaryFiles = append(bundle.BinaryFiles, binary)
	return nil
}
