# Tasks: syncd v0.2 — Polish & Usability

## Phase 1: Dependency-Aware Removal (~1.5 hours)

### 1.1 State reader: GetLeaves
> REQ-052, REQ-061 | Design: New State Queries

- [ ] Add `GetLeaves(runner) ([]string, error)` to `internal/brew/state.go`
  - Parse `brew leaves` output into string slice
- [ ] Add `Leaves` field to `plan.State` struct
- [ ] Add unit tests in `brew_test.go`
  - Test: parse multi-line leaves output
  - Test: handle empty output (no leaves)
  - Test: handle command failure
- [ ] Verify: `go test ./internal/brew/...` passes

### 1.2 Plan: leaf-only removal
> REQ-050, REQ-051 | Design: Dependency-Aware Removal

- [ ] Update `plan.Compute()` to use `state.Leaves` instead of `state.Brews` for removal candidates
- [ ] Update CLI (`plan.go`, `apply.go`) to populate `Leaves` in `plan.State`
- [ ] Add unit tests in `plan_test.go`
  - Test: dependency-only formula NOT in removal list
  - Test: explicitly-installed formula still removed when unlisted
  - Test: cask removal unchanged (no leaf concept for casks)
- [ ] Verify: `go test ./internal/plan/...` passes

### 1.3 Integration test: dependency-aware removal
> REQ-050, REQ-051 | Design: Integration Tests

- [ ] Update fake brew script to support `leaves` command
- [ ] Add integration test: formula installed as dep-only is NOT removed
- [ ] Add integration test: leaf formula not in config IS removed
- [ ] Verify: `make integration-test` passes

**Estimate:** ~1.5 hours

---

## Phase 2: Colored Output (~1 hour)

### 2.1 Output formatter
> REQ-053, REQ-054, REQ-055, REQ-056 | Design: Colored Output

- [ ] Create `internal/cli/output.go`
  - Define color constants (Green, Red, Yellow, Reset)
  - Implement `isTerminal()` TTY detection
  - Suppress colors when stdout is not a terminal
- [ ] Update `printPlan()` in `internal/cli/plan.go`
  - Green `+` for additions
  - Red `-` for removals
  - Yellow `~` for maintenance actions
- [ ] Update `FormatResults()` in `internal/brew/executor.go`
  - Green `✓` for success
  - Red `✗` for failure
- [ ] Verify: colors visible in terminal, absent when piped

### 2.2 Color tests
> REQ-056 | Design: Colored Output — TTY detection

- [ ] Add unit test: `isTerminal` returns false for non-TTY fd
- [ ] Add unit test: formatter outputs ANSI green for additions when color enabled
- [ ] Add unit test: formatter outputs ANSI red for removals when color enabled
- [ ] Add unit test: formatter outputs ANSI yellow for maintenance when color enabled
- [ ] Add unit test: correct prefix symbols (`+`, `-`, `~`)
- [ ] Add integration test: pipe `syncd plan` output, confirm no ANSI escape codes
- [ ] Add integration test: `syncd plan --no-color` in TTY, confirm no ANSI escape codes
- [ ] Verify: `make test && make integration-test` passes

**Estimate:** ~1 hour

---

## Phase 3: Upgrade Command (~2 hours)

### 3.1 Config: pin field
> REQ-040, REQ-041 | Design: Config Changes

- [ ] Add `Pin []string \`yaml:"pin"\`` to `Config` struct
- [ ] Add unit tests in `config_test.go`
  - Test: parse config with `pin` section
  - Test: invalid pin value (object instead of list) rejected by KnownFields
- [ ] Verify: `go test ./internal/config/...` passes

### 3.2 State reader: GetOutdated
> REQ-062, REQ-063 | Design: New State Queries

- [ ] Add `GetOutdated(runner) ([]string, error)` — parse `brew outdated --formula -1`
- [ ] Add `GetOutdatedCasks(runner) ([]string, error)` — parse `brew outdated --cask -1`
- [ ] Add unit tests in `brew_test.go`
  - Test: parse outdated output
  - Test: empty output (nothing outdated)
  - Test: command failure
- [ ] Verify: `go test ./internal/brew/...` passes

### 3.3 Executor: Upgrade function
> REQ-033, REQ-034, REQ-039 | Design: Upgrade Command Flow

- [ ] Add `Upgrade(runner, brews, casks []string) []Result` to `internal/brew/executor.go`
  - Run `brew upgrade <name>` for each brew
  - Run `brew upgrade --cask <name>` for each cask
  - Continue on failure, collect results
- [ ] Add unit tests in `executor_test.go`
  - Test: successful upgrade sequence
  - Test: one failure doesn't stop others
  - Test: correct command args (`upgrade` vs `upgrade --cask`)
- [ ] Verify: `go test ./internal/brew/...` passes

### 3.4 Upgrade CLI command
> REQ-035, REQ-036, REQ-037, REQ-038, REQ-042 | Design: Upgrade Command Flow

- [ ] Create `internal/cli/upgrade.go` — cobra `upgrade` subcommand
  - Load config (for pin list)
  - Query `GetOutdated()` and `GetOutdatedCasks()`
  - Filter out pinned packages
  - Print upgrade plan
  - Prompt for confirmation (unless `--yes`)
  - Execute upgrades, print results
  - Exit 0 on success, exit 1 on any failure
- [ ] Wire into root command in `cmd/syncd/main.go`
- [ ] Add `--yes` flag
- [ ] Verify: `syncd upgrade --help` works

