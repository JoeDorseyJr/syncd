# Findings: Smart Progress Display for Brew Upgrades

## Issues

### High

- [ ] **F-001: Design says `RunWithProgress` sets `HOMEBREW_NO_AUTO_UPDATE=1` but the upgrade command already runs `brew update` explicitly before checking outdated.**
  - Evidence: Design section "Progress Runner" says "Set `HOMEBREW_NO_AUTO_UPDATE=1` env." The current `upgrade.go` already calls `r.RunMutate("brew", "update")` before querying outdated packages, and `runner.go`'s `brewEnv()` already appends `HOMEBREW_NO_AUTO_UPDATE=1` to all commands. The env var in `RunWithProgress` is redundant but harmless — however, `RunWithProgress` bypasses the `CommandRunner` interface entirely (uses `exec.Command` directly), so it won't inherit `brewEnv()`.
  - Impact: If `RunWithProgress` doesn't set the env var, brew could auto-update mid-upgrade, causing unexpected output that confuses phase detection. The design is correct to include it, but this means `RunWithProgress` duplicates env setup logic from `runner.go`. This creates a maintenance risk — if env setup changes in `runner.go`, `RunWithProgress` won't pick it up.
  - Status: open
  - Resolution:
  - Follow-up: Design should document that `RunWithProgress` must replicate `brewEnv()` or import it from `internal/runner`.

- [ ] **F-002: `RunWithProgress` bypasses `CommandRunner` interface but REQ-123 says interface must not change — no mechanism for testability.**
  - Evidence: Design shows `RunWithProgress(name string, args []string, onPhase func(string)) Result` as a standalone function that directly uses `exec.Command`. REQ-123 says "Smart progress shall not change the `CommandRunner` interface." The current upgrade loop uses `r.RunMutate(...)` which goes through `CommandRunner`. The new code calls `progress.RunWithProgress(...)` directly, bypassing the interface entirely.
  - Impact: Unit testing the upgrade command's progress path becomes difficult. `MockRunner` won't intercept progress calls. Integration tests with the fake brew script will work (since they replace the `brew` binary on PATH), but unit tests of the upgrade loop logic cannot mock the progress runner without a separate abstraction.
  - Status: open
  - Resolution:
  - Follow-up: Design should address testability — either accept that progress path is only integration-tested, or add a `ProgressRunner` interface/function type that can be injected.

- [ ] **F-003: Design merges stdout and stderr via `io.MultiReader` but this loses ordering guarantees.**
  - Evidence: Design says "Get `StdoutPipe()` and `StderrPipe()`, merge into one reader." `io.MultiReader` reads one reader to completion before starting the next — it does NOT interleave. This means all stdout would be read first, then all stderr (or vice versa), which is incorrect for line-by-line phase detection.
  - Impact: Phase detection would fail if brew writes phases to stdout but errors to stderr. The goroutine would block on one pipe until the process exits, then read the other. This is a design bug.
  - Status: open
  - Resolution:
  - Follow-up: Design should specify using `io.Pipe` + two goroutines (one per pipe) writing to a shared channel, or use `cmd.CombinedOutput`-style approach with a single pipe via `cmd.Stdout = pw; cmd.Stderr = pw` where `pw` is a `*io.PipeWriter`.

### Medium

- [ ] **F-004: REQ-116 says "report failure and continue to the next" but `RunWithProgress` is per-package — "continue" is the upgrade loop's responsibility, not the progress runner's.**
  - Evidence: REQ-116: "If the retry also hangs or fails, syncd shall report the package as failed and continue to the next." The design's `RunWithProgress` returns a `Result` with `Err` set. The "continue to next" behavior is already handled by the upgrade loop (it iterates over packages). The requirement conflates two layers.
  - Impact: Low — the behavior is correct in aggregate. But the requirement wording implies `RunWithProgress` itself handles continuation, which it doesn't. Verification should test the upgrade loop, not just `RunWithProgress`.
  - Status: open
  - Resolution:
  - Follow-up: REQ-116 verification should be an integration test of the full upgrade loop, not just a unit test of `RunWithProgress`.

- [ ] **F-005: No requirement or design for what happens to the `brew update` spinner when smart progress is active.**
  - Evidence: Current `upgrade.go` has a custom spinner for `brew update` (prints dots every 500ms). The design only addresses the per-package upgrade loop. The `brew update` call still uses `r.RunMutate("brew", "update")` which streams to terminal.
  - Impact: The `brew update` phase will still dump raw output to terminal before the clean progress display starts. This creates an inconsistent UX — raw output for update, then clean progress for upgrades. Not a bug, but a UX gap.
  - Status: open
  - Resolution:
  - Follow-up: Consider whether `brew update` should also use `RunWithProgress` or at minimum suppress its output (it already has a custom spinner).

