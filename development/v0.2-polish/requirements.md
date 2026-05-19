# Requirements: syncd v0.2 — Polish & Usability

## User Stories

**US-005:** As a Mac user, I want to upgrade all my packages to the latest version with one command so that I stay current without manual effort.

**US-006:** As a Mac user, I want to pin specific packages so that they are not upgraded accidentally.

**US-007:** As a Mac user, I want to generate a config from my current system so that I don't have to write it by hand.

**US-008:** As a Mac user, I want removal to only target packages I explicitly installed so that dependencies aren't ripped out from under other packages.

**US-009:** As a Mac user, I want colored output so that I can quickly scan what's being added vs removed.

**US-010:** As a Mac user, I want to install syncd globally so that I can run it from anywhere.

---

## Functional Requirements

### Upgrade Command

**REQ-033:** `syncd upgrade` shall upgrade all installed brews to their latest version.
- Verification: Install an outdated brew, run `syncd upgrade --yes`, confirm it's upgraded.

**REQ-034:** `syncd upgrade` shall upgrade all installed casks to their latest version.
- Verification: Install an outdated cask, run `syncd upgrade --yes`, confirm it's upgraded.

**REQ-035:** `syncd upgrade` shall skip packages listed in the `pin` config section.
- Verification: Pin a brew, run `syncd upgrade --yes`, confirm pinned package version unchanged.

**REQ-036:** `syncd upgrade` shall prompt for confirmation before upgrading (unless `--yes`).
- Verification: Run `syncd upgrade` without `--yes`, confirm prompt appears. Answer "n", confirm no upgrades.

**REQ-037:** `syncd upgrade` shall show which packages will be upgraded before prompting.
- Verification: Run `syncd upgrade` with outdated packages, confirm list is printed before prompt.

**REQ-038:** `syncd upgrade` shall exit 0 when all upgrades succeed, exit 1 if any fail.
- Verification: Trigger a failure, confirm exit 1. Run clean upgrade, confirm exit 0.

**REQ-039:** `syncd upgrade` shall continue upgrading remaining packages if one fails.
- Verification: Have one bad package + good ones, confirm good ones still upgraded.

### Pin Configuration

**REQ-040:** syncd shall support a `pin` section in config as a list of package names.
- Verification: Unit test — parse config with `pin` section, confirm field populated.

**REQ-041:** syncd shall reject unknown keys in the `pin` section (must be a string list).
- Verification: Put an object in `pin`, confirm parse error.

**REQ-042:** `syncd upgrade` shall not upgrade any package whose name appears in `pin`.
- Verification: Pin `node@22`, run upgrade, confirm `node@22` not in upgrade output.

### Init Command

**REQ-043:** `syncd init` shall generate a valid YAML config file from the current system state.
- Verification: Run `syncd init`, confirm output is valid YAML that parses without error.

**REQ-044:** `syncd init` shall include only explicitly-installed formulae (leaves), not dependencies.
- Verification: Have a formula installed only as a dependency, run `syncd init`, confirm it's absent from output.

**REQ-045:** `syncd init` shall include all installed casks.
- Verification: Have casks installed, run `syncd init`, confirm they appear in output.

**REQ-046:** `syncd init` shall include all tapped repositories.
- Verification: Have taps added, run `syncd init`, confirm they appear in output.

**REQ-047:** `syncd init` shall write to stdout by default.
- Verification: Run `syncd init`, confirm output goes to stdout (pipe to file works).

**REQ-048:** `syncd init --output <path>` shall write the config to the specified file.
- Verification: Run `syncd init --output /tmp/test.yaml`, confirm file created with valid config.

**REQ-049:** `syncd init` shall not overwrite an existing file without `--force`.
- Verification: Run `syncd init --output` targeting an existing file, confirm error. Run with `--force`, confirm overwrite.

### Dependency-Aware Removal

**REQ-050:** `syncd plan` shall only list explicitly-installed formulae (leaves) as candidates for removal.
- Verification: Have a dependency-only formula installed, run `syncd plan` with `remove_unlisted: true`, confirm it does NOT appear as "to remove".

**REQ-051:** `syncd apply` shall only remove formulae that are both unlisted in config AND explicitly installed (leaves).
- Verification: Have a dependency-only formula not in config, run `syncd apply --yes`, confirm it remains installed.

**REQ-052:** syncd shall query `brew leaves` to determine which formulae are explicitly installed.
- Verification: Code review — confirm `brew leaves` is called. Unit test with mock output.

### Colored Output

**REQ-053:** syncd shall print additions with green `+` prefix.
- Verification: Run `syncd plan` with pending installs, confirm ANSI green in output.

**REQ-054:** syncd shall print removals with red `-` prefix.
- Verification: Run `syncd plan` with pending removals, confirm ANSI red in output.

**REQ-055:** syncd shall print maintenance actions with yellow `~` prefix.
- Verification: Run `syncd plan` with cleanup enabled, confirm ANSI yellow in output.

**REQ-056:** syncd shall suppress color when stdout is not a terminal (piped/redirected).
- Verification: Pipe `syncd plan` to a file, confirm no ANSI escape codes in output.

### Verbose Flag

**REQ-057:** `--verbose` flag shall print the full brew command output for each operation during apply/upgrade.
- Verification: Run `syncd apply --yes --verbose`, confirm brew stdout/stderr visible in output.

**REQ-058:** Without `--verbose`, syncd shall show only success/failure per package (current behavior).
- Verification: Run `syncd apply --yes` without `--verbose`, confirm only ✓/✗ lines shown.

---

## Technical Requirements

### Global Install

**REQ-059:** `make install` shall copy the binary to `/usr/local/bin/syncd`.
- Verification: Run `make install`, confirm `which syncd` returns `/usr/local/bin/syncd`.

**REQ-060:** `make uninstall` shall remove `/usr/local/bin/syncd`.
- Verification: Run `make uninstall`, confirm `which syncd` returns nothing.

### State Query

**REQ-061:** syncd shall query `brew leaves` to get explicitly-installed formulae.
- Verification: Unit test — mock `brew leaves` output, confirm parsed correctly.

**REQ-062:** syncd shall query `brew outdated` to determine which packages need upgrading.
- Verification: Unit test — mock `brew outdated` output, confirm parsed correctly.

**REQ-063:** syncd shall query `brew outdated --cask` to determine which casks need upgrading.
- Verification: Unit test — mock `brew outdated --cask` output, confirm parsed correctly.

---

## Traceability Matrix

| Requirement | User Story | Command |
|-------------|-----------|---------|
| REQ-033–039 | US-005 | upgrade |
| REQ-040–042 | US-006 | upgrade (pin) |
| REQ-043–049 | US-007 | init |
| REQ-050–052 | US-008 | plan/apply (removal) |
| REQ-053–056 | US-009 | plan/apply (output) |
| REQ-057–058 | US-009 | apply/upgrade (verbose) |
| REQ-059–060 | US-010 | make install |
| REQ-061–063 | US-005, US-008 | technical |
