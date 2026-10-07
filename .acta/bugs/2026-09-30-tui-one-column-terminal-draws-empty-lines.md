---
id: BUG-0017
hash: sdihchv
status: wontfix
finished: "2026-10-01 18:48:26"
---
# In a terminal one column wide, the TUI draws empty lines

## Symptom
With the terminal one column wide, many frame lines come out 0 cells wide instead of 1, so old text can stay on screen. A review probe counted 390 such lines in the narrow layout, the same on main 587107b and on the PLN-0045 branch head e48bce4.

## Root cause
Not traced yet. It sits in the narrow layout path of draw in internal/tui/view.go (the focused box drawn full width), which PLN-0045 did not change.

## Repro
Size the model to 1 x N (tea.WindowSizeMsg{Width: 1, Height: 20}), call View, and check lipgloss.Width of every line: some are 0, not 1.

## Found in
Branch tui-scroll-60fps (PLN-0045), review round 1, Standards axis; confirmed already on main 587107b.
