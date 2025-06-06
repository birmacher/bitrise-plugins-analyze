package core

type LiefSymbol struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type LiefSection struct {
	Name    string       `json:"name"`
	Size    int64        `json:"size"`
	Symbols []LiefSymbol `json:"symbols"`
}

type LiefBinary struct {
	Path     string        `json:"path"`
	Sections []LiefSection `json:"sections"`
}

func ParseLiefOutput(filePath string, liefOutput map[string]map[string]int64, bundle *AppBundle) error {
	binary := LiefBinary{
		Path:     filePath,
		Sections: []LiefSection{},
	}

	for section, symbols := range liefOutput {
		section := LiefSection{
			Name:    section,
			Size:    0,
			Symbols: []LiefSymbol{},
		}

		for symbol, size := range symbols {
			section.Size += size
			section.Symbols = append(section.Symbols, LiefSymbol{
				Name: symbol,
				Size: size,
			})
		}

		binary.Sections = append(binary.Sections, section)
	}

	bundle.BinaryFiles = append(bundle.BinaryFiles, binary)
	return nil
}
