---
id: DBT-0008
hash: szcvavw
parent: plans/2026-09-27-review-debt
---
# Review NOTEs: Review Debt Implementation Plan

- [x] appendDebt writes the file with no lock and no tmp+rename, so an append racing acta tick on the same debt file can lose one edit.
- [ ] TickLine trusts the line number from board load; a hand edit before the tick can mark the wrong box (it only checks the line is some box).
- [ ] acta set on a debt file status writes a value the board then ignores, because derive always recomputes it from the lines.
- [ ] tui fromText reads and parses the debt file on every detail render because the board drops the parent link; keep parent on the Item.
- [x] view.go is about 860 lines and model.go is over 800, past the 800-line limit.
- [ ] debt new names the file by today's date, so a second review of the same plan on a later day makes a new file instead of appending (spec section 3 says append).
- [ ] land skill only reports the debt file when step 8 ticked items; spec says the report always names the debt file the review made and its count.
- [ ] cmdDebtNew checks for a TTY before loading the board, so a bad plan id with a TTY stdin prints "pipe NOTEs on stdin" instead of the id error.
- [ ] splitNotes does not strip a leading "- [ ] ", so piping an existing checklist gives "- [ ] - [ ] text".
- [ ] Debt tick uses exit 2 for a wrong flag while a plain tick uses exit 1 for the same kind of mistake.
- [ ] title.go: when the open tab name alone is wider than the pane, topLine cuts the drawn text but tabX still makes a full-width box.
- [ ] title.go:66 comment says tabs drop "farthest from the open one first" but dropOrder drops from the right.
- [ ] TestTabTitleAlwaysShowsTheOpenTab never asserts a dropped tab has a zero-width box; pane [2] tabs have no width-by-width test.
- [ ] topLine segment-cut fallback is now unreachable at every drawn width and could hide a future title/box mismatch.
