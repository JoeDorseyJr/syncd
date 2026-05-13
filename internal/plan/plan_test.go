package plan

import (
	"testing"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
)

func TestCompute_PackagesToAdd(t *testing.T) {
	cfg := &config.Config{
		Taps:  []string{"homebrew/cask", "hashicorp/tap"},
		Brews: []string{"git", "go", "wget"},
		Casks: []string{"firefox", "iterm2"},
	}
	state := &brew.State{
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
	state := &brew.State{
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
	state := &brew.State{
		Brews: []string{"git", "wget"},
		Casks: []string{"firefox", "slack"},
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
	state := &brew.State{
		Brews: []string{"git", "wget"},
	}

	p := Compute(cfg, state)

	if len(p.BrewsToRemove) != 0 {
		t.Errorf("expected no removals, got %v", p.BrewsToRemove)
	}
	if len(p.TapsToRemove) != 0 {
		t.Errorf("expected no tap removals, got %v", p.TapsToRemove)
	}
}

func TestCompute_EmptyPlan(t *testing.T) {
	cfg := &config.Config{
		Taps:  []string{"homebrew/cask"},
		Brews: []string{"git"},
		Casks: []string{"firefox"},
	}
	state := &brew.State{
		Taps:  []string{"homebrew/cask"},
		Brews: []string{"git"},
		Casks: []string{"firefox"},
	}

	p := Compute(cfg, state)

	if !p.IsEmpty() {
		t.Error("expected empty plan")
	}
}

func TestCompute_CaseSensitivity(t *testing.T) {
	cfg := &config.Config{
		Brews: []string{"Git"},
	}
	state := &brew.State{
		Brews: []string{"git"},
	}

	p := Compute(cfg, state)

	// "Git" != "git" — case-sensitive comparison
	assertSlice(t, "BrewsToInstall", p.BrewsToInstall, []string{"Git"})
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
