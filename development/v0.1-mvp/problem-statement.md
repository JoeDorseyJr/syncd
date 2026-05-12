# Problem Statement: syncd v0.1 — MVP

Nix-darwin is broken on Intel Macs and needs to be replaced with a working package sync tool.

## Context

The current nix-darwin setup fails on every `brew bundle` and `brew cu` operation due to a `to_sym` nil error in the Nix-patched Homebrew. The upstream fix is unlikely before Nix drops x86_64-darwin entirely (26.11). The system cannot be updated through its declared configuration.

## Current Behavior

- `darwin-rebuild switch` fails during the Homebrew bundle step
- `brew cu` crashes on cask loading
- No way to declaratively install/remove packages without manual intervention
- System state drifts from what's declared

## Desired Behavior

- A YAML config declares taps, brews, and casks
- `syncd plan` shows what would be installed, removed, or upgraded
- `syncd apply` reconciles the system to match the config
- Packages not in the config are removed (cleanup)
- Works directly with Homebrew — no Nix, no patches

## Success Outcomes

1. `syncd plan` outputs a clear diff of desired vs actual packages
2. `syncd apply` installs missing taps, brews, and casks
3. `syncd apply` removes packages not declared in config (when cleanup enabled)
4. Running `syncd apply` twice in a row produces no changes (idempotent)
5. `--yes` flag skips confirmation prompt

## Scope

- Go project scaffolding (go.mod, cmd/, internal/)
- YAML config parser
- `syncd plan` command
- `syncd apply` command
- Cleanup: remove unlisted, `brew autoremove`, `brew cleanup`
- `--yes` flag

## Non-Goals (this milestone)

- macOS defaults
- Dotfile management
- Shell/oh-my-zsh setup
- Mac App Store apps
- Auto-update scheduling
- `syncd init` (snapshot current system)

## Constraints

- Single Go binary
- Only depends on Homebrew being installed
- Must not break existing installed packages on first run
- Plan mode must be safe (read-only)

## Open Questions

1. Should `syncd apply` install Homebrew if it's missing?
2. On cleanup, should it prompt per-package or batch confirm?
