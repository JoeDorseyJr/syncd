package defaults

import (
	"fmt"

	"github.com/joedorseyjr/syncd/internal/runner"
)

// WriteResult captures the outcome of a defaults write operation.
type WriteResult struct {
	Domain string
	Key    string
	Err    error
}

// KillResult captures the outcome of a killall operation.
type KillResult struct {
	App string
	Err error
}

// WriteDrifted writes all drifted defaults and kills affected apps.
func WriteDrifted(r runner.CommandRunner, drifted []DriftEntry) ([]WriteResult, []KillResult) {
	var writeResults []WriteResult
	seen := map[string]bool{}

	for _, d := range drifted {
		val := formatValue(d.Type, d.Desired)
		_, err := r.RunMutate("defaults", "write", d.Domain, d.Key, "-"+d.Type, val)
		writeResults = append(writeResults, WriteResult{Domain: d.Domain, Key: d.Key, Err: err})
		if err == nil {
			for _, app := range d.Kill {
				seen[app] = true
			}
		}
	}

	var killResults []KillResult
	for app := range seen {
		_, err := r.RunMutate("killall", app)
		killResults = append(killResults, KillResult{App: app, Err: err})
	}
	return writeResults, killResults
}

func formatValue(typ, desired string) string {
	if typ == "bool" {
		if desired == "true" {
			return "TRUE"
		}
		return "FALSE"
	}
	return fmt.Sprintf("%s", desired)
}
