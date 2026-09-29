---
id: DBT-0030
hash: yfc66vf
parent: plans/2026-09-29-tui-notes-round4
---
# Review NOTEs: TUI Notes Round 4 Implementation Plan

- [ ] OSC 52 fallback in copyText writes to os.Stdout from Update while the Bubble Tea renderer also writes there, so bytes can mix into a frame.
- [ ] When the OSC 52 write succeeds the pbcopy error is dropped, so a terminal that ignores OSC 52 shows "copied X" with no clipboard change.
- [ ] Over SSH into a Mac, pbcopy succeeds on the remote machine, so the status says copied while the local clipboard gets nothing.
- [ ] pbcopy runs synchronously in the key handler instead of as a tea.Cmd, blocking the UI until it exits.
- [ ] Activities heads sort by id number across kinds (BUG vs PLN), and those numbers come from separate sequences, so it is not creation order.
- [ ] Detail scroll offset means the middle part in sticky mode and the whole block otherwise; a resize across the limit clamps but does not convert it, so the view jumps.
- [ ] detailParts builds the whole detail 3 to 4 times per frame (linesAt, detailScroll, detailView).
- [ ] No test holds an older model's shutActs fixed after a fold on Activities.
- [ ] Help line still says "space enter open or shut a plan row"; enter on a task row now opens the detail and Activities heads can be bugs.
- [ ] treeRows reads m.openPlans directly while treeMark goes through isOpen; they agree only because the tree tab is never Activities.
- [ ] Sticky detail scrollbar thumb spans the whole wall though only the middle scrolls.
- [ ] Task 1: below about 16 columns the open tab name gets cut; widths 0 and 1 draw dashes, not a box.
- [ ] Task 10's commit 43bda74 does not name the cause the debug step proved.
- [ ] Duplicate short ids: y on the second plan's task copies an id that resolves to the first plan's task.
- [ ] TestAPopupNeverCoversTheTabBox only checks the too-short case; it stays green if popupRect always puts the box at line 0.
- [ ] TestAPopupIsWholeOnTheScreen allows the box to reach the status line at every height, not only when the box is as tall as the screen.
- [ ] A popup taller than the screen now draws a box row over the status line, which is outside the ruling and untested.
- [ ] view.go comments at View ("status line always has a row of its own") and cover ("over the middle of the body") are now wrong.
- [ ] The test comment above TestAPopupNeverCoversTheTabBox says a screen too short for the popup keeps the tab box, but popupRect moves the box over it.
- [ ] Help text lines are cut on the right at widths 50 and 80 (pre-existing).
- [ ] internal/write TestNewBug flaked once on TempDir RemoveAll cleanup (.git/objects not empty); passed on rerun.
