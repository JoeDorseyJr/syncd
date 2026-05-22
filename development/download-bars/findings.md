# Findings: Multi-Line Download Progress Bars

## Issues

### High

- [ ] **F-001: Design's `AssignSlot` is never called in the integration code — slot lifecycle is unclear.**
  - Evidence: Design defines `AssignSlot(name string, total int64)` on `DownloadDisplay`, but the integration code in the design only calls `UpdateProgress` and `MarkDone`. The `PreDownload` goroutine calls `OnProgress` immediately when bytes start flowing, but nothing assigns the package to a slot first. If `UpdateProgress` is called for a name that has no slot, it would need to auto-assign — but then `AssignSlot` is dead API.
  - Impact: Either `AssignSlot` must be called from somewhere (perhaps from `OnProgress` on first call for a new name, or from a new `OnStart` callback), or `UpdateProgress` must implicitly assign on first sight. The current design has a gap in the slot assignment lifecycle.
  - Status: open
  - Resolution:
  - Follow-up: Design should clarify: either (a) `UpdateProgress` auto-assigns on first call for unknown name, making `AssignSlot` internal-only, or (b) add an `OnStart` callback to `Options` that fires when a download begins (before first byte), which calls `AssignSlot`.

- [ ] **F-002: Design changes `downloadFile` signature but doesn't address the `OnProgress` call frequency — every `Write` call triggers a callback which triggers `UpdateProgress` with mutex lock.**
  - Evidence: `ProgressWriter.Write` calls `OnProgress` on every `Write`. HTTP response bodies are typically read in 32KB chunks (`io.Copy` default buffer). For a 100MB file, that's ~3200 calls. Each call acquires the `DownloadDisplay` mutex to update slot state. With 4 concurrent downloads, that's ~12800 mutex acquisitions during the download phase.
  - Impact: Medium — the 100ms render ticker means the display only redraws every 100ms regardless of callback frequency, so the visual output is fine. But the mutex contention could theoretically slow downloads on very fast connections. In practice, this is likely negligible since the mutex is held only for a field update.
  - Status: open
  - Resolution:
  - Follow-up: Consider whether throttling `OnProgress` calls (e.g., only call if >1% change or >100ms since last call) is worth the complexity. Likely acceptable as-is for v1.

### Medium

- [ ] **F-003: REQ-168 verification ("Non-TTY output shall show package name and final byte count per completion") relies on existing `OnComplete` callback, but the design removes the old counter and doesn't specify what non-TTY `OnComplete` prints.**
  - Evidence: Design says "Non-TTY mode: `DownloadDisplay` does nothing on progress updates. The existing `OnComplete` callback in the upgrade command prints one line per completion." But the integration code in the design sets `OnComplete` to call `dd.MarkDone(res.Package.Name, res.Err)` — which is a display method. In non-TTY mode, `MarkDone` would need to print a line, or the upgrade command needs a separate non-TTY path.
  - Impact: If `MarkDone` is a no-op in non-TTY mode (as stated for `render()`), then nothing prints per-completion in non-TTY. REQ-168 would be violated. The design needs to clarify: either `MarkDone` prints in non-TTY mode, or the `OnComplete` callback in the upgrade command has a conditional path.
  - Status: open
  - Resolution:
  - Follow-up: Design should specify: `MarkDone` prints `  ✓ name (X.X MB)` or `  ✗ name: error` directly to stdout when `isTTY` is false. This satisfies REQ-167 and REQ-168.

- [ ] **F-004: Design shows `DownloadDisplay` importing `cli.Green` and `cli.Red` — this creates `internal/download` → `internal/cli` import dependency.**
  - Evidence: REQ-176 says "reuse color variables from `internal/cli/output.go`". Design's `renderSlot` uses `Green`, `Red`, `Reset`. The `internal/download` package currently has no dependency on `internal/cli`. Adding one creates a new coupling direction (download → cli).
  - Impact: This may create an import cycle if `internal/cli` already imports `internal/download` (which it does — `upgrade.go` imports `download`). Go does not allow circular imports. This is a **build-breaking** issue.
  - Status: open
  - Resolution:
  - Follow-up: Options: (a) duplicate the color constants in `internal/download/progress.go` (simple, slight duplication), (b) extract colors to a shared package like `internal/color` or `internal/ui`, (c) pass color strings into `NewDownloadDisplay` as parameters. Option (a) is simplest and matches the "no new packages" constraint.

