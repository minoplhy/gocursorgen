package theme

import (
	"fmt"
	"gocursorgen/internal/cursors"
	"os"

	"github.com/goccy/go-yaml"
)

// Parse theme.yaml
func ParseFile(path string) (*ThemeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("theme: cannot read %q: %w", path, err)
	}
	return Parse(data)
}

func Parse(data []byte) (*ThemeFile, error) {
	var raw yamlFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("theme: invalid YAML: %w", err)
	}
	if err := validate(&raw); err != nil {
		return nil, err
	}
	return convert(&raw), nil
}

func validate(raw *yamlFile) error {
	seen := make(map[string]struct{}, len(raw.Cursor))

	for i, c := range raw.Cursor {
		if c.Name == "" {
			return fmt.Errorf("theme: cursor[%d]: name is required", i)
		}
		if _, dup := seen[c.Name]; dup {
			return fmt.Errorf("theme: cursor[%d]: duplicate name %q", i, c.Name)
		}
		seen[c.Name] = struct{}{}

		hasFiles := len(c.Files) > 0
		hasFolder := c.Folder != ""

		if !hasFiles && !hasFolder {
			return fmt.Errorf("theme: cursor %q: must have 'files' or 'folder'", c.Name)
		}
		if hasFiles && hasFolder {
			return fmt.Errorf("theme: cursor %q: 'files' and 'folder' are mutually exclusive", c.Name)
		}

		// Folder-level x/y must be either both set or both absent
		if c.YamlHotSpot.X != nil != (c.YamlHotSpot.Y != nil) {
			return fmt.Errorf("theme: cursor %q: 'x' and 'y' must both be specified or both omitted", c.Name)
		}

		for j, f := range c.Files {
			if f.Path == "" {
				return fmt.Errorf("theme: cursor %q: files[%d]: path is empty", c.Name, j)
			}
			// Per-file x/y must be either both set or both absent
			if (f.X != nil) != (f.Y != nil) {
				return fmt.Errorf("theme: cursor %q: files[%d] %q: 'x' and 'y' must both be specified or both omitted",
					c.Name, j, f.Path)
			}
		}
	}

	for sym, target := range raw.Theme {
		if sym == "" {
			return fmt.Errorf("theme: theme map contains an empty key")
		}
		if target == "" {
			return fmt.Errorf("theme: theme[%q]: target cursor name is empty", sym)
		}
		if _, ok := seen[target]; !ok {
			return fmt.Errorf("theme: theme[%q]: references unknown cursor %q", sym, target)
		}
	}

	return nil
}

// Convert yaml to ThemeFile struct
func convert(raw *yamlFile) *ThemeFile {
	CursorsEntry := make([]cursors.CursorEntry, len(raw.Cursor))
	for i, c := range raw.Cursor {
		files := make([]cursors.FileEntry, len(c.Files))
		for j, f := range c.Files {
			files[j] = cursors.FileEntry{
				Path:    f.Path,
				HotSpot: f.YamlHotSpot.toHotSpot(),
			}
		}
		CursorsEntry[i] = cursors.CursorEntry{
			Name:    c.Name,
			Files:   files,
			Folder:  c.Folder,
			HotSpot: c.YamlHotSpot.toHotSpot(),
		}
	}
	return &ThemeFile{
		Cursors: CursorsEntry,
		Theme:   raw.Theme,
	}
}
