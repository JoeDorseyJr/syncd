package brew

import (
	"fmt"
	"os/exec"
	"strings"
)

// CommandRunner abstracts command execution for testability.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

// ExecRunner executes real shell commands.
type ExecRunner struct{}

func (r *ExecRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Wrap with command context and captured output for diagnostics
		return out, &RunError{
			Cmd:    name + " " + strings.Join(args, " "),
			Output: strings.TrimSpace(string(out)),
			Err:    err,
		}
	}
	return out, nil
}

// RunError provides diagnostic context for failed commands.
type RunError struct {
	Cmd    string
	Output string
	Err    error
}

func (e *RunError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("%s: %s", e.Cmd, e.Output)
	}
	return fmt.Sprintf("%s: %v", e.Cmd, e.Err)
}

func (e *RunError) Unwrap() error { return e.Err }

// MockRunner records calls and returns preset output for tests.
type MockRunner struct {
	Outputs      []MockOutput
	Calls        []MockCall
	FailOnExtras bool // if true, return error when outputs are exhausted
	idx          int
}

type MockOutput struct {
	Out []byte
	Err error
}

type MockCall struct {
	Name string
	Args []string
}

func (m *MockRunner) Run(name string, args ...string) ([]byte, error) {
	m.Calls = append(m.Calls, MockCall{Name: name, Args: args})
	if m.idx >= len(m.Outputs) {
		if m.FailOnExtras {
			return nil, fmt.Errorf("unexpected call: %s %v", name, args)
		}
		return nil, nil
	}
	out := m.Outputs[m.idx]
	m.idx++
	return out.Out, out.Err
}
