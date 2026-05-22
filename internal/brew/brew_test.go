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

func TestGetOutdated_ParsesOutput(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte(`{"formulae":[{"name":"node","installed_versions":["20.0.0"],"current_version":"22.0.0"},{"name":"wget","installed_versions":["1.21"],"current_version":"1.24"}],"casks":[]}`)},
		},
	}

	outdated, err := GetOutdated(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outdated) != 2 {
		t.Fatalf("expected 2 outdated, got %d", len(outdated))
	}
	if outdated[0].Name != "node" || outdated[0].Current != "20.0.0" || outdated[0].Latest != "22.0.0" {
		t.Errorf("unexpected outdated[0]: %+v", outdated[0])
	}
	if outdated[1].Name != "wget" || outdated[1].Current != "1.21" || outdated[1].Latest != "1.24" {
		t.Errorf("unexpected outdated[1]: %+v", outdated[1])
	}
}

func TestGetOutdated_EmptyOutput(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte(`{"formulae":[],"casks":[]}`)},
		},
	}

	outdated, err := GetOutdated(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outdated) != 0 {
		t.Errorf("expected 0 outdated, got %d", len(outdated))
	}
}

func TestGetOutdatedCasks_ParsesOutput(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte(`{"formulae":[],"casks":[{"name":"firefox","installed_versions":"120.0","current_version":"125.0"}]}`)},
		},
	}

	outdated, err := GetOutdatedCasks(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outdated) != 1 {
		t.Fatalf("expected 1 outdated cask, got %d", len(outdated))
	}
	if outdated[0].Name != "firefox" || outdated[0].Current != "120.0" || outdated[0].Latest != "125.0" {
		t.Errorf("unexpected outdated cask: %+v", outdated[0])
	}
}
