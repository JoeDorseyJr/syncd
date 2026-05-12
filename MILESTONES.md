# Milestones

## v0.1 — MVP (replace nix-darwin)

- [ ] Go project scaffolding (go.mod, cmd/, internal/)
- [ ] Config parser (YAML → struct)
- [ ] `syncd plan` — diff desired vs installed
- [ ] `syncd apply` — install taps, brews, casks
- [ ] Cleanup: remove unlisted packages, autoremove, clear cache
- [ ] `--yes` flag to skip confirmation

## v0.1.1 — Remove Nix

- [ ] Verify all nix-managed packages are installed via Homebrew
- [ ] Stop nix-daemon
- [ ] Remove /nix, /etc/nix, /etc/nix-darwin
- [ ] Remove nix users and group
- [ ] Delete Nix APFS volume
- [ ] Clean up /etc/zshrc, /etc/bashrc, /etc/fstab, /etc/synthetic.conf
- [ ] Reboot and verify system works

## v0.2 — System config

- [ ] macOS defaults (read current, write declared, restart affected apps)
- [ ] Dotfile symlinks (backup existing, link from config dir)
- [ ] `syncd init` — snapshot current system into config.yaml
- [ ] Mac App Store apps (via mas)

## v0.3 — Shell & fonts

- [ ] Oh-my-zsh: install if missing, manage plugins and theme
- [ ] Font verification (check declared fonts are installed)
- [ ] Shell: set default shell if different

## v0.4 — Automation

- [ ] `syncd schedule` — generate and load launchd plist
- [ ] `syncd update` — upgrade all declared packages
- [ ] `syncd status` — detect drift from declared state
- [ ] Logging (write apply results to ~/.config/syncd/logs/)

## v0.5 — Distribution

- [ ] Homebrew tap formula (joedorseyjr/homebrew-syncd)
- [ ] GitHub releases with goreleaser (prebuilt binaries)
- [ ] Install script (`curl | sh` one-liner)
- [ ] Polished README + usage docs

## v1.0 — Stable

- [ ] Battle-tested on daily driver
- [ ] Cross-platform interfaces defined (Linux backends stubbed)
- [ ] Error recovery (rollback on failure)
- [ ] Config validation with helpful error messages
- [ ] `syncd doctor` — diagnose common issues
