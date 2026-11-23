---
parent: bugs/2026-11-23-plan-closes-rejects-bug-and-debt-ids
depth: minimal
id: PLN-0253
created: "2026-11-24 05:57:05"
hash: jvp7jst
started: "2026-11-24 06:01:44"
finished: "2026-11-24 06:03:50"
---
# Closes accepts a bug and a whole debt file

**Goal:** A plan whose `closes:` names a bug or a whole debt file links to it on the board with no false problem.

**Spec:** `.acta/specs/2026-11-24-closes-accepts-bug-and-debt.md` (SPC-0219)
**Tests:** fast `scripts/test ./internal/board`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched and `scripts/eval --case <name>` for one case; whole when the diff touches `plugin/hooks/` or `plugin/references/`. This plan touches neither, so the gate does not run.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- A red test is a new row in an existing table test, else an assert in the nearest test of the same unit; a new test func only for a behaviour no test checks yet.
- A bug's status never comes from the item that closes it.

## Waves

- Wave 1: Task 1

### Task 1: Accept a bug and a whole debt file in closes

**Files:** `internal/board/closes.go`, `internal/board/closes_test.go`
**Seam:** `board.Load` through `linkCloses`
**verify:** A `closes:` entry that names a bug or a whole debt file links both ways (`Closes` on the closer, `ClosedBy` on the target) and adds no problem; a plan or task entry is still a problem; the bug's status is the same as with no closer. List each kind tried: scratch, spec, debt line, debt file, bug, plan, task.
- [x] Red: the bug row in `TestClosesBadEntriesAreProblems` moves to `TestClosesLinksEveryAllowedKind` with a whole debt file row, and both fail because `linkCloses` adds a problem.
- [x] Change: add the bug and debt kinds to the allowed target kinds in `linkCloses`, and update its comment.
- [x] Commit: fix(board): closes accepts a bug and a whole debt file

## Review notes

- Review depth auto, picked quick (one reviewer, both axes): no risk signal, board display logic only. Round 1 bd8fb47..8a30a9d CLEAN. BLOCKER 0, NOTE 3, pre-existing 0. Notes: old closes spec 2026-09-29 still says a bug cannot be closed; an old wontfix debt line names the removed wording; no assert that a closed debt file keeps its status.
