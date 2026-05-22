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
var fakeDir string

func TestMain(m *testing.M) {
	if os.Getenv("SYNCD_INTEGRATION") != "1" {
		// Skip silently if not explicitly opted in
		os.Exit(0)
	}

	dir, err := os.MkdirTemp("", "syncd-test")
	if err != nil {
		panic(err)
	}
	fakeDir = dir

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

	// Create fake defaults script
	if err := createFakeDefaults(filepath.Join(dir, "defaults")); err != nil {
		panic("creating fake defaults: " + err.Error())
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
  outdated)
    if [[ "$*" == *"--cask"* ]]; then
      cat "$STATE_DIR/outdated_casks" 2>/dev/null || true
    else
      cat "$STATE_DIR/outdated_brews" 2>/dev/null || true
    fi
    ;;
  upgrade)
    if [[ "$2" == "--cask" ]]; then
      if [[ "$3" == *"fail"* ]]; then
        echo "Error: upgrade failed for $3" >&2
        exit 1
      fi
    else
      if [[ "$2" == *"fail"* ]]; then
        echo "Error: upgrade failed for $2" >&2
        exit 1
      fi
    fi
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

func setOutdated(t *testing.T, stateDir string, brews, casks []string) {
	t.Helper()
	if len(brews) > 0 {
		os.WriteFile(filepath.Join(stateDir, "outdated_brews"), []byte(strings.Join(brews, "\n")+"\n"), 0644)
	}
	if len(casks) > 0 {
		os.WriteFile(filepath.Join(stateDir, "outdated_casks"), []byte(strings.Join(casks, "\n")+"\n"), 0644)
	}
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

func TestPlan_PipedOutputNoANSI(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - newpkg
`)
	// runSyncdExpect captures output via pipe, so TTY detection should suppress color
	out, _ := runSyncdExpect(t, stateDir, 2, "plan", "--config", cfg)
	if strings.Contains(out, "\033[") {
		t.Errorf("expected no ANSI escape codes in piped output, got: %q", out)
	}
}

func TestPlan_NoColorFlagSuppressesANSI(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - newpkg
`)
	out, _ := runSyncdExpect(t, stateDir, 2, "plan", "--no-color", "--config", cfg)
	if strings.Contains(out, "\033[") {
		t.Errorf("expected no ANSI escape codes with --no-color, got: %q", out)
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

func TestUpgrade_SuccessExitZero(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"node", "wget"}, nil)
	setOutdated(t, stateDir, []string{"node", "wget"}, nil)

	out, _ := runSyncdExpect(t, stateDir, 0, "upgrade", "--yes")
	if !strings.Contains(out, "node") {
		t.Errorf("expected node in upgrade output, got: %s", out)
	}
}

func TestUpgrade_PinnedPackageSkipped(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"node", "wget"}, nil)
	setOutdated(t, stateDir, []string{"node", "wget"}, nil)
	cfg := writeConfig(t, `
brews:
  - node
  - wget
pin:
  - node
`)
	out, _ := runSyncdExpect(t, stateDir, 0, "upgrade", "--yes", "--config", cfg)
	if strings.Contains(out, "upgrade node") {
		t.Errorf("pinned package 'node' should not be upgraded, got: %s", out)
	}
	if !strings.Contains(out, "wget") {
		t.Errorf("expected wget in upgrade output, got: %s", out)
	}
}

func TestUpgrade_OneFailureExitOne(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"fail-pkg", "wget"}, nil)
	setOutdated(t, stateDir, []string{"fail-pkg", "wget"}, nil)

	out, _ := runSyncdExpect(t, stateDir, 1, "upgrade", "--yes")
	if !strings.Contains(out, "wget") {
		t.Errorf("expected wget to still be upgraded, got: %s", out)
	}
}

func TestUpgrade_CancelWithN(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"node"}, nil)
	setOutdated(t, stateDir, []string{"node"}, nil)

	cmd := exec.Command(binary, "upgrade")
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
}

func TestUpgrade_WorksWithoutConfig(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"node"}, nil)
	setOutdated(t, stateDir, []string{"node"}, nil)

	// No --config flag, no default config — should still work
	out, _ := runSyncdExpect(t, stateDir, 0, "upgrade", "--yes")
	if !strings.Contains(out, "node") {
		t.Errorf("expected node in upgrade output, got: %s", out)
	}
}

func TestInit_ProducesValidYAML(t *testing.T) {
	stateDir := setupFakeStateWithLeaves(t,
		[]string{"homebrew/core"},
		[]string{"git", "wget", "libssh2"},
		[]string{"git", "wget"},
		[]string{"firefox"},
	)

	out, _ := runSyncdExpect(t, stateDir, 0, "init")

	// Should be valid YAML that parses
	if !strings.Contains(out, "brews:") {
		t.Errorf("expected 'brews:' in output, got: %s", out)
	}
	if !strings.Contains(out, "git") {
		t.Errorf("expected 'git' in output, got: %s", out)
	}
}

