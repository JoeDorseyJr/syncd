package progress

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// DefaultHangTimeout is the production timeout for hang detection.
const DefaultHangTimeout = 60 * time.Second

// Result holds the outcome of a progress-tracked command.
type Result struct {
	Output []byte
	Err    error
	Hung   bool
}

// Process abstracts a running process for testability.
type Process interface {
	Wait() error
	Kill() error
}

// ProcessStarter creates a Process from a command name and args.
type ProcessStarter interface {
	Start(name string, args []string) (Process, io.ReadCloser, error)
}

// ExecStarter creates real OS processes.
type ExecStarter struct{}

func (s *ExecStarter) Start(name string, args []string) (Process, io.ReadCloser, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = brewEnv()

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return nil, nil, err
	}

	// Close write end after process exits so reader gets EOF
	go func() {
		cmd.Wait()
		pw.Close()
	}()

	return &execProcess{cmd: cmd}, pr, nil
}

type execProcess struct {
	cmd *exec.Cmd
}

func (p *execProcess) Wait() error { return p.cmd.Wait() }
func (p *execProcess) Kill() error { return p.cmd.Process.Kill() }

// brewEnv returns the current environment with HOMEBREW_NO_AUTO_UPDATE=1.
func brewEnv() []string {
	return append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")
}

// RunWithProgress executes a command, captures output line-by-line,
// calls onPhase for each detected phase, and kills on hang.
// Retries once if the process is killed due to timeout.
func RunWithProgress(starter ProcessStarter, name string, args []string, timeout time.Duration, onPhase func(string)) Result {
	onPhase("starting...")
	result := runOnce(starter, name, args, timeout, onPhase)
	if result.Hung {
		// Retry once
		onPhase("starting...")
		retry := runOnce(starter, name, args, timeout, onPhase)
		if retry.Hung || retry.Err != nil {
			retry.Hung = true
			return retry
		}
		return retry
	}
	return result
}

func runOnce(starter ProcessStarter, name string, args []string, timeout time.Duration, onPhase func(string)) Result {
	proc, reader, err := starter.Start(name, args)
	if err != nil {
		return Result{Err: err}
	}

	type lineResult struct {
		line string
		ok   bool
	}

	lineCh := make(chan lineResult, 1)
	var output []byte

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			lineCh <- lineResult{line: scanner.Text(), ok: true}
		}
		lineCh <- lineResult{ok: false}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case lr := <-lineCh:
			if !lr.ok {
				// EOF — process done writing
				err := proc.Wait()
				return Result{Output: output, Err: err}
			}
			output = append(output, lr.line...)
			output = append(output, '\n')
			if phase := DetectPhase(lr.line); phase != "" {
				onPhase(phase)
			}
			timer.Reset(timeout)

		case <-timer.C:
			// Hang detected — kill process
			proc.Kill()
			reader.Close()
			// Drain channel
			go func() {
				for lr := range lineCh {
					if !lr.ok {
						return
					}
				}
			}()
			return Result{Output: output, Hung: true, Err: fmt.Errorf("process hung (no output for %v)", timeout)}
		}
	}
}
