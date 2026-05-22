package download

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/joedorseyjr/syncd/internal/runner"
)

func TestCacheFilename_Formula(t *testing.T) {
	url := "https://ghcr.io/v2/homebrew/core/neovim/blobs/sha256:abc123"
	name := "neovim"
	version := "0.10.0"

	result := CacheFilename(url, name, version, false)

	hash := sha256.Sum256([]byte(url))
	expected := fmt.Sprintf("%x--neovim--0.10.0.bottle.tar.gz", hash)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestCacheFilename_CaskDMG(t *testing.T) {
	url := "https://cdn.mozilla.net/firefox-126.0.dmg"
	name := "firefox"
	version := "126.0"

	result := CacheFilename(url, name, version, true)

	hash := sha256.Sum256([]byte(url))
	expected := fmt.Sprintf("%x--firefox--126.0.dmg", hash)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestCacheFilename_CaskPKG(t *testing.T) {
	url := "https://example.com/app-1.0.pkg"
	name := "myapp"
	version := "1.0"

	result := CacheFilename(url, name, version, true)

	hash := sha256.Sum256([]byte(url))
	expected := fmt.Sprintf("%x--myapp--1.0.pkg", hash)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

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
