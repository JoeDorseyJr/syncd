package progress

import "testing"

func TestExtractError_WithErrorLine(t *testing.T) {
	output := []byte("Downloading...\nPouring...\nError: sha256 mismatch\nExpected: abc123\nActual: def456\nExtra line\n")
	got := ExtractError(output)
	want := "Error: sha256 mismatch\nExpected: abc123\nActual: def456"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractError_NoErrorLine(t *testing.T) {
	output := []byte("line1\nline2\nline3\nline4\nline5\n")
	got := ExtractError(output)
	want := "line3\nline4\nline5"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractError_EmptyOutput(t *testing.T) {
	got := ExtractError([]byte{})
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestExtractError_MaxFiveLines(t *testing.T) {
	// Error line at start with many context lines — capped at 3 (error + 2 context)
	output := []byte("Error: something\nctx1\nctx2\nctx3\nctx4\n")
	got := ExtractError(output)
	want := "Error: something\nctx1\nctx2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractError_CaseInsensitive(t *testing.T) {
	output := []byte("some output\nerror: bad thing happened\ndetails here\n")
	got := ExtractError(output)
	want := "error: bad thing happened\ndetails here"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
