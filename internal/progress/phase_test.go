package progress

import "testing"

func TestDetectPhase(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"Downloading https://ghcr.io/v2/homebrew/core/neovim/blobs/sha256:abc", "downloading"},
		{"==> downloading https://example.com/file.tar.gz", "downloading"},
		{"Pouring neovim--0.9.5.arm64_sonoma.bottle.tar.gz", "pouring"},
		{"==> pouring neovim--0.9.5", "pouring"},
		{"Installing neovim", "installing"},
		{"==> installing dependencies", "installing"},
		{"Built from source", "built"},
		{"==> built from source on 2024-01-01", "built"},
		{"Already installed", ""},
		{"==> Fetching dependencies", ""},
		{"", ""},
		{"some random output line", ""},
	}
	for _, tt := range tests {
		got := DetectPhase(tt.line)
		if got != tt.want {
			t.Errorf("DetectPhase(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}
