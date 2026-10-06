---
id: DBT-0037
hash: txc3afg
parent: plans/2026-09-30-tui-colors
---
# Review NOTEs: TUI Colors Implementation Plan

- [ ] (low) styles.go: the faint comment "faint paints every row the cursor is not on" is stale; no row uses faint now
- [ ] (low) clearStatusMsg: pressing y twice within 2s on the same row lets the first tick clear the second toast early, because both carry the same text
- [ ] (low) clearStatusMsg: when the text no longer matches, m.same stays false, so the frame is redrawn for nothing
- [ ] (low) TestDetailLabelsAndIDWearColors checks only the STATUS label, not the ID label that is painted after paintID
- [ ] (low) statusLine narrow path fit(right) stays unpainted; joinStatus has already dropped the project there
- [ ] (low) paintRight paints "live" in the accent if the repo folder is named "live" and the window is narrow
- [ ] (low) barLine builds the open-tab band style inline on every draw; a band helper on styles would match the other brushes
- [ ] (low) Detail label painting relies on lipgloss.NewStyle().Render returning the text unchanged after paintID; nothing states it
- [ ] (low) slotWork and slotBlue are both 4; a later change to one will not follow the other
- [ ] (low) TestDetailWorkLinesWearKindColoredIDs never goes through scratchLines or debtLines; they are covered only because they call workLine
- [ ] (low) planLines doc comment still says "one plain line naming each plan", but the id on that line is now in the kind color
- [ ] (low) TestDetailWorkLinesWearKindColoredIDs comment claims it covers all the routes, but it only calls workLine and planLines on one bug
- [ ] (low) styles.paintID: an item with an empty ID and no ShortID matches at 0 and writes an empty colored escape pair; text and width stay right
