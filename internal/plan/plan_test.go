package plan

import (
	"testing"

	"github.com/joedorseyjr/syncd/internal/config"
)

func TestCompute_PackagesToAdd(t *testing.T) {
	cfg := &config.Config{
		Taps:  []string{"homebrew/cask", "hashicorp/tap"},
		Brews: []string{"git", "go", "wget"},
		Casks: []string{"firefox", "iterm2"},
	}
	state := &State{
		Taps:  []string{"homebrew/cask"},
		Brews: []string{"git"},
		Casks: []string{"firefox"},
	}

	p := Compute(cfg, state)

	assertSlice(t, "TapsToAdd", p.TapsToAdd, []string{"hashicorp/tap"})
	assertSlice(t, "BrewsToInstall", p.BrewsToInstall, []string{"go", "wget"})
	assertSlice(t, "CasksToInstall", p.CasksToInstall, []string{"iterm2"})
}

func TestCompute_TapsToRemove(t *testing.T) {
	cfg := &config.Config{
		Taps:    []string{"homebrew/cask"},
		Cleanup: config.Cleanup{RemoveUnlisted: true},
	}
	state := &State{
		Taps: []string{"homebrew/cask", "old/tap"},
	}

	p := Compute(cfg, state)

	assertSlice(t, "TapsToRemove", p.TapsToRemove, []string{"old/tap"})
}

func TestCompute_PackagesToRemove(t *testing.T) {
	cfg := &config.Config{
		Brews:   []string{"git"},
		Casks:   []string{"firefox"},
		Cleanup: config.Cleanup{RemoveUnlisted: true},
	}
	state := &State{
		Brews:  []string{"git", "wget", "libssh2"},
		Leaves: []string{"git", "wget"},
		Casks:  []string{"firefox", "slack"},
	}

	p := Compute(cfg, state)

	assertSlice(t, "BrewsToRemove", p.BrewsToRemove, []string{"wget"})
	assertSlice(t, "CasksToRemove", p.CasksToRemove, []string{"slack"})
}

func TestCompute_NoRemovalsWhenDisabled(t *testing.T) {
	cfg := &config.Config{
		Brews:   []string{"git"},
		Cleanup: config.Cleanup{RemoveUnlisted: false},
	}
	state := &State{
		Brews:  []string{"git", "wget"},
		Leaves: []string{"git", "wget"},
	}

	p := Compute(cfg, state)

	if len(p.BrewsToRemove) != 0 {
		t.Errorf("expected no removals, got %v", p.BrewsToRemove)
	}
	if len(p.TapsToRemove) != 0 {
		t.Errorf("expected no tap removals, got %v", p.TapsToRemove)
	}
}

func TestCompute_DependencyOnlyNotRemoved(t *testing.T) {
	cfg := &config.Config{
		Brews:   []string{"git"},
		Cleanup: config.Cleanup{RemoveUnlisted: true},
	}
	state := &State{
		Brews:  []string{"git", "libssh2", "wget"},
		Leaves: []string{"git", "wget"},
	}

	p := Compute(cfg, state)

	// libssh2 is installed but not a leaf — should NOT be removed
	assertSlice(t, "BrewsToRemove", p.BrewsToRemove, []string{"wget"})
}

func TestCompute_EmptyPlan(t *testing.T) {
	cfg := &config.Config{
		Taps:  []string{"homebrew/cask"},
		Brews: []string{"git"},
		Casks: []string{"firefox"},
	}
	state := &State{
		Taps:  []string{"homebrew/cask"},
		Brews: []string{"git"},
		Casks: []string{"firefox"},
	}

	p := Compute(cfg, state)

	if !p.IsEmpty() {
		t.Error("expected IsEmpty() true")
	}
	if p.HasChanges() {
		t.Error("expected HasChanges() false")
	}
}

func TestCompute_CleanupOnlyIsNotDrift(t *testing.T) {
	cfg := &config.Config{
		Taps:    []string{"homebrew/cask"},
		Brews:   []string{"git"},
		Cleanup: config.Cleanup{Autoremove: true, ClearCache: true},
	}
	state := &State{
		Taps:  []string{"homebrew/cask"},
		Brews: []string{"git"},
	}

	p := Compute(cfg, state)

	// Cleanup flags are set but there are no package changes
	if p.HasChanges() {
		t.Error("expected HasChanges() false for cleanup-only plan")
	}
	// IsEmpty is false because there IS work to do (cleanup)
	if p.IsEmpty() {
		t.Error("expected IsEmpty() false when cleanup flags are set")
	}
}

func TestCompute_CaseSensitivity(t *testing.T) {
	cfg := &config.Config{
		Brews: []string{"Git"},
	}
	state := &State{
		Brews: []string{"git"},
	}

	p := Compute(cfg, state)

	assertSlice(t, "BrewsToInstall", p.BrewsToInstall, []string{"Git"})
}

func TestCompute_DuplicateConfigEntries(t *testing.T) {
	cfg := &config.Config{
		Taps:  []string{"homebrew/cask", "homebrew/cask"},
		Brews: []string{"git", "go", "git"},
		Casks: []string{"firefox", "firefox"},
	}
	state := &State{}

	p := Compute(cfg, state)

	assertSlice(t, "TapsToAdd", p.TapsToAdd, []string{"homebrew/cask"})
	assertSlice(t, "BrewsToInstall", p.BrewsToInstall, []string{"git", "go"})
	assertSlice(t, "CasksToInstall", p.CasksToInstall, []string{"firefox"})
}

func assertSlice(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: expected %v, got %v", name, want, got)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d]: expected %q, got %q", name, i, want[i], got[i])
		}
	}
}