- [ ] **F-005: Task 2.2 says "Non-TTY completion handled by existing `OnComplete` callback in upgrade command" but the design replaces the old `OnComplete` logic with `dd.MarkDone`.**
  - Evidence: Current upgrade.go `OnComplete` does `display.Status("  Downloading [%d/%d]...", c, total)`. The design replaces this with `dd.MarkDone(res.Package.Name, res.Err)`. The old single-line counter is removed. Task 2.2 claims existing `OnComplete` handles non-TTY, but the design changes what `OnComplete` does.
  - Impact: Task 2.2's verification is based on a false premise. The non-TTY behavior must be explicitly implemented in `DownloadDisplay.MarkDone` or in the `OnComplete` callback with a TTY check.
  - Status: open
  - Resolution:
  - Follow-up: Related to F-003. Task 2.2 should specify that `MarkDone` prints to stdout in non-TTY mode.

- [ ] **F-006: No requirement or task for what happens when there are more packages than concurrency slots (queue behavior).**
  - Evidence: REQ-157 says "one progress line per active download slot." If there are 8 packages and 4 slots, the display shows 4 lines. When one finishes, the next takes its slot (REQ-161). But the design's `AssignSlot` is never called from the integration code (F-001), and there's no mechanism shown for queuing packages that haven't started yet.
  - Impact: The slot transition logic is the core UX differentiator (docker pull behavior). Without clear lifecycle management, the implementation may show stale "done" entries in slots instead of transitioning to the next queued package.
  - Status: open
  - Resolution:
  - Follow-up: Design should document the full lifecycle: (1) package starts downloading → auto-assigned to free slot, (2) package completes → slot marked done, (3) next render cycle shows done state briefly, (4) when next package starts, it takes the slot. Or: done entries persist and new packages only take truly empty slots.

- [ ] **F-007: REQ-163 ("final display shall remain showing all completed packages") conflicts with slot-based display when packages > slots.**
  - Evidence: If there are 8 packages and 4 slots, the final display can only show 4 lines. REQ-163 says "remain showing all completed packages" — but with 4 slots, only the last 4 packages' final states would be visible. Earlier completions would have been overwritten by subsequent packages taking their slots.
  - Impact: REQ-163 is either impossible with the slot-based design (when packages > slots), or it means the final render after `stop()` should print all N results (not just the slot count). The requirement and design are inconsistent.
  - Status: open
  - Resolution:
  - Follow-up: Clarify REQ-163: either (a) "final display shows the last N slot states" (weaker but consistent with design), or (b) after all downloads complete, print a summary of all packages (requires additional logic beyond slot rendering).

- [ ] **F-008: Design's `renderSlot` for done+success shows `(%.1f MB)` using `s.Total` but `Total` may be 0 (unknown Content-Length).**
  - Evidence: `renderSlot` code: `fmt.Sprintf("  %s✓%s %s (%.1f MB)", Green, Reset, s.Name, float64(s.Total)/1e6)`. If `Content-Length` was unknown, `Total` is 0, and the output would show `✓ neovim (0.0 MB)` which is misleading.
  - Impact: Low-medium — the done state should use `Downloaded` (actual bytes received) rather than `Total` (declared size) for the final display. Or handle the unknown-total case separately.
  - Status: open
  - Resolution:
  - Follow-up: Design should use `s.Downloaded` for the done+success display, or show `s.Total` only when > 0, falling back to `s.Downloaded`.

### Low

