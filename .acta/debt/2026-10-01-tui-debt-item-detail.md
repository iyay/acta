---
id: DBT-0054
hash: xv2iyx2
parent: plans/2026-10-01-tui-debt-item-detail
---
# Review NOTEs: Debt Item Detail Shows the Full Note Implementation Plan

- [ ] (low) Copying a wrapped debt note from the detail puts a newline at every wrap point, and Wordwrap also breaks after hyphens, so a long path copies in pieces.
- [ ] (low) Tests of the debt item detail only sweep widths 10 to 160; nothing checks widths 2 to 9, where wide runes are most likely to overflow.
- [ ] (low) At width 1 a wide rune or emoji in a note stays 2 cells wide and fit blanks it; a 1-column detail pane is not a real layout.
- [ ] (low) Hardwrap with preserveSpace true can start a wrapped line with a space; no text is lost but it looks slightly off.
- [ ] (low) TestDetailDebtItemNoteLosesNothingAtAnyWidth compares non-space text only, so a lost or added space inside a line (a - b read as a-b) would pass; a per-line verbatim check would be stronger.
- [x] The debtLines loop was also tidied into the if-line form, a small step past the removal the plan asked for. (stale)
- [ ] (low) An empty debt note gives one blank middle line in the detail.
- [ ] (low) The TUI panes never select a debt file, so debt file detail is covered only by TestDetailDebtFileListsEveryLine calling debtLines directly.
