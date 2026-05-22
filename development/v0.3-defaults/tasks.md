# Tasks: syncd v0.3 — macOS Defaults

## Phase 1: Config Schema & Validation (~1.5 hours)

### 1.1 Config struct: DefaultEntry
> REQ-067, REQ-068, REQ-069, REQ-070, REQ-072, REQ-073 | Design: Config Changes

- [ ] Add `DefaultEntry` struct to `internal/config/config.go`
  - Fields: `Domain`, `Key`, `Type`, `Value` (interface{}), `Kill` ([]string, omitempty)
- [ ] Add `Defaults []DefaultEntry` field to `Config` struct
- [ ] Verify: existing configs without `defaults` still parse correctly
- [ ] Verify: `go build ./...` compiles

### 1.2 Validation: ValidateDefaults
> REQ-068, REQ-069, REQ-070, REQ-071 | Design: Validation

- [ ] Create `internal/config/validate.go`
  - `ValidateDefaults(entries []DefaultEntry) error`
  - Check domain, key, type required
  - Check type is one of: string, int, float, bool
  - Check value is present and matches declared type
- [ ] Call `ValidateDefaults` from `Load()` after decode
- [ ] Add unit tests in `internal/config/config_test.go`
  - Test: valid defaults section parses correctly
  - Test: missing domain returns error
  - Test: missing key returns error
  - Test: invalid type returns error
  - Test: type/value mismatch returns error (e.g., type: int, value: "hello")
  - Test: missing value returns error
  - Test: unknown field in defaults entry rejected by KnownFields
  - Test: optional kill field parses when present
  - Test: optional kill field absent → nil/empty
- [ ] Verify: `go test ./internal/config/...` passes

**Estimate:** ~1.5 hours

---

## Phase 2: Defaults State Reader & Drift (~1.5 hours)

### 2.1 State reader: ReadValue and ReadType
> REQ-074, REQ-098, REQ-101, REQ-102, REQ-103 | Design: Defaults State Reader

- [ ] Create `internal/defaults/state.go`
  - `ReadValue(runner, domain, key) (string, bool, error)` — runs `defaults read <domain> <key>`
  - `ReadType(runner, domain, key) (string, bool, error)` — runs `defaults read-type <domain> <key>`
  - Handle exit code 1 as "unset" (return "", false, nil)
- [ ] Add unit tests in `internal/defaults/state_test.go`
  - Test: successful read returns trimmed value
  - Test: command exit 1 returns ("", false, nil)
  - Test: other errors propagate
- [ ] Verify: `go test ./internal/defaults/...` passes

### 2.2 Type comparison: CompareValue
> REQ-099, REQ-100 | Design: Type Comparison Logic

- [ ] Create `internal/defaults/compare.go`
  - `CompareValue(rawOutput, configType string, configValue interface{}) bool`
  - Int: compare string representation
  - Float: compare string representation
  - Bool: normalize `1`/`0` from defaults read to `true`/`false`
  - String: direct comparison
- [ ] Add unit tests in `internal/defaults/compare_test.go`
  - Test: int 48 matches "48"
  - Test: float 0.5 matches "0.5"
  - Test: bool true matches "1"
  - Test: bool false matches "0"
  - Test: string "hello" matches "hello"
  - Test: int 48 does NOT match "49"
  - Test: bool true does NOT match "0"
- [ ] Verify: `go test ./internal/defaults/...` passes

### 2.3 Drift calculator: ComputeDrift
> REQ-075, REQ-076, REQ-077 | Design: Drift Calculator

- [ ] Create `internal/defaults/drift.go`
  - `DriftEntry` struct: Domain, Key, Type, Current, Desired, Kill
  - `ComputeDrift(runner, entries) ([]DriftEntry, error)`
  - For each entry: read value, compare, include in result if different or unset
- [ ] Add unit tests in `internal/defaults/drift_test.go`
  - Test: drifted value included in result
  - Test: unset key included with Current = "unset"
  - Test: matching value excluded from result
  - Test: multiple entries, mix of drifted and matching
- [ ] Verify: `go test ./internal/defaults/...` passes

**Estimate:** ~1.5 hours

---

## Phase 3: Plan Integration (~1 hour)

