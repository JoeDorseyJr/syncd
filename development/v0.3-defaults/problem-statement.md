# Problem Statement: syncd v0.3 — macOS Defaults

Declare macOS system preferences in config and reconcile them like packages.

## Context

syncd v0.2 manages Homebrew packages declaratively — install, remove, upgrade, pin. But macOS system preferences (Dock size, key repeat rate, Finder settings, trackpad behavior) still require manual configuration through System Settings or scattered `defaults write` commands. When setting up a new Mac or restoring after a wipe, users must remember and re-apply dozens of preference tweaks by hand.

The `defaults` command-line tool reads and writes macOS preference domains (plist files), making it automatable. But there's no declarative layer on top — no way to declare desired preferences, detect drift, or apply them idempotently.

## Current Behavior

- syncd manages Homebrew taps, formulae, and casks only
- macOS preferences are configured manually through System Settings or ad-hoc shell scripts
- No way to detect when a preference has drifted from desired state
- No way to snapshot current preferences into a reproducible config
- Some preference changes require restarting apps (Dock, Finder, SystemUIServer) to take effect — users must remember which ones

## Desired Behavior

- Declare macOS preferences in `~/.config/syncd/config.yaml` under a `defaults` section
- Each entry specifies a domain, key, type, and desired value
- `syncd plan` shows defaults that differ from desired state (current → desired)
- `syncd apply` writes drifted defaults and restarts affected apps
- `syncd init` can snapshot current defaults into config (for specified domains/keys)
- Handles missing domains gracefully (domain doesn't exist yet = drift)
- Supports all `defaults` value types: string, int, float, bool

## Success Outcomes

1. `syncd plan` shows which macOS preferences have drifted from config
2. `syncd apply` writes all drifted preferences and restarts affected apps
3. `syncd init` snapshots specified defaults into config format
4. New machine setup: one config file reproduces both packages AND preferences
5. Running `syncd apply` twice produces no changes on second run (idempotent)

## Scope

- `defaults` section in config YAML with domain/key/type/value schema
- Drift detection via `defaults read` in `syncd plan`
- Write via `defaults write` in `syncd apply`
- App restart via `killall` for affected apps
- Snapshot via `defaults read` in `syncd init`
- Support for string, int, float, bool value types

## Non-Goals (this milestone)

- Array or dictionary value types (complex plists)
- `defaults delete` (removing keys)
- Managing preferences that require `sudo` (global domain writes)
- Watching for drift in real-time
- Dotfile management (v0.4)
- Mac App Store apps (v0.4)

## Constraints

- Must work with the existing `defaults` CLI (no direct plist manipulation)
- Config schema must pass `KnownFields(true)` strict validation
- Must not break existing Homebrew management features
- Plan mode must remain read-only
- App restarts only happen during `apply`, never during `plan`

## Open Questions

1. ~~Should `syncd init` snapshot ALL defaults or only user-specified domains?~~ **Resolved:** Only specified domains/keys. The full defaults database is enormous and mostly irrelevant.
2. ~~Should app restart be automatic or require opt-in per entry?~~ **Resolved:** Automatic. Declare which app to kill per entry; syncd restarts affected apps after writing all defaults in that domain.
3. ~~How to handle `defaults read` on a key that doesn't exist?~~ **Resolved:** Treat as drift — the key should be written. Non-existent domain is also drift.
