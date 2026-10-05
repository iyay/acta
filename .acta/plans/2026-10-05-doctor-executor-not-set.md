---
parent: specs/2026-10-05-doctor-executor-not-set
depth: minimal
id: PLN-0085
created: "2026-10-05 12:26:29"
hash: cqy1o86
started: "2026-10-05 12:50:42"
---
# Doctor executor not set Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta doctor` reports an unset build executor as `ok`, since build asks each time.

**Spec:** `.acta/specs/2026-10-05-doctor-executor-not-set.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Message for the unset case, verbatim: `voice is set; build executor not set, build asks each time`.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: setup check

**Files:** Modify `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`.
**verify:** For every voice and executor state (no voice file, voice with no executor, voice with executor) the setup check gives the level, message and fix line the spec names, and no state with a voice file ever gives `warn`. List the three states checked.
- [x] Failing test: a table test over the three states; the unset executor case fails because it gives `warn` with the `/acta:setup` fix.
- [x] Code: in `checkSetup`, the empty executor case gives `OK` with the spec message and no fix; the other two cases stay as they are.
- [x] Commit: `doctor: an unset build executor is fine, build asks each time`.

### Task 2: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [ ] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [ ] Code: add 1 to the patch in all three files.
- [ ] Commit: `plugin: bump patch version`.
