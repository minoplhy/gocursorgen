package xcursorgen

import (
	"bufio"
	"fmt"
	"gocursorgen/internal/cursors"
	"os"
	"strings"
)

// This process `.cursor` file and prase into struct
func ReadCursorFile(config string) ([]cursors.CursorsEntity, int) {
	var f *os.File
	var err error

	if config == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(config)
		if err != nil {
			return nil, -1
		}
		defer f.Close()
	}

	var count int
	var out []cursors.CursorsEntity

	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Strip inline comments
		if i := strings.Index(line, "#"); i != -1 {
			line = line[:i]
		}

		// Normalize whitespace and skip blank lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)

		// Minimum required: size xhot yhot pngfile
		if len(fields) < 4 || len(fields) > 5 {
			fmt.Fprintf(os.Stderr, "xcursorgen: line %d: expected at least 4 or 5 fields, got %d\n", lineNum, len(fields))
			return nil, -1
		}

		size, ok1 := parseUint32(fields[0])
		xhot, ok2 := parseUint32(fields[1])
		yhot, ok3 := parseUint32(fields[2])

		if !ok1 || !ok2 || !ok3 {
			fmt.Fprintf(os.Stderr, "xcursorgen: line %d: size / xhot / yhot must be non-negative integers\n", lineNum)
			return nil, -1
		}

		out = append(out, cursors.CursorsEntity{
			Size:    size,
			XHot:    xhot,
			YHot:    yhot,
			PNGFile: fields[3],
		})

		// delay
		if len(fields) == 5 {
			delay, ok := parseUint32(fields[4])
			if !ok {
				fmt.Fprintf(os.Stderr, "xcursorgen: line %d: delay must be a non-negative integer\n", lineNum)
				return nil, -1
			}
			out[count].Delay = delay
		} else {
			delay, ok := parseUint32("50")
			if !ok {
				fmt.Fprintf(os.Stderr, "xcursorgen: line %d: delay must be a non-negative integer\n", lineNum)
				return nil, -1
			}
			out[count].Delay = delay
		}
		count++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "xcursorgen: error reading config: %v\n", err)
		return nil, -1
	}

	if count == 0 {
		return nil, 0
	}

	return out, count
}
