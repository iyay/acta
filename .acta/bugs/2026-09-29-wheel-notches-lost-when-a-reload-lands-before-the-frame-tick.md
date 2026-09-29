---
ref: plans/2026-09-29-tui-scroll-performance#task-4
id: BUG-0010
hash: h319txx
---
# The first reload throws away the scroll of a tab whose cursor was never moved

## Symptom
Open a tab and scroll the detail box or a list without moving the cursor first. When a
reload comes in, whether from the live watcher or from `r`, the box you scrolled jumps back
to the top, even though the item on show did not change. The same happens with key
scrolling, not only with the wheel.

## Root cause
Every pane's cursor starts as the empty string (`freshTab`, `internal/tui/sidebar.go`). The
`reloadMsg` case calls `moveTo(m.cursor())`, and the cursor now resolves to row 0. `moveTo`
sees an id that differs from the stored `""`, treats it as a new item, and sets the detail
offset to 0. It also moves the list back to the cursor row. This code was already on base
d943290 before the wheel batching work. Review round 3 probed base and got offset 0 for this
sequence, the same as the branch.

When this was first filed it said base gives 6 and blamed the wheel mark. Both claims were
wrong and have been corrected here.

## Repro
In `internal/tui`, on base d943290 or later:

    m := paneModel(t, paneDetail)   // the cursor was never moved, so m.sel is still ""
    // scroll the detail box two lines, with the wheel or with keys
    m = reloaded(m)                 // same board, nothing changed on disk
    m.off[paneDetail]               // gives 0, want the scrolled offset

Clicking the row on show first writes the cursor down, and then the reload keeps the offset.

## Fix direction
Write the cursor down when a tab opens (`New`, `openTab`), or let `moveTo` skip the reset
when the stored cursor is `""`. Either one fixes both the wheel case and the key case.

## Found in
Branch tui-scroll-perf, while mutation-verifying task 4 of PLN-0037. Review round 3 then
traced it to base.
