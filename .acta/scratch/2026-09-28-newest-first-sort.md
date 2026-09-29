---
id: SCR-0001
hash: tnwe78s
title: newest-first-sort
status: raw
created: "2026-09-28"
---
# Lists sort newest first

User, 2026-09-27: lists sort newest first, in every pane.

Not approved work yet. Brainstorm when picked up.

## Brainstorm 2026-09-29

Finding: sortItems (internal/board/board.go:596) already puts the newest file-name date first, but same-day items fall back to plan id A-Z. Pane [2] finished list already sorts by SortTime (last commit).

User answer: the most recently updated item goes first in every list, not the creation date.

Path: Bounded (one sort function, design in chat, no spec file).


User revision 2026-09-29: drop the fixed newest-first order. Each pane gets a sort toggle with two orders: oldest first and newest first. Toggle by a key binding. The pane shows a mark for the current sort.

Held: start only after PLAN-19 (top tabs) lands.

Path: upgraded to Architectural? No. Still one TUI flow, but it waits on PLAN-19 layout, so the design is re-read against that code first.


User answer 2026-09-29: sort key is the date in the file name, for every pane, so all panes agree. Brainstorm now; build still waits for PLAN-19.


## Approved design 2026-09-29

- Key o toggles the sort of the focused pane (free on main and on top-tabs; s is taken).
- Mark in the pane border title: oldest or newest, drawn as a title segment (view.go segment).
- State per pane, in memory only; every pane starts at oldest on each TUI start.
- Order: date in the file name, up or down; same date falls back to ID; tasks stay under their plan in file order.
- Pane [2] finished drops its commit-time newest-first order and follows the same rule, default oldest.
- Red tests first: o flips order and mark; other panes unchanged; same-date ties stable by ID.

Next: acta:plan after PLAN-19 lands; re-check keys and title code against the landed top-tabs code.
