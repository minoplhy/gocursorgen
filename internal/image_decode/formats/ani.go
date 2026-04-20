package formats

import (
	"bytes"
	"encoding/binary"
	"fmt"
	imagedecode "gocursorgen/internal/image_decode"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

func init() { imagedecode.Register(ANIDecoder{}) }

type ANIDecoder struct{}

func (ANIDecoder) Extensions() []string { return []string{".ani"} }
func (ANIDecoder) Magic() []byte        { return []byte{0x52, 0x49, 0x46, 0x46} } // RIFF

func (ANIDecoder) Decode(f *os.File) ([]image.Image, error) {
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("ANI: read: %w", err)
	}
	return parseANI(data)
}

func (ANIDecoder) DecodeConfig(f *os.File) (image.Config, error) {
	frames, err := ANIDecoder{}.Decode(f)
	if err != nil {
		return image.Config{}, err
	}
	b := frames[0].Bounds()
	return image.Config{Width: b.Dx(), Height: b.Dy()}, nil
}

// riffChunk represents a parsed RIFF chunk.
type riffChunk struct {
	id   string
	data []byte
}

// parseRIFF walks raw RIFF bytes and returns all direct child chunks
func parseRIFF(data []byte) ([]riffChunk, error) {
	var chunks []riffChunk
	off := 0

	for off < len(data) {
		// Skip zero padding bytes
		if data[off] == 0 {
			off++
			continue
		}

		// Need at least 8 bytes for a chunk header
		if off+8 > len(data) {
			break
		}

		id := string(data[off : off+4])
		size := int(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		off += 8

		if off+size > len(data) {
			return nil, fmt.Errorf("RIFF: chunk %q size %d overflows data", id, size)
		}

		chunks = append(chunks, riffChunk{
			id:   id,
			data: data[off : off+size],
		})

		off += size
		// RIFF word-align: chunks must start on even boundaries
		if size%2 != 0 {
			off++
		}
	}

	return chunks, nil
}

// parseANI extracts all frames from an ANI RIFF file.
func parseANI(data []byte) ([]image.Image, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("ANI: file too short")
	}
	if string(data[0:4]) != "RIFF" {
		return nil, fmt.Errorf("ANI: missing RIFF header")
	}
	if string(data[8:12]) != "ACON" {
		return nil, fmt.Errorf("ANI: not an ACON RIFF file")
	}

	// Parse top-level chunks, skip the 12-byte RIFF/ACON header
	topChunks, err := parseRIFF(data[12:])
	if err != nil {
		return nil, err
	}

	for _, chunk := range topChunks {
		// We only care about LIST frame
		if chunk.id != "LIST" || len(chunk.data) < 4 {
			continue
		}
		if string(chunk.data[0:4]) != "fram" {
			continue
		}

		// Parse inner icon chunks from the fram LIST
		innerChunks, err := parseRIFF(chunk.data[4:])
		if err != nil {
			return nil, fmt.Errorf("ANI: LIST fram: %w", err)
		}

		var frames []image.Image
		for _, inner := range innerChunks {
			if inner.id != "icon" {
				continue
			}
			img, err := decodeICOFrame(inner.data)
			if err != nil {
				return nil, fmt.Errorf("ANI: icon frame: %w", err)
			}
			frames = append(frames, img)
		}

		if len(frames) == 0 {
			return nil, fmt.Errorf("ANI: LIST fram contained no icon chunks")
		}
		return frames, nil
	}

	return nil, fmt.Errorf("ANI: no LIST fram chunk found")
}

