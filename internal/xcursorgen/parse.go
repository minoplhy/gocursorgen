package xcursorgen

import (
	"strconv"
)

// parseUint32 parses a decimal string into a uint32.
// Returns (0, false) if the string is not a valid non-negative integer
// or exceeds the uint32 range.
func parseUint32(s string) (uint32, bool) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(v), true
}
