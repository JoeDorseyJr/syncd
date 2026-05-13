package brew

import (
	"fmt"

	"github.com/joedorseyjr/syncd/internal/plan"
)

// Result captures the outcome of a single brew operation.
type Result struct {
	Action  string
	Package string
	Err     error
}

// Execute runs brew commands to reconcile state per the plan.
// Execution order: taps → brew installs → cask installs → tap removals →
// brew removals → cask removals → autoremove → cleanup.
// Continues on failure, collecting all results.
func Execute(runner CommandRunner, p *plan.Plan) []Result {
	var results []Result

	for _, name := range p.TapsToAdd {
		_, err := runner.Run("brew", "tap", name)
		results = append(results, Result{Action: "tap", Package: name, Err: err})
	}
	for _, name := range p.BrewsToInstall {
		_, err := runner.Run("brew", "install", name)
		results = append(results, Result{Action: "install", Package: name, Err: err})
	}
	for _, name := range p.CasksToInstall {
		_, err := runner.Run("brew", "install", "--cask", name)
		results = append(results, Result{Action: "install-cask", Package: name, Err: err})
	}
	for _, name := range p.TapsToRemove {
		_, err := runner.Run("brew", "untap", name)
		results = append(results, Result{Action: "untap", Package: name, Err: err})
	}
	for _, name := range p.BrewsToRemove {
		_, err := runner.Run("brew", "uninstall", name)
		results = append(results, Result{Action: "uninstall", Package: name, Err: err})
	}
	for _, name := range p.CasksToRemove {
		_, err := runner.Run("brew", "uninstall", "--cask", name)
		results = append(results, Result{Action: "uninstall-cask", Package: name, Err: err})
	}
	if p.Autoremove {
		_, err := runner.Run("brew", "autoremove")
		results = append(results, Result{Action: "autoremove", Err: err})
	}
	if p.ClearCache {
		_, err := runner.Run("brew", "cleanup")
		results = append(results, Result{Action: "cleanup", Err: err})
	}

	return results
}

// HasErrors returns true if any result has an error.
func HasErrors(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}

// FormatResults returns a human-readable summary of execution results.
func FormatResults(results []Result) string {
	var s string
	for _, r := range results {
		if r.Err != nil {
			if r.Package != "" {
				s += fmt.Sprintf("  ✗ %s %s: %v\n", r.Action, r.Package, r.Err)
			} else {
				s += fmt.Sprintf("  ✗ %s: %v\n", r.Action, r.Err)
			}
		} else {
			if r.Package != "" {
				s += fmt.Sprintf("  ✓ %s %s\n", r.Action, r.Package)
			} else {
				s += fmt.Sprintf("  ✓ %s\n", r.Action)
			}
		}
	}
	return s
}
