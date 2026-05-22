package download

import (
	"fmt"
	"testing"

	"github.com/joedorseyjr/syncd/internal/runner"
)

func TestCacheDir(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte("/Users/joe/Library/Caches/Homebrew\n")},
		},
	}

	dir, err := CacheDir(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "/Users/joe/Library/Caches/Homebrew/downloads"
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestCacheDir_Error(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Err: &runner.RunError{Cmd: "brew --cache", Err: fmt.Errorf("not found")}},
		},
	}

	_, err := CacheDir(mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
