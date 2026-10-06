---
id: DBT-0067
hash: c9qloen
parent: plans/2026-10-01-tui-detail-zoom-wheel-focus
---
# Review NOTEs: TUI Detail Zoom, Wheel Focus, Copy Toast Implementation Plan

- [ ] (low) A second copy within 2 s leaves the status text unchanged, so Update starts no new timer and the first copy's timer hides the second toast early (internal/tui/model.go, copyID and copyPicked); needs a toast serial.
- [ ] (low) Each wheel notch runs geometry() twice, once in hit and once in onPane (internal/tui/model.go); hit could return an on-pane flag so the hot scroll path builds the boxes once.
- [x] focusOnWithADoneRow(t, paneDone) in internal/tui/scroll_test.go presses tab and lands on the detail box, not Done, so the "pane2 gathers" subtests start on detail and the helper comment is wrong.
- [ ] (low) stale = true on a wheel flush in mouse() (internal/tui/model.go) has no test; a mutant without it passes, and a tab from List to detail between notches then a notch over detail is the untested redraw path.
