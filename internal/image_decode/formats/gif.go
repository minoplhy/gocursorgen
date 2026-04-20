package formats

import (
	"fmt"
	"image"
	"image/gif"
	"os"

	imagedecode "gocursorgen/internal/image_decode"
)

func init() { imagedecode.Register(GIFDecoder{}) }

type GIFDecoder struct{}

func (GIFDecoder) Extensions() []string { return []string{".gif"} }
func (GIFDecoder) Magic() []byte        { return []byte{0x47, 0x49, 0x46, 0x38} } // GIF8

func (GIFDecoder) Decode(f *os.File) ([]image.Image, error) {
	all, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}
	if len(all.Image) == 0 {
		return nil, fmt.Errorf("GIF has no frames")
	}
	frames := make([]image.Image, len(all.Image))
	for i, frame := range all.Image {
		frames[i] = frame
	}
	return frames, nil
}

func (GIFDecoder) DecodeConfig(f *os.File) (image.Config, error) {
	return gif.DecodeConfig(f)
}
