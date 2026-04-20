package libxcursor

import "crypto/md5"

const XcursorBitmapHashSize = md5.Size // 16

// XcursorImageHash computes the MD5 hash of a 1-bit LSBFirst XBM bitmap.
// data is the raw bit-packed row data
//
// bytesPerLine = (width+7)/8.
func XcursorImageHash(width, height int, data []byte) [XcursorBitmapHashSize]byte {
	bytesPerLine := (width + 7) / 8
	h := md5.New()
	pixel := make([]byte, 1)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// LSBFirst: bit 0 of each byte is the leftmost pixel
			pixel[0] = (data[y*bytesPerLine+x/8] >> uint(x%8)) & 1
			h.Write(pixel)
		}
	}

	var result [XcursorBitmapHashSize]byte
	copy(result[:], h.Sum(nil))
	return result
}
