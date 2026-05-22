package defaults

import (
	"testing"

	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/joedorseyjr/syncd/internal/runner"
)

func TestComputeDrift_DriftedValue(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: []byte("64\n"), Err: nil}},
	}
	entries := []config.DefaultEntry{
		{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Value: 48, Kill: []string{"Dock"}},
	}
	drifted, err := ComputeDrift(mock, entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drifted) != 1 {
		t.Fatalf("expected 1 drift entry, got %d", len(drifted))
	}
	if drifted[0].Current != "64" || drifted[0].Desired != "48" {
		t.Fatalf("unexpected drift: %+v", drifted[0])
	}
}

func TestComputeDrift_UnsetKey(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: nil, Err: &runner.RunError{Cmd: "defaults read", Err: nil}}},
	}
	entries := []config.DefaultEntry{
		{Domain: "NSGlobalDomain", Key: "KeyRepeat", Type: "int", Value: 2},
	}
	drifted, err := ComputeDrift(mock, entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drifted) != 1 {
		t.Fatalf("expected 1 drift entry, got %d", len(drifted))
	}
	if drifted[0].Current != "unset" {
		t.Fatalf("expected 'unset', got '%s'", drifted[0].Current)
	}
}

func TestComputeDrift_MatchingValue(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{{Out: []byte("48\n"), Err: nil}},
	}
	entries := []config.DefaultEntry{
		{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Value: 48},
	}
	drifted, err := ComputeDrift(mock, entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drifted) != 0 {
		t.Fatalf("expected 0 drift entries, got %d", len(drifted))
	}
}

func TestComputeDrift_MixedEntries(t *testing.T) {
	mock := &runner.MockRunner{
		Outputs: []runner.MockOutput{
			{Out: []byte("48\n"), Err: nil},  // matches
			{Out: []byte("0\n"), Err: nil},   // drifted (want true=1)
			{Out: nil, Err: &runner.RunError{Cmd: "defaults read", Err: nil}}, // unset
		},
	}
	entries := []config.DefaultEntry{
		{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Value: 48},
		{Domain: "com.apple.dock", Key: "autohide", Type: "bool", Value: true},
		{Domain: "NSGlobalDomain", Key: "KeyRepeat", Type: "int", Value: 2},
	}
	drifted, err := ComputeDrift(mock, entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drifted) != 2 {
		t.Fatalf("expected 2 drift entries, got %d", len(drifted))
	}
}