### 3.5 Upgrade integration tests
> REQ-035, REQ-038, REQ-039, REQ-042 | Design: Integration Tests

- [ ] Update fake brew to support `outdated` and `upgrade` commands
- [ ] Add integration test: upgrade succeeds, exit 0
- [ ] Add integration test: pinned package skipped
- [ ] Add integration test: one failure → exit 1, others still upgraded
- [ ] Add integration test: confirmation prompt (cancel with "n")
- [ ] Add integration test: upgrade works without config file (no pin filtering)
- [ ] Verify: `make integration-test` passes

**Estimate:** ~2 hours

---

## Phase 4: Init Command (~1.5 hours)

### 4.1 Init CLI command
> REQ-043, REQ-044, REQ-045, REQ-046, REQ-047 | Design: Init Command Flow

- [ ] Create `internal/cli/init.go` — cobra `init` subcommand
  - Query `GetState()` + `GetLeaves()`
  - Build Config struct: taps from state, brews from leaves, casks from state
  - Marshal to YAML
  - Print to stdout by default
- [ ] Wire into root command
- [ ] Verify: `syncd init` outputs valid YAML

### 4.2 Init output flag
> REQ-048, REQ-049 | Design: Init Command Flow

- [ ] Add `--output <path>` flag
- [ ] Add `--force` flag
- [ ] Implement: write to file, refuse overwrite without `--force`
- [ ] Verify: file created, overwrite blocked, `--force` overwrites

### 4.3 Init tests
> REQ-043, REQ-044, REQ-048, REQ-049 | Design: Integration Tests

- [ ] Add unit test: generated YAML parses back into valid Config
- [ ] Add unit test: only leaves in brews, not deps
- [ ] Add unit test: casks included in output
- [ ] Add unit test: taps included in output
- [ ] Add unit test: output goes to stdout by default
- [ ] Add unit test: output includes pin and cleanup sections with defaults
- [ ] Add integration test: `syncd init` produces valid config
- [ ] Add integration test: `--output` creates file
- [ ] Add integration test: `--output` without `--force` refuses overwrite
- [ ] Verify: `make test && make integration-test` passes

**Estimate:** ~1.5 hours

---

## Phase 5: Verbose & Install (~1 hour)

### 5.1 Verbose flag
> REQ-057, REQ-058 | Design: Verbose Flag

- [ ] Add `--verbose` persistent flag on root command
- [ ] Add `RunVerbose(name, args)` to `ExecRunner` — streams stdout/stderr
- [ ] Update executor to use `RunVerbose` when flag is set
- [ ] Verify: `syncd apply --yes --verbose` shows brew output
- [ ] Verify: `syncd upgrade --yes --verbose` shows brew output
- [ ] Verify: without `--verbose`, only ✓/✗ lines shown (apply)
- [ ] Verify: without `--verbose`, only ✓/✗ lines shown (upgrade)

### 5.2 Makefile install targets
> REQ-059, REQ-060 | Design: Makefile Additions

- [ ] Add `install` target: `cp bin/syncd /usr/local/bin/syncd`
- [ ] Add `uninstall` target: `rm -f /usr/local/bin/syncd`
- [ ] Verify: `make install && which syncd` returns `/usr/local/bin/syncd`
- [ ] Verify: `make uninstall && which syncd` returns nothing

### 5.3 End-to-end validation
> All REQs | Full user workflow

- [ ] Workflow: `make install` → `syncd init > config.yaml` → `syncd plan --config config.yaml` → exit 0
- [ ] Workflow: `syncd upgrade --yes` → packages upgraded (except pinned)
- [ ] Workflow: `syncd plan` with dep-only formula → not in removal list
- [ ] Workflow: piped output → no color codes

**Estimate:** ~1 hour

---

## Summary

| Phase | Estimate | Key Deliverable |
|-------|----------|-----------------|
| 1. Dependency-Aware Removal | ~1.5h | `remove_unlisted` only targets leaves |
| 2. Colored Output | ~1h | Scannable terminal output |
| 3. Upgrade Command | ~2h | `syncd upgrade` with pin support |
| 4. Init Command | ~1.5h | `syncd init` generates config from state |
| 5. Verbose & Install | ~1h | Global install + debug output |
| **Total** | **~7h** | |

---

## Coverage Summary

### Requirements Covered

All 34 requirements (REQ-033 through REQ-066) are covered by tasks above.

| Requirement Range | Phase | Tasks |
|-------------------|-------|-------|
| REQ-033–039 | Phase 3 | 3.3, 3.4, 3.5 |
| REQ-040–042 | Phase 3 | 3.1, 3.4, 3.5 |
| REQ-043–049 | Phase 4 | 4.1, 4.2, 4.3 |
| REQ-050–052 | Phase 1 | 1.1, 1.2, 1.3 |
| REQ-053–056 | Phase 2 | 2.1, 2.2 |
| REQ-064 | Phase 2 | 2.1, 2.2 |
| REQ-057–058 | Phase 5 | 5.1 |
| REQ-059–060 | Phase 5 | 5.2 |
| REQ-061–063 | Phase 1, 3 | 1.1, 3.2 |
| REQ-065 | Phase 4 | 4.1, 4.3 |
| REQ-066 | Phase 3 | 3.4, 3.5 |

### Uncovered Requirements

None — all 34 requirements have at least one task with a concrete verification method.
