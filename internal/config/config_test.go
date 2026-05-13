package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
taps:
  - homebrew/cask-fonts
  - hashicorp/tap
brews:
  - git
  - go
casks:
  - firefox
cleanup:
  remove_unlisted: true
  clear_cache: true
  autoremove: true
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Taps) != 2 {
		t.Errorf("expected 2 taps, got %d", len(cfg.Taps))
	}
	if len(cfg.Brews) != 2 {
		t.Errorf("expected 2 brews, got %d", len(cfg.Brews))
	}
	if len(cfg.Casks) != 1 {
		t.Errorf("expected 1 cask, got %d", len(cfg.Casks))
	}
	if !cfg.Cleanup.RemoveUnlisted {
		t.Error("expected remove_unlisted to be true")
	}
	if !cfg.Cleanup.ClearCache {
		t.Error("expected clear_cache to be true")
	}
	if !cfg.Cleanup.Autoremove {
		t.Error("expected autoremove to be true")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestLoad_EmptyPathUsesDefault(t *testing.T) {
	_, err := Load("")
	// We can't control ~/.config/syncd/config.yaml in tests, but we can
	// verify it attempts the correct default path rather than erroring on "".
	if err == nil {
		return // default config happens to exist — that's fine
	}
	home, _ := os.UserHomeDir()
	expected := home + "/.config/syncd/config.yaml"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error to reference default path %q, got: %v", expected, err)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTemp(t, "taps: [unclosed")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "invalid YAML") {
		t.Errorf("expected 'invalid YAML' in error, got: %v", err)
	}
}

func TestLoad_PartialConfig(t *testing.T) {
	content := `
brews:
  - git
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Taps) != 0 {
		t.Errorf("expected 0 taps, got %d", len(cfg.Taps))
	}
	if len(cfg.Brews) != 1 {
		t.Errorf("expected 1 brew, got %d", len(cfg.Brews))
	}
	if len(cfg.Casks) != 0 {
		t.Errorf("expected 0 casks, got %d", len(cfg.Casks))
	}
	if cfg.Cleanup.RemoveUnlisted {
		t.Error("expected remove_unlisted to default to false")
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
