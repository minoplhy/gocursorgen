package theme

import (
	"fmt"
	"gocursorgen/internal/hotspot"
)

type YamlHotSpot struct {
	X *uint32 `yaml:"x"`
	Y *uint32 `yaml:"y"`
}

func (h YamlHotSpot) complete() bool {
	return h.X != nil && h.Y != nil
}

func (h YamlHotSpot) toHotSpot() *hotspot.HotSpot {
	if !h.complete() {
		return nil
	}
	return &hotspot.HotSpot{X: *h.X, Y: *h.Y}
}

// Support mechanism:
//
//	files:
//	  - "simple.png"          # string form
//	  - path: "custom.png"    # object form
//	    x: 12
//	    y: 15
type yamlFileEntry struct {
	Path string
	YamlHotSpot
}

func (f *yamlFileEntry) UnmarshalYAML(unmarshal func(any) error) error {
	var path string
	if err := unmarshal(&path); err == nil {
		f.Path = path
		return nil
	}

	// Fall back to object form
	var obj struct {
		Path string  `yaml:"path"`
		X    *uint32 `yaml:"x"`
		Y    *uint32 `yaml:"y"`
	}
	if err := unmarshal(&obj); err != nil {
		return fmt.Errorf("file entry must be a string or {path, x, y} object: %w", err)
	}
	if obj.Path == "" {
		return fmt.Errorf("file entry object is missing 'path' field")
	}
	f.Path = obj.Path
	f.X = obj.X
	f.Y = obj.Y
	return nil
}

type yamlCursorEntry struct {
	Name        string          `yaml:"name"`
	Files       []yamlFileEntry `yaml:"files"`
	Folder      string          `yaml:"folder"`
	Sizes       []uint32        `yaml:"size"`
	YamlHotSpot `yaml:",inline"`
}

type yamlFile struct {
	Cursor []yamlCursorEntry `yaml:"cursor"`
	Theme  map[string]string `yaml:"theme"`
	Global yamlOptions       `yaml:"global"`
}

type yamlOptions struct {
	Sizes []uint32 `yaml:"size"`
}
