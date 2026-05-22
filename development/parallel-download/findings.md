# Findings: Parallel Pre-Download for syncd Upgrade

## Issues

### High

- [ ] **F-001: `CacheDir()` hardcodes `~/Library/Caches/Homebrew/downloads/` but Homebrew respects `HOMEBREW_CACHE` env var.**
  - Evidence: Design says `CacheDir() string` returns `~/Library/Caches/Homebrew/downloads/`. However, Homebrew allows users to override the cache directory via `HOMEBREW_CACHE` environment variable. If a user has this set, syncd would write to the wrong directory and brew would not get cache hits.
  - Impact: Pre-downloaded files would be invisible to brew on systems with custom `HOMEBREW_CACHE`. The feature would silently fail to provide speedup — brew would re-download everything.
  - Status: open
  - Resolution:
  - Follow-up: Design should specify: check `HOMEBREW_CACHE` env var first, fall back to `~/Library/Caches/Homebrew/downloads/`. Alternatively, run `brew --cache` to get the authoritative path.

- [ ] **F-002: Design says `GetFormulaURLs` and `GetCaskURLs` use `runner.CommandRunner` but the design's integration code passes `r` (an `*ExecRunner`) — no testability concern, but `brew info --json=v2` with many packages may exceed command-line length limits.**
  - Evidence: Design says "Query all outdated packages in one `brew info --json=v2 pkg1 pkg2 ...` call." On macOS, `ARG_MAX` is ~262144 bytes. With 100+ packages (each name ~20 chars), this is fine. But the design doesn't specify behavior when the package list is empty (0 outdated).
  - Impact: If `brewNames` or `caskNames` is empty, calling `brew info --json=v2` with no package names may produce unexpected output or an error. The design should specify: skip the call when the list is empty.
  - Status: open
  - Resolution:
  - Follow-up: Design and task 1.3/1.4 should specify: return empty slice immediately if `names` is empty.

- [ ] **F-003: Design's `PreDownload` imports `progress.Display` directly, creating a coupling between `internal/download` and `internal/progress`.**
  - Evidence: Design shows `Options` struct with `Display *progress.Display`. This means `internal/download` imports `internal/progress`. The design decision says "Keeps download logic separate from progress, runner, and CLI" but then couples it to `progress.Display`.
  - Impact: Unit testing `PreDownload` requires either a real `progress.Display` or nil. The coupling is mild (Display is simple) but contradicts the stated separation goal. More importantly, if `Display` is nil (e.g., in tests), the code must nil-check before every `display.Status()` call.
  - Status: open
  - Resolution:
  - Follow-up: Consider accepting a callback `func(completed, total int, bytes int64)` instead of `*progress.Display`, matching the callback pattern used in `RunWithProgress`. This keeps `download` decoupled from `progress`.

### Medium

- [ ] **F-004: REQ-149 verification ("pre-download a bottle, run `brew upgrade`, confirm downloading phase is skipped") is not achievable in integration tests with fake brew.**
  - Evidence: REQ-149 says "Integration test — pre-download a bottle, run `brew upgrade`, confirm 'downloading' phase is skipped." Integration tests use a fake brew script, not real Homebrew. The fake brew doesn't actually check the cache directory for pre-downloaded files — it just prints scripted output.
  - Impact: REQ-149 cannot be validated through the existing fake brew integration test infrastructure. It requires either: (a) a real Homebrew test (slow, non-deterministic), or (b) the fake brew script checking for cache files and adjusting output accordingly.
  - Status: open
  - Resolution:
  - Follow-up: Either enhance fake brew to check `$FAKE_BREW_CACHE` for files and skip "Downloading" output, or reclassify REQ-149 verification as "manual test" or "code review — confirm atomic rename produces correct filename."

- [ ] **F-005: Platform detection uses `sw_vers -productVersion` via shell command but design doesn't specify how this is testable.**
  - Evidence: Task 1.1 says "Unit test — arm64 + macOS 14.x → arm64_sonoma" but `Platform()` calls `sw_vers` directly. The function signature `Platform() string` has no dependency injection for the version query.
  - Impact: Unit tests cannot mock `sw_vers` output without either: (a) making `Platform()` accept a `CommandRunner`, (b) extracting the version query into a package-level variable/function that tests can override, or (c) using environment-based test fixtures.
  - Status: open
  - Resolution:
  - Follow-up: Design should specify testability approach — either accept a `CommandRunner` parameter or use a package-level `var platformFunc` that tests can replace.

- [ ] **F-006: Design doesn't specify HTTP timeout for downloads.**
  - Evidence: `PreDownload` uses `net/http` for downloads but no timeout is specified. Go's default `http.Client` has no timeout — it will wait indefinitely. A stalled CDN connection could block a goroutine forever.
  - Impact: A single stalled download could hold a semaphore slot indefinitely, reducing effective concurrency. In the worst case, all slots could be blocked by stalled connections, making the download phase hang.
  - Status: open
  - Resolution:
  - Follow-up: Design should specify an HTTP client timeout (e.g., 5 minutes per download) or a context with deadline per download goroutine.

