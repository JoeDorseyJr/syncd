package plan

import "github.com/joedorseyjr/syncd/internal/config"

// Plan represents the set of changes needed to reconcile state with config.
type Plan struct {
	TapsToAdd      []string
	TapsToRemove   []string
	BrewsToInstall []string
	BrewsToRemove  []string
	CasksToInstall []string
	CasksToRemove  []string
	Autoremove     bool
	ClearCache     bool
}

// IsEmpty returns true if no changes are needed.
func (p *Plan) IsEmpty() bool {
	return len(p.TapsToAdd) == 0 &&
		len(p.TapsToRemove) == 0 &&
		len(p.BrewsToInstall) == 0 &&
		len(p.BrewsToRemove) == 0 &&
		len(p.CasksToInstall) == 0 &&
		len(p.CasksToRemove) == 0 &&
		!p.Autoremove &&
		!p.ClearCache
}

// State represents installed packages (mirrors brew.State to avoid import cycle).
type State struct {
	Taps  []string
	Brews []string
	Casks []string
}

// Compute calculates the diff between desired config and actual state.
func Compute(cfg *config.Config, state *State) *Plan {
	p := &Plan{
		TapsToAdd:      diff(cfg.Taps, state.Taps),
		BrewsToInstall: diff(cfg.Brews, state.Brews),
		CasksToInstall: diff(cfg.Casks, state.Casks),
		Autoremove:     cfg.Cleanup.Autoremove,
		ClearCache:     cfg.Cleanup.ClearCache,
	}

	if cfg.Cleanup.RemoveUnlisted {
		p.TapsToRemove = diff(state.Taps, cfg.Taps)
		p.BrewsToRemove = diff(state.Brews, cfg.Brews)
		p.CasksToRemove = diff(state.Casks, cfg.Casks)
	}

	return p
}

// diff returns items in a that are not in b.
func diff(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, item := range b {
		set[item] = struct{}{}
	}
	var result []string
	for _, item := range a {
		if _, found := set[item]; !found {
			result = append(result, item)
		}
	}
	return result
}