// decodeICOFrame decodes a single ICO/CUR frame from an ANI icon chunk.
//
// ICO format:
//
//	ICONDIR   (6 bytes)  — reserved, type, image count
//	ICONDIRENTRY (16 bytes each) — width, height, dataSize, dataOffset
//	image data at dataOffset
func decodeICOFrame(data []byte) (image.Image, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("ICO: data too short for ICONDIR")
	}

	// ICONDIR
	// [0:2] reserved = 0
	// [2:4] type: 1=ICO, 2=CUR
	// [4:6] count
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	if count == 0 {
		return nil, fmt.Errorf("ICO: zero images in directory")
	}

	// ICONDIRENTRY at offset 6 (16 bytes each)
	// [0]   width  (0 = 256)
	// [1]   height (0 = 256)
	// [2]   color count
	// [3]   reserved
	// [4:6] planes / xHotspot
	// [6:8] bitCount / yHotspot
	// [8:12]  data size
	// [12:16] data offset from start of file
	const dirEntrySize = 16
	entryOff := 6
	if len(data) < entryOff+dirEntrySize {
		return nil, fmt.Errorf("ICO: data too short for ICONDIRENTRY")
	}

	dataSize := int(binary.LittleEndian.Uint32(data[entryOff+8 : entryOff+12]))
	dataOff := int(binary.LittleEndian.Uint32(data[entryOff+12 : entryOff+16]))

	if dataOff+dataSize > len(data) {
		return nil, fmt.Errorf("ICO: image data out of bounds (offset=%d size=%d total=%d)",
			dataOff, dataSize, len(data))
	}

	imgData := data[dataOff : dataOff+dataSize]

	// Detect PNG Header
	if len(imgData) >= 4 &&
		imgData[0] == 0x89 && imgData[1] == 0x50 &&
		imgData[2] == 0x4E && imgData[3] == 0x47 {
		return png.Decode(bytes.NewReader(imgData))
	}

	// Legacy DIB (BITMAPINFOHEADER) format
	return decodeDIB(imgData)
}

// decodeDIB decodes a raw 32bpp DIB (Device Independent Bitmap) image.
//
// DIB layout:
//
//	BITMAPINFOHEADER (40 bytes)
//	  [0:4]  header size = 40
//	  [4:8]  width
//	  [8:12] height (doubled in ICO — XOR + AND masks stacked)
//	  [12:14] planes
//	  [14:16] bitCount
//	  ...
//	pixel data (bottom-up, BGRA order)
func decodeDIB(data []byte) (image.Image, error) {
	const bihSize = 40
	if len(data) < bihSize {
		return nil, fmt.Errorf("DIB: data too short for BITMAPINFOHEADER")
	}

	width := int(binary.LittleEndian.Uint32(data[4:8]))
	height := int(binary.LittleEndian.Uint32(data[8:12])) / 2
	bpp := int(binary.LittleEndian.Uint16(data[14:16]))

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("DIB: invalid dimensions %dx%d", width, height)
	}

	switch bpp {
	case 32:
		return decodeDIB32(data, width, height)
	case 8:
		return decodeDIB8(data, width, height)
	case 4:
		return decodeDIB4(data, width, height)
	case 1:
		return decodeDIB1(data, width, height)
	default:
		return nil, fmt.Errorf("DIB: unsupported bit depth %d", bpp)
	}
}

func decodeDIB32(data []byte, width, height int) (image.Image, error) {
	const bihSize = 40
	if len(data) < bihSize+width*height*4 {
		return nil, fmt.Errorf("DIB: 32bpp data too short")
	}

	// Some 32bpp ICO/CUR files store A=0 in BGRA and rely on AND mask instead.
	// Detect this by checking whether all alpha bytes are zero.
	allZeroAlpha := true
	for i := 0; i < width*height && allZeroAlpha; i++ {
		if data[bihSize+i*4+3] != 0 {
			allZeroAlpha = false
		}
	}

	andStart := bihSize + width*height*4
	isTransparent := andMask(data, andStart, width, height)

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for row := 0; row < height; row++ {
		srcRow := height - 1 - row
		for col := 0; col < width; col++ {
			base := bihSize + (srcRow*width+col)*4
			r := data[base+2]
			g := data[base+1]
			b := data[base+0]
			a := data[base+3]

			if allZeroAlpha {
				if isTransparent(col, row) {
					a = 0
				} else {
					a = 255
				}
			}

			img.SetNRGBA(col, row, color.NRGBA{R: r, G: g, B: b, A: a})
		}
	}
	return img, nil
}

