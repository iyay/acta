---
parent: specs/2026-10-05-write-tests-clock-race
depth: minimal
id: PLN-0093
created: "2026-10-05 20:11:40"
hash: cr78yrn
started: "2026-10-05 20:17:16"
finished: "2026-10-05 20:24:02"
---
# write tests clock race Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** No `internal/write` test swaps the global clock while another test runs.

**Spec:** `.acta/specs/2026-10-05-write-tests-clock-race.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: no parallel clock swap

**Files:** Modify `internal/write/subject_test.go`.
**verify:** No test in `internal/write` both calls `t.Parallel()` and swaps `Now` (fixNow or fixNowAt), and `scripts/test ./internal/write -race -count=3` reports no data race and no failure. List each test that swaps `Now` and whether it is parallel.
- [x] Failing test: `scripts/test ./internal/write -race -count=3` reports `DATA RACE` at fixNow from subject_test.go.
- [x] Code: remove `t.Parallel()` from the two tests in subject_test.go, with a one-line comment saying why.
- [x] Commit: `test: clock-swapping write tests do not run in parallel`.

### Task 2: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [x] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [x] Code: add 1 to the patch in all three files.
- [x] Commit: `plugin: bump patch version`.

## Review notes

- The same "No Parallel here" comment sits in two tests; a third copy would mean the reason belongs on fixNow.
- The TUI flake TestClickOnTheBarTakesTheHighlightOff stays open under SCR-0046.
