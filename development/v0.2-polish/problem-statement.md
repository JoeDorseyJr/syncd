# Problem Statement: syncd v0.2 — Polish & Usability

## Context

v0.1 works — packages install, removals happen, the system syncs. But daily use reveals gaps: no way to upgrade packages, no way to snapshot an existing system, no colored output, and removal logic doesn't distinguish between packages you chose and their auto-installed dependencies.

## Current Behavior

- `syncd apply` installs missing packages but doesn't upgrade outdated ones
- No way to generate a config from what's currently installed
- `remove_unlisted: true` flags dependencies for removal (not just user-chosen packages)
- Output is plain text with no visual distinction between adds/removes
- Binary only available from the project build directory
- No version pinning — can't protect specific packages from upgrades

## Desired Behavior

- `syncd upgrade` upgrades all non-pinned packages to latest
- `pin` list in config prevents specific packages from being upgraded
- `syncd init` snapshots current system into a config file
- Removal only targets explicitly-installed packages (not auto-deps)
- Colored terminal output (green +, red -, yellow ~)
- `make install` puts binary on PATH
- `--verbose` flag shows brew command output in real-time

## Success Outcomes

1. `syncd upgrade` brings all packages to latest (except pinned)
2. `syncd init` produces a valid config from current state
3. `remove_unlisted: true` doesn't remove dependency-only packages
4. Output is scannable at a glance with color coding
5. New machine setup: `make install && syncd apply --config ./my-config.yaml`

## Scope

- Package upgrades with pin support
- `syncd init` command
- Dependency-aware removal
- Colored output
- Global install target
- `--verbose` flag

## Non-Goals (this milestone)

- macOS defaults management
- Dotfile symlinks
- Mac App Store apps
- Scheduled runs / launchd
- Cross-platform support

## Open Questions

1. ~~Should `upgrade: true` be in the `cleanup` section or a top-level flag?~~ **Resolved:** Neither. `syncd upgrade` is a separate command. Pin list lives in config.
2. ~~Should `syncd init` include dependencies or only explicitly-installed packages?~~ **Resolved:** Leaves-only (explicitly installed). Dependencies are managed by `autoremove`.
3. ~~For dependency-aware removal, use `brew leaves` or `brew info --json` to detect user-installed?~~ **Resolved:** Use `brew leaves`. Only remove unlisted leaves; let `autoremove` handle orphaned deps.
