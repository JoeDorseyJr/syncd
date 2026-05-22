# Tasks: Multi-Line Download Progress Bars

## Phase 1: ProgressWriter & downloadFile Integration (~1 hour)

### 1.1 Add OnProgress to Options and ProgressWriter
> REQ-164, REQ-165, REQ-169, REQ-170, REQ-171 | Design: ProgressWriter, Options Change

- [x] Add `OnProgress func(name string, downloaded, total int64)` to `Options` struct in `download.go`
- [x] Create `internal/download/progress.go`
  - `ProgressWriter` struct: Name, Total, Downloaded, OnProgress callback
  - `Write(p []byte) (int, error)` — increments Downloaded, calls OnProgress
- [x] Update `downloadFile` signature to accept `name string` and `onProgress func(string, int64, int64)`
  - Use `io.TeeReader(resp.Body, &pw)` with ProgressWriter
  - Set `pw.Total = resp.ContentLength`
- [x] Update `PreDownload` goroutine to pass `opts.OnProgress` and `p.Name` to `downloadFile`
- [x] Add unit tests in `internal/download/progress_test.go`
  - Test: ProgressWriter tracks cumulative bytes correctly
  - Test: OnProgress called with correct name, downloaded, total
  - Test: nil OnProgress does not panic
- [x] Verify: `go test ./internal/download/...` passes
- [x] Verify: existing integration tests pass (OnComplete still works)

**Estimate:** ~1 hour

---

## Phase 2: DownloadDisplay Renderer (~1.5 hours)

### 2.1 Slot management and render logic
> REQ-157, REQ-158, REQ-159, REQ-160, REQ-161, REQ-162, REQ-163, REQ-166, REQ-174, REQ-175, REQ-176 | Design: DownloadDisplay, Render Logic

- [x] Add to `internal/download/progress.go`:
  - `SlotState` struct: Name, Downloaded, Total, Done, Err
  - `DownloadDisplay` struct: mu (sync.Mutex), slots, isTTY, rendered
  - `NewDownloadDisplay(concurrency int) *DownloadDisplay` — detect TTY via `golang.org/x/term`
  - `UpdateProgress(name string, downloaded, total int64)` — find slot by name, update bytes
  - `MarkDone(name string, err error)` — mark slot done
  - `AssignSlot(name string, total int64)` — assign package to free slot
  - `Start() func()` — start 100ms ticker goroutine, return stop function
  - `render()` — cursor-up + reprint all slots
- [x] Implement `renderSlot(s SlotState) string`:
  - Active with known total: `  ↓ name  ████░░░░  45% (22/50 MB)` (20-char bar, █/░)
  - Active with unknown total: `  ↓ name  22.0 MB`
  - Done success: `  ✓ name (50.0 MB)` (green ✓)
  - Done error: `  ✗ name: error` (red ✗)
  - Empty slot: empty string
- [x] Add unit tests in `internal/download/progress_test.go`
  - Test: renderSlot at 50% → 10█ + 10░, "50%"
  - Test: renderSlot at 0% → 20░
  - Test: renderSlot at 100% → 20█
  - Test: renderSlot unknown total → bytes only, no bar
  - Test: renderSlot done success → ✓ + name
  - Test: renderSlot done error → ✗ + name + error
  - Test: AssignSlot fills first free slot
  - Test: MarkDone frees slot for next package
- [x] Verify: `go test ./internal/download/...` passes

### 2.2 Non-TTY fallback
> REQ-167, REQ-168 | Design: Non-TTY Fallback

- [x] When `isTTY` is false, `render()` is a no-op
- [x] Non-TTY completion handled by existing `OnComplete` callback in upgrade command
- [x] Add unit test: non-TTY DownloadDisplay produces no cursor codes
- [x] Verify: `go test ./internal/download/...` passes

**Estimate:** ~1.5 hours

---

## Phase 3: Integration & Polish (~1 hour)

### 3.1 Wire into upgrade command
> REQ-157, REQ-172, REQ-173 | Design: Integration with Upgrade Command

- [x] Update pre-download section in `internal/cli/upgrade.go`:
  - Create `DownloadDisplay` with `dlConcurrency`
  - Call `dd.Start()` before `PreDownload`
  - Set `OnProgress` to call `dd.UpdateProgress`
  - Set `OnComplete` to call `dd.MarkDone`
  - Call `stop()` after `PreDownload` returns
  - Remove old single-line `Downloading [N/M]...` counter
- [x] Keep verbose mode unchanged (skips pre-download entirely)
- [x] Verify: `syncd upgrade --yes` shows per-file progress bars
- [x] Verify: `syncd upgrade --verbose` skips pre-download (unchanged)

### 3.2 Integration tests
> REQ-159, REQ-167, REQ-172, REQ-173 | Design: Integration Tests

- [x] Add integration test: download with progress — output contains cursor-up codes (TTY)
- [x] Add integration test: non-TTY — no cursor codes, one line per completion
- [x] Verify: `make test && make integration-test` passes (all existing tests green)

**Estimate:** ~1 hour

---

## Summary

| Phase | Estimate | Key Deliverable |
|-------|----------|-----------------|
| 1. ProgressWriter & downloadFile | ~1h | OnProgress fires with byte counts |
| 2. DownloadDisplay renderer | ~1.5h | Multi-line slot-based display |
| 3. Integration & polish | ~1h | Wired into upgrade command |
| **Total** | **~3.5h** | |

---

## Coverage Summary

### Requirements Covered

All 20 requirements (REQ-157 through REQ-176) are covered by tasks above.

| Requirement Range | Phase | Tasks |
|-------------------|-------|-------|
| REQ-157–163 | Phase 2, 3 | 2.1, 3.1, 3.2 |
| REQ-164–166 | Phase 1, 2 | 1.1, 2.1 |
| REQ-167–168 | Phase 2 | 2.2 |
| REQ-169–173 | Phase 1, 3 | 1.1, 3.1, 3.2 |
| REQ-174–176 | Phase 2 | 2.1 |

### Uncovered Requirements

None — all 20 requirements have at least one task with a concrete verification method.
