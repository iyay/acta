---
parent: specs/2026-10-22-wiki-advisory-freshness-design
depth: minimal
id: PLN-0188
created: "2026-10-22 23:06:09"
hash: gu1s525
started: "2026-10-22 23:10:46"
finished: "2026-10-22 23:15:51"
---
# Narrow three wiki pages' paths

**Goal:** Show dispatch, ID, and parent-stamp notes on their implementation paths without unrelated state-command hints.
**Spec:** .acta/specs/2026-10-22-wiki-advisory-freshness-design.md
**Tests:** fast read-only `go run ./cmd/acta wiki match <file>` assertions described below; existing match/hint checks `scripts/test ./internal/wiki ./internal/cli ./internal/hook -run 'TestMatch|TestWikiMatch|TestWikiHints'`; full `scripts/test --full` at land.

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Executor is inline; use an isolated worktree, observe the failing CLI assertions before editing, deep review at close, and tidy landing; never push.
- Before any page edit, capture `go run ./cmd/acta wiki check` output/status, read all three full pages and their implementation code, and record claim-by-claim review evidence in this plan; a frontmatter edit counts as a check and must follow that review.
- Change only the three listed frontmatter path lists; leave their bodies unchanged if review confirms them, and report a contradictory claim before expanding scope.
- Keep every other page's paths, including the five broad TUI pages and four `plugin/omp/` pages; do not edit or confirm unrelated stale notes to clear baseline warnings. Release warning noise stays in SCR-0077.
- No Go source, new test file, new page, freshness algorithm, or workflow change belongs here; read-only CLI assertions test the actual edited pages rather than reproducing path lists in unit-test fixtures.

## File map and waves

Task 1 owns `.acta/wiki/dispatch-plan-path.md`, `.acta/wiki/fix-duplicate-ids.md`, and `.acta/wiki/tick-stamps-spec.md`. Wave 1 contains Task 1. This plan shares no edited file with the advisory-freshness plan and has no dependency on it.

### Task 1: Narrow paths while retaining every relevant entry point

**Files:** Modify only `.acta/wiki/dispatch-plan-path.md`, `.acta/wiki/fix-duplicate-ids.md`, and `.acta/wiki/tick-stamps-spec.md`; read `internal/cli/{dispatch,id,tick,specdates}.go` and `internal/write/{ids,mark,specdates}.go` to verify their claims.
**Seam:** Real repository pages loaded through `go run ./cmd/acta wiki match <file>`; assertions inspect page IDs in the output and do not require unrelated pages to disappear.
**verify:** `dispatch-plan-path` matches `internal/cli/dispatch.go`; `fix-duplicate-ids` matches both `internal/cli/id.go` and `internal/write/ids.go`; `tick-stamps-spec` matches `internal/cli/tick.go`, `internal/cli/specdates.go`, `internal/write/mark.go`, and `internal/write/specdates.go`; none of these three IDs matches `internal/cli/state.go`. Compare the final diff to prove every other page's paths stayed unchanged, and list all eight CLI invocations plus review evidence.
- [x] Red: Run `go run ./cmd/acta wiki match internal/cli/state.go` and assert that the three target IDs are absent, observing failure because all three pages currently cover `internal/cli/`, then capture the seven positive matches as the coverage baseline.
- [x] Change: After reviewing each page's claims, replace dispatch paths with `[internal/cli/dispatch.go]`, ID paths with `[internal/cli/id.go, internal/write/ids.go]`, and stamp paths with `[internal/cli/tick.go, internal/cli/specdates.go, internal/write/mark.go, internal/write/specdates.go]`, then rerun all eight real-CLI assertions and the header's focused match/hint checks until they pass.
- [x] Commit: `docs(wiki): narrow dispatch id and stamp paths` with this task's trailer, then close through the existing deep review and landing gates without blanket freshness resets.

## Verification evidence

Before edits, no-range check exited 1 with only plugin-loads-from-repo and plugin-ships-in-binary stale from ccaffcf; neither was edited or confirmed.

Read all three full pages before changing paths. Dispatch checkRecord validates a repo-relative .md path inside the planning root, while tick resolves task IDs through the board. AssignIDs uses scanIDs then increments the next ID; FixDuplicates groups equal canonical IDs and assigns later holders the next free number. TaskDates writes parent stamps except for multi-plan specs in linked worktrees; cmdSpecDates calls SpecDates to derive the earliest start and completion once all plans finish. These claims remain valid; bodies are unchanged.

Observed RED: all three page IDs matched internal/cli/state.go. Baseline positive matches existed for six paths; internal/write/ids.go lacked its ID note. After edits, all seven intended positive matches pass: dispatch.go; id.go and write/ids.go; tick.go, cli/specdates.go, write/mark.go and write/specdates.go. The eighth invocation, state.go, matches none of these IDs. Only the three listed paths fields changed.

## State

### Next

All path assertions and focused match/hint checks pass; proceed to close review and landing gates.

## Review notes

Deep, full tier because wiki guides agent behavior; Spec and Standards reviewers both CLEAN, no BLOCKERs or NOTEs. Both independently confirmed the eight real-CLI assertions; no fact-check pass needed.
