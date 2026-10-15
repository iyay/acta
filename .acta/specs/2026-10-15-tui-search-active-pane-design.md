---
id: SPC-0137
created: "2026-10-15 12:34:53"
hash: t7qvxyw
---
# TUI search filters the active pane only

## Problem

Pressing `/` in the TUI searches every item on the board. The results replace
whatever pane has focus, so a search on the Bugs tab also shows plans, specs
and done items. The reader wants the search to stay inside the pane they are
looking at.

## Design

- `/` filters the rows the active pane already shows. Build the normal rows of
  that pane first (`openRows`, `activityRows`, `doneRows`), then drop the rows
  whose item does not match the query.
- A match is the same rule as today: the query, trimmed and lower-cased, is
  inside the item's Title, Ref or Slug.
- Tree panes (Plans) and Activity keep a head row when the head matches or
  when any row under it matches. A head with no match and no matching child
  is dropped.
- Switching tab or pane while a query is set keeps the query; it now filters
  the new pane.
- The global search is removed: `searchRows` in `internal/tui/sidebar.go` and
  `Board.Search` in `internal/board/board.go` go away, since the TUI is their
  only caller.

## Testing

- Two tabs each hold an item with "foo" in the title. Searching "foo" on one
  tab shows only that tab's item.
- The Done pane with a query shows only matching done items.
- A plan tree with one matching task keeps its plan head; a plan with no match
  is gone.

## Out of scope

A global search on another key. The user ruled it dropped.
