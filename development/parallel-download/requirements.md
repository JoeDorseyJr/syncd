# Requirements: Parallel Pre-Download for syncd Upgrade

## User Stories

**US-018:** As a Mac user, I want `syncd upgrade` to download all bottles/casks in parallel before upgrading so that upgrades complete faster.

**US-019:** As a Mac user, I want to see aggregate download progress so that I know how the parallel download phase is progressing.

**US-020:** As a Mac user, I want download failures to be non-fatal so that brew can still download packages itself if pre-download fails.

---

## Functional Requirements

### URL Extraction

**REQ-128:** syncd shall query `brew info --json=v2 <names...>` to obtain bottle URLs for outdated formulae.
- Verification: Unit test — parse sample `brew info --json=v2` JSON, confirm bottle URLs extracted.

**REQ-129:** syncd shall extract the bottle URL for the current macOS version (e.g., `arm64_sonoma`) from the JSON `bottle.stable.files` object.
- Verification: Unit test — JSON with multiple OS entries, confirm correct platform URL selected.

**REQ-130:** syncd shall fall back to the `all` bottle URL if no platform-specific bottle exists.
- Verification: Unit test — JSON with only `all` entry, confirm URL extracted.

**REQ-131:** syncd shall extract cask download URLs from `brew info --json=v2 --cask <names...>`.
- Verification: Unit test — parse cask JSON, confirm URL extracted from `url` field.

**REQ-132:** syncd shall determine the current macOS architecture and version for bottle platform selection.
- Verification: Unit test — mock platform detection, confirm correct platform string produced.

### Cache Filename Generation

**REQ-133:** syncd shall generate cache filenames using the format `SHA256_of_URL--name--version.bottle.tar.gz` for formulae.
- Verification: Unit test — given URL, name, version, confirm filename matches brew's format.

**REQ-134:** syncd shall compute the SHA256 hash of the download URL (as hex string) for the filename prefix.
- Verification: Unit test — known URL produces expected SHA256 hex prefix.

**REQ-135:** syncd shall generate cache filenames using the format `SHA256_of_URL--name--version.ext` for casks, preserving the original file extension.
- Verification: Unit test — given cask URL ending in `.dmg`, confirm filename preserves extension.

**REQ-136:** syncd shall place downloaded files in `~/Library/Caches/Homebrew/downloads/`.
- Verification: Integration test — after download, file exists at expected cache path.

### Parallel Download

**REQ-137:** syncd shall download all bottles/casks in parallel using a configurable concurrency limit (default 4).
- Verification: Integration test — 4 packages download concurrently, not sequentially.

**REQ-138:** syncd shall accept a `--concurrency` flag on the upgrade command to control parallel download workers.
- Verification: Run `syncd upgrade --concurrency 2`, confirm at most 2 concurrent downloads.

**REQ-139:** syncd shall skip downloading files that already exist in brew's cache directory.
- Verification: Unit test — file already exists at cache path, confirm no HTTP request made.

**REQ-140:** syncd shall use a `.downloading` suffix for in-progress downloads and rename to final name on completion.
- Verification: Unit test — during download, file has `.downloading` suffix; after completion, final name exists.

**REQ-141:** syncd shall delete partial `.downloading` files on download failure.
- Verification: Unit test — simulate HTTP error mid-download, confirm `.downloading` file removed.

**REQ-142:** syncd shall not fail the upgrade if all downloads fail — brew falls back to its own download.
- Verification: Integration test — all downloads fail, confirm `brew upgrade` still runs and succeeds.

**REQ-143:** syncd shall report individual download failures as warnings (not errors).
- Verification: Integration test — one download fails, confirm warning printed and upgrade continues.

### Progress Display

**REQ-144:** syncd shall display aggregate download progress showing completed/total packages and total bytes downloaded.
- Verification: Integration test — during download phase, progress line shows `[2/5] 12.3 MB downloaded`.

**REQ-145:** syncd shall use `\r` overwrite for progress display in TTY mode.
- Verification: Run in TTY, confirm single-line progress updates in place.

**REQ-146:** syncd shall fall back to simple line output in non-TTY mode.
- Verification: Pipe output, confirm no `\r` characters.

**REQ-147:** `--verbose` mode shall print individual download start/complete messages per package.
- Verification: Run with `--verbose`, confirm per-package download messages visible.

### Integration with Upgrade Command

**REQ-148:** The parallel download phase shall run after the upgrade plan is confirmed but before `brew upgrade` execution.
- Verification: Integration test — confirm download phase occurs between confirmation and first `brew upgrade` call.

**REQ-149:** `brew upgrade` shall get cache hits for successfully pre-downloaded packages (no re-download).
- Verification: Integration test — pre-download a bottle, run `brew upgrade`, confirm "downloading" phase is skipped.

**REQ-150:** syncd shall skip the parallel download phase entirely when `--verbose` is set.
- Verification: Run `syncd upgrade --verbose`, confirm no parallel download phase (brew downloads itself).

**REQ-151:** The `--concurrency` flag shall default to 4 when not specified.
- Verification: Code review — confirm default value is 4.

---

## Technical Requirements

**REQ-152:** Parallel download logic shall live in a new `internal/download/` package.
- Verification: Code review — confirm package exists at `internal/download/`.

**REQ-153:** URL extraction shall use the existing `CommandRunner` interface to run `brew info`.
- Verification: Code review — confirm `runner.Run` used for `brew info` queries.

**REQ-154:** HTTP downloads shall use Go's standard `net/http` client with no external dependencies.
- Verification: Code review — confirm only `net/http` used for downloads.

**REQ-155:** syncd shall follow HTTP redirects (brew URLs often redirect through CDNs).
- Verification: Unit test — mock redirect, confirm final URL downloaded.

**REQ-156:** Existing unit and integration tests shall continue to pass without modification.
- Verification: Run `make test && make integration-test`, confirm all pass.

---

## Traceability Matrix

| Requirement | User Story | Component |
|-------------|-----------|-----------|
| REQ-128–132 | US-018 | URL extraction |
| REQ-133–136 | US-018 | cache filename |
| REQ-137–143 | US-018, US-020 | parallel download |
| REQ-144–147 | US-019 | progress display |
| REQ-148–151 | US-018 | upgrade integration |
| REQ-152–156 | US-018 | technical |