- [ ] **F-006: Design says "Extend fake brew script to support `upgrade` subcommand" but it already exists.**
  - Evidence: Task 3.2 says "Extend fake brew script to support `upgrade` subcommand." The current integration test infrastructure already has an `upgrade` case in the fake brew script (lines 136-148 of `integration_test.go`). It handles success and `*fail*` pattern matching.
  - Impact: Task 3.2 should say "extend the existing fake brew `upgrade` handler to print phase-like output lines" rather than implying it needs to be created from scratch. Minor wording issue.
  - Status: open
  - Resolution:
  - Follow-up: `tasks.md` task 3.2 — clarify that upgrade handler exists and needs enhancement.

- [ ] **F-007: REQ-114 verification says "mock a process that produces no output" but unit testing process kill requires careful design.**
  - Evidence: REQ-114 verification: "Unit test — mock a process that produces no output, confirm killed after timeout." Testing process kill in unit tests requires either: (a) spawning a real subprocess that sleeps, (b) abstracting the process creation, or (c) using a very short timeout in tests. The design doesn't specify how to make the 60s timeout testable.
  - Impact: If the timeout is hardcoded at 60s with no test override, unit tests will either take 60s per hang test or require a different approach. Task 1.2 tests say "use short timeout for test" but the design says `const HangTimeout = 60 * time.Second` with no injection mechanism.
  - Status: open
  - Resolution:
  - Follow-up: Design should either make timeout a parameter of `RunWithProgress` or provide a package-level test helper to override it.

- [ ] **F-008: No requirement for what the status line shows BEFORE the first phase is detected.**
  - Evidence: REQ-109 specifies format `[N/T] name: phase` but brew may output several lines before any phase keyword appears (e.g., "Updating Homebrew..." or dependency resolution). The design's `DetectPhase` returns `""` for non-matching lines, and the callback is only called when a phase is detected.
  - Impact: The user sees no progress indication between starting a package and the first phase detection. The status line would show the package name but no phase. This is a UX gap — should it show "starting..." or similar?
  - Status: open
  - Resolution:
  - Follow-up: Consider adding an initial status display before the first phase callback fires.

- [ ] **F-009: Design's phase detection patterns are case-sensitive in the table but implementation says "Contains `Downloading` or `downloading`".**
  - Evidence: Design Phase Detector table shows both capitalized and lowercase variants as separate patterns. The implementation note says "Contains `Downloading` or `downloading`". This is ambiguous — should it be case-insensitive matching or explicit dual-case checks?
  - Impact: If brew changes its output casing (e.g., "DOWNLOADING"), the detector would miss it. Low risk since brew output is stable, but the design should be explicit about the matching strategy.
  - Status: open
  - Resolution:
  - Follow-up: Design should specify: use `strings.Contains(strings.ToLower(line), pattern)` for case-insensitive matching.

- [ ] **F-010: REQ-118 verification is weak — "Run upgrade, confirm no raw brew output appears in terminal."**
  - Evidence: REQ-118 verification says to run upgrade and confirm no raw output. But in integration tests with a fake brew script, the fake brew's output IS the raw output. The test would need to verify that the output is reformatted (shows progress format, not raw lines). This is really testing the same thing as REQ-108/109.
  - Impact: REQ-118 verification overlaps with REQ-108/109 and doesn't add independent validation. The real verification is that pipes are used (code review) and that output is reformatted (integration test).
  - Status: open
  - Resolution:
  - Follow-up: REQ-118 verification should be "Code review — confirm `cmd.StdoutPipe()` used instead of `cmd.Stdout = os.Stdout`" plus the integration test from REQ-108.

- [ ] **F-011: Task 1.2 test "hung command killed after timeout (use short timeout for test)" contradicts design's hardcoded constant.**
  - Evidence: Task 1.2 says "use short timeout for test" but design declares `const HangTimeout = 60 * time.Second`. A Go `const` cannot be overridden in tests. Either it needs to be a `var` (less safe) or `RunWithProgress` needs a timeout parameter.
  - Impact: Direct conflict between design and task. Implementation will need to deviate from one or the other.
  - Status: open
  - Resolution:
  - Follow-up: Design should change `const` to a package-level `var` or add a `RunWithProgressTimeout` variant for testing. Alternatively, use a test helper that creates a short-lived process.

### Low

