# Traceability Matrix: Parallel Pre-Download for syncd Upgrade

## Forward Trace: Requirements → Design → Tasks

| Req ID | Requirement Summary | Design Section | Task IDs | Verification Method |
|--------|-------------------|----------------|----------|-------------------|
| REQ-128 | Query `brew info --json=v2` for bottle URLs | URL Extraction — `GetFormulaURLs` | 1.3 | Unit test — parse sample JSON, confirm URLs extracted |
| REQ-129 | Extract bottle URL for current platform | URL Extraction — platform selection from `bottle.stable.files` | 1.3 | Unit test — JSON with multiple OS entries, correct platform selected |
| REQ-130 | Fall back to `all` bottle URL | URL Extraction — `all` fallback | 1.3 | Unit test — JSON with only `all` entry, URL extracted |
| REQ-131 | Extract cask download URLs | URL Extraction — `GetCaskURLs` | 1.4 | Unit test — parse cask JSON, URL from `url` field |
| REQ-132 | Determine macOS arch + version for platform | Platform Detection — `Platform()` | 1.1 | Unit test — mock platform detection, correct string |
| REQ-133 | Cache filename format for formulae | Cache Filename Generation — `CacheFilename` (formula) | 1.2 | Unit test — given URL/name/version, filename matches |
| REQ-134 | SHA256 hash of URL as filename prefix | Cache Filename Generation — SHA256 of URL | 1.2 | Unit test — known URL produces expected hex prefix |
| REQ-135 | Cache filename format for casks (preserve ext) | Cache Filename Generation — `CacheFilename` (cask) | 1.2 | Unit test — cask URL `.dmg` preserved in filename |
| REQ-136 | Place files in `~/Library/Caches/Homebrew/downloads/` | Cache Filename Generation — `CacheDir()` | 1.2, 3.3 | Integration test — file exists at expected cache path |
| REQ-137 | Parallel download with configurable concurrency | Parallel Downloader — worker pool with semaphore | 2.1 | Integration test — 4 packages download concurrently |
| REQ-138 | `--concurrency` flag on upgrade command | Integration — `--concurrency` flag | 3.2 | Run with `--concurrency 2`, confirm limit |
| REQ-139 | Skip already-cached files | Parallel Downloader — skip existing files | 2.1 | Unit test — file exists, no HTTP request |
| REQ-140 | `.downloading` suffix + rename on completion | Parallel Downloader — `.downloading` suffix + rename | 2.1 | Unit test — suffix during download, final name after |
| REQ-141 | Delete `.downloading` on failure | Parallel Downloader — delete on failure | 2.1 | Unit test — HTTP error, `.downloading` removed |
| REQ-142 | Upgrade not failed if all downloads fail | Design Decision 4 — non-fatal failures | 2.2, 3.3 | Integration test — all fail, brew upgrade still runs |
| REQ-143 | Download failures as warnings | Design Decision 4 — warnings not errors | 2.2, 3.3 | Integration test — one fails, warning printed |
| REQ-144 | Aggregate progress: completed/total + bytes | Progress Display — aggregate status line | 3.1 | Integration test — progress line shows counts + MB |
| REQ-145 | `\r` overwrite in TTY mode | Progress Display — `\r` overwrite via `progress.Display` | 3.1 | Run in TTY, confirm in-place updates |
| REQ-146 | Simple line output in non-TTY | Progress Display — non-TTY fallback | 3.1 | Pipe output, confirm no `\r` |
| REQ-147 | Verbose: per-package download messages | Progress Display — verbose per-package messages | 3.1 | Run with `--verbose`, per-package messages visible |
| REQ-148 | Download phase after confirm, before upgrade | Integration — after confirm, before loop | 3.2, 3.3 | Integration test — download between confirm and upgrade |
| REQ-149 | Cache hits for pre-downloaded packages | Design Decision 3 — atomic placement = cache hit | 2.1 | Integration test — pre-download, brew skips download |
| REQ-150 | Skip download phase when `--verbose` | Design Decision 5 — skip in verbose mode | 3.2 | Run `--verbose`, no parallel download phase |
| REQ-151 | Default concurrency = 4 | Integration — default concurrency 4 | 3.2 | Code review — default value is 4 |
| REQ-152 | Logic in `internal/download/` package | Design Decision 1 — `internal/download/` package | 1.1, 1.2, 1.3, 1.4, 2.1 | Code review — package exists |
| REQ-153 | URL extraction uses `CommandRunner` | URL Extraction — uses `CommandRunner` | 1.3, 1.4 | Code review — `runner.Run` used |
| REQ-154 | HTTP via `net/http` only | Parallel Downloader — `net/http` only | 2.1 | Code review — only `net/http` |
| REQ-155 | Follow HTTP redirects | Parallel Downloader — Go http.Client follows redirects | 2.1 | Unit test — mock redirect, final URL downloaded |
| REQ-156 | Existing tests pass without modification | Testing Strategy — existing tests pass | 3.3 | Run `make test && make integration-test`, all pass |

