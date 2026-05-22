# Tasks: Smart Progress Display for Brew Upgrades

## Phase 1: Progress Runner & Phase Detection (~2 hours)

### 1.1 Phase detector
> REQ-110 | Design: Phase Detector

- [x] Create `internal/progress/phase.go`
  - `DetectPhase(line string) string`
  - Match: `Downloading`/`downloading` → `"downloading"`
  - Match: `Pouring`/`pouring` → `"pouring"`
  - Match: `Installing`/`installing` → `"installing"`
  - Match: `Built`/`built from source` → `"built"`
  - Return `""` for non-matching lines
- [x] Add unit tests in `internal/progress/phase_test.go`
  - Test: "Downloading https://..." → "downloading"
  - Test: "Pouring neovim--0.9.5" → "pouring"
  - Test: "Installing neovim" → "installing"
  - Test: "Built from source" → "built"
  - Test: random line → ""
  - Test: empty line → ""
- [x] Verify: `go test ./internal/progress/...` passes

### 1.2 Progress runner with hang detection
> REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-125, REQ-126 | Design: Progress Runner

- [x] Create `internal/progress/runner.go`
  - `const HangTimeout = 60 * time.Second`
  - `type Result struct { Output []byte; Err error; Hung bool }`
  - `func RunWithProgress(name string, args []string, onPhase func(string)) Result`
  - Set `HOMEBREW_NO_AUTO_UPDATE=1` env
  - Capture stdout+stderr via pipes, merge with `io.MultiReader`
  - Goroutine: scan lines, detect phase, call callback, reset timer, buffer output
  - Kill on timer expiry (60s no output)
  - Retry once on hang; if retry fails/hangs, return failure
- [x] Add unit tests in `internal/progress/runner_test.go`
  - Test: successful command returns output and no error
  - Test: failing command returns error with captured output
  - Test: hung command killed after timeout (use short timeout for test)
  - Test: retry succeeds after first hang
  - Test: retry also hangs → failure reported
  - Test: callback called with detected phases
- [x] Verify: `go test ./internal/progress/...` passes

**Estimate:** ~2 hours

---

## Phase 2: Error Extraction & Display (~1.5 hours)

### 2.1 Error extractor
> REQ-113, REQ-120 | Design: Error Extractor

- [x] Create `internal/progress/errors.go`
  - `func ExtractError(output []byte) string`
  - Scan for `Error:` line (case-insensitive)
  - Include error line + up to 2 following context lines
  - Fallback: last 3 lines if no `Error:` found
  - Max 5 lines total
- [x] Add unit tests in `internal/progress/errors_test.go`
  - Test: output with "Error: ..." extracts that line + context
  - Test: output without "Error:" returns last 3 lines
  - Test: empty output returns empty string
  - Test: long output truncated to 5 lines
- [x] Verify: `go test ./internal/progress/...` passes

### 2.2 Display with TTY detection
> REQ-108, REQ-122, REQ-127 | Design: Display

- [x] Create `internal/progress/display.go`
  - `type Display struct { IsTTY bool; last int }`
  - `func NewDisplay() *Display` — detect TTY via `os.Stdout.Fd()`
  - `func (d *Display) Status(format string, args ...interface{})` — `\r` overwrite in TTY, newline in non-TTY
  - `func (d *Display) Finish(format string, args ...interface{})` — always newline-terminated
  - Pad with spaces to clear previous longer line
- [x] Add unit tests in `internal/progress/display_test.go`
  - Test: TTY mode output contains `\r` prefix
  - Test: non-TTY mode output contains `\n` suffix
  - Test: Finish always ends with `\n`
  - Test: padding clears previous longer content
- [x] Verify: `go test ./internal/progress/...` passes

**Estimate:** ~1.5 hours

---

## Phase 3: Upgrade Integration (~1.5 hours)

### 3.1 Wire progress into upgrade command
> REQ-108, REQ-109, REQ-111, REQ-112, REQ-121, REQ-123 | Design: Integration with Upgrade Command

- [ ] Update `internal/cli/upgrade.go`
  - Create `progress.NewDisplay()` before upgrade loop
  - When `runner.Verbose` is true: use `RunMutate` (unchanged behavior)
  - When not verbose: use `progress.RunWithProgress` with phase callback
  - On success: `display.Finish("  %s✓%s %s", Green, Reset, name)`
  - On failure: extract error, `display.Finish("  %s✗%s %s: %s", Red, Reset, name, errMsg)`
  - Same pattern for cask upgrades
- [ ] Verify: `syncd upgrade --yes` shows progress lines (not raw brew output)
- [ ] Verify: `syncd upgrade --yes --verbose` streams raw output (unchanged)
- [ ] Verify: `CommandRunner` interface unchanged (REQ-123)

### 3.2 Integration tests
> REQ-108, REQ-109, REQ-111, REQ-112, REQ-121, REQ-124 | Design: Integration Tests

- [ ] Extend fake brew script to support `upgrade` subcommand
  - Print phase-like lines: "Downloading...", "Pouring...", etc.
  - Support `$FAKE_BREW_UPGRADE_FAIL` env to simulate failure with "Error:" output
- [ ] Add integration test: upgrade shows `✓` on success (no raw brew output)
- [ ] Add integration test: upgrade shows `✗` + error summary on failure
- [ ] Add integration test: verbose mode streams raw output
- [ ] Verify: `make test && make integration-test` passes (all existing tests green)

**Estimate:** ~1.5 hours

---

## Summary

| Phase | Estimate | Key Deliverable |
|-------|----------|-----------------|
| 1. Progress Runner & Phase Detection | ~2h | Core engine: capture, detect, timeout, retry |
| 2. Error Extraction & Display | ~1.5h | Clean error messages and `\r` rendering |
| 3. Upgrade Integration | ~1.5h | `syncd upgrade` uses smart progress |
| **Total** | **~5h** | |

---

## Coverage Summary

### Requirements Covered

All 20 requirements (REQ-108 through REQ-127) are covered by tasks above.

| Requirement Range | Phase | Tasks |
|-------------------|-------|-------|
| REQ-108–109 | Phase 2, 3 | 2.2, 3.1, 3.2 |
| REQ-110 | Phase 1 | 1.1 |
| REQ-111–112 | Phase 3 | 3.1, 3.2 |
| REQ-113 | Phase 2 | 2.1 |
| REQ-114–120 | Phase 1 | 1.2 |
| REQ-121 | Phase 3 | 3.1, 3.2 |
| REQ-122 | Phase 2 | 2.2 |
| REQ-123–124 | Phase 3 | 3.1, 3.2 |
| REQ-125–126 | Phase 1 | 1.2 |
| REQ-127 | Phase 2 | 2.2 |

### Uncovered Requirements

None — all 20 requirements have at least one task with a concrete verification method.