### 3.1 Plan command: show defaults drift
> REQ-074, REQ-075, REQ-076, REQ-077, REQ-078, REQ-079, REQ-080 | Design: CLI Changes — Plan

- [ ] Update `NewPlanCmd` in `internal/cli/plan.go`
  - After brew plan, compute defaults drift if `cfg.Defaults` is non-empty
  - Print drift section with colored output
  - Yellow `~` for value change, green `+` for unset → desired
  - Format: `domain key: current → desired`
  - Update exit code: exit 2 if `p.HasChanges() || len(drifted) > 0`
- [ ] Verify: `syncd plan` with no defaults in config → unchanged behavior
- [ ] Verify: `syncd plan` with drifted defaults → shows drift, exit 2
- [ ] Verify: `syncd plan` with matching defaults → no drift shown

### 3.2 Integration tests: plan with defaults
> REQ-075, REQ-076, REQ-077, REQ-079, REQ-080 | Design: Integration Tests

- [ ] Add fake `defaults` script to integration test infrastructure
  - Stores values in `$FAKE_DEFAULTS_STATE` directory as `domain/key` files
  - Supports `read`, `read-type`, `write` subcommands
  - Returns exit 1 for missing domain/key
- [ ] Add integration test: plan shows drift for changed value
- [ ] Add integration test: plan shows "unset" for missing key
- [ ] Add integration test: plan hides matching values
- [ ] Add integration test: plan exit 2 with only defaults drift (no brew changes)
- [ ] Add integration test: plan does not call `defaults write`
- [ ] Verify: `make integration-test` passes

**Estimate:** ~1 hour

---

## Phase 4: Defaults Executor & Apply (~2 hours)

### 4.1 Executor: WriteDrifted
> REQ-081, REQ-082, REQ-083, REQ-084, REQ-085, REQ-086, REQ-104, REQ-105, REQ-106 | Design: Defaults Executor

- [ ] Create `internal/defaults/executor.go`
  - `WriteResult` struct: Domain, Key, Err
  - `KillResult` struct: App, Err
  - `WriteDrifted(runner, drifted) ([]WriteResult, []KillResult)`
  - For each entry: `defaults write <domain> <key> -<type> <value>`
  - Bool value mapping: true → `TRUE`, false → `FALSE`
  - Collect kill apps from successful writes, deduplicate
  - Run `killall <app>` for each unique app (non-fatal)
- [ ] Add unit tests in `internal/defaults/executor_test.go`
  - Test: correct command args for int write
  - Test: correct command args for bool write (TRUE/FALSE)
  - Test: correct command args for string write
  - Test: correct command args for float write
  - Test: kill deduplication (same app listed twice → one killall)
  - Test: killall failure does not mark apply as failed
  - Test: write failure continues to next entry
  - Test: no killall when no entries drifted
- [ ] Verify: `go test ./internal/defaults/...` passes

### 4.2 Apply command: write defaults
> REQ-081, REQ-087, REQ-088, REQ-089, REQ-090 | Design: CLI Changes — Apply

- [ ] Update `NewApplyCmd` in `internal/cli/apply.go`
  - After brew execution, compute defaults drift
  - Include defaults drift in plan display (before confirmation prompt)
  - Write drifted defaults via `WriteDrifted`
  - Print write results (✓/✗ per entry)
  - Print kill results
  - Exit 1 if any write failed
- [ ] Verify: `syncd apply --yes` with drifted defaults → writes them
- [ ] Verify: `syncd apply` without `--yes` → shows drift before prompt

### 4.3 Integration tests: apply with defaults
> REQ-081, REQ-083, REQ-084, REQ-086, REQ-090 | Design: Integration Tests

- [ ] Add integration test: apply writes drifted defaults
- [ ] Add integration test: apply kills affected apps
- [ ] Add integration test: apply deduplicates kills
- [ ] Add integration test: apply continues on write failure
- [ ] Add integration test: apply idempotent (second run = no drift)
- [ ] Verify: `make integration-test` passes

**Estimate:** ~2 hours

---

## Phase 5: Init Snapshot (~1.5 hours)

