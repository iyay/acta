---
id: DBT-0038
hash: svhzrjd
parent: plans/2026-09-30-popup-dim-contrast
---
# Review NOTEs: Popup Dim Contrast Implementation Plan

- [ ] (low) view.go cover comment still says lines behind the box go "grey and faint"; dim is no longer faint
- [ ] (low) No view-level test checks that a short line still ends in \x1b[K; only the paintFrame unit test does, so a draw call passing width 0 would stay green
- [ ] (low) TestTerminalThemeUsesANSISlots calls paintFrame("a\nb", 1), so both lines are full and the terminal-theme check never reaches the erase branch
- [ ] (low) TestFullLinesKeepTheirLastCell checks "not a blank" rather than the exact wall rune the spec names
- [ ] (low) TestFullLinesKeepTheirLastCell runs only at 80, 120 and 204 wide and only with ? open; odd widths, very narrow windows and the n/s popups are not covered
- [ ] (low) TestHelpKeysWearTheAccent checks the box width only at 120 wide; at 80 and below fit cuts the help lines
- [ ] (low) paintFrame and fit count ambiguous-width runes (…, ←, →, box lines) as one cell; a terminal that draws them two wide would disagree (not new)