func decodeDIB8(data []byte, width, height int) (image.Image, error) {
	const (
		bihSize       = 40
		paletteColors = 256
		paletteSize   = paletteColors * 4
	)
	if len(data) < bihSize+paletteSize {
		return nil, fmt.Errorf("DIB: 8bpp data too short for palette")
	}

	palette := make([]color.NRGBA, paletteColors)
	for i := 0; i < paletteColors; i++ {
		base := bihSize + i*4
		palette[i] = color.NRGBA{
			B: data[base+0],
			G: data[base+1],
			R: data[base+2],
			A: data[base+3],
		}
	}

	rowStride := (width + 3) &^ 3
	pixelStart := bihSize + paletteSize
	andStart := pixelStart + rowStride*height
	isTransparent := andMask(data, andStart, width, height)

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for row := 0; row < height; row++ {
		srcRow := height - 1 - row
		for col := 0; col < width; col++ {
			idx := data[pixelStart+srcRow*rowStride+col]
			c := palette[idx]

			if isTransparent(col, row) {
				c.A = 0
			} else if c.A == 0 {
				// AND mask says opaque but palette alpha is 0
				c.A = 255
			}

			img.SetNRGBA(col, row, c)
		}
	}
	return img, nil
}

func decodeDIB4(data []byte, width, height int) (image.Image, error) {
	const (
		bihSize       = 40
		paletteColors = 16
		paletteSize   = paletteColors * 4
	)
	if len(data) < bihSize+paletteSize {
		return nil, fmt.Errorf("DIB: 4bpp data too short for palette")
	}

	palette := make([]color.NRGBA, paletteColors)
	for i := 0; i < paletteColors; i++ {
		base := bihSize + i*4
		palette[i] = color.NRGBA{
			B: data[base+0],
			G: data[base+1],
			R: data[base+2],
			A: data[base+3],
		}
	}

	rowStride := ((width+1)/2 + 3) &^ 3
	pixelStart := bihSize + paletteSize
	andStart := pixelStart + rowStride*height
	isTransparent := andMask(data, andStart, width, height)

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for row := 0; row < height; row++ {
		srcRow := height - 1 - row
		for col := 0; col < width; col++ {
			b := data[pixelStart+srcRow*rowStride+col/2]
			var idx byte
			if col%2 == 0 {
				idx = (b >> 4) & 0x0F
			} else {
				idx = b & 0x0F
			}
			c := palette[idx]

			if isTransparent(col, row) {
				c.A = 0
			} else if c.A == 0 {
				c.A = 255
			}

			img.SetNRGBA(col, row, c)
		}
	}
	return img, nil
}

func decodeDIB1(data []byte, width, height int) (image.Image, error) {
	const (
		bihSize       = 40
		paletteColors = 2
		paletteSize   = paletteColors * 4
	)
	if len(data) < bihSize+paletteSize {
		return nil, fmt.Errorf("DIB: 1bpp data too short for palette")
	}

	palette := make([]color.NRGBA, paletteColors)
	for i := 0; i < paletteColors; i++ {
		base := bihSize + i*4
		palette[i] = color.NRGBA{
			B: data[base+0],
			G: data[base+1],
			R: data[base+2],
			A: data[base+3],
		}
	}

	rowStride := ((width + 31) / 32) * 4
	pixelStart := bihSize + paletteSize
	andStart := pixelStart + rowStride*height
	isTransparent := andMask(data, andStart, width, height)

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for row := 0; row < height; row++ {
		srcRow := height - 1 - row
		for col := 0; col < width; col++ {
			b := data[pixelStart+srcRow*rowStride+col/8]
			bit := (b >> uint(7-col%8)) & 1
			c := palette[bit]

			if isTransparent(col, row) {
				c.A = 0
			} else if c.A == 0 {
				c.A = 255
			}

			img.SetNRGBA(col, row, c)
		}
	}
	return img, nil
}

// andMask reads the 1bpp AND mask that follows XOR pixel data in ICO/CUR DIB.
// Returns a function that reports whether pixel (col, row) is transparent.
// AND mask: bit=1 -> transparent, bit=0 -> opaque.
func andMask(data []byte, start, width, height int) func(col, row int) bool {
	andRowStride := ((width + 31) / 32) * 4
	return func(col, row int) bool {
		// DIB rows are bottom-up
		srcRow := height - 1 - row
		off := start + srcRow*andRowStride + col/8
		if off >= len(data) {
			return false
		}
		return (data[off]>>uint(7-col%8))&1 == 1
	}
}
