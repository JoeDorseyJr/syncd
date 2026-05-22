# Milestones

## v0.1 — MVP ✅

- [x] Go project scaffolding (go.mod, cmd/, internal/)
- [x] Config parser (YAML → struct, strict field validation)
- [x] `syncd plan` — diff desired vs installed
- [x] `syncd apply` — install taps, brews, casks
- [x] Cleanup: remove unlisted packages, autoremove, clear cache
- [x] `--yes` flag to skip confirmation
- [x] `--config` flag to override default path
- [x] `--version` flag
- [x] Homebrew availability check with install guidance
- [x] Continue-on-error execution
- [x] Integration test suite (fake brew, no real system modification)

## v0.1.1 — Remove Nix ✅

- [x] Verify all nix-managed packages installed via Homebrew
- [x] Automated uninstall script (uninstall-nix.sh)
- [x] Remove nix daemon, users, group, volume, config
- [x] Clean shell hooks
- [x] Reboot and verify system works

## v0.2 — Polish & Usability ✅

- [x] Dependency-aware removal (only remove leaves, not deps)
- [x] Colored output (green +, red -, yellow ~) with TTY detection
- [x] `--no-color` flag
- [x] `syncd upgrade` — upgrade all outdated packages
- [x] Pin list in config (skip pinned packages during upgrade)
- [x] `syncd init` — snapshot current system into config.yaml
- [x] `--verbose` flag for debugging
- [x] Global install (`make install` → /usr/local/bin)

## v0.3 — macOS Defaults

- [ ] Declare macOS preferences in config (dock size, key repeat, etc.)
- [ ] `syncd plan` shows defaults drift
- [ ] `syncd apply` writes defaults and restarts affected apps
- [ ] Snapshot current defaults into config

## v0.4 — Dotfile Management

- [ ] Declare dotfile symlinks in config
- [ ] `syncd apply` creates symlinks (backup existing)
- [ ] `syncd plan` shows missing/broken symlinks
- [ ] Mac App Store apps via `mas`

## v0.5 — Distribution

- [ ] Homebrew tap formula (joedorseyjr/homebrew-syncd)
- [ ] GitHub releases with goreleaser (prebuilt binaries)
- [ ] Install script (`curl | sh` one-liner)
- [ ] Shell completions (zsh, bash, fish)

## v1.0 — Stable

- [ ] Battle-tested on daily driver
- [ ] Error recovery (rollback on failure)
- [ ] `syncd status` — detect drift without full plan output
- [ ] Logging (write apply results to ~/.config/syncd/logs/)
- [ ] `syncd doctor` — diagnose common issues