- [ ] **F-012: Problem statement says "no config needed for v1" for timeout but REQ-117 says "hardcoded, no config" — consistent but worth noting the timeout is not user-tunable.**
  - Evidence: Both problem statement and REQ-117 agree: 60s, no config. This is fine for initial implementation but may need revisiting if users have slow connections.
  - Impact: Informational. No action needed for this milestone.
  - Status: open
  - Resolution:
  - Follow-up: None needed.

- [ ] **F-013: Design doesn't specify behavior when brew process is killed — does it send SIGTERM or SIGKILL?**
  - Evidence: Design says "kill process" on timeout. Go's `cmd.Process.Kill()` sends SIGKILL on Unix. `cmd.Process.Signal(syscall.SIGTERM)` would be gentler. SIGKILL doesn't allow cleanup (temp files, partial downloads).
  - Impact: Low — brew handles interruption gracefully regardless. But SIGTERM → wait → SIGKILL is a more robust pattern.
  - Status: open
  - Resolution:
  - Follow-up: Design could specify SIGTERM with 5s grace period before SIGKILL, or accept SIGKILL as simpler.

- [ ] **F-014: No requirement for progress display during cask upgrades specifically — casks may have different output patterns.**
  - Evidence: REQ-110 lists phases: downloading, installing, pouring, built. Cask upgrades may output different patterns (e.g., "Moving App to /Applications", "Quarantine", "Linking"). These won't be detected as phases.
  - Impact: Cask upgrades may show no phase updates (just the package name with no phase). Functional but less informative.
  - Status: open
  - Resolution:
  - Follow-up: Consider adding cask-specific phase patterns in a future iteration, or document as known limitation.

- [ ] **F-015: Design's `ExtractError` fallback of "last 3 lines" may include progress/phase lines rather than actual error context.**
  - Evidence: Design says "If no `Error:` found, return last 3 lines of output as fallback." If brew fails without printing "Error:" explicitly, the last 3 lines might be phase output (e.g., "Pouring...", "Installing...", then the process exits with non-zero). This would show unhelpful context.
  - Impact: Low — most brew failures do include "Error:" in output. Edge case for unusual failures.
  - Status: open
  - Resolution:
  - Follow-up: None needed for this milestone.

---

## Open Clarification Questions

1. **Should `RunWithProgress` accept a timeout parameter for testability, or should the package use a `var` instead of `const`?**
   - Context: Task 1.2 says "use short timeout for test" but design declares `const HangTimeout`. These are incompatible. Options: (a) make it a `var`, (b) add timeout as a function parameter, (c) use a test-only build tag to override.
   - Answer: (b) Pass timeout as a function parameter.

2. **How should stdout and stderr be merged for line-by-line reading?**
   - Context: Design says `io.MultiReader` but that reads sequentially, not interleaved. Options: (a) `cmd.Stdout = pw; cmd.Stderr = pw` with a shared pipe writer, (b) two goroutines feeding a channel, (c) redirect stderr to stdout via `cmd.Stderr = cmd.Stdout` (not supported by Go).
   - Answer: (a) Shared pipe writer — both cmd.Stdout and cmd.Stderr point to the same pipe writer.

3. **Should there be an initial status display (e.g., "starting...") before the first phase is detected?**
   - Context: Between starting a package and the first phase keyword appearing in output, the user sees no progress. This could be seconds of apparent inactivity.
   - Answer: Yes, show "starting..." immediately.

4. **Should `brew update` also use smart progress, or is the current dot-spinner sufficient?**
   - Context: The upgrade command runs `brew update` with a custom spinner before checking outdated packages. Smart progress only applies to the per-package upgrade loop. The UX transition from raw spinner to clean progress is slightly jarring.
   - Answer: Keep the dot-spinner for brew update.

5. **How should the progress runner be tested in unit tests — real subprocesses with short timeouts, or an abstracted process interface?**
   - Context: Testing hang detection requires either spawning real processes (slow, flaky) or abstracting process creation behind an interface (more code, more testable). The design doesn't specify.
   - Answer: Abstract process creation behind an interface for unit tests.

---

## Validation Notes

- All REQ-108 through REQ-127 IDs are present in requirements, design, and tasks.
- All 20 requirements have at least one task with a verification method.
- All tasks trace back to at least one requirement.
- No orphan tasks or orphan requirements found.
- Traceability is complete across all four planning artifacts.
- Problem statement scope, non-goals, and constraints are reflected in requirements.
- The strongest validation paths are unit tests for phase detection/error extraction and integration tests with enhanced fake brew for CLI behavior.
- The weakest validation paths are REQ-117 (code review only), REQ-119 (code review only), and REQ-123 (code review only) — these are structural requirements that can only be verified by inspection.
