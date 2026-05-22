# Traceability Matrix: syncd v0.2 — Polish & Usability

## Forward Trace: Requirements -> Design -> Tasks

| Requirement | User Story | Design Section | Task IDs | Validation Path |
|-------------|------------|----------------|----------|-----------------|
| REQ-033 | US-005 | Upgrade Command Flow; Executor addition | 3.3, 3.4, 3.5 | CLI/integration: `syncd upgrade --yes` upgrades formulae |
| REQ-034 | US-005 | Upgrade Command Flow; Executor addition | 3.3, 3.4, 3.5 | CLI/integration: `syncd upgrade --yes` upgrades casks |
| REQ-035 | US-006 | Config Changes; Upgrade Command Flow | 3.1, 3.4, 3.5 | CLI/integration: pinned package is skipped |
| REQ-036 | US-005 | Upgrade Command Flow | 3.4, 3.5 | CLI/integration: prompt blocks upgrade unless confirmed |
| REQ-037 | US-005 | Upgrade Command Flow | 3.4 | CLI/integration: upgrade plan printed before prompt |
| REQ-038 | US-005 | Upgrade Command Flow | 3.4, 3.5 | CLI/integration: exit 0 on success, 1 on failure |
| REQ-039 | US-005 | Upgrade Command Flow; Executor addition | 3.3, 3.5 | Unit/integration: failure does not stop remaining upgrades |
| REQ-040 | US-006 | Config Changes | 3.1 | Unit: config parses `pin` list |
| REQ-041 | US-006 | Config Changes | 3.1 | Unit: config rejects invalid `pin` shape/value types |
| REQ-042 | US-006 | Config Changes; Upgrade Command Flow | 3.4, 3.5 | CLI/integration: pinned package absent from upgrade output/execution |
| REQ-043 | US-007 | Init Command Flow | 4.1, 4.3 | CLI/integration: `syncd init` emits valid YAML |
| REQ-044 | US-007 | Init Command Flow; New State Queries | 4.1, 4.3 | CLI/integration: generated brews use leaves only |
| REQ-045 | US-007 | Init Command Flow | 4.1 | CLI/integration: generated config includes installed casks |
| REQ-046 | US-007 | Init Command Flow | 4.1 | CLI/integration: generated config includes taps |
| REQ-047 | US-007 | Init Command Flow | 4.1 | CLI/integration: stdout default can be piped |
| REQ-048 | US-007 | Init Command Flow | 4.2, 4.3 | CLI/integration: `--output` creates requested file |
| REQ-049 | US-007 | Init Command Flow | 4.2, 4.3 | CLI/integration: overwrite refused without `--force` |
| REQ-050 | US-008 | Dependency-Aware Removal | 1.2, 1.3 | Plan/integration: dependency-only formula not listed for removal |
| REQ-051 | US-008 | Dependency-Aware Removal | 1.2, 1.3 | Apply/integration: dependency-only formula not removed |
| REQ-052 | US-008 | New State Queries; Dependency-Aware Removal | 1.1 | Unit/code review: `brew leaves` queried and parsed |
| REQ-053 | US-009 | Colored Output | 2.1 | CLI/integration: additions use green `+` in terminal |
| REQ-054 | US-009 | Colored Output | 2.1 | CLI/integration: removals use red `-` in terminal |
| REQ-055 | US-009 | Colored Output | 2.1 | CLI/integration: maintenance actions use yellow `~` in terminal |
| REQ-056 | US-009 | Colored Output | 2.1, 2.2 | CLI/integration: piped output has no ANSI codes |
| REQ-057 | US-009 | Verbose Flag | 5.1 | CLI/integration: `--verbose` streams brew output for apply/upgrade |
| REQ-058 | US-009 | Verbose Flag | 5.1 | CLI/integration: default output stays summarized |
| REQ-059 | US-010 | Makefile Additions | 5.2, 5.3 | Real command: `make install` makes `syncd` available on PATH |
| REQ-060 | US-010 | Makefile Additions | 5.2 | Real command: `make uninstall` removes installed binary |
| REQ-061 | US-008 | New State Queries | 1.1 | Unit/code review: `brew leaves` output parsed into state |
| REQ-062 | US-005 | New State Queries | 3.2 | Unit/code review: `brew outdated --formula -1` parsed |
| REQ-063 | US-005 | New State Queries | 3.2 | Unit/code review: `brew outdated --cask -1` parsed |
| REQ-064 | US-009 | Colored Output; Design Decision 7 | 2.1, 2.2 | CLI/integration: `--no-color` suppresses ANSI in TTY |
| REQ-065 | US-007 | Init Command Flow | 4.1, 4.3 | Unit/integration: init output includes pin and cleanup sections |
| REQ-066 | US-005 | Upgrade Command Flow; Design Decision 6 | 3.4, 3.5 | CLI/integration: upgrade works without config file |

