# Multi-Line Download Progress Bars

Replace single-line aggregate download counter with per-file progress bars like docker pull.

## Context

`syncd upgrade` pre-downloads bottles/casks in parallel before running `brew upgrade`. The current progress display shows a single aggregate counter (`Downloading [3/8]...`) that updates in-place with `\r`. This gives no visibility into individual download progress — users can't tell which package is slow, how large each download is, or how far along each one is.

Docker's `pull` command shows one line per concurrent layer download with a progress bar, percentage, and byte counts. This is the gold standard for parallel download UX.

## Current Behavior

- `PreDownload` calls `OnComplete` callback after each download finishes
- Upgrade command shows: `  Downloading [3/8]...` (single `\r` line)
- No per-file progress, no byte counts during download, no indication of which packages are active
- `io.Copy` streams the entire response body with no progress tracking

## Desired Behavior

- Show N lines (one per concurrency slot), each with a progress bar:
  ```
  ↓ neovim       ██████████░░░░░░░░  45% (22/50 MB)
  ↓ firefox      ████░░░░░░░░░░░░░░  18% (45/250 MB)
  ↓ ripgrep      ██████████████████ 100% (3.2/3.2 MB)
  ↓ node@22      ░░░░░░░░░░░░░░░░░░   2% (1.1/48 MB)
  ```
- Use ANSI cursor-up (`\033[NA`) to redraw all N lines every 100ms
- When a download finishes (`✓`), the next queued package takes its slot
- When all done, leave final state showing all completed
- Handle unknown `Content-Length`: show bytes only, no bar
- Non-TTY: one line per completion, no redraws

## Success Outcomes

1. Users see real-time per-file download progress during `syncd upgrade`
2. Slow downloads are immediately visible (which package, how far along)
3. The display is clean and doesn't scroll — fixed N lines that update in place
4. Non-TTY environments get sensible fallback output
5. No functional change to download behavior — purely a display improvement

## Scope

- New multi-line progress renderer using ANSI cursor-up codes
- Progress-tracking writer wrapper for `io.Copy`
- Integration into `PreDownload` via progress callback
- TTY/non-TTY detection and fallback

## Non-Goals (this milestone)

- Download speed/ETA display
- Bandwidth throttling
- Resumable downloads
- Changes to download concurrency logic
- Changes to cache filename generation

## Constraints

- Must not change `PreDownload` function signature (additive changes to `Options` only)
- Must reuse existing TTY detection (`golang.org/x/term`)
- Must reuse existing color variables from `internal/cli/output.go`
- Verbose mode still skips pre-download entirely (no change)
- 100ms refresh rate to avoid terminal flicker

## Open Questions

None — all resolved in the task description.
