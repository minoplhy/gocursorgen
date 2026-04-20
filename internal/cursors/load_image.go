package cursors

import (
	"fmt"
	imagedecode "gocursorgen/internal/image_decode"
	libxcursor "gocursorgen/internal/libXCursor"
	"path/filepath"
)

// load image convert `CursorsEntity` entry to XCursorImage with Pixel arranged
func (entity CursorsEntity) loadImage(prefix *string) (*libxcursor.XcursorImage, error) {
	var path string
	if prefix != nil && *prefix != "" {
		path = filepath.Join(*prefix, entity.PNGFile)
	} else {
		path = entity.PNGFile
	}

	frames, err := imagedecode.Decode(path)
	if err != nil {
		return nil, err
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("load_image: %q decoded to zero frames", path)
	}

	// For multi-frame files we use the first frame
	img := frames[0]
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	out := &libxcursor.XcursorImage{
		Width:  uint32(width),
		Height: uint32(height),
		Size:   entity.Size,
		XHot:   entity.XHot,
		YHot:   entity.YHot,
		Delay:  entity.Delay,
		Pixels: make([]uint32, width*height),
	}

	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			out.Pixels[idx] = (a>>8)<<24 | (r>>8)<<16 | (g>>8)<<8 | (b >> 8)
			idx++
		}
	}

	return out, nil
}
