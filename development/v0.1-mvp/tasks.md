# Tasks: syncd v0.1 — MVP

## Phase 1: Scaffolding & Config (~2 hours)

### 1.1 Project initialization
> REQ-025, REQ-028, REQ-029, REQ-032 | Design: Project Layout, Config Resolution

- [x] Initialize Go module (`go mod init github.com/joedorseyjr/syncd`)
- [x] Add dependencies: `cobra`, `gopkg.in/yaml.v3`
- [x] Create `cmd/syncd/main.go` with cobra root command
- [x] Add `--config <path>` persistent flag on root command (REQ-032)
- [x] Set cobra `Version` field for `--version` output
- [x] Create `Makefile` with `build`, `test`, `lint` targets (`CGO_ENABLED=0` in build target)
- [x] Verify: `go build ./cmd/syncd` produces single binary

### 1.2 Config parser
> REQ-001, REQ-002, REQ-003, REQ-004 | Design: Config Resolution, Core Types

- [x] Create `internal/config/config.go` — `Config` and `Cleanup` structs
- [x] Implement `Load(path string) (*Config, error)`
  - Read from `~/.config/syncd/config.yaml`
  - Return clear error if file missing (REQ-002)
  - Return clear error if YAML invalid (REQ-003)
  - Treat missing sections as empty slices
- [x] Create `internal/config/config_test.go`
  - Test: valid config parses all sections correctly
  - Test: missing file returns descriptive error
  - Test: malformed YAML returns descriptive error
  - Test: config with only some sections still parses
- [x] Verify: `go test ./internal/config/...` passes

**Estimate:** ~1.5 hours

---

## Phase 2: State Reader (~1 hour)

### 2.1 CommandRunner interface
> REQ-027 | Design: Test Infrastructure, Homebrew Interaction

- [x] Create `internal/brew/runner.go` — `CommandRunner` interface
- [x] Implement `ExecRunner` (real `os/exec`) and `MockRunner` (for tests)
- [x] Verify: interface compiles, mock satisfies it

### 2.2 Homebrew state query
> REQ-027 | Design: Homebrew Interaction, Core Types

- [x] Create `internal/brew/state.go` — `State` struct + `GetState(runner) (*State, error)`
  - Parse `brew tap` output → `State.Taps`
  - Parse `brew list --formula -1` output → `State.Brews`
  - Parse `brew list --cask -1` output → `State.Casks`
- [x] Create `internal/brew/brew_test.go`
  - Test: parse multi-line brew output into string slices
  - Test: handle empty output (nothing installed)
  - Test: handle command failure gracefully
- [x] Verify: `go test ./internal/brew/...` passes

**Estimate:** ~1 hour

---

## Phase 3: Plan Command (~2 hours)

### 3.1 Diff calculator
> REQ-005, REQ-006, REQ-007, REQ-008, REQ-009, REQ-019, REQ-030 | Design: Diff Calculator

- [x] Create `internal/plan/plan.go` — `Plan` struct + `Compute(config, state) *Plan`
  - Taps in config but not in state → `TapsToAdd`
  - Taps in state but not in config → `TapsToRemove` (only if `remove_unlisted`)
  - Brews in config but not in state → `BrewsToInstall`
  - Casks in config but not in state → `CasksToInstall`
  - Brews in state but not in config → `BrewsToRemove` (only if `remove_unlisted`)
  - Casks in state but not in config → `CasksToRemove` (only if `remove_unlisted`)
  - Set `Autoremove` and `ClearCache` from cleanup config
- [x] Implement `Plan.IsEmpty() bool`
- [x] Create `internal/plan/plan_test.go`
  - Test: packages to add (tap, brew, cask)
  - Test: taps to remove when `remove_unlisted: true`
  - Test: packages to remove when `remove_unlisted: true`
  - Test: no removals when `remove_unlisted: false`
  - Test: empty plan when system matches config
  - Test: case sensitivity handling
- [x] Verify: `go test ./internal/plan/...` passes

### 3.2 Plan CLI command
> REQ-010, REQ-011 | Design: Command Flow, Exit Codes

- [x] Create `internal/cli/plan.go` — cobra `plan` subcommand
  - Load config, get state, compute plan
  - Print formatted diff (adds in green, removes in red)
  - Exit 0 if plan empty, exit 2 if changes pending
- [x] Wire into root command in `cmd/syncd/main.go`
- [x] Verify: `syncd plan` runs without modifying system; exit code correct

**Estimate:** ~2 hours

---

## Phase 4: Apply Command (~3 hours)

### 4.1 Executor
> REQ-014, REQ-015, REQ-016, REQ-017, REQ-018, REQ-020, REQ-021, REQ-023, REQ-024, REQ-031 | Design: Homebrew Interaction, Executor

