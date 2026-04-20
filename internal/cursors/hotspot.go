package cursors

import (
	"fmt"
	"gocursorgen/internal/hotspot"

	_ "image/gif"
	_ "image/jpeg"
)

// FindHotSpotForEntry returns the hotspot for the first (canonical) file of
// a CursorEntry, respecting the override priority:
//
//  1. Per-file x/y on the first file
//  2. Folder-level x/y on the entry
//  3. Auto-detection from the image
func (entry *CursorEntry) FindHotSpotForEntry(prefix string) (hotspot.HotSpot, error) {
	files, err := entry.ResolveFiles(prefix)
	if err != nil {
		return hotspot.HotSpot{}, err
	}
	if len(files) == 0 {
		return hotspot.HotSpot{}, fmt.Errorf("hotspot: cursor %q has no files", entry.Name)
	}

	// Folder-based entries have no FileEntry slice - only global or auto-detect
	if len(entry.Files) == 0 {
		if entry.HotSpot != nil {
			return *entry.HotSpot, nil
		}
		return hotspot.FindHotSpot(files[0])
	}

	return entry.Files[0].FindHotSpotForFile(entry.HotSpot, prefix)
}

// FindHotSpotForFile returns the hotspot for a single FileEntry,
// using the entry-level hotspot as a fallback if no per-file override exists.
func (fe FileEntry) FindHotSpotForFile(entryHotSpot *hotspot.HotSpot, prefix string) (hotspot.HotSpot, error) {
	if fe.HotSpot != nil {
		return *fe.HotSpot, nil
	}
	if entryHotSpot != nil {
		return *entryHotSpot, nil
	}
	return hotspot.FindHotSpotWithPrefix(prefix, fe.Path)
}

// resolveHotSpot applies the three-level priority:
//  1. Per-file override
//  2. Entry-level override
//  3. Auto-detect from image
func (entry *CursorEntry) resolveHotSpot(i int, path string) (hotspot.HotSpot, error) {
	// Per-file override - only available on files-based entries
	if i < len(entry.Files) && entry.Files[i].HotSpot != nil {
		return *entry.Files[i].HotSpot, nil
	}
	// Entry-level override
	if entry.HotSpot != nil {
		return *entry.HotSpot, nil
	}
	// Auto-detect
	return hotspot.FindHotSpot(path)
}
