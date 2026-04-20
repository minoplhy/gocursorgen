package libxcursor

import (
	"encoding/binary"
	"io"
)

// Xcursor binary file format constants
// https://gitlab.freedesktop.org/xorg/lib/libxcursor/-/blob/master/src/xcursorint.h
const (
	xcursorMagic          = 0x72756358 // "Xcur"
	xcursorFileVersion    = 0x00010000
	xcursorFileHeaderLen  = 16 // magic + header_len + version + ntoc
	xcursorFileTocLen     = 12 // type + subtype + position
	xcursorImageType      = 0xfffd0002
	xcursorImageVersion   = 1
	xcursorImageHeaderLen = 36 // header + type + subtype + version + w + h + xhot + yhot + delay
)

// XcursorFileSaveImages writes an XcursorImages set to w in the Xcursor binary format.
// Returns true on success, false on failure.
func XcursorFileSaveImages(w io.Writer, images XcursorImages) bool {
	n := images.NImage

	// Calculate the offset where image chunks begin:
	//   file header (16) + TOC entries (n * 12)
	chunkStart := uint32(xcursorFileHeaderLen + n*xcursorFileTocLen)

	// Pre-calculate the file position of each image chunk so we can write the TOC.
	offsets := make([]uint32, n)
	pos := chunkStart
	for i := 0; i < n; i++ {
		offsets[i] = pos
		img := images.Images[i]
		pixelBytes := uint32(img.Width * img.Height * 4)
		pos += xcursorImageHeaderLen + pixelBytes
	}

	le := binary.LittleEndian

	// File header (16 bytes)
	fileHeader := [4]uint32{
		xcursorMagic,
		xcursorFileHeaderLen,
		xcursorFileVersion,
		uint32(n),
	}
	for _, v := range fileHeader {
		buf := make([]byte, 4)
		le.PutUint32(buf, v)
		if _, err := w.Write(buf); err != nil {
			return false
		}
	}

	// TOC entries (12 bytes each)
	for i := 0; i < n; i++ {
		img := &images.Images[i]
		toc := [3]uint32{
			xcursorImageType,
			uint32(img.Size), // subtype = nominal size
			offsets[i],
		}
		for _, v := range toc {
			buf := make([]byte, 4)
			le.PutUint32(buf, v)
			if _, err := w.Write(buf); err != nil {
				return false
			}
		}
	}

	// Image chunks
	for i := 0; i < n; i++ {
		img := &images.Images[i]

		chunkHeader := [9]uint32{
			xcursorImageHeaderLen,
			xcursorImageType,
			uint32(img.Size), // subtype = nominal size
			xcursorImageVersion,
			uint32(img.Width),
			uint32(img.Height),
			uint32(img.XHot),
			uint32(img.YHot),
			uint32(img.Delay),
			/* Default of Delay should be 50, as per https://gitlab.freedesktop.org/xorg/app/xcursorgen/-/blob/master/xcursorgen.c#L95
			   But is responsibility of implementation package to handle */
		}

		for _, v := range chunkHeader {
			buf := make([]byte, 4)
			le.PutUint32(buf, v)
			if _, err := w.Write(buf); err != nil {
				return false
			}
		}

		// Pixel data - each pixel is a uint32 ARGB, little-endian
		buf := make([]byte, 4)
		for _, px := range img.Pixels {
			le.PutUint32(buf, px)
			if _, err := w.Write(buf); err != nil {
				return false
			}
		}
	}

	return true
}
