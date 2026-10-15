---
title: TUI enter no longer opens the detail
status: draft
id: SPC-0140
created: "2026-10-15 16:51:40"
hash: afwu9x7
started: "2026-10-15 16:53:23"
finished: "2026-10-15 16:55:28"
---

# TUI enter no longer opens the detail

## Why

In the list panes, enter does two jobs: it folds a plan row, and on any other
row it moves the focus to the detail pane. The user wants enter to fold only.
Tab already reaches the detail, so no path is lost.

## Design

- `listKey` in `internal/tui/model.go`: enter still calls `toggleRow`. When the
  row is a legacy group row, enter still folds the group. On every other row,
  enter does nothing.
- `enter()` keeps only the group fold. The focus move and the "branch is not
  checked out" status go away. Rename it `toggleGroup` so the name says what it
  does.
- Hints: drop `Detail: enter` in `internal/tui/hints.go`. In the help screen
  (`internal/tui/view.go`) the enter line reads "open or shut a plan row or
  group". The comments in `sidebar.go` (`paneKey`) and on `toggleRow` drop the
  enter-to-detail wording.
- `detail.go` "enter opens the group" stays true and stays.
- The popup and form enter keys are not touched.

## Tests

- Enter on a non-tree, non-group row keeps the focus on the list.
- Enter on a plan head still folds; enter on a group row still folds.
- Hints no longer list `Detail: enter`.
- Existing tests that press enter to reach the detail switch to tab.
