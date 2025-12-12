package utils

import (
	"fmt"
	"io"
)

// ANSI escape codes for text colors
const (
	ColorBlack   = "\033[30m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"

	ColorBrightBlack   = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	ColorReset = "\033[0m" // Resets all attributes to their default state
)

// ANSI escape codes for text styles
const (
	StyleBold            = "\033[1m"
	StyleFaint           = "\033[2m"
	StyleItalic          = "\033[3m" // Not widely supported
	StyleUnderline       = "\033[4m"
	StyleSlowBlink       = "\033[5m" // Not widely supported
	StyleRapidBlink      = "\033[6m" // Not widely supported
	StyleReverse         = "\033[7m" // Inverse foreground/background
	StyleConceal         = "\033[8m" // Not widely supported
	StyleCrossedOut      = "\033[9m" // Not widely supported
	StyleDefaultFont     = "\033[10m"
	StyleDoublyUnderline = "\033[21m" // Not widely supported, often just bold
	StyleNormal          = "\033[22m" // Neither bold nor faint
	StyleNoItalic        = "\033[23m"
	StyleNoUnderline     = "\033[24m"
	StyleNoBlink         = "\033[25m"
	StyleNoReverse       = "\033[27m"
	StyleReveal          = "\033[28m" // Conceal off
	StyleNoCrossedOut    = "\033[29m"
)

// Colorize wraps a string with ANSI color codes.
func Colorize(text, color string) string {
	return color + text + ColorReset
}

// FprintColor prints a string with ANSI color codes to the given writer.
func FprintColor(w io.Writer, text, color string) (int, error) {
	return fmt.Fprint(w, Colorize(text, color))
}

// FprintlnColor prints a string with ANSI color codes and a newline to the given writer.
func FprintlnColor(w io.Writer, text, color string) (int, error) {
	return fmt.Fprintln(w, Colorize(text, color))
}
