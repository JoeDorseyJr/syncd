package brew

import (
	"errors"
	"testing"

	"github.com/joedorseyjr/syncd/internal/plan"
)

func TestExecute_SuccessfulSequence(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte("")}, // tap
			{Out: []byte("")}, // install
			{Out: []byte("")}, // install --cask
		},
	}

	p := &plan.Plan{
		TapsToAdd:      []string{"hashicorp/tap"},
		BrewsToInstall: []string{"go"},
		CasksToInstall: []string{"firefox"},
	}

	results := Execute(mock, p)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s %s: %v", r.Action, r.Package, r.Err)
		}
	}

	// Verify execution order
	if mock.Calls[0].Args[0] != "tap" {
		t.Errorf("expected first call to be tap, got %v", mock.Calls[0].Args)
	}
	if mock.Calls[1].Args[0] != "install" {
		t.Errorf("expected second call to be install, got %v", mock.Calls[1].Args)
	}
	if mock.Calls[2].Args[0] != "install" && mock.Calls[2].Args[1] != "--cask" {
		t.Errorf("expected third call to be install --cask, got %v", mock.Calls[2].Args)
	}
}

func TestExecute_OneFailureDoesNotStopOthers(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Err: errors.New("tap failed")}, // tap fails
			{Out: []byte("")},               // install succeeds
		},
	}

	p := &plan.Plan{
		TapsToAdd:      []string{"bad/tap"},
		BrewsToInstall: []string{"go"},
	}

	results := Execute(mock, p)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("expected first result to have error")
	}
	if results[1].Err != nil {
		t.Errorf("expected second result to succeed, got: %v", results[1].Err)
	}
}

func TestExecute_TapRemoval(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte("")}, // untap
		},
	}

	p := &plan.Plan{
		TapsToRemove: []string{"old/tap"},
	}

	results := Execute(mock, p)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Action != "untap" {
		t.Errorf("expected action 'untap', got %q", results[0].Action)
	}
	if mock.Calls[0].Args[0] != "untap" {
		t.Errorf("expected brew untap call, got %v", mock.Calls[0].Args)
	}
}

func TestExecute_AutoremoveCleanupOnlyWhenFlagged(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{},
	}

	p := &plan.Plan{
		Autoremove: false,
		ClearCache: false,
	}

	results := Execute(mock, p)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}

	// Now with flags enabled
	mock2 := &MockRunner{
		Outputs: []MockOutput{
			{Out: []byte("")}, // autoremove
			{Out: []byte("")}, // cleanup
		},
	}

	p2 := &plan.Plan{
		Autoremove: true,
		ClearCache: true,
	}

	results2 := Execute(mock2, p2)
	if len(results2) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results2))
	}
	if results2[0].Action != "autoremove" {
		t.Errorf("expected 'autoremove', got %q", results2[0].Action)
	}
	if results2[1].Action != "cleanup" {
		t.Errorf("expected 'cleanup', got %q", results2[1].Action)
	}
}
