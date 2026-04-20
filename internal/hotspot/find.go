package hotspot

import (
	"fmt"
	imagedecode "gocursorgen/internal/image_decode"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"path/filepath"
)

// findHotSpotInImage, An actual algorithm to find the hotspot
// It bound check against zero(transparent) from y --> x
// The first found win
func findHotSpotInImage(img image.Image) (HotSpot, error) {
	bounds := img.Bounds()
	if bounds.Empty() {
		return HotSpot{}, fmt.Errorf("hotspot: image has zero size")
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				return HotSpot{
					X: uint32(x - bounds.Min.X),
					Y: uint32(y - bounds.Min.Y),
				}, nil
			}
		}
	}

	return HotSpot{X: 0, Y: 0}, nil
}

// FindHotSpot opens an image file and returns the top-leftmost non-transparent
// pixel scanning top→bottom, left→right.
//
// Falls back to (0,0) for fully opaque or fully transparent images.
func FindHotSpot(filename string) (HotSpot, error) {
	img, err := imagedecode.DecodeFirst(filename)
	if err != nil {
		return HotSpot{}, fmt.Errorf("hotspot: %w", err)
	}
	return findHotSpotInImage(img)
}

// FindHotSpotWithPrefix joins prefix with path before -> FindHotSpot
func FindHotSpotWithPrefix(prefix, filename string) (HotSpot, error) {
	path := filename
	if prefix != "" {
		path = filepath.Join(prefix, filename)
	}
	return FindHotSpot(path)
}
