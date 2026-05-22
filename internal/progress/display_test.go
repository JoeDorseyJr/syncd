package progress

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func captureOutput(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestDisplay_StatusTTY(t *testing.T) {
	d := &Display{IsTTY: true}
	out := captureOutput(func() {
		d.Status("hello")
	})
	if !strings.HasPrefix(out, "\r") {
		t.Errorf("TTY status should start with \\r, got %q", out)
	}
	if strings.Contains(out, "\n") {
		t.Errorf("TTY status should not contain newline, got %q", out)
	}
}

func TestDisplay_StatusNonTTY(t *testing.T) {
	d := &Display{IsTTY: false}
	out := captureOutput(func() {
		d.Status("hello")
	})
	if strings.Contains(out, "\r") {
		t.Errorf("non-TTY status should not contain \\r, got %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("non-TTY status should end with newline, got %q", out)
	}
}

func TestDisplay_FinishAlwaysNewline(t *testing.T) {
	for _, isTTY := range []bool{true, false} {
		t.Run(fmt.Sprintf("IsTTY=%v", isTTY), func(t *testing.T) {
			d := &Display{IsTTY: isTTY}
			out := captureOutput(func() {
				d.Finish("done")
			})
			if !strings.HasSuffix(out, "\n") {
				t.Errorf("Finish should end with newline, got %q", out)
			}
		})
	}
}

func TestDisplay_PaddingClearsPreviousLongerLine(t *testing.T) {
	d := &Display{IsTTY: true}
	out := captureOutput(func() {
		d.Status("long message here")
		d.Status("short")
	})
	// Second status should have padding spaces to clear the longer first line
	parts := strings.Split(out, "\r")
	// parts[0] is empty (before first \r), parts[1] is first status, parts[2] is second
	if len(parts) < 3 {
		t.Fatalf("expected at least 3 parts split by \\r, got %d: %q", len(parts), out)
	}
	lastPart := parts[2]
	if !strings.Contains(lastPart, "short") {
		t.Errorf("expected 'short' in last part, got %q", lastPart)
	}
	// Should have trailing spaces
	afterShort := strings.TrimPrefix(lastPart, "short")
	if len(afterShort) == 0 {
		t.Errorf("expected padding spaces after 'short', got none")
	}
}
