# Traceability Matrix: syncd v0.3 — macOS Defaults

## Forward Trace: Requirements → Design → Tasks

| Req ID | Requirement Summary | Design Section | Task IDs | Verification Method |
|--------|-------------------|----------------|----------|-------------------|
| REQ-067 | `defaults` section in config as list of entries | Config Changes — `Defaults` field | 1.1 | Unit test — parse config with `defaults` section |
| REQ-068 | Each entry requires `domain` field | Config Changes — `DefaultEntry.Domain` | 1.1, 1.2 | Unit test — missing domain returns error |
| REQ-069 | Each entry requires `key` field | Config Changes — `DefaultEntry.Key` | 1.1, 1.2 | Unit test — missing key returns error |
| REQ-070 | Each entry requires `type` field (string/int/float/bool) | Validation — `allowedTypes` | 1.1, 1.2 | Unit test — invalid type returns error |
| REQ-071 | Each entry requires `value` matching declared type | Validation — `validateTypeMatch` | 1.2 | Unit test — type/value mismatch returns error |
| REQ-072 | Optional `kill` field (list of app names) | Config Changes — `DefaultEntry.Kill` | 1.1, 1.2 | Unit test — parse with/without kill field |
| REQ-073 | Reject unknown fields via `KnownFields(true)` | Config Changes — `KnownFields(true)` | 1.2 | Unit test — extra field rejected |
| REQ-074 | Plan reads current value via `defaults read` | Defaults State Reader — `ReadValue` | 2.1, 3.1, 3.2 | Integration test — `defaults read` invoked |
| REQ-075 | Plan shows defaults with differing values | Drift Calculator — `ComputeDrift` | 2.3, 3.1, 3.2 | Integration test — drift shown with current → desired |
| REQ-076 | Plan shows unset domain/key as drift | Drift Calculator — unset handling | 2.3, 3.1, 3.2 | Integration test — non-existent key appears as drift |
| REQ-077 | Plan hides matching defaults | Drift Calculator — match = no drift | 2.3, 3.1, 3.2 | Integration test — matching value not in output |
| REQ-078 | Plan displays drift as `domain key: current → desired` | CLI Changes — Plan output format | 3.1 | Integration test — output format matches |
| REQ-079 | Defaults drift counts for exit code 2 | CLI Changes — `hasChanges` includes drift | 3.1, 3.2 | Integration test — exit 2 with only defaults drift |
| REQ-080 | Plan does not modify defaults (read-only) | CLI Changes — plan is read-only | 3.1, 3.2 | Integration test — no `defaults write` called |
| REQ-081 | Apply writes drifted defaults via `defaults write` | Defaults Executor — `WriteDrifted` | 4.1, 4.2, 4.3 | Integration test — value written after apply |
| REQ-082 | Apply uses correct type flag (-string/-int/-float/-bool) | Defaults Executor — type flag mapping | 4.1 | Unit test — correct flag per type |
| REQ-083 | Apply restarts apps in `kill` field | Defaults Executor — kill after write | 4.1, 4.3 | Integration test — `killall` executed |
| REQ-084 | Apply deduplicates app restarts | Defaults Executor — kill deduplication | 4.1, 4.3 | Integration test — `killall` called once |
| REQ-085 | Apply does not restart when no drift | Defaults Executor — no kill when no drift | 4.1 | Unit test — no killall when no drift |
| REQ-086 | Apply continues on write failure | Defaults Executor — continue on error | 4.1, 4.3 | Integration test — good entries still written |
| REQ-087 | Apply reports which defaults failed | Defaults Executor — `WriteResult.Err` | 4.1, 4.2 | Integration test — error reported with domain/key |
| REQ-088 | Apply exits 1 if any write fails | CLI Changes — exit 1 on failure | 4.2 | Integration test — exit code 1 on failure |
| REQ-089 | Apply includes defaults in confirmation prompt | CLI Changes — Apply confirmation includes drift | 4.2 | Integration test — drift shown before prompt |
| REQ-090 | Apply is idempotent | Defaults Executor — idempotent writes | 4.3 | Integration test — second run = no drift |
| REQ-091 | Init accepts `--defaults` flag with domain:key pairs | CLI Changes — Init `--defaults` flag | 5.2, 5.3 | Integration test — output includes entry |
| REQ-092 | Init reads type and value via `defaults read-type`/`read` | Defaults State Reader — `ReadType` + `ReadValue` | 2.1, 5.2 | Integration test — type and value correct |
| REQ-093 | Init includes defaults in generated YAML | CLI Changes — Init YAML output | 5.2, 5.3 | Integration test — `defaults:` section in output |
| REQ-094 | Init infers `kill` for well-known domains | Well-Known Domain Map — `KnownApps` | 5.1, 5.2, 5.3 | Integration test — `kill: [Dock]` in output |
| REQ-095 | Init omits `kill` for unknown domains | Well-Known Domain Map — omit unknown | 5.1, 5.2, 5.3 | Integration test — no `kill` for unknown domain |
| REQ-096 | Init skips unreadable defaults with warning | CLI Changes — Init skip + warning | 5.2, 5.3 | Integration test — warning printed, entry absent |
| REQ-097 | Init `--defaults` works alongside brew snapshot | CLI Changes — Init combined output | 5.2, 5.3 | Integration test — both `brews` and `defaults` in output |
| REQ-098 | Read `defaults read-type` for stored type | Defaults State Reader — `ReadType` | 2.1 | Unit test — type parsed from mock output |
| REQ-099 | Type-aware comparison (int 48 = "48") | Type Comparison Logic — `CompareValue` | 2.2 | Unit test — int match confirmed |
| REQ-100 | Bool mapping: `1`/`0` ↔ `true`/`false` | Type Comparison Logic — bool mapping | 2.2 | Unit test — `1` matches `true`, `0` matches `false` |
| REQ-101 | Query `defaults read <domain> <key>` for value | Defaults State Reader — `ReadValue` | 2.1 | Unit test — value parsed from mock output |
| REQ-102 | Query `defaults read-type <domain> <key>` for type | Defaults State Reader — `ReadType` | 2.1 | Unit test — type string parsed |
| REQ-103 | Handle exit code 1 as "unset" | Defaults State Reader — exit 1 = unset | 2.1 | Unit test — command failure = unset |
| REQ-104 | Execute `defaults write <domain> <key> -<type> <value>` | Defaults Executor — write command | 4.1 | Unit test — correct command args |
| REQ-105 | Execute `killall <app>` to restart apps | Defaults Executor — killall command | 4.1 | Unit test — killall called with app name |
| REQ-106 | `killall` failure is non-fatal | Defaults Executor — killall non-fatal | 4.1 | Unit test — error does not fail apply |
| REQ-107 | `--verbose` prints defaults read/write output | CLI Changes — verbose output | 6.1 | Manual test — command output visible |

