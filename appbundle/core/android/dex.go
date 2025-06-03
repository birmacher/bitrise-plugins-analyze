package android

import "path/filepath"

type DexClass struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Shasum string `json:"shasum"`
}

type DexPackage struct {
	Name    string     `json:"name"`
	Size    int64      `json:"size"`
	Classes []DexClass `json:"classes"`
}

func (dexPackage DexPackage) GetPath() string {
	return dexPackage.Name

}

func (dexClass DexClass) GetPath(dexPackage DexPackage) string {
	return filepath.Join(dexPackage.GetPath(), dexClass.Name)

}
