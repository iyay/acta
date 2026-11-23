---
parent: bugs/2026-11-23-plan-closes-rejects-bug-and-debt-ids
id: SPC-0219
created: "2026-11-24 05:56:53"
hash: gp4itib
---
# Closes accepts a bug and a whole debt file

## Why

The land bookkeeping of SPC-0217 reads `BUG-nnnn` and whole `DBT-nnnn` ids from a plan's `closes:` list. The board still treats those kinds as targets that cannot be closed: `linkCloses` in `internal/board/closes.go` accepts only scratch items, specs and debt lines, so every such plan shows `closes BUG-0001: a bug cannot be closed` in `acta show` and the TUI.

## Design

`linkCloses` also accepts a bug and a whole debt file as a target. Both get the usual `Closes` and `ClosedBy` links and no problem line.

Nothing else changes:
- A bug's status still comes from its own `fixed_in` or written status. `closedByStatus` only runs for specs, so a closer never sets a bug's status.
- A plan or a task in `closes:` is still a problem.
- `closesOf` in `internal/write/landnotes.go` stays as it is.

Files: `internal/board/closes.go`, `internal/board/closes_test.go`.

## Verification

Red first, in existing tests:
- `TestClosesBadEntriesAreProblems`: the bug row moves out, since a bug is no longer a bad entry.
- `TestClosesLinksEveryAllowedKind`: rows for a bug and a whole debt file, each linked both ways with no problem.

Then `scripts/test ./internal/board`. The full suite runs at land. No skill changes, so the eval gate does not run.
