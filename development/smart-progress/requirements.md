# Requirements: Smart Progress Display for Brew Upgrades

## User Stories

**US-015:** As a Mac user, I want `syncd upgrade` to show clean, single-line progress per package so that I can see what's happening without a wall of brew output.

**US-016:** As a Mac user, I want syncd to detect and recover from hung brew processes so that upgrades don't stall indefinitely.

**US-017:** As a Mac user, I want to see only the relevant error message when a package fails to upgrade so that I can quickly diagnose the issue.

---

## Functional Requirements

### Progress Display

**REQ-108:** `syncd upgrade` shall display a single status line per package that updates in-place using `\r` carriage return.
- Verification: Run `syncd upgrade --yes` in a TTY, confirm status line overwrites itself (no scrolling).

**REQ-109:** The status line shall show the package index, total count, package name, and current phase (e.g., `  [1/5] neovim: downloading...`).
- Verification: Run upgrade with multiple packages, confirm format matches `[N/T] name: phase`.

**REQ-110:** syncd shall detect the following phases from brew output: `downloading`, `installing`, `pouring`, `built`.
- Verification: Unit test — feed known brew output lines, confirm correct phase detected.

**REQ-111:** On successful completion of a package, syncd shall print a final `✓ name` line (not overwritten).
- Verification: Run upgrade, confirm `✓` lines persist after completion.

**REQ-112:** On failure of a package, syncd shall print a `✗ name` line followed by extracted error summary.
- Verification: Trigger a failure, confirm `✗` line and error summary shown without full brew dump.

**REQ-113:** syncd shall extract only relevant error lines from brew output on failure (not the full dump).
- Verification: Unit test — feed brew failure output, confirm only error-relevant lines extracted.

### Hang Detection & Retry

**REQ-114:** syncd shall kill a brew process if it produces no stdout/stderr output for 60 seconds.
- Verification: Unit test — mock a process that produces no output, confirm killed after timeout.

**REQ-115:** After killing a hung process, syncd shall retry the same package once.
- Verification: Unit test — first attempt hangs, second succeeds, confirm package reported as success.

**REQ-116:** If the retry also hangs or fails, syncd shall report the package as failed and continue to the next.
- Verification: Unit test — both attempts hang, confirm failure reported and next package proceeds.

**REQ-117:** The hang timeout shall be 60 seconds (hardcoded, no config).
- Verification: Code review — confirm 60-second constant.

### Output Capture

**REQ-118:** syncd shall capture brew's stdout and stderr to pipes (not stream to terminal) during upgrade.
- Verification: Run upgrade, confirm no raw brew output appears in terminal (only syncd's progress lines).

**REQ-119:** syncd shall read captured output line-by-line in a goroutine for phase detection.
- Verification: Code review — confirm goroutine reads from pipe.

**REQ-120:** syncd shall store captured output in memory for error extraction on failure.
- Verification: Unit test — after failure, confirm full output available for error extraction.

### Compatibility

**REQ-121:** `--verbose` mode shall bypass smart progress and stream raw brew output (existing `RunMutate` behavior).
- Verification: Run `syncd upgrade --yes --verbose`, confirm raw brew output streams to terminal.

**REQ-122:** When stdout is not a TTY, syncd shall fall back to simple line-per-phase output (no `\r` overwrite).
- Verification: Pipe `syncd upgrade --yes` output, confirm no `\r` characters, one line per phase change.

**REQ-123:** Smart progress shall not change the `CommandRunner` interface.
- Verification: Code review — confirm `CommandRunner` interface unchanged.

**REQ-124:** Existing unit and integration tests shall continue to pass without modification.
- Verification: Run `make test && make integration-test`, confirm all pass.

---

## Technical Requirements

**REQ-125:** Smart progress logic shall live in a new `internal/progress` package.
- Verification: Code review — confirm package exists at `internal/progress/`.

**REQ-126:** The progress runner shall accept a callback or channel for phase updates (decoupled from display).
- Verification: Unit test — run progress with mock callback, confirm phases reported.

**REQ-127:** The progress display shall use `\r` + space-padding to clear previous line content.
- Verification: Unit test — confirm output contains `\r` and trailing spaces.

---

## Traceability Matrix

| Requirement | User Story | Component |
|-------------|-----------|-----------|
| REQ-108–109 | US-015 | display |
| REQ-110–113 | US-015, US-017 | phase detection, error extraction |
| REQ-114–117 | US-016 | hang detection |
| REQ-118–120 | US-015 | output capture |
| REQ-121–124 | US-015 | compatibility |
| REQ-125–127 | US-015 | technical |
