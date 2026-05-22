package download

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestProgressWriter_TracksCumulativeBytes(t *testing.T) {
	var pw ProgressWriter
	pw.Name = "test"
	pw.Total = 100

	pw.Write([]byte("hello"))
	if pw.Downloaded != 5 {
		t.Fatalf("expected 5, got %d", pw.Downloaded)
	}
	pw.Write([]byte("world!"))
	if pw.Downloaded != 11 {
		t.Fatalf("expected 11, got %d", pw.Downloaded)
	}
}

func TestProgressWriter_CallsOnProgress(t *testing.T) {
	var calls []struct {
		name       string
		downloaded int64
		total      int64
	}
	pw := ProgressWriter{
		Name:  "pkg",
		Total: 50,
		OnProgress: func(name string, downloaded, total int64) {
			calls = append(calls, struct {
				name       string
				downloaded int64
				total      int64
			}{name, downloaded, total})
		},
	}

	pw.Write([]byte("abc"))
	pw.Write([]byte("de"))

	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].name != "pkg" || calls[0].downloaded != 3 || calls[0].total != 50 {
		t.Fatalf("call 0 wrong: %+v", calls[0])
	}
	if calls[1].downloaded != 5 {
		t.Fatalf("call 1 downloaded wrong: %d", calls[1].downloaded)
	}
}

func TestProgressWriter_NilOnProgress(t *testing.T) {
	pw := ProgressWriter{Name: "test", Total: 10}
	// Should not panic
	n, err := pw.Write([]byte("data"))
	if err != nil || n != 4 {
		t.Fatalf("unexpected: n=%d, err=%v", n, err)
	}
}

func TestRenderSlot_50Percent(t *testing.T) {
	s := SlotState{Name: "neovim", Downloaded: 25000000, Total: 50000000}
	out := renderSlot(s)
	if !strings.Contains(out, "██████████░░░░░░░░░░") {
		t.Fatalf("expected 10 filled + 10 empty, got: %s", out)
	}
	if !strings.Contains(out, "50%") {
		t.Fatalf("expected 50%%, got: %s", out)
	}
	if !strings.Contains(out, "neovim") {
		t.Fatalf("expected name, got: %s", out)
	}
}

func TestRenderSlot_0Percent(t *testing.T) {
	s := SlotState{Name: "pkg", Downloaded: 0, Total: 100000000}
	out := renderSlot(s)
	if !strings.Contains(out, "░░░░░░░░░░░░░░░░░░░░") {
		t.Fatalf("expected 20 empty, got: %s", out)
	}
	if !strings.Contains(out, "0%") {
		t.Fatalf("expected 0%%, got: %s", out)
	}
}

func TestRenderSlot_100Percent(t *testing.T) {
	s := SlotState{Name: "pkg", Downloaded: 50000000, Total: 50000000}
	out := renderSlot(s)
	if !strings.Contains(out, "████████████████████") {
		t.Fatalf("expected 20 filled, got: %s", out)
	}
	if !strings.Contains(out, "100%") {
		t.Fatalf("expected 100%%, got: %s", out)
	}
}

func TestRenderSlot_UnknownTotal(t *testing.T) {
	s := SlotState{Name: "pkg", Downloaded: 22000000, Total: 0}
	out := renderSlot(s)
	if strings.Contains(out, "█") || strings.Contains(out, "░") {
		t.Fatalf("should not have bar for unknown total: %s", out)
	}
	if !strings.Contains(out, "22.0 MB") {
		t.Fatalf("expected bytes display, got: %s", out)
	}
	if !strings.Contains(out, "↓") {
		t.Fatalf("expected arrow, got: %s", out)
	}
}

func TestRenderSlot_DoneSuccess(t *testing.T) {
	s := SlotState{Name: "neovim", Downloaded: 50000000, Total: 50000000, Done: true}
	out := renderSlot(s)
	if !strings.Contains(out, "✓") {
		t.Fatalf("expected ✓, got: %s", out)
	}
	if !strings.Contains(out, "neovim") {
		t.Fatalf("expected name, got: %s", out)
	}
	if !strings.Contains(out, "50.0 MB") {
		t.Fatalf("expected size, got: %s", out)
	}
}

func TestRenderSlot_DoneError(t *testing.T) {
	s := SlotState{Name: "bad", Done: true, Err: errors.New("timeout")}
	out := renderSlot(s)
	if !strings.Contains(out, "✗") {
		t.Fatalf("expected ✗, got: %s", out)
	}
	if !strings.Contains(out, "bad") {
		t.Fatalf("expected name, got: %s", out)
	}
	if !strings.Contains(out, "timeout") {
		t.Fatalf("expected error, got: %s", out)
	}
}

