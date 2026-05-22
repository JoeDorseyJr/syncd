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
	pending  []SlotState // completed items waiting to be printed permanently
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
		if d.slots[i].Name == name {
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
			} else {
				d.pending = append(d.pending, d.slots[i])
			}
			// Free the slot for the next package
			d.slots[i] = SlotState{}
			return
		}
	}
	// Not in a slot yet — print directly
	s := SlotState{Name: name, Done: true, Err: err}
	if !d.isTTY {
		d.printNonTTY(s)
	} else {
		d.pending = append(d.pending, s)
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
		d.render() // final render to flush pending completions
		// Clear the active slot lines since we're done
		if d.rendered {
			fmt.Fprintf(os.Stdout, "\033[%dA", len(d.slots))
			for range d.slots {
				fmt.Fprintf(os.Stdout, "\r\033[K\n")
			}
		}
	}
}

func (d *DownloadDisplay) render() {
	if !d.isTTY {
		return
	}
	d.mu.Lock()
	// Count active (non-empty) slots for cursor movement
	activeLines := make([]string, 0, len(d.slots))
	for _, s := range d.slots {
		activeLines = append(activeLines, renderSlot(s))
	}
	pendingItems := d.pending
	d.pending = nil
	d.mu.Unlock()

	// If we previously rendered, move cursor up to overwrite active slot lines
	if d.rendered {
		fmt.Fprintf(os.Stdout, "\033[%dA", len(activeLines))
		// Clear the active lines
		for range activeLines {
			fmt.Fprintf(os.Stdout, "\r\033[K\n")
		}
		fmt.Fprintf(os.Stdout, "\033[%dA", len(activeLines))
	}

	// Print completed items permanently (these won't be overwritten)
	for _, s := range pendingItems {
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

	// Print active slot lines (these will be overwritten on next render)
	for _, line := range activeLines {
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