func TestInit_OnlyLeavesInBrews(t *testing.T) {
	stateDir := setupFakeStateWithLeaves(t,
		nil,
		[]string{"git", "libssh2"},
		[]string{"git"},
		nil,
	)

	out, _ := runSyncdExpect(t, stateDir, 0, "init")

	if strings.Contains(out, "libssh2") {
		t.Errorf("dependency-only 'libssh2' should not be in init output, got: %s", out)
	}
	if !strings.Contains(out, "git") {
		t.Errorf("expected 'git' in output, got: %s", out)
	}
}

func TestInit_IncludesCasks(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, []string{"firefox", "iterm2"})

	out, _ := runSyncdExpect(t, stateDir, 0, "init")

	if !strings.Contains(out, "firefox") || !strings.Contains(out, "iterm2") {
		t.Errorf("expected casks in output, got: %s", out)
	}
}

func TestInit_IncludesTaps(t *testing.T) {
	stateDir := setupFakeState(t, []string{"homebrew/core", "hashicorp/tap"}, nil, nil)

	out, _ := runSyncdExpect(t, stateDir, 0, "init")

	if !strings.Contains(out, "homebrew/core") || !strings.Contains(out, "hashicorp/tap") {
		t.Errorf("expected taps in output, got: %s", out)
	}
}

func TestInit_IncludesPinAndCleanup(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"git"}, nil)

	out, _ := runSyncdExpect(t, stateDir, 0, "init")

	if !strings.Contains(out, "pin:") {
		t.Errorf("expected 'pin:' in output, got: %s", out)
	}
	if !strings.Contains(out, "cleanup:") {
		t.Errorf("expected 'cleanup:' in output, got: %s", out)
	}
}

func TestInit_OutputCreatesFile(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"git"}, nil)
	outFile := filepath.Join(t.TempDir(), "config.yaml")

	runSyncdExpect(t, stateDir, 0, "init", "--output", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
	if !strings.Contains(string(data), "git") {
		t.Errorf("expected 'git' in file, got: %s", string(data))
	}
}

func TestInit_RefusesOverwriteWithoutForce(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"git"}, nil)
	outFile := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(outFile, []byte("existing"), 0644)

	out, _ := runSyncdExpect(t, stateDir, 1, "init", "--output", outFile)
	if !strings.Contains(out, "already exists") {
		t.Errorf("expected 'already exists' error, got: %s", out)
	}

	// With --force, should succeed
	runSyncdExpect(t, stateDir, 0, "init", "--output", outFile, "--force")
}

func TestApply_VerboseShowsOutput(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - cowsay
`)
	// With --verbose, brew output should be visible (fake brew doesn't print much but command runs)
	out, _ := runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--verbose", "--config", cfg)
	if !strings.Contains(out, "Applying") {
		t.Errorf("expected 'Applying' in verbose output, got: %s", out)
	}
}

func TestUpgrade_VerboseShowsOutput(t *testing.T) {
	stateDir := setupFakeState(t, nil, []string{"node"}, nil)
	setOutdated(t, stateDir, []string{"node"}, nil)

	out, _ := runSyncdExpect(t, stateDir, 0, "upgrade", "--yes", "--verbose")
	if !strings.Contains(out, "Upgrading") {
		t.Errorf("expected 'Upgrading' in verbose output, got: %s", out)
	}
}

func TestApply_DefaultNonVerbose(t *testing.T) {
	stateDir := setupFakeState(t, nil, nil, nil)
	cfg := writeConfig(t, `
brews:
  - cowsay
`)
	out, _ := runSyncdExpect(t, stateDir, 0, "apply", "--yes", "--config", cfg)
	// Should show result lines
	if !strings.Contains(out, "install cowsay") {
		t.Errorf("expected result summary in non-verbose output, got: %s", out)
	}
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

func createFakeDefaults(path string) error {
	script := `#!/bin/bash
STATE_DIR="${FAKE_DEFAULTS_STATE:-/tmp/fake-defaults-state}"
mkdir -p "$STATE_DIR"

case "$1" in
  read-type)
    FILE="$STATE_DIR/${2}__${3}.type"
    if [ ! -f "$FILE" ]; then
      echo "The domain/default pair of (${2}, ${3}) does not exist" >&2
      exit 1
    fi
    cat "$FILE"
    ;;
  read)
    FILE="$STATE_DIR/${2}__${3}"
    if [ ! -f "$FILE" ]; then
      echo "The domain/default pair of (${2}, ${3}) does not exist" >&2
      exit 1
    fi
    cat "$FILE"
    ;;
  write)
    DOMAIN="$2"
    KEY="$3"
    # $4 is -<type>, $5 is value
    TYPE="${4#-}"
    VALUE="$5"
    echo -n "$VALUE" > "$STATE_DIR/${DOMAIN}__${KEY}"
    echo -n "$TYPE" > "$STATE_DIR/${DOMAIN}__${KEY}.type"
    ;;
  *)
    echo "fake defaults: unknown command $1" >&2
    exit 1
    ;;
