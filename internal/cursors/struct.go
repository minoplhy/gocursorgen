package cursors

import (
	"gocursorgen/internal/hotspot"
)

// .cursor format should be like this:
//
//		[image_size] [xhot] [yhot] [png_name] [delay(animated)]
//
//	 image_size: 16 22 32 48 64 128 256
//	 Via: https://develop.kde.org/docs/features/additional-features/cursor/#creating-the-image-file
type CursorsEntity struct {
	Size    uint32
	XHot    uint32
	YHot    uint32
	Delay   uint32
	PNGFile string
	//Next    *File_list
}

type CursorEntities []CursorsEntity

// CursorEntry is one named cursor from the "cursor:" block.
type CursorEntry struct {
	Name    string
	Files   []FileEntry
	Folder  string
	HotSpot *hotspot.HotSpot
	Options Options
}

// FileEntry is a single file within a cursor entry.
// HotSpot is nil when no override is specified - detection runs instead.
type FileEntry struct {
	Path    string
	HotSpot *hotspot.HotSpot // nil = auto-detect; non-nil = skip detection entirely
}
