// Package bytesize formats a size with units.
package bytesize

import "fmt"

//-----------------------------------------------------------[ BYTESIZE ]--

// ByteSize formats size into human-readable content.
//
type ByteSize float64

// Size constants.
//
const (
	_           = iota // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

// String formats the value to an usable range (1-999) with related unit.
// No decimal under MB, otherwise three digits are provided : 123 or 12.3 or 1.23.
//
func (b ByteSize) String() string {
	val, unit := b.unit()
	digits := "2"
	switch {
	case b < MB:
		digits = "0"
	case val >= 100:
		digits = "0"
	case val >= 10:
		digits = "1"
	}

	return fmt.Sprintf("%."+digits+"f %s", val, unit)
}

func (b ByteSize) unit() (ByteSize, string) {
	switch {
	case b >= YB:
		return b / YB, "YB"
	case b >= ZB:
		return b / ZB, "ZB"
	case b >= EB:
		return b / EB, "EB"
	case b >= PB:
		return b / PB, "PB"
	case b >= TB:
		return b / TB, "TB"
	case b >= GB:
		return b / GB, "GB"
	case b >= MB:
		return b / MB, "MB"
	case b >= KB:
		return b / KB, "KB"
	}
	return b, "B"
}
