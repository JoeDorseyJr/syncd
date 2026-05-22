# Traceability Matrix: Smart Progress Display for Brew Upgrades

## Forward Trace: Requirements → Design → Tasks

| Req ID | Requirement Summary | Design Section | Task IDs | Verification Method |
|--------|-------------------|----------------|----------|-------------------|
| REQ-108 | Single status line per package using `\r` overwrite | Display — `\r` overwrite | 2.2, 3.1, 3.2 | Run upgrade in TTY, confirm status line overwrites |
| REQ-109 | Status line format: `[N/T] name: phase` | Integration — status format | 3.1, 3.2 | Run upgrade with multiple packages, confirm format |
| REQ-110 | Detect phases: downloading, installing, pouring, built | Phase Detector — pattern table | 1.1 | Unit test — known brew output lines → correct phase |
| REQ-111 | Success shows final `✓ name` line (not overwritten) | Integration — `display.Finish` on success | 3.1, 3.2 | Run upgrade, confirm `✓` lines persist |
| REQ-112 | Failure shows `✗ name` + error summary | Integration — `display.Finish` on failure | 3.1, 3.2 | Trigger failure, confirm `✗` + error summary |
| REQ-113 | Extract only relevant error lines from brew output | Error Extractor — `ExtractError` | 2.1 | Unit test — brew failure output → relevant lines only |
| REQ-114 | Kill brew process if no output for 60 seconds | Progress Runner — timer kill | 1.2 | Unit test — mock slow process, confirm killed |
| REQ-115 | Retry once after killing hung process | Progress Runner — single retry | 1.2 | Unit test — first hangs, second succeeds |
| REQ-116 | If retry also hangs/fails, report failure and continue | Progress Runner — retry failure | 1.2 | Unit test — both hang, confirm failure + next proceeds |
| REQ-117 | Hang timeout is 60 seconds (hardcoded) | Progress Runner — `HangTimeout` constant | 1.2 | Code review — confirm 60s constant |
| REQ-118 | Capture brew stdout/stderr to pipes (not stream) | Progress Runner — pipe capture | 1.2 | Run upgrade, confirm no raw brew output in terminal |
| REQ-119 | Read captured output line-by-line in goroutine | Progress Runner — goroutine reader | 1.2 | Code review — confirm goroutine reads from pipe |
| REQ-120 | Store captured output in memory for error extraction | Progress Runner — output buffer | 1.2, 2.1 | Unit test — after failure, full output available |
| REQ-121 | `--verbose` bypasses smart progress (uses RunMutate) | Design Decision 7 — verbose bypass | 3.1, 3.2 | Run `upgrade --verbose`, confirm raw output streams |
| REQ-122 | Non-TTY falls back to simple line-per-phase output | Display — TTY detection | 2.2, 3.2 | Pipe output, confirm no `\r`, one line per phase |
| REQ-123 | Smart progress does not change CommandRunner interface | Design Decision 1 — interface unchanged | 3.1 | Code review — CommandRunner unchanged |
| REQ-124 | Existing tests pass without modification | Testing Strategy — existing tests pass | 3.2 | Run `make test && make integration-test`, all pass |
| REQ-125 | Logic lives in new `internal/progress` package | New Components — `internal/progress/` | 1.1, 1.2, 2.1, 2.2 | Code review — package exists |
| REQ-126 | Progress runner accepts callback for phase updates | Design Decision 2 — callback | 1.2 | Unit test — mock callback receives phases |
| REQ-127 | Display uses `\r` + space-padding to clear previous line | Display — `\r` + padding | 2.2 | Unit test — output contains `\r` and trailing spaces |

## Reverse Trace: Design Sections → Requirement IDs

| Design Section | Requirement IDs |
|----------------|-----------------|
| System Architecture | REQ-108, REQ-114, REQ-118, REQ-125 |
| New Components | REQ-125 |
| Design Decision 1 (new package, interface unchanged) | REQ-123, REQ-125 |
| Design Decision 2 (callback-based) | REQ-126 |
| Design Decision 3 (pipe capture, not PTY) | REQ-118 |
| Design Decision 4 (goroutine + timer) | REQ-114, REQ-119 |
| Design Decision 5 (single retry) | REQ-115, REQ-116 |
| Design Decision 6 (TTY detection) | REQ-122 |
| Design Decision 7 (verbose bypass) | REQ-121 |
| Progress Runner | REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-125, REQ-126 |
| Phase Detector | REQ-110 |
| Error Extractor | REQ-113, REQ-120 |
| Display | REQ-108, REQ-122, REQ-127 |
| Integration with Upgrade Command | REQ-108, REQ-109, REQ-111, REQ-112, REQ-121, REQ-123 |

## Reverse Trace: Task IDs → Requirement IDs

| Task ID | Task Title | Requirement IDs |
|---------|------------|-----------------|
| 1.1 | Phase detector | REQ-110, REQ-125 |
| 1.2 | Progress runner with hang detection | REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-125, REQ-126 |
| 2.1 | Error extractor | REQ-113, REQ-120 |
| 2.2 | Display with TTY detection | REQ-108, REQ-122, REQ-127 |
| 3.1 | Wire progress into upgrade command | REQ-108, REQ-109, REQ-111, REQ-112, REQ-121, REQ-123 |
| 3.2 | Integration tests | REQ-108, REQ-109, REQ-111, REQ-112, REQ-121, REQ-122, REQ-124 |

## Coverage Summary

- **Requirements with tasks:** 20/20 (100%)
- **Requirements with design mapping:** 20/20 (100%)
- **Requirements with verification method:** 20/20 (100%)
- **Orphan tasks (no requirement):** None
- **Orphan requirements (no task):** None
