# Findings: syncd v0.1 — MVP

## Issues

### High

- [ ] **F-001: No tap removal in plan/apply.** Design has `TapsToAdd` but no `TapsToRemove` field. Requirements define removal for brews (REQ-008/017) and casks (REQ-009/018) but not taps. If a user removes a tap from config, syncd won't detect or remove it. Either add tap removal requirements or document that taps are add-only in v0.1.

### Medium

- [ ] **F-002: REQ-025 says "single static binary" but Go default is dynamically linked on macOS.** `go build` on macOS produces a dynamically-linked binary (links libc). True static linking requires `CGO_ENABLED=0`. The Makefile should set this explicitly, or the requirement should say "single binary" without "static".

- [ ] **F-003: No config path override mechanism.** REQ-001 hardcodes `~/.config/syncd/config.yaml`. There's no `--config` flag or env var. Integration tests will need to use a test config without polluting the user's real config. Design should add a `--config` flag or `SYNCD_CONFIG` env var.

- [ ] **F-004: Open question 1 (install Homebrew if missing) is unresolved.** Problem-statement asks whether syncd should install Homebrew. Requirements say "Homebrew being installed" is a prerequisite (REQ-026) but don't specify behavior when it's absent. Design should define: exit with error + install instructions, or attempt auto-install.

- [ ] **F-005: No version/help command specified.** Standard CLI tools provide `--version` and `--help`. Cobra provides `--help` by default, but `--version` needs explicit wiring. Not in requirements or tasks.

### Low

- [ ] **F-006: Cleanup section naming inconsistency.** Problem-statement says "Cleanup: remove unlisted, `brew autoremove`, `brew cleanup`" (scope). Requirements use `cleanup.remove_unlisted`, `cleanup.autoremove`, `cleanup.clear_cache`. Design uses `Autoremove` and `ClearCache` struct fields. The mapping is clear but `clear_cache` → `brew cleanup` naming could confuse contributors.

- [ ] **F-007: No explicit ordering guarantee for operations.** Design says "commands are executed sequentially" but doesn't specify order (taps before brews? installs before removals?). Tap-add must precede brew-install for packages from new taps. This ordering should be documented in design.

- [ ] **F-008: Task time estimates may be optimistic.** Phase 4 (Apply) estimates 3 hours for executor + prompt + CLI + idempotency. Given error handling complexity (REQ-023/024) and prompt testing, 4–5 hours is more realistic.

---

## Improvements

- [ ] **I-001:** Add a `--config` flag to all commands for testability and multi-config workflows.
- [ ] **I-002:** Add `--dry-run` as an alias for `plan` to match common CLI conventions (Terraform, kubectl).
- [ ] **I-003:** Define explicit execution order in design: taps → brew installs → cask installs → brew removals → cask removals → autoremove → cleanup.
- [ ] **I-004:** Add `TapsToRemove` to the Plan struct and corresponding requirements if tap cleanup is desired in v0.1.
- [ ] **I-005:** Add `CGO_ENABLED=0` to Makefile build target for reproducible cross-platform builds.
- [ ] **I-006:** Consider adding a `--verbose` flag for debugging brew command output (not required for MVP but low-cost with cobra).

---

## Open Questions

1. **Should syncd remove taps that are no longer in config?**
   - Context: Brews and casks have explicit removal requirements. Taps do not. Removing a tap could break formulae from that tap.
   - Answer:

2. **What should syncd do when Homebrew is not installed?**
   - Context: REQ-026 says Homebrew is a prerequisite. Problem-statement open question #1 is unresolved. Options: (a) exit with error + install instructions, (b) auto-install via official script, (c) prompt user.
   - Answer:

3. **Should `--config` be added to v0.1 scope?**
   - Context: Without it, integration tests must use the real config path or mock the filesystem. Adding it is low-effort with cobra but expands scope.
   - Answer:

4. **Is the execution order (taps → installs → removals → cleanup) correct?**
   - Context: Installing a formula from a new tap requires the tap to be added first. Removing before installing could break dependencies temporarily.
   - Answer:
