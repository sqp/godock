// Package color formats colored terminal output.
package color

// Terminal colors.
const (
	Reset      = "\x1b[0m"  // Reset terminal color.
	Bright     = "\x1b[1m"  // Set bright display.
	Dim        = "\x1b[2m"  // ??.
	Underscore = "\x1b[4m"  // Underline format.
	Blink      = "\x1b[5m"  // Blinking text.
	Reverse    = "\x1b[7m"  // Reverse fg and bg colors.
	Hidden     = "\x1b[8m"  // Hide text.
	FgBlack    = "\x1b[30m" // Foreground black.
	FgRed      = "\x1b[31m" // Foreground Red.
	FgGreen    = "\x1b[32m" // Foreground Green.
	FgYellow   = "\x1b[33m" // Foreground Yellow.
	FgBlue     = "\x1b[34m" // Foreground Blue.
	FgMagenta  = "\x1b[35m" // Foreground Magenta.
	FgCyan     = "\x1b[36m" // Foreground Cyan.
	FgWhite    = "\x1b[37m" // Foreground White.
	BgBlack    = "\x1b[40m" // Background Black.
	BgRed      = "\x1b[41m" // Background Red.
	BgGreen    = "\x1b[42m" // Background Green.
	BgYellow   = "\x1b[43m" // Background Yellow.
	BgBlue     = "\x1b[44m" // Background Blue.
	BgMagenta  = "\x1b[45m" // Background Magenta.
	BgCyan     = "\x1b[46m" // Background Cyan.
	BgWhite    = "\x1b[47m" // Background White.
)

// Yellow formatting of text.
func Yellow(msg string) string { return Colored(msg, FgYellow) }

// Magenta formatting of text.
func Magenta(msg string) string { return Colored(msg, FgMagenta) }

// Green formatting of text.
func Green(msg string) string { return Colored(msg, FgGreen) }

// Red formatting of text.
func Red(msg string) string { return Colored(msg, FgRed) }

// Colored returns a colored message if any.
func Colored(msg, col string) string {
	if msg == "" {
		return ""
	}
	return col + msg + Reset
}
