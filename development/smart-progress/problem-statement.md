# Smart Progress Display for Brew Upgrades

Replace raw brew output streaming with captured, parsed progress display and hang detection.

## Context

`syncd upgrade` currently uses `RunMutate` to stream brew's stdout/stderr directly to the terminal. This produces ugly, verbose output that users can't easily scan. Brew sometimes hangs during downloads or post-install steps with no indication to the user. There's no timeout, no retry, and no way to extract just the error message when something fails.

The upgrade loop in `internal/cli/upgrade.go` calls `r.RunMutate("brew", "upgrade", name)` per package, which pipes raw brew output to the terminal. This gives zero control over presentation.

## Current Behavior

- `RunMutate` connects brew's stdout/stderr directly to `os.Stdout`/`os.Stderr`
- Raw brew output floods the terminal (download progress bars, dependency info, compilation logs)
- If brew hangs (network stall, stuck post-install script), syncd waits forever
- No way to show a clean per-package progress summary
- Error messages buried in pages of brew output

## Desired Behavior

- Capture brew's stdout/stderr to a pipe instead of streaming raw
- Read output line-by-line in a goroutine
- Parse lines to detect phases: downloading, installing, done
- Display single-line progress updates using `\r` (carriage return overwrite)
- Detect hangs: if no output for N seconds, kill the process and retry once
- On failure, extract and display only the relevant error lines
- Clean UX: one status line per package that updates in-place

## Success Outcomes

1. `syncd upgrade` shows clean single-line progress per package (e.g., `  [1/5] neovim: downloading...`)
2. Status line updates in-place using `\r` — no scrolling wall of text
3. If brew produces no output for 60 seconds, the process is killed and retried once
4. On failure, only the error summary is shown (not full brew dump)
5. On success, final line shows `✓ package` — same as current but without the noise before it

## Scope

- New `RunProgress` method on `ExecRunner` (or new package `internal/progress`)
- Line-by-line output capture with phase detection
- Hang detection with configurable timeout
- Kill + single retry on hang
- `\r` carriage-return progress display
- Integration into `upgrade` command loop
- Verbose mode (`--verbose`) bypasses smart progress and streams raw (existing behavior)

## Non-Goals (this milestone)

- Progress bars with actual percentage (brew doesn't reliably report %)
- Parallel package upgrades
- Applying smart progress to `apply` command (future)
- Custom progress for `brew update` (already has its own spinner)
- Windows/Linux support

## Constraints

- Must not break `--verbose` mode (verbose = raw stream, same as today)
- Must not break `MockRunner` or existing tests
- `CommandRunner` interface should not change (add new method or use separate type)
- Timeout/retry only applies to upgrade operations, not reads
- Must handle non-TTY gracefully (fall back to simple line output)

## Open Questions

1. ~~What timeout value for hang detection?~~ **Resolved:** 60 seconds default, no config needed for v1.
2. ~~How many retries on hang?~~ **Resolved:** One retry. If it hangs again, report failure.
3. ~~Should smart progress apply to `apply` command too?~~ **Resolved:** No, upgrade only for now. Apply can adopt later.
