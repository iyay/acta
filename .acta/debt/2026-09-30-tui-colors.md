---
id: DBT-0037
hash: txc3afg
parent: plans/2026-09-30-tui-colors
---
# Review NOTEs: TUI Colors Implementation Plan

- [ ] styles.go: the faint comment "faint paints every row the cursor is not on" is stale; no row uses faint now
- [ ] clearStatusMsg: pressing y twice within 2s on the same row lets the first tick clear the second toast early, because both carry the same text
- [ ] clearStatusMsg: when the text no longer matches, m.same stays false, so the frame is redrawn for nothing
- [ ] TestDetailLabelsAndIDWearColors checks only the STATUS label, not the ID label that is painted after paintID
- [ ] statusLine narrow path fit(right) stays unpainted; joinStatus has already dropped the project there
- [ ] paintRight paints "live" in the accent if the repo folder is named "live" and the window is narrow
- [ ] barLine builds the open-tab band style inline on every draw; a band helper on styles would match the other brushes
- [ ] Detail label painting relies on lipgloss.NewStyle().Render returning the text unchanged after paintID; nothing states it
- [ ] slotWork and slotBlue are both 4; a later change to one will not follow the other
- [ ] TestDetailWorkLinesWearKindColoredIDs never goes through scratchLines or debtLines; they are covered only because they call workLine
- [ ] planLines doc comment still says "one plain line naming each plan", but the id on that line is now in the kind color
- [ ] TestDetailWorkLinesWearKindColoredIDs comment claims it covers all the routes, but it only calls workLine and planLines on one bug
- [ ] styles.paintID: an item with an empty ID and no ShortID matches at 0 and writes an empty colored escape pair; text and width stay right
