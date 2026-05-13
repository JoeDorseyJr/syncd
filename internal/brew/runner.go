package brew

import "os/exec"

// CommandRunner abstracts command execution for testability.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

// ExecRunner executes real shell commands.
type ExecRunner struct{}

func (r *ExecRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// MockRunner records calls and returns preset output for tests.
type MockRunner struct {
	Outputs []MockOutput
	Calls   []MockCall
	idx     int
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
		return nil, nil
	}
	out := m.Outputs[m.idx]
	m.idx++
	return out.Out, out.Err
}
