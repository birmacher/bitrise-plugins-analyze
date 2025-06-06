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
