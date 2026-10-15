---
created: "2026-10-15 17:17:23"
id: SPC-0141
hash: z7x13ls
started: "2026-10-15 17:20:35"
finished: "2026-10-15 17:28:07"
---
# TUI list filter by priority and status

## Problem

The List box shows every open item of a tab. Bugs and debt carry a priority
(high, medium, low) and every kind carries a status, but the reader cannot
narrow the list to one of them.

## Design

- `f` cycles the priority filter of the active tab: All, high, medium, low,
  none, back to All. "none" keeps only items with no priority. It works on the
  Bugs and Debt tabs.
- `F` cycles the status filter of the active tab over the open statuses that
  tab shows: Bugs `open, fixing`; Specs `draft, approved, in-progress`;
  Scratch `raw, brainstorming`; Plans `todo, in-progress`. Debt has only
  `open` open, so `F` there changes nothing and the status bar says why.
- On a tab with no priority, `f` does the same as `F`.
- Both filters can be on together. The List box title shows the active ones,
  for example `Bugs [H · fixing]`.
- Each tab keeps its own filters while the TUI runs. They are not saved; a new
  TUI starts at All.
- Filtering runs on the rows `openRows` builds, after the sort. An empty
  result shows "no items match" and keeps the cursor in range.
- Plans: a plan head stays when the head or any task under it matches the
  status filter, like the search rule in SPC-0137.
- `f` and `F` appear in the key hints and in the `?` help.

## Testing

- Model tests per key: after each press, check the rows of the List box and
  its title.
- Bugs with H, M, L and unset items: `f` steps through each level and "none".
- `f` then `F` on Bugs keeps only items matching both.
- A plan with one `in-progress` task stays under `F` = in-progress; a plan
  with none is gone.
- `F` on Debt leaves the rows unchanged and sets the status bar text.

## Out of scope

Filters on worktree, author or agent. Saving filters between runs.
