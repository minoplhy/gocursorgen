package xcursorgen

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// unused
// readXBMFile parses a standard X11 XBM file and returns its dimensions and
// raw bit-packed pixel data (LSBFirst, 8-bit aligned rows).
func readXBMFile(filename string) (width, height int, data []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, 0, nil, err
	}
	defer f.Close()

	width, height = -1, -1
	scanner := bufio.NewScanner(f)

	// parse #define width / height
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#define") {
			fields := strings.Fields(line)
			if len(fields) != 3 {
				continue
			}
			val, err := strconv.Atoi(fields[2])
			if err != nil {
				continue
			}
			switch {
			case strings.HasSuffix(fields[1], "_width"):
				width = val
			case strings.HasSuffix(fields[1], "_height"):
				height = val
			}
			continue
		}

		// parse the bits[] array
		if strings.Contains(line, "[]") {
			if width < 0 || height < 0 {
				return 0, 0, nil, fmt.Errorf("xbm: dimensions missing before data array")
			}

			// Accumulate all hex tokens from the remainder of the file
			var tokens []string
			for scanner.Scan() {
				l := strings.NewReplacer("{", "", "}", "", ";", "").Replace(scanner.Text())
				for _, part := range strings.Split(l, ",") {
					if t := strings.TrimSpace(part); t != "" {
						tokens = append(tokens, t)
					}
				}
			}

			bytesPerLine := (width + 7) / 8
			data = make([]byte, 0, height*bytesPerLine)
			for _, tok := range tokens {
				tok = strings.TrimPrefix(strings.TrimPrefix(tok, "0x"), "0X")
				v, err := strconv.ParseUint(tok, 16, 8)
				if err != nil {
					continue
				}
				data = append(data, byte(v))
			}
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, nil, err
	}
	if width < 0 || height < 0 || len(data) == 0 {
		return 0, 0, nil, fmt.Errorf("xbm: invalid or empty file")
	}
	return width, height, data, nil
}