### 5.1 Well-known domain map
> REQ-094, REQ-095 | Design: Well-Known Domain Map

- [ ] Create `internal/defaults/apps.go`
  - `KnownApps` map: domain → app name
  - Include: com.apple.dock → Dock, com.apple.finder → Finder, com.apple.systemuiserver → SystemUIServer, NSGlobalDomain → ""
- [ ] Verify: compiles

### 5.2 Init command: --defaults flag
> REQ-091, REQ-092, REQ-093, REQ-094, REQ-095, REQ-096, REQ-097 | Design: CLI Changes — Init

- [ ] Update `NewInitCmd` in `internal/cli/init.go`
  - Add `--defaults` flag (string, comma-separated `domain:key` pairs)
  - Parse pairs, call `ReadType` and `ReadValue` for each
  - Skip unreadable entries with warning to stderr
  - Look up domain in `KnownApps` for kill inference
  - Omit kill field for unknown domains
  - Add entries to generated config's `Defaults` field
  - Works alongside existing brew snapshot
- [ ] Verify: `syncd init --defaults "com.apple.dock:tilesize"` outputs defaults section

### 5.3 Integration tests: init with defaults
> REQ-091, REQ-094, REQ-096, REQ-097 | Design: Integration Tests

- [ ] Add integration test: init snapshots specified defaults
- [ ] Add integration test: init infers kill for known domains
- [ ] Add integration test: init skips unreadable with warning
- [ ] Add integration test: init combined with brew snapshot
- [ ] Verify: `make integration-test` passes

**Estimate:** ~1.5 hours

---

## Phase 6: Verbose & Polish (~0.5 hours)

### 6.1 Verbose output for defaults
> REQ-107 | Design: CLI Changes — verbose

- [ ] Ensure `--verbose` prints `defaults read` and `defaults write` command output
  - Use `RunMutate` for writes (already streams when verbose)
  - For reads during plan, print command + output when verbose flag set
- [ ] Verify: `syncd plan --verbose` shows defaults read output
- [ ] Verify: `syncd apply --yes --verbose` shows defaults write output

### 6.2 End-to-end validation
> All REQs | Full user workflow

- [ ] Workflow: config with defaults → `syncd plan` → shows drift → `syncd apply --yes` → defaults written
- [ ] Workflow: `syncd plan` after apply → exit 0, no drift
- [ ] Workflow: `syncd init --defaults "com.apple.dock:tilesize"` → valid YAML with defaults section
- [ ] Workflow: config with only defaults (no brews) → plan/apply work correctly
- [ ] Workflow: config with both brews and defaults → both handled in single run

**Estimate:** ~0.5 hours

---

## Summary

| Phase | Estimate | Key Deliverable |
|-------|----------|-----------------|
| 1. Config Schema & Validation | ~1.5h | Config accepts and validates `defaults` section |
| 2. Defaults State Reader & Drift | ~1.5h | Can detect drift between config and system state |
| 3. Plan Integration | ~1h | `syncd plan` shows defaults drift |
| 4. Defaults Executor & Apply | ~2h | `syncd apply` writes defaults and restarts apps |
| 5. Init Snapshot | ~1.5h | `syncd init --defaults` snapshots preferences |
| 6. Verbose & Polish | ~0.5h | Verbose output and end-to-end validation |
| **Total** | **~8h** | |

---

## Coverage Summary

### Requirements Covered

All 41 requirements (REQ-067 through REQ-107) are covered by tasks above.

| Requirement Range | Phase | Tasks |
|-------------------|-------|-------|
| REQ-067–073 | Phase 1 | 1.1, 1.2 |
| REQ-074–080 | Phase 2, 3 | 2.1, 2.3, 3.1, 3.2 |
| REQ-081–090 | Phase 4 | 4.1, 4.2, 4.3 |
| REQ-091–097 | Phase 5 | 5.1, 5.2, 5.3 |
| REQ-098–100 | Phase 2 | 2.1, 2.2 |
| REQ-101–103 | Phase 2 | 2.1 |
| REQ-104–106 | Phase 4 | 4.1 |
| REQ-107 | Phase 6 | 6.1 |

### Uncovered Requirements

None — all 41 requirements have at least one task with a concrete verification method.
