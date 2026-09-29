---
parent: scratch/2026-09-28-newest-first-sort
id: SPC-0014
hash: q9ccjo0
---
# Sort toggle in every pane

Status: design approved by the user in chat on 2026-09-29. Bounded. Built only after PLAN-19 (top tabs) lands.

## Why

The user wants to choose the order of each list. One fixed order does not fit every pane, so each pane gets two orders and a key to switch between them.

## Non-goals

- No sort by last update or commit time. All panes use one rule so they agree.
- No more than two orders.
- The choice is not saved across restarts.

## Behaviour

- Key `o` flips the sort of the focused pane between `oldest` and `newest`. `o` is free on main and on the `top-tabs` branch. `s` is taken.
- The pane border title shows the current sort as a word, `oldest` or `newest`, drawn as one more title segment (see `segment` in `internal/tui/view.go`).
- Each pane keeps its own sort. Flipping one pane leaves the others as they are.
- Every pane starts at `oldest` each time the TUI starts. The state lives in memory only.
- The help popup lists `o`.

## Order rule

- Sort by the date in the file name, up for `oldest`, down for `newest`.
- Same date: fall back to the item ID, in the same direction, so the order never jumps between loads.
- Tasks stay under their plan, in file order, in both directions.
- An item with no date in its name goes last in both directions.

## Changed behaviour

Pane [2] (finished) sorts by last commit time today (`SortTime` in `internal/board/closed.go`). It drops that and follows the rule above, starting at `oldest`. If nothing else reads `CommitAt` after this, remove it and `fillCommitTimes`.

## Depends on

PLAN-19 changes the pane layout and titles. Before writing the plan, merge main into this branch (no rebase) and re-check the key and the title code against the landed top-tabs code.

## Tests

Write each one red first:

1. Pressing `o` in a pane reverses its order, and the title word changes from `oldest` to `newest` and back.
2. Pressing `o` in one pane leaves every other pane's order and title unchanged.
3. Items with the same date keep a stable order by ID, in both directions.
4. Tasks stay under their plan in file order after a flip.
5. Pane [2] follows the file-name date, not commit time.
