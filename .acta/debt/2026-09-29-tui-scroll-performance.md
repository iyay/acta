---
id: DBT-0029
hash: jpyauw3
parent: plans/2026-09-29-tui-scroll-performance
---
# Review NOTEs: TUI Scroll Performance Implementation Plan

- [x] Task 1 was committed in two commits with the same message (595dc38, 93b3bb3), and fix round 1 in two (c3c169e, 08c87df); each plan step asked for one commit. (stale)
- [ ] (low) view_bench_test.go uses &testing.T{} instead of passing b as testing.TB to paneModel, so a t.Fatal inside the helper does not stop the benchmark cleanly and its TempDir is never cleaned up.
- [x] view_bench_test.go flips wheel direction on every notch, so each frame adds up to zero and the benchmark never scrolls; the plan flipped once per frame.
- [ ] (low) No test turns the wheel while a popup, the slug prompt or search is open; only help is tested, and the guard for all four is one shared line.
- [ ] (low) fromText reads the debt file from disk and is now cached per board, so in manual mode an edit to the debt parent only shows after r.
- [ ] (low) boardModel and other tests swap m.render after New or sized without a fresh dcache; they pass only because nothing builds detail lines first.
- [ ] (low) The pageLines comment at internal/tui/model.go:28 says the wheel jumps pageLines; this was wrong before this plan and the wheel now uses wheelStep.
- [ ] (low) A tick that arrives after help or a popup opens still scrolls the pane under it.
- [ ] (low) The m.dcache reset in newModel right after New does nothing, because New already makes an empty cache.
- [ ] (low) The query field in mark() has no test of its own; removing it keeps every test green.
- [ ] (low) mark() compares the stored cursor id, not the item on screen, so in a fresh tab both list panes read "" and a focus change keeps the mark equal.
- [ ] (low) Pressing g or G in the detail box inside a frame runs before the pending notches, so the box ends at the notch total instead of the top.
- [ ] (low) List notches are dropped when the cursor moves or the focus switches list panes inside the same frame; base scrolled them first.
- [ ] (low) The Done sub-tab case in detail_test.go stays green when the nil-item cache fix is reverted; the other three cases cover it.
- [ ] (low) The if m.wheelDelta == 0 guard around the wheelMark write can never change the result after the drop line; its comment suggests a risk that is not there.
- [ ] (low) The reload-that-kept-the-item test clicks the row first without a comment saying why (the cursor must be written down, see BUG-0010).
- [ ] (low) The key-focus flush path (focus moves to another pane with the same item, then a notch) has no test.
