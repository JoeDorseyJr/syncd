# Requirements: syncd v0.1 — MVP

## User Stories

**US-001:** As a Mac user, I want to declare my desired packages in a config file so that I can reproduce my setup on any machine.

**US-002:** As a Mac user, I want to preview what changes will be made before they happen so that I don't accidentally break my system.

**US-003:** As a Mac user, I want undeclared packages removed automatically so that my system stays clean.

**US-004:** As a Mac user, I want the tool to be idempotent so that I can run it repeatedly without side effects.

---

## Functional Requirements

### Config Parsing

**REQ-001:** syncd shall read a YAML config file from `~/.config/syncd/config.yaml`.
- Verification: Unit test — parse a valid config, confirm all fields populated in struct.

**REQ-002:** syncd shall exit with a clear error message if the config file is missing.
- Verification: Run `syncd plan` with no config file, confirm non-zero exit and error message.

**REQ-003:** syncd shall exit with a clear error message if the config file has invalid YAML.
- Verification: Run with malformed YAML, confirm non-zero exit and error pointing to the issue.

**REQ-004:** syncd shall support the following config sections: `taps`, `brews`, `casks`, `cleanup`.
- Verification: Unit test — parse config with all sections, confirm struct fields match.

### Plan Command

**REQ-005:** `syncd plan` shall compare declared taps against installed taps and list taps to add.
- Verification: Declare a tap not currently installed, run `syncd plan`, confirm it appears in output as "to add".

**REQ-006:** `syncd plan` shall compare declared brews against installed brews and list brews to install.
- Verification: Declare a brew not currently installed, run `syncd plan`, confirm it appears as "to install".

**REQ-007:** `syncd plan` shall compare declared casks against installed casks and list casks to install.
- Verification: Declare a cask not currently installed, run `syncd plan`, confirm it appears as "to install".

**REQ-008:** `syncd plan` shall list installed brews not in the config as "to remove" (when `cleanup.remove_unlisted` is true).
- Verification: Have a brew installed that's not in config, run `syncd plan`, confirm it appears as "to remove".

**REQ-009:** `syncd plan` shall list installed casks not in the config as "to remove" (when `cleanup.remove_unlisted` is true).
- Verification: Have a cask installed that's not in config, run `syncd plan`, confirm it appears as "to remove".

**REQ-010:** `syncd plan` shall not modify the system (read-only).
- Verification: Run `syncd plan`, confirm no packages installed or removed (compare `brew list` before and after).

**REQ-011:** `syncd plan` shall exit with code 0 if no changes needed, code 2 if changes pending. Cleanup-only actions (`autoremove`, `clear_cache`) are recurring maintenance and do not count as pending changes for exit-code purposes.
- Verification: Run on a synced system → exit 0. Add a new brew to config → exit 2. Enable only cleanup flags on synced system → exit 0.

### Apply Command

**REQ-012:** `syncd apply` shall prompt for confirmation before making changes.
- Verification: Run `syncd apply` with pending changes, confirm prompt appears, answer "n", confirm no changes made.

**REQ-013:** `syncd apply --yes` shall skip the confirmation prompt.
- Verification: Run `syncd apply --yes`, confirm changes applied without prompt.

**REQ-014:** `syncd apply` shall add taps not currently tapped.
- Verification: Declare a new tap, run `syncd apply --yes`, confirm `brew tap` shows it.

**REQ-015:** `syncd apply` shall install brews not currently installed.
- Verification: Declare a new brew, run `syncd apply --yes`, confirm `brew list` includes it.

**REQ-016:** `syncd apply` shall install casks not currently installed.
- Verification: Declare a new cask, run `syncd apply --yes`, confirm `brew list --cask` includes it.

**REQ-017:** `syncd apply` shall remove installed brews not in the config (when `cleanup.remove_unlisted` is true).
- Verification: Install a brew not in config, run `syncd apply --yes`, confirm it's no longer in `brew list`.

**REQ-018:** `syncd apply` shall remove installed casks not in the config (when `cleanup.remove_unlisted` is true).
- Verification: Install a cask not in config, run `syncd apply --yes`, confirm it's no longer in `brew list --cask`.

**REQ-019:** `syncd apply` shall not remove packages when `cleanup.remove_unlisted` is false.
- Verification: Set `remove_unlisted: false`, have an undeclared package installed, run `syncd apply --yes`, confirm package remains.

**REQ-020:** `syncd apply` shall run `brew autoremove` when `cleanup.autoremove` is true.
- Verification: Run `syncd apply --yes` with `autoremove: true`, confirm autoremove was executed (check output).

**REQ-021:** `syncd apply` shall run `brew cleanup` when `cleanup.clear_cache` is true.
- Verification: Run `syncd apply --yes` with `clear_cache: true`, confirm cleanup was executed (check output).

**REQ-022:** `syncd apply` shall be idempotent — running it twice with no config changes produces no modifications on the second run. Cleanup actions (`autoremove`, `clear_cache`) are recurring maintenance that run on every apply when enabled; idempotency applies to package install/remove operations.
- Verification: Run `syncd apply --yes`, then run `syncd plan`, confirm exit code 0 (no changes).

### Error Handling

**REQ-023:** syncd shall report which specific package failed if a brew/cask install fails, and continue with remaining packages.
- Verification: Declare a nonexistent brew alongside valid ones, run `syncd apply --yes`, confirm error reported for bad package and others still installed.

**REQ-024:** syncd shall exit with non-zero code if any operation fails during apply.
- Verification: Trigger a failure, confirm exit code is non-zero.

---

## Technical Requirements

**REQ-025:** syncd shall be written in Go and compile to a single binary.
- Verification: Run `go build`, confirm single binary output.

**REQ-026:** syncd shall have no runtime dependencies beyond Homebrew being installed.
- Verification: Run on a clean Mac with only Homebrew, confirm it works.

**REQ-027:** syncd shall invoke Homebrew via shell commands (`brew tap`, `brew install`, `brew uninstall`, `brew list`).
- Verification: Code review — confirm no Homebrew Ruby API usage, only CLI invocations.

**REQ-028:** syncd shall use cobra for CLI command structure.
- Verification: Code review — confirm cobra usage in cmd/.

**REQ-029:** syncd shall use `gopkg.in/yaml.v3` for config parsing.
- Verification: Code review — confirm yaml.v3 in go.mod.

---

## Additional Requirements

### Tap Cleanup

**REQ-030:** `syncd plan` shall list installed taps not in the config as "to remove" (when `cleanup.remove_unlisted` is true).
- Verification: Have a tap installed that's not in config, run `syncd plan`, confirm it appears as "to remove".

**REQ-031:** `syncd apply` shall remove installed taps not in the config (when `cleanup.remove_unlisted` is true).
- Verification: Tap a repo not in config, run `syncd apply --yes`, confirm `brew tap` no longer shows it.

### CLI Flags

**REQ-032:** syncd shall accept a `--config <path>` flag that overrides the default config file path.
- Verification: Create a config at a non-default path, run `syncd plan --config /tmp/test.yaml`, confirm it reads from the specified path.

---

## Traceability Matrix

| Requirement | User Story | Command |
|-------------|-----------|---------|
| REQ-001–004 | US-001 | config |
| REQ-005–011 | US-002 | plan |
| REQ-012–022 | US-001, US-003, US-004 | apply |
| REQ-023–024 | US-002 | apply (errors) |
| REQ-025–029 | — | technical |
| REQ-030 | US-003 | plan (tap cleanup) |
| REQ-031 | US-003 | apply (tap cleanup) |
| REQ-032 | — | CLI flag |
