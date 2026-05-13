# Findings: syncd v0.1 — MVP

## Issues

### High

- [x] **F-001: No tap removal in plan/apply.** Design has `TapsToAdd` but no `TapsToRemove` field. Requirements define removal for brews (REQ-008/017) and casks (REQ-009/018) but not taps. If a user removes a tap from config, syncd won't detect or remove it. Either add tap removal requirements or document that taps are add-only in v0.1.
  - Status: closed
  - Resolution: Add tap removal. When `cleanup.remove_unlisted` is true, taps not in config are removed — same behavior as brews/casks.
  - Follow-up: ~~Add REQ-030 (plan: list taps to remove) and REQ-031 (apply: remove undeclared taps) to `requirements.md`. Add `TapsToRemove` to Plan struct in `design.md`. Add tap-removal subtasks to task 3.1 and 4.1 in `tasks.md`. Update `traceability-matrix.md`.~~ **Applied.**

### Medium

- [x] **F-002: REQ-025 says "single static binary" but Go default is dynamically linked on macOS.** `go build` on macOS produces a dynamically-linked binary (links libc). True static linking requires `CGO_ENABLED=0`. The Makefile should set this explicitly, or the requirement should say "single binary" without "static".
  - Status: closed
  - Resolution: Change REQ-025 wording to "single binary" (drop "static"). Add `CGO_ENABLED=0` to Makefile as a build optimization, not a hard requirement.
  - Follow-up: ~~Update REQ-025 in `requirements.md` to remove "static". Add `CGO_ENABLED=0` note to Makefile target in `tasks.md` task 1.1.~~ **Applied.**

- [x] **F-003: No config path override mechanism.** REQ-001 hardcodes `~/.config/syncd/config.yaml`. There's no `--config` flag or env var. Integration tests will need to use a test config without polluting the user's real config. Design should add a `--config` flag or `SYNCD_CONFIG` env var.
  - Status: closed
  - Resolution: Add `--config <path>` flag to v0.1 scope. Low-effort with cobra, unblocks integration testing.
  - Follow-up: ~~Add REQ-032 (`--config` flag overrides default path) to `requirements.md`. Add flag to CLI Layer in `design.md`. Add subtask to task 1.1 or 3.2 in `tasks.md`. Update `traceability-matrix.md`.~~ **Applied.**

- [x] **F-004: Open question 1 (install Homebrew if missing) is unresolved.** Problem-statement asks whether syncd should install Homebrew. Requirements say "Homebrew being installed" is a prerequisite (REQ-026) but don't specify behavior when it's absent. Design should define: exit with error + install instructions, or attempt auto-install.
  - Status: closed
  - Resolution: Exit with a clear error message and print the official Homebrew install URL. No auto-install.
  - Follow-up: ~~Close open question #1 in `problem-statement.md`. Add a note to `design.md` Config Resolution section specifying early Homebrew check + error behavior.~~ **Applied.**

- [x] **F-005: No version/help command specified.** Standard CLI tools provide `--version` and `--help`. Cobra provides `--help` by default, but `--version` needs explicit wiring. Not in requirements or tasks.
  - Status: closed
  - Resolution: Cobra provides `--help` for free. Add `--version` wiring to task 1.1 (trivial with cobra's `Version` field). No new requirement needed — this is standard CLI scaffolding.
  - Follow-up: ~~Add `--version` subtask to task 1.1 in `tasks.md`.~~ **Applied.**

### Low

- [x] **F-006: Cleanup section naming inconsistency.** Problem-statement says "Cleanup: remove unlisted, `brew autoremove`, `brew cleanup`" (scope). Requirements use `cleanup.remove_unlisted`, `cleanup.autoremove`, `cleanup.clear_cache`. Design uses `Autoremove` and `ClearCache` struct fields. The mapping is clear but `clear_cache` → `brew cleanup` naming could confuse contributors.
  - Status: closed
  - Resolution: Accept as-is. The YAML key `clear_cache` is user-facing and descriptive of intent; `brew cleanup` is the implementation detail. The mapping is documented in the Homebrew Interaction table in design.md. No change needed.

- [x] **F-007: No explicit ordering guarantee for operations.** Design says "commands are executed sequentially" but doesn't specify order (taps before brews? installs before removals?). Tap-add must precede brew-install for packages from new taps. This ordering should be documented in design.
  - Status: closed
  - Resolution: Execution order is: taps → brew installs → cask installs → brew removals → cask removals → autoremove → cleanup cache.
  - Follow-up: ~~Document this order in `design.md` Homebrew Interaction section or Executor description.~~ **Applied.**

- [x] **F-008: Task time estimates may be optimistic.** Phase 4 (Apply) estimates 3 hours for executor + prompt + CLI + idempotency. Given error handling complexity (REQ-023/024) and prompt testing, 4–5 hours is more realistic.
  - Status: closed
  - Resolution: Acknowledged. Estimates are guidance, not commitments. No doc change needed.

---

## Improvements

- [x] **I-001:** Add a `--config` flag to all commands for testability and multi-config workflows.
  - Status: closed (covered by F-003 resolution)

- [ ] **I-002:** Add `--dry-run` as an alias for `plan` to match common CLI conventions (Terraform, kubectl).
  - Status: deferred
  - Resolution: Nice-to-have but not needed for v0.1. `plan` is already the dry-run equivalent. Revisit post-MVP.

- [x] **I-003:** Define explicit execution order in design: taps → brew installs → cask installs → brew removals → cask removals → autoremove → cleanup.
  - Status: closed (covered by F-007 resolution)

- [x] **I-004:** Add `TapsToRemove` to the Plan struct and corresponding requirements if tap cleanup is desired in v0.1.
  - Status: closed (covered by F-001 resolution)

- [x] **I-005:** Add `CGO_ENABLED=0` to Makefile build target for reproducible cross-platform builds.
  - Status: closed (covered by F-002 resolution)

- [ ] **I-006:** Consider adding a `--verbose` flag for debugging brew command output (not required for MVP but low-cost with cobra).
  - Status: deferred
  - Resolution: Not needed for v0.1. Revisit post-MVP.

---

## Open Questions

1. **Should syncd remove taps that are no longer in config?**
   - Context: Brews and casks have explicit removal requirements. Taps do not. Removing a tap could break formulae from that tap.
   - Answer: Yes. When `cleanup.remove_unlisted` is true, remove taps not declared in config — same as brews/casks. Formulae from removed taps would also be flagged for removal if undeclared.

2. **What should syncd do when Homebrew is not installed?**
   - Context: REQ-026 says Homebrew is a prerequisite. Problem-statement open question #1 is unresolved. Options: (a) exit with error + install instructions, (b) auto-install via official script, (c) prompt user.
   - Answer: (a) Exit with a clear error message and print the official Homebrew install URL (`https://brew.sh`). No auto-install.

3. **Should `--config` be added to v0.1 scope?**
   - Context: Without it, integration tests must use the real config path or mock the filesystem. Adding it is low-effort with cobra but expands scope.
   - Answer: Yes. Add `--config <path>` as a persistent flag on the root command.

4. **Is the execution order (taps → installs → removals → cleanup) correct?**
   - Context: Installing a formula from a new tap requires the tap to be added first. Removing before installing could break dependencies temporarily.
   - Answer: Yes. Confirmed order: taps → brew installs → cask installs → brew removals → cask removals → autoremove → cleanup cache.
