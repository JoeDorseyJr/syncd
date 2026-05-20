package brew

import (
	"errors"
	"testing"
)

func TestGetState_ParsesOutput(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte("homebrew/core\nhomebrew/cask\n")},
			{Out: []byte("git\ngo\nwget\n")},
			{Out: []byte("firefox\niterm2\n")},
		},
	}

	state, err := GetState(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state.Taps) != 2 {
		t.Errorf("expected 2 taps, got %d", len(state.Taps))
	}
	if len(state.Brews) != 3 {
		t.Errorf("expected 3 brews, got %d", len(state.Brews))
	}
	if len(state.Casks) != 2 {
		t.Errorf("expected 2 casks, got %d", len(state.Casks))
	}
}

func TestGetState_EmptyOutput(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte("")},
			{Out: []byte("")},
			{Out: []byte("")},
		},
	}

	state, err := GetState(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state.Taps) != 0 {
		t.Errorf("expected 0 taps, got %d", len(state.Taps))
	}
	if len(state.Brews) != 0 {
		t.Errorf("expected 0 brews, got %d", len(state.Brews))
	}
	if len(state.Casks) != 0 {
		t.Errorf("expected 0 casks, got %d", len(state.Casks))
	}
}

func TestGetState_CommandFailure(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Err: errors.New("brew not found")},
		},
	}

	_, err := GetState(mock)
	if err == nil {
		t.Fatal("expected error on command failure")
	}
}

func TestGetLeaves_ParsesOutput(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte("git\nwget\n")},
		},
	}

	leaves, err := GetLeaves(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(leaves) != 2 {
		t.Errorf("expected 2 leaves, got %d", len(leaves))
	}
	if leaves[0] != "git" || leaves[1] != "wget" {
		t.Errorf("unexpected leaves: %v", leaves)
	}
}

func TestGetLeaves_EmptyOutput(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte("")},
		},
	}

	leaves, err := GetLeaves(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(leaves) != 0 {
		t.Errorf("expected 0 leaves, got %d", len(leaves))
	}
}

func TestGetLeaves_CommandFailure(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Err: errors.New("brew leaves failed")},
		},
	}

	_, err := GetLeaves(mock)
	if err == nil {
		t.Fatal("expected error on command failure")
	}
}
