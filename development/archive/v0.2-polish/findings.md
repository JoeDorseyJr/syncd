# Findings: syncd v0.2 — Polish & Usability

## Checklist of Issues and Improvements

- [x] **High: Problem statement still describes upgrade as `syncd apply` behavior while requirements/design/tasks define a separate `syncd upgrade` command.**
  - Evidence: `problem-statement.md` Desired Behavior says "`syncd apply` with `upgrade: true` upgrades all non-pinned packages"; the resolved open question says "`syncd upgrade` is a separate command"; `requirements.md` REQ-033 through REQ-039 and `design.md` both define `syncd upgrade`.
  - Improvement: Normalize the problem statement to say `syncd upgrade` handles upgrades and remove the stale `upgrade: true` wording.
  - Status: closed
  - Resolution: Update problem-statement.md Desired Behavior to reference `syncd upgrade` command instead of `syncd apply` with `upgrade: true`.
  - Follow-up: `development/v0.2-polish/problem-statement.md` — Desired Behavior section

- [x] **High: Real-usage validation for upgrade behavior is weak if it depends on naturally outdated Homebrew packages.**
  - Evidence: REQ-033 and REQ-034 say to install an outdated brew/cask and confirm it upgrades, but the task plan primarily calls for fake brew integration tests and does not define a deterministic real-usage fixture.
  - Improvement: Validate command behavior with fake brew integration tests, and separately document any manual real-Homebrew smoke test as optional because creating an outdated package state is not reliably reproducible.
  - Status: closed
  - Resolution: Fake brew integration tests are the primary validation. Real-Homebrew smoke tests are optional and not gated. No change to requirements needed — verification wording describes intent, not mandatory CI steps.
  - Follow-up: none

- [x] **Medium: REQ-041 uses the wrong acceptance language for the `pin` schema.**
  - Evidence: `requirements.md` says "reject unknown keys in the `pin` section" even though `pin` is a flat list, and the verification describes rejecting an object value rather than an unknown key.
  - Improvement: Rename the requirement to "syncd shall reject non-list or non-string `pin` values" and keep the unit test focused on invalid YAML shapes.
  - Status: closed
  - Resolution: Reword REQ-041 to "syncd shall reject non-list or non-string values in the `pin` field."
  - Follow-up: `development/v0.2-polish/requirements.md` — REQ-041 text

- [x] **Medium: Init command requirements for casks, taps, and stdout have task coverage but no explicit test cases.**
  - Evidence: REQ-045, REQ-046, and REQ-047 map to task 4.1, but task 4.3 only names tests for valid YAML, leaves-only brews, `--output`, and overwrite protection.
  - Improvement: Add explicit tests for casks included, taps included, and stdout default behavior.
  - Status: closed
  - Resolution: Add test cases to task 4.3 for casks, taps, and stdout.
  - Follow-up: `development/v0.2-polish/tasks.md` — task 4.3 test list

- [x] **Medium: Color requirements should validate both semantic prefixes and ANSI behavior.**
  - Evidence: REQ-053 through REQ-055 require green `+`, red `-`, and yellow `~`; tasks verify visible colors manually and test only non-TTY suppression.
  - Improvement: Add deterministic formatter tests that assert prefix choice and ANSI wrapping when color is enabled, plus the existing piped-output suppression test.
  - Status: closed
  - Resolution: Add unit tests to task 2.2 that assert ANSI codes present when color enabled, and correct prefix symbols.
  - Follow-up: `development/v0.2-polish/tasks.md` — task 2.2 test list

- [x] **Medium: Verbose flag implementation needs a cross-command contract.**
  - Evidence: REQ-057 says `--verbose` prints brew output during apply/upgrade; design adds a root persistent flag and executor switch, but tasks do not explicitly require tests for both apply and upgrade.
  - Improvement: Add tests or integration cases for `syncd apply --yes --verbose`, `syncd upgrade --yes --verbose`, and default non-verbose behavior.
  - Status: closed
  - Resolution: Add integration test cases to task 5.1 for both apply and upgrade with verbose.
  - Follow-up: `development/v0.2-polish/tasks.md` — task 5.1 verification list

- [x] **Low: `make install` and `make uninstall` verification may require permissions outside normal test sandboxes.**
  - Evidence: REQ-059 and REQ-060 target `/usr/local/bin/syncd`, which can require elevated permissions depending on the machine.
  - Improvement: Keep the real command verification, but make the Makefile support `PREFIX ?= /usr/local` so tests can run against a temporary prefix.
  - Status: closed
  - Resolution: Use `PREFIX ?= /usr/local` in Makefile. Tests can override with a temp dir.
  - Follow-up: `development/v0.2-polish/design.md` — Makefile Additions section; `development/v0.2-polish/tasks.md` — task 5.2

- [x] **Low: Existing implementation names `plan.State`, but v0.2 design says "Add `Leaves` field to `plan.State`" while state reader returns `brew.State`.**
  - Evidence: `tasks.md` task 1.1 says add `Leaves` to `plan.State`; `design.md` describes `internal/brew/state.go` additions and a `State` struct with leaves; current code has separate `brew.State` and `plan.State`.
  - Improvement: Specify whether `Leaves` belongs in both structs or whether `plan.State` should be replaced with/import from `brew.State` to avoid duplicated drift.
  - Status: closed
  - Resolution: Keep both structs (avoids import cycle per v0.1 design decision). Add `Leaves` to `plan.State`. CLI layer converts `brew.State` → `plan.State` including leaves, same pattern as v0.1.
  - Follow-up: none (design already implies this; task 1.2 covers the CLI conversion)

## Open Clarification Questions

1. Should `syncd upgrade` require a config file when the only config-owned behavior is pin filtering?

   Answer: No. Load config if it exists; skip pin filtering if config is missing. This makes `syncd upgrade` usable without any config file.

2. Should the `pin` list apply to casks as well as formulae when names overlap?

   Answer: Yes. Pin is universal — applies to both brews and casks.

3. Should `syncd init` emit `pin: []` and `cleanup` defaults, or only the currently discovered `taps`, `brews`, and `casks`?

   Answer: Full template — include empty `pin: []` and `cleanup` section with sensible defaults so the user sees all available options.

4. Should colored output be disabled by a user-facing flag such as `--no-color`, or only by non-TTY detection?

   Answer: Both. Add `--no-color` flag in addition to TTY detection for explicit user control.

5. Should `make install` support a configurable `PREFIX` for non-`/usr/local` installs and test isolation?

   Answer: Yes. Use `PREFIX ?= /usr/local` so users can override.

## Validation Notes

- All REQ-033 through REQ-066 IDs are present and have task coverage in `tasks.md`.
- All findings closed. No new issues discovered during re-review.
- Traceability matrix updated with REQ-064, REQ-065, REQ-066.
- The strongest validation paths are CLI/integration tests with fake brew for command behavior, unit tests for parser/state/diff behavior, and a small number of real command smoke tests for install/uninstall.
- Requirements that mention code review only should be paired with a unit or CLI-level test where feasible so behavior is validated through usage, not just document alignment.
