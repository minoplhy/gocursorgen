package formats

import (
	"image"
	"image/jpeg"
	"os"

	imagedecode "gocursorgen/internal/image_decode"
)

func init() { imagedecode.Register(JPEGDecoder{}) }

type JPEGDecoder struct{}

func (JPEGDecoder) Extensions() []string { return []string{".jpg", ".jpeg"} }
func (JPEGDecoder) Magic() []byte        { return []byte{0xFF, 0xD8, 0xFF} }

func (JPEGDecoder) Decode(f *os.File) ([]image.Image, error) {
	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, err
	}
	return []image.Image{img}, nil
}

func (JPEGDecoder) DecodeConfig(f *os.File) (image.Config, error) {
	return jpeg.DecodeConfig(f)
}
