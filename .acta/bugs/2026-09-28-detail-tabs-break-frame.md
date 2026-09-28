---
id: BUG-2
hash: s1nx
fixed_in: cd5bbfd
---
# Scrolling the detail pane breaks the screen when the body has tabs

## Symptom
Open a plan with Go code (for example PLAN-16 or PLAN-17) in the detail pane and scroll. Lines spill over the pane borders into the sidebar, and old text stays on screen, stacked over the new text.

## Root cause
internal/tui has no tab expansion: detail body lines keep literal "\t". lipgloss.Width("\tab") returns 2, so a tab counts as 0 columns, while the terminal draws it up to the next tab stop (8 columns). Lines are wider than the TUI thinks, overflow the pane, and Bubble Tea's line diff leaves stale cells.

## Repro
1. acta (TUI) in this repo, focus the Plans list, select PLAN-16 (.acta/plans/2026-09-27-tui-polish.md has 67 lines with tabs).
2. Focus the detail pane and scroll down to a Go code block.
3. The frame breaks as in the user's screenshot of 2026-09-28.

## Found in
main at 994065b, found by the user while reading PLAN-17 in the TUI; root cause checked with a throwaway lipgloss.Width probe.
