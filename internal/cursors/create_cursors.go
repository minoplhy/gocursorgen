package cursors

import (
	"fmt"
	libxcursor "gocursorgen/internal/libXCursor"
)

// Write Cursor into a libxcursor.XCursorImages
func (e CursorEntities) CreateCursors(count int, filename string, prefix *string) (libxcursor.XcursorImages, error) {
	cimages := libxcursor.XcursorImagesCreate(count)

	cur := e
	for i, v := range cur {
		if cur == nil {
			return libxcursor.XcursorImages{}, fmt.Errorf("xcursorgen: file list shorter than count\n")
		}

		image, err := e[i].loadImage(prefix)
		if err != nil {
			return libxcursor.XcursorImages{}, fmt.Errorf("xcursorgen: error while reading %s: %v\n", v.PNGFile, err)
		}

		cimages.Images[i] = *image
	}

	//if !libxcursor.XcursorFileSaveImages(fp, cimages) {
	//	return 1
	//}
	return cimages, nil
}
