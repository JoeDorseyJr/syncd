package download

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/term"
)

// Color constants (duplicated from cli to avoid import cycle).
var (
	dlGreen = "\033[32m"
	dlRed   = "\033[31m"
	dlReset = "\033[0m"
)

// ProgressWriter wraps writes and reports byte progress.
type ProgressWriter struct {
	Name       string
	Total      int64
	Downloaded int64
	OnProgress func(name string, downloaded, total int64)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Downloaded += int64(n)
	if pw.OnProgress != nil {
		pw.OnProgress(pw.Name, pw.Downloaded, pw.Total)
	}
	return n, nil
}

// SlotState represents the state of one download slot.
type SlotState struct {
	Name       string
	Downloaded int64
	Total      int64
	Done       bool
	Err        error
}

// DownloadDisplay manages multi-line progress rendering.
type DownloadDisplay struct {
	mu       sync.Mutex
	slots    []SlotState
	isTTY    bool
	rendered bool
	stopped  int32
}

// NewDownloadDisplay creates a display with the given number of slots.
func NewDownloadDisplay(concurrency int) *DownloadDisplay {
	return &DownloadDisplay{
		slots: make([]SlotState, concurrency),
		isTTY: term.IsTerminal(int(os.Stdout.Fd())),
	}
}

// UpdateProgress updates a slot's byte counts. Auto-assigns if name not found.
func (d *DownloadDisplay) UpdateProgress(name string, downloaded, total int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.slots {
		if d.slots[i].Name == name && !d.slots[i].Done {
			d.slots[i].Downloaded = downloaded
			d.slots[i].Total = total
			return
		}
	}
	// Auto-assign to first empty slot
	for i := range d.slots {
		if d.slots[i].Name == "" {
			d.slots[i] = SlotState{Name: name, Downloaded: downloaded, Total: total}
			return
		}
	}
	// Take over first done slot
	for i := range d.slots {
		if d.slots[i].Done {
			d.slots[i] = SlotState{Name: name, Downloaded: downloaded, Total: total}
			return
		}
	}
}

// MarkDone marks a slot as complete.
func (d *DownloadDisplay) MarkDone(name string, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.slots {
		if d.slots[i].Name == name && !d.slots[i].Done {
			d.slots[i].Done = true
			d.slots[i].Err = err
			if !d.isTTY {
				d.printNonTTY(d.slots[i])
			}
			return
		}
	}
	// Not in a slot yet — print directly for non-TTY
	if !d.isTTY {
		s := SlotState{Name: name, Done: true, Err: err}
		d.printNonTTY(s)
	}
}

func (d *DownloadDisplay) printNonTTY(s SlotState) {
	if s.Err != nil {
		fmt.Fprintf(os.Stdout, "  %s✗%s %s: %v\n", dlRed, dlReset, s.Name, s.Err)
	} else {
		bytes := s.Downloaded
		if s.Total > 0 && s.Total > bytes {
			bytes = s.Total
		}
		fmt.Fprintf(os.Stdout, "  %s✓%s %s (%.1f MB)\n", dlGreen, dlReset, s.Name, float64(bytes)/1e6)
	}
}

// Start begins the 100ms render loop. Returns a stop function.
func (d *DownloadDisplay) Start() func() {
	if !d.isTTY {
		return func() {}
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				d.render()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()
	return func() {
		atomic.StoreInt32(&d.stopped, 1)
		close(done)
		d.render() // final render
		// Move past the slots
		fmt.Println()
	}
}

func (d *DownloadDisplay) render() {
	if !d.isTTY {
		return
	}
	d.mu.Lock()
	lines := make([]string, len(d.slots))
	for i, s := range d.slots {
		lines[i] = renderSlot(s)
	}
	d.mu.Unlock()

	if d.rendered {
		// Move cursor up N lines
		fmt.Fprintf(os.Stdout, "\033[%dA", len(lines))
	}
	for _, line := range lines {
		// Clear line and print
		fmt.Fprintf(os.Stdout, "\r\033[K%s\n", line)
	}
	d.rendered = true
}

// renderSlot renders a single slot state to a string.
func renderSlot(s SlotState) string {
	if s.Name == "" {
		return ""
	}
	if s.Done && s.Err != nil {
		return fmt.Sprintf("  %s✗%s %s: %v", dlRed, dlReset, s.Name, s.Err)
	}
	if s.Done {
		bytes := s.Downloaded
		if s.Total > 0 {
			bytes = s.Total
		}
		return fmt.Sprintf("  %s✓%s %s (%.1f MB)", dlGreen, dlReset, s.Name, float64(bytes)/1e6)
	}
	if s.Total <= 0 {
		return fmt.Sprintf("  ↓ %s  %.1f MB", s.Name, float64(s.Downloaded)/1e6)
	}
	pct := int(s.Downloaded * 100 / s.Total)
	if pct > 100 {
		pct = 100
	}
	filled := pct * 20 / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 20-filled)
	return fmt.Sprintf("  ↓ %s  %s %3d%% (%.1f/%.1f MB)",
		s.Name, bar, pct, float64(s.Downloaded)/1e6, float64(s.Total)/1e6)
}
