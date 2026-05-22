# Traceability Matrix: Multi-Line Download Progress Bars

## Forward Trace: Requirements → Design → Tasks

| Req ID | Requirement Summary | Design Section | Task IDs | Verification Method |
|--------|-------------------|----------------|----------|-------------------|
| REQ-157 | One progress line per active download slot | DownloadDisplay — slot-based display | 2.1, 3.1 | Run `syncd upgrade --yes` with 4+ packages, confirm 4 lines |
| REQ-158 | Progress line format: arrow, name, bar, %, bytes | Render Logic — `renderSlot` format | 2.1 | Unit test — render at 45% of 50 MB, confirm format |
| REQ-159 | ANSI cursor-up escape codes for redraw | Render Logic — cursor-up codes | 2.1, 3.2 | Integration test — output contains `\033[` in TTY |
| REQ-160 | Refresh display every 100ms | DownloadDisplay — 100ms ticker | 2.1 | Code review — ticker interval is 100ms |
| REQ-161 | Completed slot shows ✓, next package takes slot | DownloadDisplay — `MarkDone` + slot transition | 2.1 | Unit test — complete download, slot transitions |
| REQ-162 | Failed slot shows ✗ + error, next package takes slot | DownloadDisplay — `MarkDone` with error | 2.1 | Unit test — fail download, slot shows error |
| REQ-163 | Final display persists showing all completed | Render Logic — final state persists | 2.1, 3.2 | Integration test — final lines persist after phase |
| REQ-164 | Track bytes per file via writer wrapper | ProgressWriter — byte tracking | 1.1 | Unit test — write through wrapper, byte count updates |
| REQ-165 | Read Content-Length for total file size | ProgressWriter — `resp.ContentLength` | 1.1 | Unit test — mock response with Content-Length |
| REQ-166 | Unknown Content-Length: bytes only, no bar | Render Logic — unknown total branch | 2.1 | Unit test — render with unknown total, no bar |
| REQ-167 | Non-TTY: one line per completion, no cursor codes | DownloadDisplay — non-TTY skip | 2.2, 3.2 | Pipe output, confirm no ANSI cursor codes |
| REQ-168 | Non-TTY output shows name + byte count | Integration — OnComplete prints name + size | 2.2, 3.2 | Pipe output, confirm name and size per line |
| REQ-169 | OnProgress callback in Options struct | Options Change — `OnProgress` field | 1.1 | Code review — Options has OnProgress |
| REQ-170 | PreDownload calls OnProgress during download | ProgressWriter — calls OnProgress | 1.1 | Unit test — mock OnProgress, confirm called |
| REQ-171 | Logic in `internal/download/progress.go` | Technical — file location | 1.1, 2.1 | Code review — file exists |
| REQ-172 | OnComplete callback unchanged | Integration — OnComplete unchanged | 1.1, 3.1, 3.2 | Existing integration tests pass |
| REQ-173 | Verbose mode skips pre-download (no change) | Integration — verbose bypass | 3.1, 3.2 | Run `--verbose`, no download progress |
| REQ-174 | Bar: 20-char width, █ filled, ░ empty | Render Logic — 20-char bar with █/░ | 2.1 | Unit test — 50% = 10█ + 10░ |
| REQ-175 | TTY detection via `golang.org/x/term` | DownloadDisplay — `golang.org/x/term` | 2.1 | Code review — no new TTY dependency |
| REQ-176 | Reuse `cli.Green`, `cli.Red` for indicators | Render Logic — `cli.Green`, `cli.Red` | 2.1 | Code review — imports cli colors |

## Reverse Trace: Design Sections → Requirement IDs

| Design Section | Requirement IDs |
|----------------|-----------------|
| System Architecture | REQ-157, REQ-164, REQ-169 |
| New Components | REQ-164, REQ-157, REQ-171 |
| Design Decision 1 (single file) | REQ-171 |
| Design Decision 2 (OnProgress on Options) | REQ-169, REQ-172 |
| Design Decision 3 (ProgressWriter wraps body) | REQ-164, REQ-165, REQ-170 |
| Design Decision 4 (slot-based display) | REQ-157, REQ-161, REQ-162 |
| Design Decision 5 (100ms ticker) | REQ-160 |
| Design Decision 6 (ANSI cursor-up) | REQ-159 |
| Design Decision 7 (non-TTY fallback) | REQ-167, REQ-168 |
| Options Change | REQ-169 |
| ProgressWriter | REQ-164, REQ-165, REQ-170 |
| DownloadDisplay | REQ-157, REQ-160, REQ-161, REQ-162, REQ-163, REQ-167, REQ-175 |
| Render Logic | REQ-158, REQ-159, REQ-163, REQ-166, REQ-174, REQ-176 |
| Integration with Upgrade Command | REQ-168, REQ-172, REQ-173 |

## Reverse Trace: Task IDs → Requirement IDs

| Task ID | Task Title | Requirement IDs |
|---------|------------|-----------------|
| 1.1 | Add OnProgress to Options and ProgressWriter | REQ-164, REQ-165, REQ-169, REQ-170, REQ-171, REQ-172 |
| 2.1 | Slot management and render logic | REQ-157, REQ-158, REQ-159, REQ-160, REQ-161, REQ-162, REQ-163, REQ-166, REQ-174, REQ-175, REQ-176 |
| 2.2 | Non-TTY fallback | REQ-167, REQ-168 |
| 3.1 | Wire into upgrade command | REQ-157, REQ-172, REQ-173 |
| 3.2 | Integration tests | REQ-159, REQ-163, REQ-167, REQ-168, REQ-172, REQ-173 |

## Coverage Summary

- **Requirements with tasks:** 20/20 (100%)
- **Requirements with design mapping:** 20/20 (100%)
- **Requirements with verification method:** 20/20 (100%)
- **Orphan tasks (no requirement):** None
- **Orphan requirements (no task):** None
