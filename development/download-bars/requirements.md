# Requirements: Multi-Line Download Progress Bars

## User Stories

**US-021:** As a Mac user, I want to see per-file download progress bars during `syncd upgrade` so that I can tell which packages are downloading and how far along each one is.

---

## Functional Requirements

### Multi-Line Progress Display

**REQ-157:** syncd shall display one progress line per active download slot during the pre-download phase.
- Verification: Run `syncd upgrade --yes` with 4+ outdated packages, confirm 4 lines shown simultaneously.

**REQ-158:** Each progress line shall show: arrow indicator, package name, progress bar, percentage, and byte counts (e.g., `↓ neovim  ██████░░░░  45% (22/50 MB)`).
- Verification: Unit test — render a slot at 45% of 50 MB, confirm format matches.

**REQ-159:** syncd shall use ANSI cursor-up escape codes (`\033[NA`) to redraw all N slot lines in place.
- Verification: Integration test — confirm output contains `\033[` sequences in TTY mode.

**REQ-160:** syncd shall refresh the progress display every 100ms.
- Verification: Code review — confirm ticker interval is 100ms.

**REQ-161:** When a download completes successfully, its slot shall show `✓ name` and the next queued package shall take the slot.
- Verification: Unit test — complete a download, confirm slot transitions to next package.

**REQ-162:** When a download fails, its slot shall show `✗ name: error` and the next queued package shall take the slot.
- Verification: Unit test — fail a download, confirm slot shows error and transitions.

**REQ-163:** When all downloads complete, the final display shall remain showing all completed packages.
- Verification: Integration test — after download phase, confirm final lines persist (no clear).

### Progress Tracking

**REQ-164:** syncd shall track bytes downloaded per file in real-time using a writer wrapper around `io.Copy`.
- Verification: Unit test — write through wrapper, confirm byte count updates.

**REQ-165:** syncd shall read `Content-Length` from the HTTP response header to determine total file size.
- Verification: Unit test — mock HTTP response with Content-Length, confirm total captured.

**REQ-166:** When `Content-Length` is unknown (0 or absent), syncd shall display bytes downloaded only (no bar, no percentage).
- Verification: Unit test — render slot with unknown total, confirm format is `↓ name  22 MB` (no bar).

### Non-TTY Fallback

**REQ-167:** When stdout is not a TTY, syncd shall print one line per download completion (no cursor movement).
- Verification: Pipe output, confirm one `✓`/`✗` line per package, no ANSI cursor codes.

**REQ-168:** Non-TTY output shall show package name and final byte count per completion.
- Verification: Pipe output, confirm each line includes name and size.

### Integration

**REQ-169:** The multi-line display shall be activated via a new `OnProgress` callback in `download.Options`.
- Verification: Code review — `Options` struct has `OnProgress` field.

**REQ-170:** `PreDownload` shall call `OnProgress` with package name, bytes downloaded, and total bytes during download.
- Verification: Unit test — mock OnProgress, confirm called with correct values during download.

**REQ-171:** The progress display logic shall live in `internal/download/progress.go`.
- Verification: Code review — file exists.

**REQ-172:** Existing `OnComplete` callback shall continue to work unchanged.
- Verification: Run existing integration tests, confirm all pass.

**REQ-173:** Verbose mode shall continue to skip the pre-download phase entirely (no change).
- Verification: Run `syncd upgrade --verbose`, confirm no download progress shown.

---

## Technical Requirements

**REQ-174:** The progress bar shall use `█` for filled and `░` for empty, with a fixed width of 20 characters.
- Verification: Unit test — render at 50%, confirm 10 `█` and 10 `░`.

**REQ-175:** syncd shall use `golang.org/x/term` for TTY detection (reuse existing dependency).
- Verification: Code review — no new TTY detection dependency.

**REQ-176:** syncd shall reuse color variables from `internal/cli/output.go` for `✓`/`✗` indicators.
- Verification: Code review — imports `cli.Green`, `cli.Red`.

---

## Traceability Matrix

| Requirement | User Story | Component |
|-------------|-----------|-----------|
| REQ-157–163 | US-021 | multi-line display |
| REQ-164–166 | US-021 | progress tracking |
| REQ-167–168 | US-021 | non-TTY fallback |
| REQ-169–173 | US-021 | integration |
| REQ-174–176 | US-021 | technical |
