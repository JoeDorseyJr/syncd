# Parallel Pre-Download for syncd Upgrade

Pre-download brew bottles and casks in parallel before running brew upgrade to eliminate sequential download bottleneck.

## Context

`syncd upgrade` currently runs `brew upgrade <name>` sequentially per package. Each invocation downloads its bottle/cask one at a time before installing. For users with many outdated packages, the download phase dominates wall-clock time — even though the downloads are independent and could run concurrently.

Homebrew uses a cache directory (`~/Library/Caches/Homebrew/downloads/`) with a deterministic filename format: `SHA256_of_URL--name--version.bottle.tar.gz` for formulae and `SHA256_of_URL--name--version.ext` for casks. If a file already exists in this directory with the correct name, `brew upgrade` skips the download entirely and pours/installs from cache.

This means syncd can pre-download all bottles/casks in parallel using Go's HTTP client and goroutines, place them in brew's cache with the correct filenames, and then run `brew upgrade` which will get instant cache hits.

## Current Behavior

- `syncd upgrade` runs `brew upgrade <name>` one package at a time
- Each `brew upgrade` downloads its bottle sequentially (one HTTP connection)
- Total download time = sum of all individual download times
- No parallelism in the download phase
- Smart progress shows per-package phase detection but can't speed up downloads

## Desired Behavior

- Before running `brew upgrade`, syncd queries `brew info --json=v2` for all outdated packages
- Extracts bottle/cask download URLs from the JSON response
- Downloads all files in parallel (configurable concurrency, default 4)
- Places files in `~/Library/Caches/Homebrew/downloads/` with correct SHA256-prefixed filenames
- Shows aggregate download progress (total bytes, packages completed)
- Then runs `brew upgrade` per package — each gets a cache hit and skips download
- Download failures are non-fatal: brew falls back to its own download for that package

## Success Outcomes

1. `syncd upgrade` downloads all bottles/casks in parallel before upgrading
2. Total download time ≈ time of slowest single download (not sum of all)
3. `brew upgrade` gets cache hits and skips its own download phase
4. Failed downloads don't block the upgrade — brew handles them normally
5. Partial downloads are cleaned up (no corrupt cache entries)
6. User sees aggregate download progress during the parallel phase

## Scope

- New `internal/download/` package for parallel download logic
- URL extraction from `brew info --json=v2` output (formulae and casks)
- SHA256-based cache filename generation
- Parallel HTTP downloads with configurable concurrency
- Aggregate progress display
- Integration into upgrade command (between plan display and brew upgrade loop)
- Partial download cleanup (atomic rename or delete on failure)

## Non-Goals (this milestone)

- Resumable downloads (HTTP Range headers) — delete and retry instead
- Download verification (checksum of file contents) — brew does this itself
- Caching download URLs across runs
- Parallel `brew upgrade` execution (still sequential for safety)
- Applying parallel download to `syncd apply` (future)
- Bandwidth throttling or rate limiting

## Constraints

- Must not corrupt brew's cache directory
- Must use brew's exact filename format (SHA256 of URL as hex prefix)
- Must not break existing upgrade flow if download phase fails entirely
- Must work with the existing `CommandRunner` interface for `brew info` queries
- Must handle non-TTY gracefully (no `\r` progress in pipes)
- Verbose mode should show individual download details

## Open Questions

1. ~~Should partial downloads use atomic rename or temp suffix?~~ **Resolved:** Use `.downloading` suffix, rename on completion, delete on failure.
2. ~~Should concurrency be configurable via flag or config?~~ **Resolved:** Flag `--concurrency` on upgrade command, default 4.
3. ~~Should already-cached files be skipped?~~ **Resolved:** Yes. Check if file exists with correct name before downloading.