- [ ] **F-007: No requirement or design for handling `brew info --json=v2` failure.**
  - Evidence: Design shows `GetFormulaURLs` and `GetCaskURLs` returning `([]PackageURL, error)`. But the integration code in the design doesn't show error handling for these calls. If `brew info` fails (e.g., network error during API call to GitHub), the entire download phase would error.
  - Impact: Since the design principle is "non-fatal failures," a `brew info` failure should also be non-fatal — skip pre-download and let brew handle everything. But this isn't explicitly stated.
  - Status: open
  - Resolution:
  - Follow-up: Design integration section should specify: if `GetFormulaURLs` or `GetCaskURLs` returns an error, print a warning and skip pre-download (don't fail the upgrade).

- [ ] **F-008: REQ-137 verification says "Integration test — 4 packages download concurrently, not sequentially" but proving concurrency in tests is inherently flaky.**
  - Evidence: Verifying that downloads happen concurrently (not sequentially) requires timing-based assertions. If the test server responds instantly, 4 sequential downloads are indistinguishable from 4 concurrent ones.
  - Impact: The verification method is weak. A better approach: use a test server with artificial delay (e.g., 100ms per download), measure total time, assert it's less than N*100ms. Or verify via concurrency counter (atomic int tracking active downloads).
  - Status: open
  - Resolution:
  - Follow-up: Task 2.1 test "concurrency limit respected (use test server + timing)" is the right approach but should be more specific about the assertion strategy.

- [ ] **F-009: Design shows `progress.Display` being created in the non-verbose path of upgrade, but `PreDownload` also needs it — potential double-creation or ordering issue.**
  - Evidence: Current `upgrade.go` creates `display := progress.NewDisplay()` inside the `else` (non-verbose) block, before the upgrade loop. The design's integration code shows `display` being passed to `PreDownload` options. But `display` is created after the verbose check — the design code snippet shows `if !runner.Verbose { ... display ... }` which is correct, but the existing code creates `display` deeper in the block.
  - Impact: Low — the integration is straightforward. But the design should clarify that `display` is created once and shared between the download phase and the upgrade loop.
  - Status: open
  - Resolution:
  - Follow-up: Minor — design integration section could note "reuse the same Display instance for both download progress and upgrade progress."

- [ ] **F-010: No requirement for what happens when `--concurrency` is set to 0 or a negative number.**
  - Evidence: REQ-138 says `--concurrency` controls workers, REQ-151 says default is 4. No requirement specifies validation of the flag value.
  - Impact: If a user passes `--concurrency 0` or `--concurrency -1`, the buffered channel semaphore would have size 0 or panic. The code needs input validation.
  - Status: open
  - Resolution:
  - Follow-up: Design should specify: validate `concurrency >= 1`, default to 4 if invalid. Or clamp to minimum 1.

- [ ] **F-011: Design's verbose mode shows download messages (`↓ neovim`, `✓ neovim downloaded`) but REQ-150 says "skip the parallel download phase entirely when `--verbose` is set."**
  - Evidence: REQ-150: "syncd shall skip the parallel download phase entirely when `--verbose` is set." Design Progress Display section shows verbose output format: `↓ neovim (12.3 MB)` / `✓ neovim downloaded`. These are contradictory — if verbose skips pre-download, there's nothing to show verbose messages for.
  - Impact: The design's verbose progress display section is dead code/spec. REQ-147 ("verbose mode shall print individual download start/complete messages per package") conflicts with REQ-150 ("skip download phase when verbose").
  - Status: open
  - Resolution:
  - Follow-up: Either REQ-147 should be removed/reworded (since verbose skips download), or REQ-150 should be changed to "verbose mode shows per-package messages instead of aggregate progress." These two requirements are mutually exclusive as written.

### Low

- [ ] **F-012: Design's macOS version mapping only covers 12.x–15.x. macOS 16.x will need an update.**
  - Evidence: Platform detection maps 15.x→sequoia, 14.x→sonoma, 13.x→ventura, 12.x→monterey. No fallback for future versions.
  - Impact: When macOS 16 ships, `Platform()` would return an empty/incorrect string, causing bottle URL extraction to fail (no matching platform key). Task 1.1 mentions "unknown version → fallback to most recent known" which addresses this, but the design doesn't specify this fallback.
  - Status: open
  - Resolution:
  - Follow-up: Design should document the fallback behavior for unknown versions (use most recent known codename).

- [ ] **F-013: `CacheFilename` for formulae uses `.bottle.tar.gz` but modern Homebrew uses `.bottle.tar.gz` or `.bottle_manifest` depending on version.**
  - Evidence: Design says format is `SHA256--name--version.bottle.tar.gz`. Modern Homebrew (4.x+) uses GitHub Packages (GHCR) for bottles, and the actual filename in the cache may differ. The URL itself determines the filename extension.
  - Impact: Low — the SHA256 prefix is the primary cache key. If the extension doesn't match exactly, brew may not recognize the cache hit. This needs validation against actual brew cache behavior.
  - Status: open
  - Resolution:
  - Follow-up: Validate against real brew cache entries. The safest approach may be to derive the extension from the URL path rather than hardcoding `.bottle.tar.gz`.

- [ ] **F-014: REQ-145 and REQ-146 verification methods are weak — "Run in TTY, confirm in-place updates" and "Pipe output, confirm no `\r`."**
  - Evidence: These are essentially manual tests. The existing `progress.Display` already handles TTY/non-TTY, and task 3.1 reuses it. Since Display is already tested in the smart-progress milestone, these requirements are implicitly covered.
  - Impact: Informational. No additional test needed beyond confirming Display is used correctly.
  - Status: open
  - Resolution:
  - Follow-up: None needed — covered by existing Display tests.

- [ ] **F-015: Task 3.3 integration test "already-cached files skipped" overlaps with unit test in task 2.1.**
  - Evidence: Task 2.1 has "Test: already-cached file skipped (no HTTP request)" as a unit test. Task 3.3 has "Add integration test: already-cached files skipped." Both test the same behavior at different levels.
  - Impact: Redundant but not harmful. The integration test adds confidence that the full flow respects caching.
  - Status: open
  - Resolution:
  - Follow-up: None needed — defense in depth is acceptable.

---

## Open Clarification Questions

1. **Should `CacheDir()` respect `HOMEBREW_CACHE` env var or use `brew --cache` to determine the correct path?**
   - Context: Homebrew allows overriding the cache directory. Hardcoding `~/Library/Caches/Homebrew/downloads/` will fail for users with custom cache paths. Options: (a) check `HOMEBREW_CACHE` env, append `/downloads/`; (b) run `brew --cache` and use its output; (c) hardcode and document as limitation.
   - Answer: (b) Run `brew --cache` to get the correct path.

2. **Should `GetFormulaURLs`/`GetCaskURLs` return early with empty slice when given an empty names list?**
   - Context: If there are no outdated formulae but there are outdated casks (or vice versa), one call would receive an empty list. Running `brew info --json=v2` with no arguments may produce unexpected output.
   - Answer: Yes, return early with empty slice when given empty names.

3. **How should `brew info --json=v2` failure be handled — skip pre-download silently, print warning, or propagate error?**
   - Context: The design principle is non-fatal failures for downloads, but `brew info` is a prerequisite step, not a download. If it fails, no URLs are available.
   - Answer: Print warning, skip pre-download, let brew handle downloads normally.

4. **Are REQ-147 and REQ-150 mutually exclusive? If verbose skips pre-download, what does REQ-147's "verbose per-package messages" refer to?**
   - Context: REQ-150 says skip download phase when verbose. REQ-147 says verbose shows per-package download messages. These cannot both be true simultaneously.
   - Answer: Remove REQ-147. Verbose skips pre-download entirely — brew handles it with its own output.

5. **Should there be an HTTP timeout per download to prevent goroutine leaks on stalled connections?**
   - Context: Go's default `http.Client` has no timeout. A stalled CDN could block a worker indefinitely. Options: (a) per-request timeout (e.g., 5 min), (b) context with deadline, (c) accept the risk since brew will eventually handle it.
   - Answer: (a) Per-request timeout of 5 minutes.

6. **Should `--concurrency` validate its input (reject 0 or negative values)?**
   - Context: A buffered channel with size 0 would deadlock. Negative values would panic. The flag needs bounds checking.
   - Answer: Yes, clamp to minimum 1.

---

## Validation Notes

- All REQ-128 through REQ-156 IDs are present in requirements, design, and tasks.
- All 29 requirements have at least one task with a verification method.
- All tasks trace back to at least one requirement.
- No orphan tasks or orphan requirements found.
- Traceability is complete across all four planning artifacts.
- Problem statement scope, non-goals, and constraints are reflected in requirements.
- Design decisions are well-reasoned and follow established project patterns (same as v0.3, smart-progress).
- The strongest validation paths are unit tests for JSON parsing/filename generation and integration tests with fake brew + test HTTP server.
- The weakest validation paths are REQ-149 (cache hit verification requires real brew or enhanced fake), REQ-145/146 (TTY behavior — covered by existing Display tests), and REQ-151 (code review only).
- REQ-147 and REQ-150 are contradictory and need resolution before implementation.
