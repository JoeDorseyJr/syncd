package progress

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

// mockProcess simulates a process for testing.
type mockProcess struct {
	waitErr error
	killed  bool
	killCh  chan struct{}
}

func (p *mockProcess) Wait() error { return p.waitErr }
func (p *mockProcess) Kill() error {
	p.killed = true
	close(p.killCh)
	return nil
}

// mockStarter creates mock processes with configurable behavior.
type mockStarter struct {
	outputs  []string // output to produce per attempt
	waitErrs []error  // error to return from Wait per attempt
	hangs    []bool   // whether to hang (produce no output) per attempt
	attempt  int
}

func (s *mockStarter) Start(name string, args []string) (Process, io.ReadCloser, error) {
	idx := s.attempt
	s.attempt++

	proc := &mockProcess{killCh: make(chan struct{})}
	if idx < len(s.waitErrs) {
		proc.waitErr = s.waitErrs[idx]
	}

	pr, pw := io.Pipe()

	if idx < len(s.hangs) && s.hangs[idx] {
		// Hang: close pipe only when killed
		go func() {
			<-proc.killCh
			pw.Close()
		}()
	} else {
		// Normal: write output then close
		var output string
		if idx < len(s.outputs) {
			output = s.outputs[idx]
		}
		go func() {
			if output != "" {
				io.WriteString(pw, output)
			}
			pw.Close()
		}()
	}

	return proc, pr, nil
}

func TestRunWithProgress_Success(t *testing.T) {
	starter := &mockStarter{
		outputs:  []string{"Downloading https://example.com\nPouring neovim\n"},
		waitErrs: []error{nil},
	}

	var phases []string
	result := RunWithProgress(starter, "brew", []string{"upgrade", "neovim"}, time.Second, func(phase string) {
		phases = append(phases, phase)
	})

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Hung {
		t.Fatal("unexpected hung")
	}
	if !strings.Contains(string(result.Output), "Downloading") {
		t.Errorf("output missing 'Downloading': %s", result.Output)
	}
	// Should have: starting..., downloading, pouring
	expected := []string{"starting...", "downloading", "pouring"}
	if len(phases) != len(expected) {
		t.Fatalf("phases = %v, want %v", phases, expected)
	}
	for i, want := range expected {
		if phases[i] != want {
			t.Errorf("phases[%d] = %q, want %q", i, phases[i], want)
		}
	}
}

func TestRunWithProgress_Failure(t *testing.T) {
	starter := &mockStarter{
		outputs:  []string{"Error: something went wrong\n"},
		waitErrs: []error{fmt.Errorf("exit status 1")},
	}

	result := RunWithProgress(starter, "brew", []string{"upgrade", "bad"}, time.Second, func(string) {})

	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Hung {
		t.Fatal("should not be hung")
	}
	if !strings.Contains(string(result.Output), "Error:") {
		t.Errorf("output should contain error: %s", result.Output)
	}
}

func TestRunWithProgress_HungKilled(t *testing.T) {
	starter := &mockStarter{
		hangs:    []bool{true, true}, // both attempts hang
		waitErrs: []error{nil, nil},
	}

	result := RunWithProgress(starter, "brew", []string{"upgrade", "stuck"}, 50*time.Millisecond, func(string) {})

	if !result.Hung {
		t.Fatal("expected hung=true")
	}
	if result.Err == nil {
		t.Fatal("expected error for hung process")
	}
}

func TestRunWithProgress_RetrySucceeds(t *testing.T) {
	starter := &mockStarter{
		hangs:    []bool{true, false},
		outputs:  []string{"", "Installing neovim\n"},
		waitErrs: []error{nil, nil},
	}

	var phases []string
	result := RunWithProgress(starter, "brew", []string{"upgrade", "neovim"}, 50*time.Millisecond, func(phase string) {
		phases = append(phases, phase)
	})

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Hung {
		t.Fatal("should not be hung after successful retry")
	}
	if !strings.Contains(string(result.Output), "Installing") {
		t.Errorf("output should contain retry output: %s", result.Output)
	}
}

func TestRunWithProgress_CallbackCalledWithPhases(t *testing.T) {
	starter := &mockStarter{
		outputs:  []string{"Downloading foo\nPouring foo\nInstalling foo\n"},
		waitErrs: []error{nil},
	}

	var phases []string
	RunWithProgress(starter, "brew", []string{"upgrade", "foo"}, time.Second, func(phase string) {
		phases = append(phases, phase)
	})

	expected := []string{"starting...", "downloading", "pouring", "installing"}
	if len(phases) != len(expected) {
		t.Fatalf("phases = %v, want %v", phases, expected)
	}
	for i, want := range expected {
		if phases[i] != want {
			t.Errorf("phases[%d] = %q, want %q", i, phases[i], want)
		}
	}
}