- [ ] **F-009: REQ-160 verification is "Code review — confirm ticker interval is 100ms" — weakest validation.**
  - Evidence: Code review is the only verification for the 100ms refresh rate. No automated test confirms the ticker fires at the correct interval.
  - Impact: Low — the value is a constant and unlikely to regress. Code review is appropriate for this type of requirement.
  - Status: open
  - Resolution:
  - Follow-up: None needed — code review is acceptable for timing constants.

- [ ] **F-010: Problem statement says "Must not change `PreDownload` function signature" but design changes `downloadFile` (internal function) signature.**
  - Evidence: Constraint says "additive changes to `Options` only." Design adds `OnProgress` to `Options` (additive, correct) and changes `downloadFile` signature (internal, not exported). This is consistent — the constraint applies to the public API only.
  - Impact: None — internal function changes are fine. The constraint is satisfied.
  - Status: open
  - Resolution:
  - Follow-up: None needed — constraint is correctly interpreted.

- [ ] **F-011: No task for removing the old single-line counter code.**
  - Evidence: Task 3.1 says "Remove old single-line `Downloading [N/M]...` counter" as a bullet point. This is present but could be missed since it's a sub-bullet rather than a standalone verification item.
  - Impact: Low — it's listed in the task. Just noting it's easy to overlook during implementation.
  - Status: open
  - Resolution:
  - Follow-up: None needed — already in task 3.1.

- [ ] **F-012: Design doesn't specify thread safety for `ProgressWriter.Downloaded` field.**
  - Evidence: `ProgressWriter.Write` increments `Downloaded` and calls `OnProgress`. Each download runs in its own goroutine, so each `ProgressWriter` instance is single-goroutine (one per download). No race condition within a single writer. However, `OnProgress` calls `dd.UpdateProgress` which acquires the display mutex — this is correct.
  - Impact: None — each `ProgressWriter` is owned by one goroutine. No race. The design is correct but doesn't explicitly state this safety property.
  - Status: open
  - Resolution:
  - Follow-up: None needed — single-owner pattern is implicit and correct.

---

## Open Clarification Questions

1. **How does slot assignment work when `AssignSlot` is never called from the integration code?**
   - Context: Design defines `AssignSlot` but the integration code only calls `UpdateProgress` and `MarkDone`. Either `UpdateProgress` must auto-assign, or there's a missing call site.
   - Answer:

2. **What does `MarkDone` do in non-TTY mode — print a line or no-op?**
   - Context: Design says `render()` is a no-op in non-TTY. But REQ-167/168 require per-completion output. If `MarkDone` is also a no-op, nothing prints.
   - Answer:

3. **How should `renderSlot` handle done+success when `Total` is 0 (unknown Content-Length)?**
   - Context: Current design uses `s.Total` for the done display. If Content-Length was unknown, this shows "0.0 MB".
   - Answer:

4. **How does the import cycle between `internal/download` and `internal/cli` get resolved for color constants?**
   - Context: `cli/upgrade.go` imports `download`. If `download/progress.go` imports `cli` for colors, Go will reject the circular import.
   - Answer:

5. **What does REQ-163 mean when packages > slots — show last N slot states, or print all results after completion?**
   - Context: With 8 packages and 4 slots, earlier completions are overwritten. "Final display showing all completed" is ambiguous.
   - Answer:

---

## Validation Notes

- All REQ-157 through REQ-176 IDs are present in requirements, design, and tasks.
- All 20 requirements have at least one task with a verification method.
- All tasks trace back to at least one requirement.
- No orphan tasks or orphan requirements found.
- Traceability is complete across all four planning artifacts.
- Problem statement scope, non-goals, and constraints are reflected in requirements.
- The strongest validation paths are unit tests for `renderSlot` (all states) and `ProgressWriter` byte tracking.
- The weakest validation paths are REQ-160 (code review only) and REQ-163 (ambiguous when packages > slots).
- **Critical blocker:** F-004 (import cycle) must be resolved before implementation — it will cause a build failure.
