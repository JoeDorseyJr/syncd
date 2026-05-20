//go:build integration

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Integration tests use a fake `brew` script on PATH to avoid
// destructive operations against the real Homebrew installation.
// Set SYNCD_INTEGRATION=1 to run these tests.

var binary string
var fakeBrew string

func TestMain(m *testing.M) {
	if os.Getenv("SYNCD_INTEGRATION") != "1" {
		// Skip silently if not explicitly opted in
		os.Exit(0)
	}

	dir, err := os.MkdirTemp("", "syncd-test")
	if err != nil {
		panic(err)
	}

	// Build syncd binary
	binary = filepath.Join(dir, "syncd")
	cmd := exec.Command("go", "build", "-o", binary, "../cmd/syncd")
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}

	// Create fake brew script
	fakeBrew = filepath.Join(dir, "brew")
	if err := createFakeBrew(fakeBrew); err != nil {
		panic("creating fake brew: " + err.Error())
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func createFakeBrew(path string) error {
	// Fake brew that tracks state in a temp dir via env var FAKE_BREW_STATE
	script := `#!/bin/bash
set -e
STATE_DIR="${FAKE_BREW_STATE:-/tmp/fake-brew-state}"
mkdir -p "$STATE_DIR"

case "$1" in
  tap)
    if [ -z "$2" ]; then
      # List taps
      cat "$STATE_DIR/taps" 2>/dev/null || true
    else
      # Add tap
      echo "$2" >> "$STATE_DIR/taps"
    fi
    ;;
  untap)
    if [ -f "$STATE_DIR/taps" ]; then
      grep -v "^$2$" "$STATE_DIR/taps" > "$STATE_DIR/taps.tmp" || true
      mv "$STATE_DIR/taps.tmp" "$STATE_DIR/taps"
    fi
    ;;
  leaves)
    cat "$STATE_DIR/leaves" 2>/dev/null || true
    ;;
  list)
    if [[ "$*" == *"--cask"* ]]; then
      cat "$STATE_DIR/casks" 2>/dev/null || true
    else
      cat "$STATE_DIR/brews" 2>/dev/null || true
    fi
    ;;
  install)
    if [[ "$2" == "--cask" ]]; then
      if [[ "$3" == *"nonexistent"* ]]; then
        echo "Error: No available formula or cask with the name \"$3\"" >&2
        exit 1
      fi
      echo "$3" >> "$STATE_DIR/casks"
    else
      if [[ "$2" == *"nonexistent"* ]]; then
        echo "Error: No available formula or cask with the name \"$2\"" >&2
        exit 1
      fi
      echo "$2" >> "$STATE_DIR/brews"
    fi
    ;;
  uninstall)
    if [[ "$2" == "--cask" ]]; then
      if [ -f "$STATE_DIR/casks" ]; then
        grep -v "^$3$" "$STATE_DIR/casks" > "$STATE_DIR/casks.tmp" || true
        mv "$STATE_DIR/casks.tmp" "$STATE_DIR/casks"
      fi
    else
      if [ -f "$STATE_DIR/brews" ]; then
        grep -v "^$2$" "$STATE_DIR/brews" > "$STATE_DIR/brews.tmp" || true
        mv "$STATE_DIR/brews.tmp" "$STATE_DIR/brews"
      fi
    fi
    ;;
  autoremove|cleanup)
    # No-op for fake
    ;;
  --version)
    echo "Homebrew 4.0.0 (fake)"
    ;;
  *)
    echo "fake brew: unknown command $1" >&2
    exit 1
    ;;
esac
`
	return os.WriteFile(path, []byte(script), 0755)
}

func setupFakeState(t *testing.T, taps, brews, casks []string) string {
	return setupFakeStateWithLeaves(t, taps, brews, brews, casks)
}

func setupFakeStateWithLeaves(t *testing.T, taps, brews, leaves, casks []string) string {
	t.Helper()
	dir := t.TempDir()
	if len(taps) > 0 {
		os.WriteFile(filepath.Join(dir, "taps"), []byte(strings.Join(taps, "\n")+"\n"), 0644)
	}
	if len(brews) > 0 {
		os.WriteFile(filepath.Join(dir, "brews"), []byte(strings.Join(brews, "\n")+"\n"), 0644)
	}
	if len(leaves) > 0 {
		os.WriteFile(filepath.Join(dir, "leaves"), []byte(strings.Join(leaves, "\n")+"\n"), 0644)
	}
	if len(casks) > 0 {
		os.WriteFile(filepath.Join(dir, "casks"), []byte(strings.Join(casks, "\n")+"\n"), 0644)
	}
	return dir
}

func TestPlan_DoesNotModifyState(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"git", "wget"}, nil)
	cfg := writeConfig(t, `
brews:
  - git
  - cowsay
`)
	out, code := runSyncdExpect(t, stateDir, 2, "plan", "--config", cfg)
	if !strings.Contains(out, "cowsay") {
		t.Errorf("expected plan to show cowsay, got: %s", out)
	}

	// Verify state unchanged
	brews := readFakeState(t, stateDir, "brews")
	if !strings.Contains(brews, "wget") {
		t.Error("plan should not have modified state")
	}
	_ = code
}

