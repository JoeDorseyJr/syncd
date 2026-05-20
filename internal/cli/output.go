package cli

import (
	"os"

	"golang.org/x/term"
)

var (
	Green  = "\033[32m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Reset  = "\033[0m"
)

// NoColor is set by the --no-color flag.
var NoColor bool

// InitColor disables color if stdout is not a terminal or --no-color is set.
func InitColor() {
	if NoColor || !term.IsTerminal(int(os.Stdout.Fd())) {
		DisableColor()
	}
}

// DisableColor removes all ANSI codes.
func DisableColor() {
	Green, Red, Yellow, Reset = "", "", "", ""
}
