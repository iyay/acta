---
id: BUG-0046
hash: bwkrdzr
started: "2026-10-15 13:00:28"
---
# Search in the Open box also filters the Done box

## Symptom
With the focus on the Open box, typing a `/` query also narrows the Done box
under it, and opens every plan in the Done tree. The search should touch only
the box that has the focus.

## Root cause
internal/tui/sidebar.go:173 `rowsOf` runs `matchRows` for every pane it is
asked about, not only the focused one. `isOpen` and the legacy group check
`filtering()` with no pane, so the Done tree opens every plan too.

## Repro
Open the TUI on a tab with a Done box (Debt or Plans), focus Open, press `/`
and type a word. The Done box shrinks to matching rows.

## Found in
main after PLN-0147 landed (b552ea5); reported by the user from a screenshot.