func TestPlan_ExitZeroWhenSynced(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"git"}, nil)
	cfg := writeConfig(t, `
brews:
  - git
`)
	runSyncdExpect(t, stateDir, 0, "plan", "--config", cfg)
}

func TestPlan_CleanupOnlyExitsZero(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"git"}, nil)
	cfg := writeConfig(t, `
brews:
  - git
cleanup:
  autoremove: true
  clear_cache: true
`)
	// Cleanup-only plan should exit 0 (not drift)
	runSyncdExpect(t, stateDir, 0, "plan", "--config", cfg)
}

func TestApply_InstallsPackage(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - cowsay
`)
	out, _ := runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--config", cfg)
	if !strings.Contains(out, "install cowsay") {
		t.Errorf("expected install cowsay in output, got: %s", out)
	}

	brews := readFakeState(t, stateDir, "brews")
	if !strings.Contains(brews, "cowsay") {
		t.Error("cowsay should be in fake state after apply")
	}
}

func TestApply_RemovesUndeclaredPackage(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"git", "cowsay"}, nil)
	cfg := writeConfig(t, `
brews:
  - git
cleanup:
  remove_unlisted: true
`)
	runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--config", cfg)

	brews := readFakeState(t, stateDir, "brews")
	if strings.Contains(brews, "cowsay") {
		t.Error("cowsay should have been removed")
	}
	if !strings.Contains(brews, "git") {
		t.Error("git should remain")
	}
}

func TestApply_DependencyOnlyNotRemoved(t *testing.T) {
	// libssh2 is installed but NOT a leaf (it's a dependency)
	stateDir := setupFakeStateWithLeaves(t, nil,
		[]string{"git", "libssh2"},
		[]string{"git"},
		nil,
	)
	cfg := writeConfig(t, `
brews:
  - git
cleanup:
  remove_unlisted: true
`)
	runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--config", cfg)

	brews := readFakeState(t, stateDir, "brews")
	if !strings.Contains(brews, "libssh2") {
		t.Error("libssh2 (dependency-only) should NOT have been removed")
	}
}

func TestPlan_LeafNotInConfigIsRemoved(t *testing.T) {
	stateDir := setupFakeStateWithLeaves(t, nil,
		[]string{"git", "wget", "libssh2"},
		[]string{"git", "wget"},
		nil,
	)
	cfg := writeConfig(t, `
brews:
  - git
cleanup:
  remove_unlisted: true
`)
	out, _ := runSyncdExpect(t, stateDir, 2, "plan", "--config", cfg)
	if !strings.Contains(out, "wget") {
		t.Errorf("expected wget in removal plan, got: %s", out)
	}
	if strings.Contains(out, "libssh2") {
		t.Errorf("libssh2 (dep-only) should NOT be in removal plan, got: %s", out)
	}
}

func TestApply_NoRemovalWhenDisabled(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"git", "cowsay"}, nil)
	cfg := writeConfig(t, `
brews:
  - git
cleanup:
  remove_unlisted: false
`)
	runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--config", cfg)

	brews := readFakeState(t, stateDir, "brews")
	if !strings.Contains(brews, "cowsay") {
		t.Error("cowsay should remain when remove_unlisted is false")
	}
}

func TestApply_Idempotent(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - cowsay
`)
	runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--config", cfg)

	// Second plan should show no changes
	runSyncdExpect(t, stateDir, 0, "plan", "--config", cfg)
}

