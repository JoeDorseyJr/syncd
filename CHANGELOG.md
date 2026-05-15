# Changelog

## v0.1.0 — MVP

Initial release. Declarative Homebrew package management for macOS.

### Added

- `syncd plan` — show what would change without modifying the system
- `syncd apply` — reconcile system state to match config (with confirmation prompt)
- `--yes` flag to skip confirmation
- `--config <path>` flag to override default config path
- `--version` flag
- YAML config with `taps`, `brews`, `casks`, `cleanup` sections
- Strict config validation (unknown keys rejected)
- Tap, formula, and cask install/removal
- Tap removal after package removal (safe ordering)
- `cleanup.remove_unlisted` — remove undeclared packages
- `cleanup.autoremove` — prune unused dependencies
- `cleanup.clear_cache` — clear Homebrew download cache
- Continue-on-error execution (partial failures don't abort)
- Duplicate config entry deduplication
- Homebrew availability check with install guidance
- Exit code 2 for pending changes, 1 for errors
- Integration test suite with fake brew (no real system modification)
- `uninstall-nix.sh` utility for migrating from nix-darwin
