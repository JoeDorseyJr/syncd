package download

import (
	"testing"

	"github.com/joedorseyjr/syncd/internal/runner"
)

func TestPlatform_ARM64Sonoma(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte("14.5\n")},
		},
	}
	old := platformFunc
	platformFunc = func(r runner.CommandRunner) string {
		// Simulate arm64 by calling detectPlatform logic directly
		codename := macOSCodename(r)
		return "arm64_" + codename
	}
	defer func() { platformFunc = old }()

	result := Platform(mock)
	if result != "arm64_sonoma" {
		t.Errorf("expected arm64_sonoma, got %s", result)
	}
}

func TestPlatform_IntelSonoma(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte("14.3\n")},
		},
	}
	old := platformFunc
	platformFunc = func(r runner.CommandRunner) string {
		codename := macOSCodename(r)
		return codename // simulate amd64
	}
	defer func() { platformFunc = old }()

	result := Platform(mock)
	if result != "sonoma" {
		t.Errorf("expected sonoma, got %s", result)
	}
}

func TestMacOSCodename(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{"15.1\n", "sequoia"},
		{"14.5\n", "sonoma"},
		{"13.2\n", "ventura"},
		{"12.7\n", "monterey"},
		{"16.0\n", "sequoia"}, // unknown → fallback
	}
	for _, tt := range tests {
		mock := &runner.MockRunner{
			Outputs: []runner.MockOutput{
				{Out: []byte(tt.version)},
			},
		}
		result := macOSCodename(mock)
		if result != tt.expected {
			t.Errorf("version %q: expected %s, got %s", tt.version, tt.expected, result)
		}
	}
}

func TestMacOSCodename_Error(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Err: &runner.RunError{Cmd: "sw_vers", Err: nil}},
		},
	}
	result := macOSCodename(mock)
	if result != "sequoia" {
		t.Errorf("expected sequoia fallback, got %s", result)
	}
}
