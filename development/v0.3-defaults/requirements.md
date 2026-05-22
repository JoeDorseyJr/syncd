# Requirements: syncd v0.3 — macOS Defaults

## User Stories

**US-011:** As a Mac user, I want to declare my macOS preferences in config so that I can reproduce my system settings on any machine.

**US-012:** As a Mac user, I want to see which preferences have drifted from my desired state so that I know what needs fixing.

**US-013:** As a Mac user, I want syncd to write drifted preferences and restart affected apps so that changes take effect without manual intervention.

**US-014:** As a Mac user, I want to snapshot my current macOS preferences into config so that I don't have to look up domain/key names by hand.

---

## Functional Requirements

### Config Schema

**REQ-067:** syncd shall support a `defaults` section in config as a list of entries with `domain`, `key`, `type`, and `value` fields.
- Verification: Unit test — parse config with `defaults` section, confirm all entries populated.

**REQ-068:** Each defaults entry shall require a `domain` field (string, e.g., `com.apple.dock`).
- Verification: Unit test — config with missing `domain` returns parse error.

**REQ-069:** Each defaults entry shall require a `key` field (string, e.g., `tilesize`).
- Verification: Unit test — config with missing `key` returns parse error.

**REQ-070:** Each defaults entry shall require a `type` field with allowed values: `string`, `int`, `float`, `bool`.
- Verification: Unit test — config with invalid type value returns error.

**REQ-071:** Each defaults entry shall require a `value` field matching the declared type.
- Verification: Unit test — config with type/value mismatch returns error.

**REQ-072:** Each defaults entry shall support an optional `kill` field (list of app names to restart after write).
- Verification: Unit test — parse config with `kill` field, confirm populated. Parse without `kill`, confirm nil/empty.

**REQ-073:** syncd shall reject unknown fields within defaults entries via `KnownFields(true)`.
- Verification: Unit test — config with extra field in defaults entry returns parse error.

### Plan Command — Defaults Drift

**REQ-074:** `syncd plan` shall read the current value of each declared default via `defaults read <domain> <key>`.
- Verification: Integration test — declare a default, run `syncd plan`, confirm `defaults read` is invoked.

**REQ-075:** `syncd plan` shall show defaults whose current value differs from the declared value.
- Verification: Set a default to a different value, run `syncd plan`, confirm drift shown with current → desired.

**REQ-076:** `syncd plan` shall show defaults where the domain or key does not exist as drift (desired value shown, current shown as "unset").
- Verification: Declare a default for a non-existent domain/key, run `syncd plan`, confirm it appears as drift.

**REQ-077:** `syncd plan` shall not show defaults whose current value matches the declared value.
- Verification: Set a default to match config, run `syncd plan`, confirm it does NOT appear in output.

**REQ-078:** `syncd plan` shall display defaults drift with the format: `domain key: current → desired`.
- Verification: Run `syncd plan` with drifted defaults, confirm output format matches.

**REQ-079:** Defaults drift shall count as pending changes for exit code 2.
- Verification: Run `syncd plan` with only defaults drift, confirm exit code 2.

**REQ-080:** `syncd plan` shall not modify any defaults (read-only).
- Verification: Run `syncd plan`, confirm `defaults read` called but never `defaults write`.

### Apply Command — Defaults Write

**REQ-081:** `syncd apply` shall write each drifted default via `defaults write <domain> <key> -<type> <value>`.
- Verification: Integration test — declare a drifted default, run `syncd apply --yes`, confirm value written.

**REQ-082:** `syncd apply` shall write defaults with the correct type flag: `-string`, `-int`, `-float`, `-bool`.
- Verification: Unit test — confirm correct type flag passed for each type.

**REQ-083:** `syncd apply` shall restart apps listed in the `kill` field after writing all defaults for that domain.
- Verification: Integration test — declare defaults with `kill: [Dock]`, apply, confirm `killall Dock` executed.

**REQ-084:** `syncd apply` shall deduplicate app restarts (kill each app at most once per apply run).
- Verification: Declare multiple defaults for same domain with same `kill` app, apply, confirm `killall` called once.

**REQ-085:** `syncd apply` shall not restart apps when no defaults in that domain have drifted.
- Verification: All defaults match, apply, confirm no `killall` executed.

**REQ-086:** `syncd apply` shall continue writing remaining defaults if one write fails.
- Verification: Integration test — one bad domain + good ones, confirm good ones still written.