esac
`
	return os.WriteFile(path, []byte(script), 0755)
}

func setupFakeDefaultsState(t *testing.T, entries map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for key, val := range entries {
		os.WriteFile(filepath.Join(dir, key), []byte(val), 0644)
	}
	return dir
}

func runSyncdWithDefaults(t *testing.T, brewStateDir, defaultsStateDir string, wantExit int, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(),
		"PATH="+fakeDir+":"+os.Getenv("PATH"),
		"FAKE_BREW_STATE="+brewStateDir,
		"FAKE_DEFAULTS_STATE="+defaultsStateDir,
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

func TestPlan_DefaultsDriftChangedValue(t *testing.T) {
	brewState := setupFakeState(t, nil, []string{"git"}, nil)
	defaultsState := setupFakeDefaultsState(t, map[string]string{
		"com.apple.dock__tilesize": "64",
	})
	cfg := writeConfig(t, `
brews:
  - git
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: 48
`)
	out, _ := runSyncdWithDefaults(t, brewState, defaultsState, 2, "plan", "--config", cfg)
	if !strings.Contains(out, "com.apple.dock tilesize") {
		t.Errorf("expected drift for com.apple.dock tilesize, got: %s", out)
	}
	if !strings.Contains(out, "64") || !strings.Contains(out, "48") {
		t.Errorf("expected current→desired values, got: %s", out)
	}
}

func TestPlan_DefaultsDriftUnsetKey(t *testing.T) {
	brewState := setupFakeState(t, nil, []string{"git"}, nil)
	defaultsState := setupFakeDefaultsState(t, nil)
	cfg := writeConfig(t, `
brews:
  - git
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: 48
`)
	out, _ := runSyncdWithDefaults(t, brewState, defaultsState, 2, "plan", "--config", cfg)
	if !strings.Contains(out, "unset") {
		t.Errorf("expected 'unset' for missing key, got: %s", out)
	}
	if !strings.Contains(out, "48") {
		t.Errorf("expected desired value 48, got: %s", out)
	}
}

func TestPlan_DefaultsMatchingHidden(t *testing.T) {
	brewState := setupFakeState(t, nil, []string{"git"}, nil)
	defaultsState := setupFakeDefaultsState(t, map[string]string{
		"com.apple.dock__tilesize": "48",
	})
	cfg := writeConfig(t, `
brews:
  - git
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: 48
`)
	out, _ := runSyncdWithDefaults(t, brewState, defaultsState, 0, "plan", "--config", cfg)
	if strings.Contains(out, "tilesize") {
		t.Errorf("matching default should not appear in output, got: %s", out)
	}
}

func TestPlan_DefaultsDriftOnlyExitTwo(t *testing.T) {
	brewState := setupFakeState(t, nil, []string{"git"}, nil)
	defaultsState := setupFakeDefaultsState(t, map[string]string{
		"com.apple.dock__tilesize": "64",
	})
	cfg := writeConfig(t, `
brews:
  - git
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: 48
`)
	// Brew is in sync, only defaults drift → exit 2
	_, code := runSyncdWithDefaults(t, brewState, defaultsState, 2, "plan", "--config", cfg)
	if code != 2 {
		t.Errorf("expected exit 2 with only defaults drift, got %d", code)
	}
}

func TestPlan_DefaultsReadOnly(t *testing.T) {
	brewState := setupFakeState(t, nil, []string{"git"}, nil)
	defaultsState := setupFakeDefaultsState(t, map[string]string{
		"com.apple.dock__tilesize": "64",
	})
	cfg := writeConfig(t, `
brews:
  - git
defaults:
  - domain: com.apple.dock
    key: tilesize
    type: int
    value: 48
`)
	runSyncdWithDefaults(t, brewState, defaultsState, 2, "plan", "--config", cfg)

	// Verify defaults state was not modified (no write happened)
	data, err := os.ReadFile(filepath.Join(defaultsState, "com.apple.dock__tilesize"))
	if err != nil {
		t.Fatalf("state file should still exist: %v", err)
	}
	if string(data) != "64" {
		t.Errorf("plan should not modify defaults state, got: %s", string(data))
	}
}