func TestApply_NonexistentPackageReportsError(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - syncd-test-nonexistent-pkg
  - cowsay
`)
	// Should exit 1 due to partial failure
	out, _ := runSyncdExpect(t, stateDir, 1, "apply", "--yes", "--config", cfg)
	if !strings.Contains(out, "nonexistent") {
		t.Errorf("expected error about nonexistent package, got: %s", out)
	}

	// cowsay should still be installed
	brews := readFakeState(t, stateDir, "brews")
	if !strings.Contains(brews, "cowsay") {
		t.Error("cowsay should be installed despite other package failing")
	}
}

func TestApply_TapRemovalAfterPackageRemoval(t *testing.T) {
	stateDir := setupFakeState(t, []string{"custom/tap"}, []string{"custom-pkg"}, nil)
	cfg := writeConfig(t, `
taps: []
brews: []
cleanup:
  remove_unlisted: true
`)
	runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--config", cfg)

	taps := readFakeState(t, stateDir, "taps")
	if strings.Contains(taps, "custom/tap") {
		t.Error("custom/tap should have been removed")
	}
	brews := readFakeState(t, stateDir, "brews")
	if strings.Contains(brews, "custom-pkg") {
		t.Error("custom-pkg should have been removed")
	}
}

func TestApply_CancelDoesNotExecute(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - cowsay
`)
	// Pipe "n" to stdin to cancel
	cmd := exec.Command(binary, "apply", "--config", cfg)
	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Dir(fakeBrew)+":"+os.Getenv("PATH"),
		"FAKE_BREW_STATE="+stateDir,
	)
	cmd.Stdin = strings.NewReader("n\n")
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
	}
	if code != 0 {
		t.Errorf("expected exit 0 on cancel, got %d", code)
	}
	if !strings.Contains(string(out), "Cancelled") {
		t.Errorf("expected 'Cancelled' in output, got: %s", string(out))
	}
	// Verify nothing was installed
	brews := readFakeState(t, stateDir, "brews")
	if strings.Contains(brews, "cowsay") {
		t.Error("cowsay should NOT be installed after cancel")
	}
}

func TestApply_ConfirmExecutes(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - cowsay
`)
	cmd := exec.Command(binary, "apply", "--config", cfg)
	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Dir(fakeBrew)+":"+os.Getenv("PATH"),
		"FAKE_BREW_STATE="+stateDir,
	)
	cmd.Stdin = strings.NewReader("y\n")
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
	}
	if code != 0 {
		t.Errorf("expected exit 0 on confirm, got %d: %s", code, string(out))
	}
	brews := readFakeState(t, stateDir, "brews")
	if !strings.Contains(brews, "cowsay") {
		t.Error("cowsay should be installed after confirm")
	}
}

func TestPlan_ExitTwoOnDrift(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - newpkg
`)
	out, _ := runSyncdExpect(t, stateDir, 2, "plan", "--config", cfg)
	if !strings.Contains(out, "newpkg") {
		t.Errorf("expected plan to show newpkg, got: %s", out)
	}
}

func TestApply_FailureExitOne(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - syncd-test-nonexistent-pkg
`)
	runSyncdExpect(t, stateDir, 1, "apply", "--yes", "--config", cfg)
}

func TestPlan_MissingBrewShowsGuidance(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - git
`)
	// Use empty PATH so brew is not found
	cmd := exec.Command(binary, "plan", "--config", cfg)
	cmd.Env = []string{
		"PATH=/nonexistent",
		"HOME=" + os.Getenv("HOME"),
		"FAKE_BREW_STATE=" + stateDir,
	}
	out, _ := cmd.CombinedOutput()
	if !strings.Contains(string(out), "https://brew.sh") {
		t.Errorf("expected install guidance with brew.sh URL, got: %s", string(out))
	}
}

func TestApply_IdempotentWithCleanup(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - cowsay
cleanup:
  autoremove: true
  clear_cache: true
`)
	runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--config", cfg)

	// Second plan should exit 0 (cleanup-only is not drift)
	runSyncdExpect(t, stateDir, 0, "plan", "--config", cfg)
}

// Helpers

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func runSyncdExpect(t *testing.T, stateDir string, wantExit int, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	// Put fake brew first on PATH
	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Dir(fakeBrew)+":"+os.Getenv("PATH"),
		"FAKE_BREW_STATE="+stateDir,
	)
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run syncd: %v", err)
		}
	}
	if code != wantExit {
		t.Errorf("expected exit %d, got %d\noutput: %s", wantExit, code, string(out))
	}
	return string(out), code
}

func readFakeState(t *testing.T, stateDir, file string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(stateDir, file))
	if err != nil {
		return ""
	}
	return string(data)
}