## Reverse Trace: Design Sections → Requirement IDs

| Design Section | Requirement IDs |
|----------------|-----------------|
| System Architecture | REQ-067, REQ-074, REQ-081 |
| New Components | REQ-074, REQ-075, REQ-081, REQ-094 |
| Design Decision 1 (new package) | REQ-067, REQ-074, REQ-081 |
| Design Decision 2 (reuse CommandRunner) | REQ-074, REQ-101, REQ-104 |
| Design Decision 3 (drift struct) | REQ-075, REQ-076, REQ-077 |
| Design Decision 4 (type-aware comparison) | REQ-099, REQ-100 |
| Design Decision 5 (kill deduplication) | REQ-083, REQ-084 |
| Design Decision 6 (killall non-fatal) | REQ-106 |
| Design Decision 7 (init snapshots keys) | REQ-091, REQ-092 |
| Config Changes | REQ-067, REQ-068, REQ-069, REQ-070, REQ-071, REQ-072, REQ-073 |
| Validation | REQ-068, REQ-069, REQ-070, REQ-071 |
| Defaults State Reader | REQ-074, REQ-092, REQ-098, REQ-101, REQ-102, REQ-103 |
| Drift Calculator | REQ-075, REQ-076, REQ-077 |
| Type Comparison Logic | REQ-099, REQ-100 |
| Defaults Executor | REQ-081, REQ-082, REQ-083, REQ-084, REQ-085, REQ-086, REQ-087, REQ-104, REQ-105, REQ-106 |
| Well-Known Domain Map | REQ-094, REQ-095 |
| CLI Changes — Plan | REQ-074, REQ-075, REQ-076, REQ-077, REQ-078, REQ-079, REQ-080 |
| CLI Changes — Apply | REQ-081, REQ-087, REQ-088, REQ-089, REQ-090 |
| CLI Changes — Init | REQ-091, REQ-092, REQ-093, REQ-094, REQ-095, REQ-096, REQ-097 |
| CLI Changes — Verbose | REQ-107 |

## Reverse Trace: Task IDs → Requirement IDs

| Task ID | Task Title | Requirement IDs |
|---------|------------|-----------------|
| 1.1 | Config struct: DefaultEntry | REQ-067, REQ-068, REQ-069, REQ-070, REQ-072, REQ-073 |
| 1.2 | Validation: ValidateDefaults | REQ-068, REQ-069, REQ-070, REQ-071, REQ-072, REQ-073 |
| 2.1 | State reader: ReadValue and ReadType | REQ-074, REQ-098, REQ-101, REQ-102, REQ-103 |
| 2.2 | Type comparison: CompareValue | REQ-099, REQ-100 |
| 2.3 | Drift calculator: ComputeDrift | REQ-075, REQ-076, REQ-077 |
| 3.1 | Plan command: show defaults drift | REQ-074, REQ-075, REQ-076, REQ-077, REQ-078, REQ-079, REQ-080 |
| 3.2 | Integration tests: plan with defaults | REQ-075, REQ-076, REQ-077, REQ-079, REQ-080 |
| 4.1 | Executor: WriteDrifted | REQ-081, REQ-082, REQ-083, REQ-084, REQ-085, REQ-086, REQ-104, REQ-105, REQ-106 |
| 4.2 | Apply command: write defaults | REQ-081, REQ-087, REQ-088, REQ-089, REQ-090 |
| 4.3 | Integration tests: apply with defaults | REQ-081, REQ-083, REQ-084, REQ-086, REQ-090 |
| 5.1 | Well-known domain map | REQ-094, REQ-095 |
| 5.2 | Init command: --defaults flag | REQ-091, REQ-092, REQ-093, REQ-094, REQ-095, REQ-096, REQ-097 |
| 5.3 | Integration tests: init with defaults | REQ-091, REQ-094, REQ-096, REQ-097 |
| 6.1 | Verbose output for defaults | REQ-107 |
| 6.2 | End-to-end validation | All REQs (workflow validation) |

## Coverage Summary

- **Requirements with tasks:** 41/41 (100%)
- **Requirements with design mapping:** 41/41 (100%)
- **Requirements with verification method:** 41/41 (100%)
- **Orphan tasks (no requirement):** None
- **Orphan requirements (no task):** None
