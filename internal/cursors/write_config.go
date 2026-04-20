package cursors

import (
	"fmt"
	"os"
)

// WriteConfig generates a xcursorgen-compatible .cursor config file
// from a theme.CursorEntry
//
// Output format per line:
//
//	<size> <xhot> <yhot> <file> [delay]
func (entry *CursorEntry) WriteConfig(filename string, prefix string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("xcursorgen: cannot create config %q: %w", filename, err)
	}
	defer f.Close()

	out, err := entry.GetConfig(prefix)
	if err != nil {
		return fmt.Errorf("xcursorgen: get config %q: %w", filename, err)
	}

	_, err = fmt.Fprintf(f, "%s", out)
	if err != nil {
		return fmt.Errorf("xcursorgen: write config %q: %w", filename, err)
	}

	return nil
}

// GetConfig returns the config as a string
func (entry *CursorEntry) GetConfig(prefix string) (string, error) {
	list, _, err := entry.entryToFileList(prefix)
	if err != nil {
		return "", err
	}

	var out string
	for _, v := range list {
		if v.Delay > 0 {
			out += fmt.Sprintf("%d %d %d %s %d\n",
				v.Size, v.XHot, v.YHot, v.PNGFile, v.Delay)
		} else {
			out += fmt.Sprintf("%d %d %d %s\n",
				v.Size, v.XHot, v.YHot, v.PNGFile)
		}
	}
	return out, nil
}
