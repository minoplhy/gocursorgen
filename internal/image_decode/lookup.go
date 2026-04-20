package imagedecode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type registry struct {
	byExt   map[string]FormatDecoder
	byMagic []FormatDecoder // ordered, checked in registration order
}

var reg = &registry{
	byExt:   map[string]FormatDecoder{},
	byMagic: []FormatDecoder{},
}

// Register a FormatDecoder
func Register(d FormatDecoder) {
	for _, ext := range d.Extensions() {
		ext = strings.ToLower(ext)
		if _, exists := reg.byExt[ext]; exists {
			panic(fmt.Sprintf("imaging: decoder already registered for extension %q", ext))
		}
		reg.byExt[ext] = d
	}
	if len(d.Magic()) > 0 {
		reg.byMagic = append(reg.byMagic, d)
	}
}

// Lookup resolves the correct FormatDecoder for a given file using:
//  1. Magic bytes - reads the file header, matches against registered signatures
//  2. Extension   - falls back to file extension if no magic matched
func Lookup(f *os.File) (FormatDecoder, error) {
	// Read enough bytes to cover the longest magic signature
	header := make([]byte, maxMagicLen())
	n, _ := f.Read(header)
	header = header[:n]

	// Seek back so the decoder sees the full file
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("imaging: seek %q: %w", f.Name(), err)
	}

	for _, d := range reg.byMagic {
		magic := d.Magic()
		if len(magic) <= len(header) && matchMagic(header, magic) {
			return d, nil
		}
	}

	ext := strings.ToLower(filepath.Ext(f.Name()))
	if d, ok := reg.byExt[ext]; ok {
		return d, nil
	}

	return nil, fmt.Errorf("imaging: no decoder found for %q (tried magic and extension %q)", f.Name(), ext)
}

func matchMagic(header, magic []byte) bool {
	for i, b := range magic {
		if header[i] != b {
			return false
		}
	}
	return true
}

func maxMagicLen() int {
	max := 0
	for _, d := range reg.byMagic {
		if l := len(d.Magic()); l > max {
			max = l
		}
	}
	if max == 0 {
		return 16 // sensible default when no magic registered yet
	}
	return max
}
