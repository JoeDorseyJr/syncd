package progress

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Display handles progress rendering.
type Display struct {
	IsTTY bool
	last  int
}

// NewDisplay creates a display, detecting TTY status.
func NewDisplay() *Display {
	return &Display{IsTTY: term.IsTerminal(int(os.Stdout.Fd()))}
}

// Status prints a progress status line. In TTY mode, overwrites previous.
func (d *Display) Status(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if d.IsTTY {
		pad := ""
		if len(line) < d.last {
			pad = strings.Repeat(" ", d.last-len(line))
		}
		fmt.Fprintf(os.Stdout, "\r%s%s", line, pad)
		d.last = len(line)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", line)
	}
}

// Finish prints a final line that won't be overwritten (newline terminated).
func (d *Display) Finish(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if d.IsTTY {
		pad := ""
		if len(line) < d.last {
			pad = strings.Repeat(" ", d.last-len(line))
		}
		fmt.Fprintf(os.Stdout, "\r%s%s\n", line, pad)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", line)
	}
	d.last = 0
}
