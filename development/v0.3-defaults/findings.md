# Findings: syncd v0.3 — macOS Defaults

## Issues

### High

- [ ] **F-001: Config struct already has `DefaultEntry` and `Defaults` field in code, but no `ValidateDefaults` call in `Load()`.**
  - Evidence: `internal/config/config.go` already defines `DefaultEntry` struct and `Defaults []DefaultEntry` field on `Config`. However, `Load()` only calls `dec.Decode(&cfg)` with `KnownFields(true)` — there is no call to `ValidateDefaults()`. Task 1.1 marks adding the struct as TODO, but it's already done. Task 1.2 correctly identifies that `ValidateDefaults` must be called from `Load()`.
  - Impact: Phase 1 task 1.1 is partially complete (struct exists), but task 1.2 (validation logic) is not implemented. The task list should reflect that 1.1 is done.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/tasks.md` — task 1.1 should be marked complete or noted as pre-implemented.

- [ ] **F-002: Design says `defaults` package imports `brew` only for `CommandRunner` interface, but `RunMutate` semantics differ from `Run` for defaults operations.**
  - Evidence: Design says defaults uses `CommandRunner` from `internal/brew`. The `RunMutate` method in `runner.go` streams output when `Verbose` is true. For defaults reads during `plan`, the code should use `Run` (capture output for parsing). For defaults writes during `apply`, it should use `RunMutate` (stream when verbose). However, the design doesn't specify which method to use for reads vs writes — it just says "accepts a `CommandRunner` parameter."
  - Impact: If `ReadValue`/`ReadType` accidentally use `RunMutate`, verbose mode would stream read output to stdout instead of capturing it for comparison. The design should clarify: reads use `Run`, writes use `RunMutate`.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/design.md` — Defaults State Reader and Executor sections should specify `Run` vs `RunMutate`.

### Medium

- [ ] **F-003: REQ-107 verification is "Manual test" — weakest validation of all requirements.**
  - Evidence: REQ-107 (`--verbose` prints defaults read/write output) has verification: "Run with `--verbose`, confirm command output visible." All other requirements have unit or integration test verification. This is the only requirement relying on manual testing.
  - Impact: Regression risk. If verbose behavior breaks, no automated test catches it.
  - Improvement: Add an integration test that runs `syncd plan --verbose` with a fake defaults script and asserts that `defaults read` command output appears in stdout.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/tasks.md` — task 6.1 should include an integration test.

- [ ] **F-004: Design specifies `defaults write` bool mapping as `TRUE`/`FALSE` but doesn't document what `defaults read` returns for bools set this way.**
  - Evidence: Design says "Bool value mapping for write: config `true` → write `TRUE`, config `false` → write `FALSE`." The comparison logic says `defaults read` returns `1`/`0`. However, macOS `defaults read` actually returns `1` or `0` for boolean values regardless of whether they were written as `TRUE`/`FALSE`, `YES`/`NO`, or `1`/`0`. This is correct but the design doesn't explicitly confirm the round-trip: write `TRUE` → read back `1` → compare against config `true` → match.
  - Impact: Low risk since the behavior is correct, but the round-trip should be documented to prevent future confusion.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/design.md` — Type Comparison Logic section.

- [ ] **F-005: `KnownApps` map in design is minimal — only 4 domains listed.**
  - Evidence: Design lists only `com.apple.dock`, `com.apple.finder`, `com.apple.systemuiserver`, `com.apple.menuextra.clock`, and `NSGlobalDomain`. Common defaults like `com.apple.Safari`, `com.apple.Terminal`, `com.apple.screencapture` are absent.
  - Impact: `syncd init` will omit `kill` for many common domains where restart is needed. Users will need to manually add `kill` fields.
  - Improvement: Expand the map or document that it's intentionally minimal and users should specify `kill` explicitly for unlisted domains.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/design.md` — Well-Known Domain Map section.

- [ ] **F-006: Task 3.2 requires a fake `defaults` script but doesn't specify how it integrates with the existing fake `brew` script infrastructure.**
  - Evidence: Task 3.2 says "Add fake `defaults` script to integration test infrastructure" with `$FAKE_DEFAULTS_STATE` directory. The existing integration tests use `$FAKE_BREW_STATE` and prepend a fake `brew` to PATH. The task doesn't specify whether the fake `defaults` script goes in the same temp directory, how PATH is managed for both fakes simultaneously, or whether `TestMain` needs modification.
  - Impact: Implementation ambiguity. Developer may need to refactor `TestMain` to handle both fake scripts.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/tasks.md` — task 3.2 should specify PATH setup for both fakes.

- [ ] **F-007: Design says "No import cycle: `defaults` imports `brew` only for the `CommandRunner` interface" but this creates a coupling concern.**
  - Evidence: The `defaults` package imports `brew` for `CommandRunner` and `MockRunner`. If `brew` ever needs to reference defaults (unlikely but possible), this creates a one-way dependency. A cleaner approach would be to extract `CommandRunner` into a shared `internal/runner` package.
  - Impact: Low immediate risk. The current design works. But it's a design smell — a domain package (`defaults`) importing another domain package (`brew`) for an infrastructure interface.
  - Improvement: Consider extracting `CommandRunner` to `internal/runner` or `internal/exec` package. This is a nice-to-have, not a blocker.
  - Status: open
  - Resolution:
  - Follow-up: Design decision — accept or refactor.

- [ ] **F-008: REQ-083 says "restart apps listed in the `kill` field after writing all defaults for that domain" but design says "kill once after all writes complete."**
  - Evidence: REQ-083 wording: "restart apps listed in the `kill` field after writing all defaults for that domain." Design Decision 5: "Collect all apps to kill from drifted entries, deduplicate, then kill once after all writes complete." The requirement implies per-domain kill timing; the design implements global kill-after-all-writes.
  - Impact: The design approach is better (fewer restarts), but the requirement wording is slightly inconsistent with the implementation. The behavior is functionally equivalent since kill is deduplicated anyway.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/requirements.md` — REQ-083 wording could be clarified to say "after writing all drifted defaults."

