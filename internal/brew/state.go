package brew

import "strings"

// State represents the current Homebrew installation state.
type State struct {
	Taps  []string
	Brews []string
	Casks []string
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
