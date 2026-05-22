package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandRunner abstracts command execution for testability.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
	RunMutate(name string, args ...string) ([]byte, error)
	RunSilent(name string, args ...string) ([]byte, error)
}

// Verbose controls whether mutating commands stream output to stdout/stderr.
var Verbose bool

// ExecRunner executes real shell commands.
type ExecRunner struct{}

func (r *ExecRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = brewEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, &RunError{
			Cmd:    name + " " + strings.Join(args, " "),
			Output: strings.TrimSpace(string(out)),
			Err:    err,
		}
	}
	return out, nil
}

func (r *ExecRunner) RunMutate(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = brewEnv()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, &RunError{
			Cmd: name + " " + strings.Join(args, " "),
			Err: err,
		}
	}
	return nil, nil
}

// RunSilent captures stdout/stderr but connects stdin for password prompts.
func (r *ExecRunner) RunSilent(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = brewEnv()
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, &RunError{
			Cmd:    name + " " + strings.Join(args, " "),
			Output: strings.TrimSpace(string(out)),
			Err:    err,
		}
	}
	return out, nil
}

// brewEnv returns the current environment with HOMEBREW_NO_AUTO_UPDATE=1
// to prevent brew from polluting stdout with update messages.
func brewEnv() []string {
	env := os.Environ()
	return append(env, "HOMEBREW_NO_AUTO_UPDATE=1")
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
	FailOnExtras bool
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

func (m *MockRunner) RunMutate(name string, args ...string) ([]byte, error) {
	return m.Run(name, args...)
}

func (m *MockRunner) RunSilent(name string, args ...string) ([]byte, error) {
	return m.Run(name, args...)
}
