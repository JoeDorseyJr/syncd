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
	if err == nil {
		return // default config happens to exist
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
	if !strings.Contains(err.Error(), "invalid config") {
		t.Errorf("expected 'invalid config' in error, got: %v", err)
	}
}

func TestLoad_UnknownTopLevelKey(t *testing.T) {
	content := `
brew:
  - git
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown key 'brew' (should be 'brews')")
	}
	if !strings.Contains(err.Error(), "brew") {
		t.Errorf("expected error to mention 'brew', got: %v", err)
	}
}

func TestLoad_UnknownCleanupKey(t *testing.T) {
	content := `
brews:
  - git
cleanup:
  remove_unlistd: true
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for misspelled cleanup key 'remove_unlistd'")
	}
	if !strings.Contains(err.Error(), "remove_unlistd") {
		t.Errorf("expected error to mention 'remove_unlistd', got: %v", err)
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

func TestLoad_PinField(t *testing.T) {
	content := `
brews:
  - git
pin:
  - node@22
  - firefox
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Pin) != 2 {
		t.Errorf("expected 2 pins, got %d", len(cfg.Pin))
	}
	if cfg.Pin[0] != "node@22" || cfg.Pin[1] != "firefox" {
		t.Errorf("unexpected pin values: %v", cfg.Pin)
	}
}

func TestLoad_InvalidPinObject(t *testing.T) {
	content := `
brews:
  - git
pin:
  name: node
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for object in pin field")
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

func TestLoad_ValidDefaults(t *testing.T) {
	content := `
brews:
  - git
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: 48
    kill:
      - Dock
  - domain: NSGlobalDomain
    key: KeyRepeat
    type: int
    value: 2
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Defaults) != 2 {
		t.Fatalf("expected 2 defaults, got %d", len(cfg.Defaults))
	}
	if cfg.Defaults[0].Domain != "com.apple.dock" {
		t.Errorf("unexpected domain: %s", cfg.Defaults[0].Domain)
	}
	if len(cfg.Defaults[0].Kill) != 1 || cfg.Defaults[0].Kill[0] != "Dock" {
		t.Errorf("unexpected kill: %v", cfg.Defaults[0].Kill)
	}
	if cfg.Defaults[1].Kill != nil {
		t.Errorf("expected nil kill, got: %v", cfg.Defaults[1].Kill)
	}
}

func TestLoad_DefaultsMissingDomain(t *testing.T) {
	content := `
defaults:
  - key: tilesize
    type: int
    value: 48
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing domain")
	}
	if !strings.Contains(err.Error(), "domain is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsMissingKey(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    type: int
    value: 48
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "key is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsInvalidType(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: integer
    value: 48
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "type must be string, int, float, or bool") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsTypeMismatch(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: hello
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for type/value mismatch")
	}
	if !strings.Contains(err.Error(), "not a valid int") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsMissingValue(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing value")
	}
	if !strings.Contains(err.Error(), "value is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsUnknownField(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: 48
    extra: bad
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unknown field in defaults entry")
	}
	if !strings.Contains(err.Error(), "extra") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsKillPresent(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: autohide
    type: bool
    value: true
    kill:
      - Dock
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Defaults[0].Kill) != 1 || cfg.Defaults[0].Kill[0] != "Dock" {
		t.Errorf("unexpected kill: %v", cfg.Defaults[0].Kill)
	}
}

func TestLoad_DefaultsKillAbsent(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: 48
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Defaults[0].Kill != nil {
		t.Errorf("expected nil kill, got: %v", cfg.Defaults[0].Kill)
	}
}

func TestLoad_DefaultsBoolTypeMismatch(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: autohide
    type: bool
    value: 1
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for bool type with int value")
	}
	if !strings.Contains(err.Error(), "not a valid bool") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsArrayValueRejected(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: string
    value:
      - a
      - b
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for array value")
	}
	if !strings.Contains(err.Error(), "scalar") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsMapValueRejected(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: string
    value:
      nested: thing
`
	path := writeTemp(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for map value")
	}
	if !strings.Contains(err.Error(), "scalar") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsFloatType(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: magnification
    type: float
    value: 0.5
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Defaults[0].Value != 0.5 {
		t.Errorf("unexpected value: %v", cfg.Defaults[0].Value)
	}
}

func TestLoad_DefaultsStringType(t *testing.T) {
	content := `
defaults:
  - domain: com.apple.dock
    key: orientation
    type: string
    value: left
`
	path := writeTemp(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Defaults[0].Value != "left" {
		t.Errorf("unexpected value: %v", cfg.Defaults[0].Value)
	}
}
