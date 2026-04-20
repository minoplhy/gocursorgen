package imagedecode

import (
	"image"
	"os"
)

// FormatDecoder handles a specific image format.
type FormatDecoder interface {
	Extensions() []string
	Magic() []byte
	// Decode returns all frames. Single-frame formats return a slice of one.
	Decode(f *os.File) ([]image.Image, error)
	// DecodeConfig returns image dimensions from the first frame.
	DecodeConfig(f *os.File) (image.Config, error)
}
