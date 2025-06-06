package android

import (
	"bitrise-plugins-analyze/appbundle/core"
	"fmt"
)

func ParseLiefOutput(liefOutput map[string]map[string]int64, bundle *core.AppBundle) error {
	sections := []core.LiefSection{}

	for section, symbols := range liefOutput {
		section := core.LiefSection{
			Name:    section,
			Size:    0,
			Symbols: []core.LiefSymbol{},
		}
		fmt.Println("Processing section:", section)
		for symbol, size := range symbols {
			section.Size += size
			section.Symbols = append(section.Symbols, core.LiefSymbol{
				Name: symbol,
				Size: size,
			})
		}

		sections = append(sections, section)
	}

	bundle.BinaryFiles = sections
	return nil
}
