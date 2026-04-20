package theme

import "gocursorgen/internal/cursors"

// ThemeFile is the parsed top-level YAML document
type ThemeFile struct {
	Cursors []cursors.CursorEntry
	Theme   map[string]string // XC_name -> CursorEntry.Name
}
