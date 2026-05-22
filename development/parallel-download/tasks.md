# Tasks: Parallel Pre-Download for syncd Upgrade

## Phase 1: URL Extraction & Cache Naming (~2 hours)

### 1.1 Platform detection
> REQ-132 | Design: Platform Detection

- [ ] Create `internal/download/platform.go`
  - `Platform() string` — returns brew platform string (e.g., `arm64_sonoma`)
  - Use `runtime.GOARCH` for architecture
  - Use `sw_vers -productVersion` for macOS version → codename mapping
  - Map: 15.x→sequoia, 14.x→sonoma, 13.x→ventura, 12.x→monterey
- [ ] Add unit tests in `internal/download/platform_test.go`
  - Test: arm64 + macOS 14.x → "arm64_sonoma"
  - Test: amd64 + macOS 14.x → "sonoma"
  - Test: unknown version → fallback to most recent known
- [ ] Verify: `go test ./internal/download/...` passes

### 1.2 Cache filename generation
> REQ-133, REQ-134, REQ-135, REQ-136 | Design: Cache Filename Generation

- [ ] Create `internal/download/cache.go`
  - `CacheFilename(url, name, version string, isCask bool) string`
  - SHA256 hash of URL string as lowercase hex prefix
  - Formula: `SHA256--name--version.bottle.tar.gz`
  - Cask: `SHA256--name--version.ext` (preserve extension from URL)
  - `CacheDir() string` — returns `~/Library/Caches/Homebrew/downloads/`
- [ ] Add unit tests in `internal/download/cache_test.go`
  - Test: known URL produces expected SHA256 prefix
  - Test: formula filename format correct
  - Test: cask filename preserves `.dmg` extension
  - Test: cask filename preserves `.pkg` extension
  - Test: CacheDir returns correct path
- [ ] Verify: `go test ./internal/download/...` passes

### 1.3 Formula URL extraction
> REQ-128, REQ-129, REQ-130, REQ-153 | Design: URL Extraction

- [ ] Create `internal/download/info.go`
  - `PackageURL` struct: Name, Version, URL, Filename, IsCask
  - `GetFormulaURLs(r runner.CommandRunner, names []string) ([]PackageURL, error)`
  - Run `brew info --json=v2 <names...>`, parse JSON
  - Extract bottle URL for current platform from `bottle.stable.files`
  - Fall back to `all` if platform-specific entry missing
  - Generate `Filename` using `CacheFilename`
- [ ] Add unit tests in `internal/download/info_test.go`
  - Test: extract bottle URL for arm64_sonoma
  - Test: fall back to `all` when platform missing
  - Test: multiple formulae parsed correctly
  - Test: formula with no bottle → skip (no error)
- [ ] Verify: `go test ./internal/download/...` passes

### 1.4 Cask URL extraction
> REQ-131 | Design: URL Extraction

- [ ] Add `GetCaskURLs(r runner.CommandRunner, names []string) ([]PackageURL, error)` to `info.go`
  - Run `brew info --json=v2 --cask <names...>`, parse JSON
  - Extract `url` and `version` fields from each cask
  - Generate `Filename` using `CacheFilename` (isCask=true)
- [ ] Add unit tests
  - Test: extract cask URL with `.dmg` extension
  - Test: extract cask URL with `.pkg` extension
  - Test: multiple casks parsed correctly
- [ ] Verify: `go test ./internal/download/...` passes

**Estimate:** ~2 hours

---

## Phase 2: Parallel Downloader (~2 hours)

### 2.1 Download with atomic placement
> REQ-137, REQ-139, REQ-140, REQ-141, REQ-154, REQ-155 | Design: Parallel Downloader

- [ ] Create `internal/download/download.go`
  - `Result` struct: Package, Err, Bytes, Cached
  - `Options` struct: Concurrency, Display, Verbose
  - `PreDownload(packages []PackageURL, opts Options) []Result`
  - Semaphore via buffered channel (size = Concurrency)
  - Per-package goroutine:
    1. Check if final file exists → skip (Cached=true)
    2. HTTP GET (follows redirects by default)
    3. Write to `filename.downloading`
    4. On success: `os.Rename` to final filename
    5. On error: `os.Remove` the `.downloading` file
  - sync.WaitGroup for completion
  - Thread-safe result collection
