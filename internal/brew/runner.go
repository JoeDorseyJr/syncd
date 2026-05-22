package brew

import "github.com/joedorseyjr/syncd/internal/runner"

// CommandRunner abstracts command execution for testability.
type CommandRunner = runner.CommandRunner

// ExecRunner executes real shell commands.
type ExecRunner = runner.ExecRunner

// RunError provides diagnostic context for failed commands.
type RunError = runner.RunError

// MockRunner records calls and returns preset output for tests.
type MockRunner = runner.MockRunner

// MockOutput holds preset output for MockRunner.
type MockOutput = runner.MockOutput

// MockCall records a single command invocation.
type MockCall = runner.MockCall