## Reverse Trace: Design Sections -> Requirement IDs

| Design Section | Requirement IDs |
|----------------|-----------------|
| System Architecture | REQ-033, REQ-034, REQ-043, REQ-050, REQ-053, REQ-057 |
| New/Modified Components | REQ-033, REQ-034, REQ-039, REQ-040, REQ-042, REQ-043, REQ-049, REQ-050, REQ-052, REQ-053, REQ-056, REQ-061, REQ-062, REQ-063 |
| Design Decisions — `brew leaves` for removal filtering | REQ-050, REQ-051, REQ-052, REQ-061 |
| Design Decisions — Separate `upgrade` command | REQ-033, REQ-034, REQ-035, REQ-036, REQ-037, REQ-038, REQ-039, REQ-042 |
| Design Decisions — Color via ANSI with TTY detection | REQ-053, REQ-054, REQ-055, REQ-056, REQ-064 |
| Design Decisions — Pin in config | REQ-035, REQ-040, REQ-041, REQ-042 |
| Design Decisions — Upgrade is config-optional | REQ-066 |
| Design Decisions — `--no-color` flag | REQ-064 |
| Config Changes | REQ-035, REQ-040, REQ-041, REQ-042 |
| New State Queries | REQ-044, REQ-052, REQ-061, REQ-062, REQ-063 |
| Dependency-Aware Removal | REQ-050, REQ-051, REQ-052 |
| Upgrade Command Flow | REQ-033, REQ-034, REQ-035, REQ-036, REQ-037, REQ-038, REQ-039, REQ-042, REQ-062, REQ-063, REQ-066 |
| Init Command Flow | REQ-043, REQ-044, REQ-045, REQ-046, REQ-047, REQ-048, REQ-049, REQ-065 |
| Colored Output | REQ-053, REQ-054, REQ-055, REQ-056 |
| Verbose Flag | REQ-057, REQ-058 |
| Makefile Additions | REQ-059, REQ-060 |
| Testing Strategy | REQ-039, REQ-042, REQ-043, REQ-048, REQ-049, REQ-050, REQ-051, REQ-056 |

## Reverse Trace: Task IDs -> Requirement IDs

| Task ID | Task Title | Requirement IDs |
|---------|------------|-----------------|
| 1.1 | State reader: GetLeaves | REQ-052, REQ-061 |
| 1.2 | Plan: leaf-only removal | REQ-050, REQ-051 |
| 1.3 | Integration test: dependency-aware removal | REQ-050, REQ-051 |
| 2.1 | Output formatter | REQ-053, REQ-054, REQ-055, REQ-056, REQ-064 |
| 2.2 | Color tests | REQ-056, REQ-064 |
| 3.1 | Config: pin field | REQ-040, REQ-041 |
| 3.2 | State reader: GetOutdated | REQ-062, REQ-063 |
| 3.3 | Executor: Upgrade function | REQ-033, REQ-034, REQ-039 |
| 3.4 | Upgrade CLI command | REQ-035, REQ-036, REQ-037, REQ-038, REQ-042, REQ-066 |
| 3.5 | Upgrade integration tests | REQ-035, REQ-038, REQ-039, REQ-042, REQ-066 |
| 4.1 | Init CLI command | REQ-043, REQ-044, REQ-045, REQ-046, REQ-047, REQ-065 |
| 4.2 | Init output flag | REQ-048, REQ-049 |
| 4.3 | Init tests | REQ-043, REQ-044, REQ-048, REQ-049, REQ-065 |
| 5.1 | Verbose flag | REQ-057, REQ-058 |
| 5.2 | Makefile install targets | REQ-059, REQ-060 |
| 5.3 | End-to-end validation | REQ-043, REQ-050, REQ-056, REQ-059 |

## Coverage Summary

- All requirements REQ-033 through REQ-066 have at least one design section and one task ID.
- All task IDs trace back to one or more requirement IDs.
- All open questions resolved; decisions reflected in requirements and design.
