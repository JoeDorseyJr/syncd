package brew

import (
	"encoding/json"
	"strings"
)

// State represents the current Homebrew installation state.
type State struct {
	Taps  []string
	Brews []string
	Casks []string
}

// OutdatedPkg represents a package with an available upgrade.
type OutdatedPkg struct {
	Name    string
	Current string
	Latest  string
}

// GetState queries Homebrew for installed taps, brews, and casks.
func GetState(runner CommandRunner) (*State, error) {
	taps, err := runAndParse(runner, "brew", "tap")
	if err != nil {
		return nil, err
	}

	brews, err := runAndParse(runner, "brew", "list", "--formula", "-1")
	if err != nil {
		return nil, err
	}

	casks, err := runAndParse(runner, "brew", "list", "--cask", "-1")
	if err != nil {
		return nil, err
	}

	return &State{Taps: taps, Brews: brews, Casks: casks}, nil
}

// GetLeaves returns explicitly-installed formulae (not auto-deps).
func GetLeaves(runner CommandRunner) ([]string, error) {
	return runAndParse(runner, "brew", "leaves")
}

// GetOutdated returns formulae with available upgrades.
func GetOutdated(runner CommandRunner) ([]OutdatedPkg, error) {
	return getOutdatedJSON(runner, "brew", "outdated", "--formula", "--json=v2")
}

// GetOutdatedCasks returns casks with available upgrades.
func GetOutdatedCasks(runner CommandRunner) ([]OutdatedPkg, error) {
	return getOutdatedJSON(runner, "brew", "outdated", "--cask", "--greedy", "--json=v2")
}

func getOutdatedJSON(runner CommandRunner, name string, args ...string) ([]OutdatedPkg, error) {
	out, err := runner.Run(name, args...)
	if err != nil {
		return nil, err
	}

	var result struct {
		Formulae []struct {
			Name             string `json:"name"`
			InstalledVersions []string `json:"installed_versions"`
			CurrentVersion   string `json:"current_version"`
		} `json:"formulae"`
		Casks []struct {
			Name             string `json:"name"`
			InstalledVersions string `json:"installed_versions"`
			CurrentVersion   string `json:"current_version"`
		} `json:"casks"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}

	var pkgs []OutdatedPkg
	for _, f := range result.Formulae {
		current := ""
		if len(f.InstalledVersions) > 0 {
			current = f.InstalledVersions[len(f.InstalledVersions)-1]
		}
		pkgs = append(pkgs, OutdatedPkg{Name: f.Name, Current: current, Latest: f.CurrentVersion})
	}
	for _, c := range result.Casks {
		pkgs = append(pkgs, OutdatedPkg{Name: c.Name, Current: c.InstalledVersions, Latest: c.CurrentVersion})
	}
	return pkgs, nil
}

func runAndParse(runner CommandRunner, name string, args ...string) ([]string, error) {
	out, err := runner.Run(name, args...)
	if err != nil {
		return nil, err
	}
	return parseLines(string(out)), nil
}

func parseLines(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}
	return strings.Split(output, "\n")
}
