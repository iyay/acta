---
id: DBT-0120
hash: enjd4b8
parent: plans/2026-10-15-search-skips-unfocused-pane
---
# Review NOTEs: Search leaves the unfocused box alone

- [ ] (medium) Switching focus clears the query, but the box just left keeps the scroll offset of its filtered rows, so its cursor can sit off screen until it gets the focus back (internal/tui/model.go focusPane).
- [ ] (medium) The "Filter: " label takes 7 more cells than "/", so on a narrow window the typed text is cut sooner; no test runs the status line width loop with a query (internal/tui/view.go:604).