- [ ] **F-009: No requirement or task for handling `defaults write` to a domain that doesn't exist yet.**
  - Evidence: REQ-076 handles `defaults read` on non-existent domain/key (treated as drift). But there's no explicit requirement or test for what happens when `defaults write` targets a domain that has never been used. macOS `defaults write` creates the plist file automatically, so this should work, but it's not tested.
  - Impact: Low — macOS handles this gracefully. But an integration test confirming write-to-new-domain succeeds would strengthen coverage.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/tasks.md` — task 4.3 could add a test case.

### Low

- [ ] **F-010: Problem statement Non-Goals lists "Array or dictionary value types" but design doesn't validate against them.**
  - Evidence: Problem statement says array/dict types are non-goals. Design validation checks `type` is one of `string`, `int`, `float`, `bool`. But YAML allows `value: [1, 2, 3]` which would parse as `[]interface{}`. The `validateTypeMatch` function needs to reject array/dict values even when type is declared as something else.
  - Impact: If a user writes `type: string` with `value: [a, b]`, the YAML parser will set `Value` to a slice. Without explicit validation, this could cause a runtime panic or confusing error during `defaults write`.
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/design.md` — Validation section should note that `validateTypeMatch` rejects non-scalar values.

- [ ] **F-011: Task estimates total 8 hours but Phase 1 struct work is already done.**
  - Evidence: `internal/config/config.go` already has `DefaultEntry` and `Defaults` field. Task 1.1 estimates time for work that's complete. Actual remaining work is ~6.5-7 hours.
  - Impact: Informational only — estimates are guidance.
  - Status: open
  - Resolution:
  - Follow-up: None needed.

- [ ] **F-012: Design uses `interface{}` for `DefaultEntry.Value` but Go 1.18+ convention is `any`.**
  - Evidence: Design code samples use `interface{}`. The existing codebase (`config.go`) also uses `interface{}`. This is consistent but slightly dated.
  - Impact: Cosmetic. No functional issue.
  - Status: open
  - Resolution:
  - Follow-up: None needed — consistency with existing code is correct.

- [ ] **F-013: REQ-091 specifies comma-separated `domain:key` pairs but doesn't address edge cases.**
  - Evidence: `--defaults "com.apple.dock:tilesize,NSGlobalDomain:KeyRepeat"` — what if a domain contains a colon? What if the value contains a comma? macOS domains don't typically contain colons, but the parsing logic should be documented.
  - Impact: Very low — macOS preference domains use reverse-DNS notation without colons. But the parser should split on first `:` only (not all colons).
  - Status: open
  - Resolution:
  - Follow-up: `development/v0.3-defaults/design.md` — Init section should note "split on first `:` per pair."

---

## Open Clarification Questions

1. **Should `defaults` reads during `plan` use `Run` (capture) or `RunMutate` (stream when verbose)?**
   - Context: The verbose flag should show what commands are being run, but `ReadValue` needs to capture output for comparison. Options: (a) always use `Run` for reads, print command name separately when verbose; (b) use `Run` but add verbose logging around it.
   - Answer: Always use `Run` for reads (must capture output for parsing). Print the command being run to stdout when verbose is enabled, separately from the captured output.

2. **Should the `CommandRunner` interface be extracted to a shared package to avoid `defaults` importing `brew`?**
   - Context: Current design has `defaults` import `brew` for the interface. This works but creates cross-domain coupling. Extracting to `internal/runner` is cleaner but adds a package.
   - Answer: Yes, extract to `internal/runner`. Cleaner separation, avoids cross-domain coupling.

3. **Should `KnownApps` be expanded beyond the 4 listed domains, or is it intentionally minimal?**
   - Context: Many common preference domains (Safari, Terminal, screencapture, trackpad) are absent. Users must manually specify `kill` for these.
   - Answer: Expand to ~10-15 common domains (Safari, Terminal, screencapture, trackpad, etc.).

4. **Should REQ-107 have an automated integration test instead of manual verification?**
   - Context: Every other requirement has automated verification. Manual testing is the weakest validation path.
   - Answer: Yes, add an integration test.

5. **Should task 1.1 be marked as complete given the struct already exists in code?**
   - Context: `DefaultEntry` and `Defaults` field are already in `internal/config/config.go`. The task list shows them as TODO.
   - Answer: Yes, mark as complete. Proceed from task 1.2.

---

## Validation Notes

- All REQ-067 through REQ-107 IDs are present in requirements, design, and tasks.
- All 41 requirements have at least one task with a verification method.
- All tasks trace back to at least one requirement.
- No orphan tasks or orphan requirements found.
- The traceability is complete and consistent across all four planning artifacts.
- Problem statement scope, non-goals, and constraints are all reflected in requirements.
- Design decisions are well-reasoned and follow the established v0.1/v0.2 patterns.
- The strongest validation paths are unit tests for comparison/validation logic and integration tests with fake scripts for CLI behavior.
