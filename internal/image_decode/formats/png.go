package formats

import (
	"image"
	"image/png"
	"os"

	imagedecode "gocursorgen/internal/image_decode"
)

func init() { imagedecode.Register(PNGDecoder{}) }

type PNGDecoder struct{}

func (PNGDecoder) Extensions() []string { return []string{".png"} }
func (PNGDecoder) Magic() []byte        { return []byte{0x89, 0x50, 0x4E, 0x47} }

func (PNGDecoder) Decode(f *os.File) ([]image.Image, error) {
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return []image.Image{img}, nil
}

func (PNGDecoder) DecodeConfig(f *os.File) (image.Config, error) {
	return png.DecodeConfig(f)
}
