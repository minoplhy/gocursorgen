package xcursorgen

import (
	"fmt"
	libxcursor "gocursorgen/internal/libXCursor"
	"os"
)

// unused
// CheckImage mirrors the C check_image(): reads an XBM bitmap, computes its
// XcursorImageHash, and prints "<filename>: <hexhash>".
// Returns 0 on success, 1 on failure.
func CheckImage(image string) int {
	width, height, data, err := readXBMFile(image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "xcursorgen: Can't open bitmap file %q\n", image)
		return 1
	}

	hash := libxcursor.XcursorImageHash(width, height, data)

	fmt.Printf("%s: ", image)
	for _, b := range hash {
		fmt.Printf("%02x", b)
	}
	fmt.Println()
	return 0
}
