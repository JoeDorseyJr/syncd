package brew

import (
	"errors"
	"reflect"
	"testing"

	"github.com/joedorseyjr/syncd/internal/plan"
)

func TestExecute_SuccessfulSequence(t *testing.T) {
	mock := &MockRunner{
		FailOnExtras: true,
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

	// Verify exact command calls
	wantCalls := []MockCall{
		{Name: "brew", Args: []string{"tap", "hashicorp/tap"}},
		{Name: "brew", Args: []string{"install", "go"}},
		{Name: "brew", Args: []string{"install", "--cask", "firefox"}},
	}
	assertCalls(t, mock.Calls, wantCalls)
}

func TestExecute_FullOrderWithRemovals(t *testing.T) {
	mock := &MockRunner{
		FailOnExtras: true,
		Outputs: []MockOutput{
			{Out: []byte("")}, // tap add
			{Out: []byte("")}, // brew install
			{Out: []byte("")}, // cask install
			{Out: []byte("")}, // brew uninstall
			{Out: []byte("")}, // cask uninstall
			{Out: []byte("")}, // untap (after package removals)
			{Out: []byte("")}, // autoremove
			{Out: []byte("")}, // cleanup
		},
	}

	p := &plan.Plan{
		TapsToAdd:      []string{"new/tap"},
		BrewsToInstall: []string{"git"},
		CasksToInstall: []string{"firefox"},
		BrewsToRemove:  []string{"wget"},
		CasksToRemove:  []string{"slack"},
		TapsToRemove:   []string{"old/tap"},
		Autoremove:     true,
		ClearCache:     true,
	}

	results := Execute(mock, p)
	if len(results) != 8 {
		t.Fatalf("expected 8 results, got %d", len(results))
	}

	// Verify order: taps → installs → cask installs → brew removals → cask removals → tap removals → autoremove → cleanup
	wantCalls := []MockCall{
		{Name: "brew", Args: []string{"tap", "new/tap"}},
		{Name: "brew", Args: []string{"install", "git"}},
		{Name: "brew", Args: []string{"install", "--cask", "firefox"}},
		{Name: "brew", Args: []string{"uninstall", "wget"}},
		{Name: "brew", Args: []string{"uninstall", "--cask", "slack"}},
		{Name: "brew", Args: []string{"untap", "old/tap"}},
		{Name: "brew", Args: []string{"autoremove"}},
		{Name: "brew", Args: []string{"cleanup"}},
	}
	assertCalls(t, mock.Calls, wantCalls)
}

func TestExecute_OneFailureDoesNotStopOthers(t *testing.T) {
	mock := &MockRunner{
		Outputs: []MockOutput{
			{Err: errors.New("tap failed")},
			{Out: []byte("")},
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

func TestExecute_TapRemovalAfterPackageRemovals(t *testing.T) {
	mock := &MockRunner{
		FailOnExtras: true,
		Outputs: []MockOutput{
			{Out: []byte("")}, // uninstall brew
			{Out: []byte("")}, // untap
		},
	}

	p := &plan.Plan{
		BrewsToRemove: []string{"pkg-from-tap"},
		TapsToRemove:  []string{"old/tap"},
	}

	results := Execute(mock, p)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify untap comes after uninstall
	wantCalls := []MockCall{
		{Name: "brew", Args: []string{"uninstall", "pkg-from-tap"}},
		{Name: "brew", Args: []string{"untap", "old/tap"}},
	}
	assertCalls(t, mock.Calls, wantCalls)
}

func TestExecute_AutoremoveCleanupOnlyWhenFlagged(t *testing.T) {
	mock := &MockRunner{FailOnExtras: true}

	p := &plan.Plan{
		Autoremove: false,
		ClearCache: false,
	}

	results := Execute(mock, p)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}

	mock2 := &MockRunner{
		FailOnExtras: true,
		Outputs: []MockOutput{
			{Out: []byte("")},
			{Out: []byte("")},
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

	wantCalls := []MockCall{
		{Name: "brew", Args: []string{"autoremove"}},
		{Name: "brew", Args: []string{"cleanup"}},
	}
	assertCalls(t, mock2.Calls, wantCalls)
}

func TestMockRunner_FailOnExtras(t *testing.T) {
	mock := &MockRunner{FailOnExtras: true}
	_, err := mock.Run("brew", "unexpected")
	if err == nil {
		t.Error("expected error for unexpected call with FailOnExtras")
	}
}

func assertCalls(t *testing.T, got, want []MockCall) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("call count: expected %d, got %d\ngot:  %v\nwant: %v", len(want), len(got), got, want)
		return
	}
	for i := range got {
		if got[i].Name != want[i].Name || !reflect.DeepEqual(got[i].Args, want[i].Args) {
			t.Errorf("call[%d]: expected %v %v, got %v %v", i, want[i].Name, want[i].Args, got[i].Name, got[i].Args)
		}
	}
}
