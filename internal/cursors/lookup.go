package cursors

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResolveFiles returns the concrete file paths for a CursorEntry.
// For folder-based entries it reads the directory and returns all
// .png and .ani files in sorted order.
// For file-based entries it joins prefix to each FileEntry.Path.
func (ce *CursorEntry) ResolveFiles(prefix string) ([]string, error) {
	if len(ce.Files) > 0 {
		out := make([]string, len(ce.Files))
		for i, fe := range ce.Files {
			if prefix != "" {
				out[i] = filepath.Join(prefix, fe.Path)
			} else {
				out[i] = fe.Path
			}
		}
		return out, nil
	}

	// Folder-based: list directory contents
	dir := ce.Folder
	if prefix != "" {
		dir = filepath.Join(prefix, dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("theme: cursor %q: cannot read folder %q: %w", ce.Name, dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lower := strings.ToLower(e.Name())
		if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".ani") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}

	sort.Strings(files)

	if len(files) == 0 {
		return nil, fmt.Errorf("theme: cursor %q: folder %q contains no .png or .ani files", ce.Name, dir)
	}

	return files, nil
}