- [x] Create `internal/brew/executor.go` — `Execute(runner, plan) []Result`
  - Run `brew tap` for each tap to add
  - Run `brew install` for each brew to install
  - Run `brew install --cask` for each cask to install
  - Run `brew untap` for each tap to remove (REQ-031)
  - Run `brew uninstall` for each brew to remove
  - Run `brew uninstall --cask` for each cask to remove
  - Run `brew autoremove` if flagged
  - Run `brew cleanup` if flagged
  - Execution order: taps → brew installs → cask installs → brew removals → cask removals → tap removals → autoremove → cleanup
  - Capture errors per operation, continue on failure (REQ-023)
  - Return all results including failures
- [x] Create `internal/brew/executor_test.go`
  - Test: successful install sequence
  - Test: one failure doesn't stop others
  - Test: tap removal executes `brew untap`
  - Test: autoremove/cleanup only run when flagged
- [x] Verify: `go test ./internal/brew/...` passes

### 4.2 Confirmation prompt
> REQ-012, REQ-013 | Design: Apply Command

- [x] Implement confirmation prompt in apply flow (print plan, ask y/n)
- [x] Implement `--yes` flag to skip prompt
- [x] Test: prompt blocks execution until confirmed (internal/cli/cli_test.go)
- [x] Test: `--yes` bypasses prompt (internal/cli/cli_test.go)

### 4.3 Apply CLI command
> REQ-012–022 | Design: Command Flow

- [x] Create `internal/cli/apply.go` — cobra `apply` subcommand
  - Load config, get state, compute plan
  - If plan empty, print "Already in sync" and exit 0
  - Show plan, prompt for confirmation (unless `--yes`)
  - Execute plan, print results
  - Exit 0 on full success, exit 1 if any operation failed
- [x] Wire into root command
- [x] Verify: full apply cycle works end-to-end

### 4.4 Idempotency validation
> REQ-022 | Design: Testing — idempotency

- [x] Run `syncd apply --yes`, then `syncd plan` → confirm exit 0
- [x] Verify: no side effects on second run

**Estimate:** ~3 hours

---

## Phase 5: Polish & Integration Testing (~1.5 hours)

### 5.1 Integration test suite
> REQ-010, REQ-022, REQ-023 | Design: Integration Tests

- [x] Create `test/integration_test.go` with `//go:build integration` tag
  - Test: `plan` doesn't modify system (compare `brew list` before/after)
  - Test: `apply` installs a declared package
  - Test: `apply` removes an undeclared package (with `cleanup.remove_unlisted: true`)
  - Test: `apply` does NOT remove an undeclared package (with `cleanup.remove_unlisted: false`)
  - Test: idempotency (apply twice, second plan is empty)
  - Test: nonexistent package reports error, others still succeed
- [x] Verify: `go test -tags integration ./test/...` passes

### 5.2 Build & release prep
> REQ-025, REQ-026 | Design: Single binary

- [x] Makefile targets: `build`, `test`, `integration-test`, `lint`, `clean`
- [x] Verify: `make build` produces static binary
- [x] Verify: binary runs on clean system with only Homebrew installed
- [x] Update `.gitignore` for build artifacts

### 5.3 End-to-end validation
> All REQs | Full user workflow

- [x] Workflow: fresh config → `syncd plan` → shows adds → `syncd apply --yes` → packages installed
- [x] Workflow: with `cleanup.remove_unlisted: true`, add package not in config → `syncd plan` → shows removal → `syncd apply --yes` → removed
- [x] Workflow: with `cleanup.remove_unlisted: false`, add package not in config → `syncd plan` → no removal shown
- [x] Workflow: `syncd plan` on synced system → exit 0, no output
- [x] Workflow: invalid config → clear error message, non-zero exit

**Estimate:** ~1.5 hours

---

## Summary

| Phase | Estimate | Key Deliverable |
|-------|----------|-----------------|
| 1. Scaffolding & Config | ~2h | Binary loads and validates config |
| 2. State Reader | ~1h | Queries Homebrew for installed packages |
| 3. Plan Command | ~2h | `syncd plan` shows diff without changes |
| 4. Apply Command | ~3h | `syncd apply` reconciles system state |
| 5. Polish & Integration | ~1.5h | Tested, stable v0.1 binary |
| **Total** | **~9.5h** | |

---

## Coverage Summary

### Requirements Covered

All 32 requirements (REQ-001 through REQ-032) are covered by tasks above.

| Requirement Range | Phase | Tasks |
|-------------------|-------|-------|
| REQ-001–004 | Phase 1 | 1.2 |
| REQ-005–011 | Phase 3 | 3.1, 3.2 |
| REQ-012–022 | Phase 4 | 4.1, 4.2, 4.3, 4.4 |
| REQ-023–024 | Phase 4 | 4.1 |
| REQ-025–029 | Phase 1, 2, 5 | 1.1, 2.1, 5.2 |
| REQ-030 | Phase 3 | 3.1 |
| REQ-031 | Phase 4 | 4.1 |
| REQ-032 | Phase 1 | 1.1 |

### Uncovered Requirements

None — all 32 requirements have at least one task with a concrete verification method.
