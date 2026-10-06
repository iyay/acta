---
id: DBT-0043
hash: r9yao9j
parent: plans/2026-09-30-tui-drag-select
---
# Review NOTEs: TUI Drag Select and Pink Scratch Implementation Plan

- [x] copyPicked cuts the toast to m.width minus "copied ", ignoring the rest of the status line (select.go:141); y does not cut at all.
- [ ] (low) dragText is uneven on a wide character at the edge: a start inside it takes the whole character, an end inside it drops it.
- [ ] (low) Only a left press clears the highlight; a right or middle press leaves it (plan says left press, spec says press).
- [x] Plan text names the helper withColor; the package has withColors. (stale)
- [ ] (low) TestNoDragWhileSomethingIsOpen stays green with Task 3 reverted: it guards the old mouse guard, not the new wiring.
- [ ] (low) paintDrag puts the band right after the left cut with no reset, so a drag starting inside faint or bold text keeps that attribute on the band (select.go:102).
- [ ] (low) paintDrag tail cut repeats every escape code from before the cut, so band rows carry duplicate codes and the frame gets bigger.
- [ ] (low) The highlight survives a reloadMsg and a clearStatusMsg; after a reload moves rows the band sits over other words until the next key, press, wheel or resize.
- [ ] (low) copyDrag calls bare.draw(), which writes frame.dots on the shared frame pointer; same value, but a hidden side effect of a copy.
- [ ] (low) A wheel turn while the button is held clears the drag; the following motion and release fall through with same=true; not tested.
- [ ] (low) styles.go:79 comment and the scratchColor doc comment repeat the same reason.
- [ ] (low) A right-click press leaves the highlight on and keeps same=true; matches old right-click behaviour, plan does not name it.