func TestRenderSlot_Empty(t *testing.T) {
	s := SlotState{}
	out := renderSlot(s)
	if out != "" {
		t.Fatalf("expected empty string, got: %q", out)
	}
}

func TestDownloadDisplay_AssignSlot(t *testing.T) {
	dd := &DownloadDisplay{slots: make([]SlotState, 2)}
	dd.UpdateProgress("pkg1", 100, 1000)
	dd.UpdateProgress("pkg2", 200, 2000)

	if dd.slots[0].Name != "pkg1" {
		t.Fatalf("slot 0 should be pkg1, got %s", dd.slots[0].Name)
	}
	if dd.slots[1].Name != "pkg2" {
		t.Fatalf("slot 1 should be pkg2, got %s", dd.slots[1].Name)
	}
}

func TestDownloadDisplay_MarkDone(t *testing.T) {
	dd := &DownloadDisplay{slots: make([]SlotState, 2), isTTY: true}
	dd.UpdateProgress("pkg1", 100, 1000)
	dd.MarkDone("pkg1", nil)

	// Slot should be cleared after marking done
	if dd.slots[0].Name != "" {
		t.Fatal("slot 0 should be cleared after MarkDone")
	}
	// Item should be in pending list
	if len(dd.pending) != 1 || dd.pending[0].Name != "pkg1" {
		t.Fatal("expected pkg1 in pending")
	}
}

func TestDownloadDisplay_SlotTransition(t *testing.T) {
	dd := &DownloadDisplay{slots: make([]SlotState, 1), isTTY: true}
	dd.UpdateProgress("pkg1", 100, 1000)
	dd.MarkDone("pkg1", nil)

	// Slot is freed immediately, so a new package can take it
	dd.UpdateProgress("pkg2", 50, 500)
	if dd.slots[0].Name != "pkg2" {
		t.Fatalf("slot 0 should be pkg2, got %s", dd.slots[0].Name)
	}
}

func TestDownloadDisplay_NonTTY_NoCursorCodes(t *testing.T) {
	dd := &DownloadDisplay{slots: make([]SlotState, 2), isTTY: false}
	// render should be a no-op for non-TTY
	dd.UpdateProgress("pkg1", 500, 1000)
	dd.render()
	// If we got here without writing cursor codes, the test passes.
	// The real verification is that render() returns early when !isTTY.
	if dd.rendered {
		t.Fatal("non-TTY should not set rendered=true")
	}
}

func TestPreDownload_OnProgressCalled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		w.Write([]byte("0123456789"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	pkgs := []PackageURL{{Name: "prog", Version: "1.0", URL: srv.URL + "/f", Filename: "prog.tar.gz"}}

	var called int64
	results := PreDownload(pkgs, Options{
		Concurrency: 1,
		CacheDir:    dir,
		OnProgress: func(name string, downloaded, total int64) {
			atomic.AddInt64(&called, 1)
			if name != "prog" {
				t.Errorf("expected name 'prog', got %q", name)
			}
			if total != 10 {
				t.Errorf("expected total 10, got %d", total)
			}
		},
	})

	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if atomic.LoadInt64(&called) == 0 {
		t.Fatal("OnProgress was never called")
	}
}

func TestPreDownload_WithDownloadDisplay(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "20")
		w.Write([]byte("01234567890123456789"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	pkgs := []PackageURL{
		{Name: "pkg1", Version: "1.0", URL: srv.URL + "/a", Filename: "pkg1.tar.gz"},
		{Name: "pkg2", Version: "1.0", URL: srv.URL + "/b", Filename: "pkg2.tar.gz"},
	}

	// Create display in non-TTY mode (test environment)
	dd := &DownloadDisplay{slots: make([]SlotState, 2), isTTY: false}

	var doneCount int64
	results := PreDownload(pkgs, Options{
		Concurrency: 2,
		CacheDir:    dir,
		OnProgress: func(name string, downloaded, total int64) {
			dd.UpdateProgress(name, downloaded, total)
		},
		OnComplete: func(res Result) {
			dd.MarkDone(res.Package.Name, res.Err)
			atomic.AddInt64(&doneCount, 1)
		},
	})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error: %v", r.Err)
		}
	}
	if atomic.LoadInt64(&doneCount) != 2 {
		t.Fatalf("expected 2 done callbacks, got %d", doneCount)
	}
}
