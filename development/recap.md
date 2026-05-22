# Recap

## v0.3 — macOS Defaults

### Phase 1 — 2026-05-21
- Config schema & validation complete: `DefaultEntry` struct, `ValidateDefaults()`, `CommandRunner` extracted to `internal/runner/`
- All 7 reqs (REQ-067–073) verified with 14 unit tests; no regressions across 34 integration tests

### Phase 2 — 2026-05-21
- State reader, type comparison, and drift calculator complete; 17 unit tests pass (REQ-074–077, 098–103)

### Phase 3 — 2026-05-22
- Plan integration complete: drift display, exit code 2, read-only enforcement; 5 new integration tests (REQ-075–080)

### Phase 4 — 2026-05-22
- Executor & apply complete: `WriteDrifted`, type flags, kill dedup, continue-on-error (REQ-081–090, 104–106)

### Phase 5 — 2026-05-22
- Init snapshot complete: `--defaults` flag, `KnownApps` map (14 domains), kill inference, skip+warn (REQ-091–097)

### Phase 6 — 2026-05-22
- Verbose output complete; v0.3 feature done. All tests green.

---

## Smart Progress Display

### Phases 1–3 — 2026-05-22
- `internal/progress` package: phase detection, hang timeout (60s), single retry, `\r` display, error extraction
- Integrated into upgrade command; verbose bypasses to raw output; non-TTY falls back to line-per-phase
- All tests green: 99 unit + 52 integration tests pass (REQ-108–127)

---

## Parallel Pre-Download

### Phases 1–3 — 2026-05-22
- `internal/download` package: platform detection, `brew info --json=v2` URL extraction, SHA256 cache naming, parallel HTTP with semaphore pool
- Atomic `.downloading` → rename placement; non-fatal failures; `--concurrency` flag (default 4); verbose skips pre-download
- All tests green: unit + integration pass with no regressions (REQ-128–156)

---

## Download Progress Bars

### Phases 1–3 — 2026-05-22
- `ProgressWriter` + `DownloadDisplay` in `internal/download/progress.go`: per-file bars (█/░), slot-based multi-line renderer with 100ms ANSI cursor-up redraw
- Integrated into upgrade command via `OnProgress` callback; non-TTY falls back to one line per completion
- All tests green: no regressions across unit + 52 integration tests (REQ-157–176)