## Reverse Trace: Design Sections → Requirement IDs

| Design Section | Requirement IDs |
|----------------|-----------------|
| System Architecture | REQ-128, REQ-137, REQ-148, REQ-152 |
| New Components | REQ-128, REQ-131, REQ-132, REQ-133, REQ-137, REQ-152 |
| Design Decision 1 (new package) | REQ-152 |
| Design Decision 2 (worker pool semaphore) | REQ-137 |
| Design Decision 3 (atomic file placement) | REQ-140, REQ-141, REQ-149 |
| Design Decision 4 (non-fatal failures) | REQ-142, REQ-143 |
| Design Decision 5 (skip verbose) | REQ-150 |
| Design Decision 6 (reuse progress.Display) | REQ-144, REQ-145, REQ-146 |
| Design Decision 7 (single brew info call) | REQ-128, REQ-131 |
| URL Extraction | REQ-128, REQ-129, REQ-130, REQ-131, REQ-153 |
| Platform Detection | REQ-132 |
| Cache Filename Generation | REQ-133, REQ-134, REQ-135, REQ-136 |
| Parallel Downloader | REQ-137, REQ-139, REQ-140, REQ-141, REQ-154, REQ-155 |
| Progress Display | REQ-144, REQ-145, REQ-146, REQ-147 |
| Integration with Upgrade Command | REQ-138, REQ-148, REQ-150, REQ-151 |

## Reverse Trace: Task IDs → Requirement IDs

| Task ID | Task Title | Requirement IDs |
|---------|------------|-----------------|
| 1.1 | Platform detection | REQ-132, REQ-152 |
| 1.2 | Cache filename generation | REQ-133, REQ-134, REQ-135, REQ-136, REQ-152 |
| 1.3 | Formula URL extraction | REQ-128, REQ-129, REQ-130, REQ-152, REQ-153 |
| 1.4 | Cask URL extraction | REQ-131, REQ-152, REQ-153 |
| 2.1 | Download with atomic placement | REQ-137, REQ-139, REQ-140, REQ-141, REQ-149, REQ-152, REQ-154, REQ-155 |
| 2.2 | Non-fatal error handling | REQ-142, REQ-143 |
| 3.1 | Aggregate progress display | REQ-144, REQ-145, REQ-146, REQ-147 |
| 3.2 | Wire into upgrade command | REQ-138, REQ-148, REQ-150, REQ-151 |
| 3.3 | Integration tests | REQ-136, REQ-142, REQ-143, REQ-148, REQ-150, REQ-156 |

## Coverage Summary

- **Requirements with tasks:** 29/29 (100%)
- **Requirements with design mapping:** 29/29 (100%)
- **Requirements with verification method:** 29/29 (100%)
- **Orphan tasks (no requirement):** None
- **Orphan requirements (no task):** None
