package defaults

import (
	"fmt"

	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/joedorseyjr/syncd/internal/runner"
)

// DriftEntry represents a single default that differs from desired state.
type DriftEntry struct {
	Domain  string
	Key     string
	Type    string
	Current string
	Desired string
	Kill    []string
}

// ComputeDrift compares declared defaults against current system state.
func ComputeDrift(r runner.CommandRunner, entries []config.DefaultEntry) ([]DriftEntry, error) {
	var drifted []DriftEntry
	for _, e := range entries {
		raw, found, err := ReadValue(r, e.Domain, e.Key)
		if err != nil {
			return nil, err
		}
		if !found {
			drifted = append(drifted, DriftEntry{
				Domain:  e.Domain,
				Key:     e.Key,
				Type:    e.Type,
				Current: "unset",
				Desired: fmt.Sprintf("%v", e.Value),
				Kill:    e.Kill,
			})
			continue
		}
		if !CompareValue(raw, e.Type, e.Value) {
			drifted = append(drifted, DriftEntry{
				Domain:  e.Domain,
				Key:     e.Key,
				Type:    e.Type,
				Current: raw,
				Desired: fmt.Sprintf("%v", e.Value),
				Kill:    e.Kill,
			})
		}
	}
	return drifted, nil
}