**REQ-087:** `syncd apply` shall report which defaults failed to write.
- Verification: Trigger a write failure, confirm error reported with domain/key.

**REQ-088:** `syncd apply` shall exit 1 if any defaults write fails.
- Verification: Trigger a failure, confirm exit code 1.

**REQ-089:** `syncd apply` shall include defaults in the confirmation prompt (show drift before asking).
- Verification: Run `syncd apply` without `--yes`, confirm defaults drift shown before prompt.

**REQ-090:** `syncd apply` shall be idempotent — writing a default that already matches is a no-op.
- Verification: Apply twice, confirm second run shows no defaults drift.

### Init Command — Defaults Snapshot

**REQ-091:** `syncd init` shall accept a `--defaults` flag with a comma-separated list of `domain:key` pairs to snapshot.
- Verification: Run `syncd init --defaults "com.apple.dock:tilesize"`, confirm output includes that entry.

**REQ-092:** `syncd init` shall read the current value and type of each specified default via `defaults read-type` and `defaults read`.
- Verification: Integration test — snapshot a known default, confirm type and value correct.

**REQ-093:** `syncd init` shall include snapshotted defaults in the generated YAML config under the `defaults` section.
- Verification: Run `syncd init --defaults ...`, confirm output has `defaults:` section with entries.

**REQ-094:** `syncd init` shall infer the `kill` field for well-known domains (e.g., `com.apple.dock` → `Dock`).
- Verification: Snapshot a dock default, confirm `kill: [Dock]` in output.

**REQ-095:** `syncd init` shall omit the `kill` field for domains without a known app mapping.
- Verification: Snapshot an unknown domain, confirm no `kill` field in output.

**REQ-096:** `syncd init` shall skip defaults that cannot be read (domain/key doesn't exist) and print a warning.
- Verification: Specify a non-existent domain:key, confirm warning printed and entry absent from output.

**REQ-097:** `syncd init --defaults` shall work alongside existing brew snapshot (combined output).
- Verification: Run `syncd init --defaults ...` with brews installed, confirm both `brews` and `defaults` in output.

### Type Handling

**REQ-098:** syncd shall read `defaults read-type <domain> <key>` to determine the stored type of a value.
- Verification: Unit test — mock `defaults read-type` output, confirm type parsed correctly.

**REQ-099:** syncd shall compare values using type-aware comparison (e.g., `48` int equals `48` int, not `48.0` float).
- Verification: Unit test — int 48 in config vs "48" from `defaults read` → no drift.

**REQ-100:** syncd shall handle bool values as `1`/`0` from `defaults read` mapped to `true`/`false` in config.
- Verification: Unit test — `defaults read` returns `1`, config says `true` → no drift.

---

## Technical Requirements

### State Query

**REQ-101:** syncd shall query `defaults read <domain> <key>` to get the current value of a preference.
- Verification: Unit test — mock `defaults read` output, confirm value parsed.

**REQ-102:** syncd shall query `defaults read-type <domain> <key>` to get the stored type.
- Verification: Unit test — mock `defaults read-type` output, confirm type string parsed.

**REQ-103:** syncd shall handle `defaults read` returning exit code 1 (domain/key not found) as "unset".
- Verification: Unit test — mock command failure, confirm treated as unset (drift).

### Execution

**REQ-104:** syncd shall execute `defaults write <domain> <key> -<type> <value>` to set a preference.
- Verification: Unit test — confirm correct command args constructed for each type.

**REQ-105:** syncd shall execute `killall <app>` to restart affected applications.
- Verification: Unit test — confirm `killall` called with correct app name.

**REQ-106:** syncd shall not fail the entire apply if `killall` fails (app may not be running).
- Verification: Unit test — `killall` returns error, confirm apply continues without failure.

### Verbose Output

**REQ-107:** `--verbose` shall print the full `defaults read` and `defaults write` command output during plan/apply.
- Verification: Run with `--verbose`, confirm command output visible.

---

## Traceability Matrix

| Requirement | User Story | Command |
|-------------|-----------|---------|
| REQ-067–073 | US-011 | config |
| REQ-074–080 | US-012 | plan |
| REQ-081–090 | US-013 | apply |
| REQ-091–097 | US-014 | init |
| REQ-098–100 | US-011, US-012 | type handling |
| REQ-101–103 | US-012 | technical (state query) |
| REQ-104–106 | US-013 | technical (execution) |
| REQ-107 | US-012, US-013 | technical (verbose) |
