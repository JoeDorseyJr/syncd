# Traceability Matrix: syncd v0.1 — MVP

## Forward Trace: Requirement → Design → Task

| Req ID | Requirement Summary | Design Section | Task ID | Verification Method |
|--------|-------------------|----------------|---------|-------------------|
| REQ-001 | Read YAML config from `~/.config/syncd/config.yaml` | Config Resolution | 1.2 | Unit test — parse valid config |
| REQ-002 | Exit with error if config missing | Config Resolution — validation on load | 1.2 | Run with no config, confirm error + non-zero exit |
| REQ-003 | Exit with error if config invalid YAML | Config Resolution — validation on load | 1.2 | Run with malformed YAML, confirm error |
| REQ-004 | Support `taps`, `brews`, `casks`, `cleanup` sections | Core Types — Config struct | 1.2 | Unit test — all sections parsed |
| REQ-005 | Plan: list taps to add | Diff Calculator | 3.1 | Declare undeclared tap, run plan, confirm output |
| REQ-006 | Plan: list brews to install | Diff Calculator | 3.1 | Declare undeclared brew, run plan, confirm output |
| REQ-007 | Plan: list casks to install | Diff Calculator | 3.1 | Declare undeclared cask, run plan, confirm output |
| REQ-008 | Plan: list brews to remove (cleanup enabled) | Diff Calculator (cleanup flag) | 3.1 | Have undeclared brew installed, run plan, confirm output |
| REQ-009 | Plan: list casks to remove (cleanup enabled) | Diff Calculator (cleanup flag) | 3.1 | Have undeclared cask installed, run plan, confirm output |
| REQ-010 | Plan: read-only, no system modification | Command Flow — plan is read-only | 3.2, 5.1 | Compare `brew list` before/after plan |
| REQ-011 | Plan: exit 0 if no changes, exit 2 if pending | Exit Codes | 3.2 | Run on synced system → 0; add brew → 2 |
| REQ-012 | Apply: prompt for confirmation | Apply Command — confirmation prompt | 4.2, 4.3 | Run apply, confirm prompt, answer "n", no changes |
| REQ-013 | Apply: `--yes` skips prompt | Apply Command — `--yes` flag | 4.2, 4.3 | Run apply --yes, confirm no prompt |
| REQ-014 | Apply: add taps | Homebrew Interaction — `brew tap` | 4.1, 4.3 | Declare tap, apply, verify `brew tap` output |
| REQ-015 | Apply: install brews | Homebrew Interaction — `brew install` | 4.1, 4.3 | Declare brew, apply, verify `brew list` |
| REQ-016 | Apply: install casks | Homebrew Interaction — `brew install --cask` | 4.1, 4.3 | Declare cask, apply, verify `brew list --cask` |
| REQ-017 | Apply: remove brews (cleanup enabled) | Homebrew Interaction — `brew uninstall` | 4.1, 4.3 | Install undeclared brew, apply, verify removed |
| REQ-018 | Apply: remove casks (cleanup enabled) | Homebrew Interaction — `brew uninstall --cask` | 4.1, 4.3 | Install undeclared cask, apply, verify removed |
| REQ-019 | Apply: no removals when cleanup disabled | Plan Command — cleanup flag gating | 3.1, 4.3 | Set `remove_unlisted: false`, verify package remains |
| REQ-020 | Apply: run `brew autoremove` | Homebrew Interaction — `brew autoremove` | 4.1 | Apply with `autoremove: true`, confirm in output |
| REQ-021 | Apply: run `brew cleanup` | Homebrew Interaction — `brew cleanup` | 4.1 | Apply with `clear_cache: true`, confirm in output |
| REQ-022 | Apply: idempotent | Testing — idempotency validation | 4.4, 5.1 | Apply twice, second plan shows no changes |
| REQ-023 | Continue on error, report failed package | Design Decisions — continue on error | 4.1 | Declare bad package + good ones, verify partial success |
| REQ-024 | Non-zero exit on any failure | Exit Codes — code 1 on failure | 4.1, 4.3 | Trigger failure, confirm exit code 1 |
| REQ-025 | Single Go binary | Design Decisions — single binary | 1.1, 5.2 | `go build` produces one binary |
| REQ-026 | No runtime deps beyond Homebrew | Design Decisions — no runtime deps | 5.2 | Run on clean Mac with only Homebrew |
| REQ-027 | Shell out to Homebrew CLI | Design Decisions — shell-out | 2.1, 2.2 | Code review — no Ruby API, only CLI |
| REQ-028 | Use cobra for CLI | Project Layout — cobra in cmd/ | 1.1 | Code review — cobra in cmd/ |
| REQ-029 | Use `gopkg.in/yaml.v3` | Core Types — yaml.v3 tags | 1.1, 1.2 | Code review — yaml.v3 in go.mod |
| REQ-030 | Plan: list taps to remove (cleanup enabled) | Diff Calculator (cleanup flag) — TapsToRemove | 3.1 | Have undeclared tap installed, run plan, confirm output |
| REQ-031 | Apply: remove undeclared taps (cleanup enabled) | Homebrew Interaction — `brew untap` | 4.1, 4.3 | Tap a repo not in config, apply, verify `brew tap` no longer shows it |
| REQ-032 | `--config <path>` overrides default config path | Config Resolution — `--config` flag | 1.1 | Run `syncd plan --config /tmp/test.yaml`, confirm reads from specified path |

## Reverse Trace: Task → Requirements

| Task ID | Task Name | Requirements Covered |
|---------|-----------|---------------------|
| 1.1 | Project initialization | REQ-025, REQ-028, REQ-029, REQ-032 |
| 1.2 | Config parser | REQ-001, REQ-002, REQ-003, REQ-004 |
| 2.1 | CommandRunner interface | REQ-027 |
| 2.2 | Homebrew state query | REQ-027 |
| 3.1 | Diff calculator | REQ-005, REQ-006, REQ-007, REQ-008, REQ-009, REQ-019, REQ-030 |
| 3.2 | Plan CLI command | REQ-010, REQ-011 |
| 4.1 | Executor | REQ-014, REQ-015, REQ-016, REQ-017, REQ-018, REQ-020, REQ-021, REQ-023, REQ-024, REQ-031 |
| 4.2 | Confirmation prompt | REQ-012, REQ-013 |
| 4.3 | Apply CLI command | REQ-012, REQ-013, REQ-014, REQ-015, REQ-016, REQ-017, REQ-018, REQ-019, REQ-022, REQ-024, REQ-031 |
| 4.4 | Idempotency validation | REQ-022 |
| 5.1 | Integration test suite | REQ-010, REQ-022, REQ-023 |
| 5.2 | Build & release prep | REQ-025, REQ-026 |
| 5.3 | End-to-end validation | All REQs (workflow validation) |

## Coverage Summary

- **Requirements with tasks:** 32/32 (100%)
- **Requirements with design mapping:** 32/32 (100%)
- **Requirements with verification method:** 32/32 (100%)
- **Orphan tasks (no requirement):** None
- **Orphan requirements (no task):** None