- [ ] Add unit tests in `internal/download/download_test.go`
  - Test: successful download creates file at correct path
  - Test: already-cached file skipped (no HTTP request)
  - Test: failed download cleans up `.downloading` file
  - Test: HTTP redirect followed to final URL
  - Test: concurrency limit respected (use test server + timing)
- [ ] Verify: `go test ./internal/download/...` passes

### 2.2 Non-fatal error handling
> REQ-142, REQ-143 | Design: Design Decision 4

- [ ] Ensure `PreDownload` never returns an error — all failures are per-package in `Result.Err`
- [ ] Add unit tests
  - Test: all downloads fail → all Results have Err set, function returns normally
  - Test: mix of success and failure → successful files exist, failed files cleaned up
- [ ] Verify: `go test ./internal/download/...` passes

**Estimate:** ~2 hours

---

## Phase 3: Progress & Integration (~1.5 hours)

### 3.1 Aggregate progress display
> REQ-144, REQ-145, REQ-146, REQ-147 | Design: Progress Display

- [ ] Add progress callback to `PreDownload` (update display after each completion)
  - Normal mode: `display.Status("  Downloading [%d/%d] %.1f MB", completed, total, bytes)`
  - On finish: `display.Finish("  Downloaded %d/%d packages (%.1f MB)", ok, total, bytes)`
  - Verbose mode: print per-package `↓ name` / `✓ name` / `✗ name: error`
- [ ] Verify: progress updates visible during download phase

### 3.2 Wire into upgrade command
> REQ-138, REQ-148, REQ-150, REQ-151 | Design: Integration with Upgrade Command

- [ ] Update `internal/cli/upgrade.go`
  - Add `--concurrency` flag (int, default 4)
  - After confirmation, before upgrade loop:
    - If not verbose and packages exist:
      - Call `download.GetFormulaURLs(r, brewNames)`
      - Call `download.GetCaskURLs(r, caskNames)`
      - Combine URLs
      - Call `download.PreDownload(urls, opts)`
      - Print warnings for any failures
  - Skip entirely when `runner.Verbose` is true
- [ ] Verify: `syncd upgrade --yes` pre-downloads before upgrading
- [ ] Verify: `syncd upgrade --verbose` skips pre-download
- [ ] Verify: `syncd upgrade --concurrency 2` limits workers

### 3.3 Integration tests
> REQ-136, REQ-142, REQ-148, REQ-150, REQ-156 | Design: Integration Tests

- [ ] Add fake brew support for `info --json=v2` subcommand
  - Return JSON with bottle URLs pointing to a test HTTP server
- [ ] Add integration test: upgrade pre-downloads files to cache directory
- [ ] Add integration test: download failure → upgrade still succeeds
- [ ] Add integration test: `--verbose` skips pre-download phase
- [ ] Add integration test: already-cached files skipped
- [ ] Verify: `make test && make integration-test` passes (all existing tests green)

**Estimate:** ~1.5 hours

---

## Summary

| Phase | Estimate | Key Deliverable |
|-------|----------|-----------------|
| 1. URL Extraction & Cache Naming | ~2h | Extract URLs, generate correct cache filenames |
| 2. Parallel Downloader | ~2h | Download files in parallel with atomic placement |
| 3. Progress & Integration | ~1.5h | Wire into upgrade command with progress display |
| **Total** | **~5.5h** | |

---

## Coverage Summary

### Requirements Covered

All 29 requirements (REQ-128 through REQ-156) are covered by tasks above.

| Requirement Range | Phase | Tasks |
|-------------------|-------|-------|
| REQ-128–131 | Phase 1 | 1.3, 1.4 |
| REQ-132 | Phase 1 | 1.1 |
| REQ-133–136 | Phase 1 | 1.2 |
| REQ-137–141 | Phase 2 | 2.1 |
| REQ-142–143 | Phase 2 | 2.2 |
| REQ-144–147 | Phase 3 | 3.1 |
| REQ-148–151 | Phase 3 | 3.2 |
| REQ-152–156 | Phase 1, 2, 3 | 1.3, 2.1, 3.3 |

### Uncovered Requirements

None — all 29 requirements have at least one task with a concrete verification method.
