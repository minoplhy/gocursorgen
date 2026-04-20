package theme

import (
	"fmt"
	"gocursorgen/internal/cursors"
)

// CursorByName returns the CursorEntry with the given name, or an error.
func (tf *ThemeFile) CursorByName(name string) (*cursors.CursorEntry, error) {
	for i := range tf.Cursors {
		if tf.Cursors[i].Name == name {
			return &tf.Cursors[i], nil
		}
	}
	return nil, fmt.Errorf("theme: no cursor named %q", name)
}

// ResolveCursor looks up a theme symbol and returns its resolved file paths.
func (tf *ThemeFile) ResolveCursor(xcSymbol string, prefix string) ([]string, error) {
	name, ok := tf.Theme[xcSymbol]
	if !ok {
		return nil, fmt.Errorf("theme: unknown symbol %q", xcSymbol)
	}
	entry, err := tf.CursorByName(name)
	if err != nil {
		return nil, err
	}
	return entry.ResolveFiles(prefix)
}
