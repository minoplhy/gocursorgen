package imagedecode

import (
	"fmt"
	"image"
	"os"
)

// Decode returns all frames from image input
func Decode(path string) ([]image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("imaging: open %q: %w", path, err)
	}
	defer f.Close()

	d, err := Lookup(f)
	if err != nil {
		return nil, err
	}
	frames, err := d.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("imaging: decode %q: %w", path, err)
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("imaging: %q decoded to zero frames", path)
	}
	return frames, nil
}

// DecodeFirst always returns only the first frame.
func DecodeFirst(path string) (image.Image, error) {
	frames, err := Decode(path)
	if err != nil {
		return nil, err
	}
	return frames[0], nil
}

// DecodeConfig returns image dimensions from the first frame
func DecodeConfig(path string) (image.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return image.Config{}, fmt.Errorf("imaging: open %q: %w", path, err)
	}
	defer f.Close()

	d, err := Lookup(f)
	if err != nil {
		return image.Config{}, err
	}
	cfg, err := d.DecodeConfig(f)
	if err != nil {
		return image.Config{}, fmt.Errorf("imaging: config %q: %w", path, err)
	}
	return cfg, nil
}
