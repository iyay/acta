---
id: DBT-0042
hash: r8pcoq5
parent: plans/2026-09-30-tui-status-colors
---
# Review NOTEs: TUI Status Colors Implementation Plan

- [ ] model_test.go TestAStartedTaskKeepsTheAccentAndItsCountAndAgent name is stale; its body asserts the row is not the accent
- [ ] hasWork (inProgress checks Started first) and dotOf (Closed first) disagree: a closed item with a leftover started keeps the 120 ms tick alive with no dot on screen (cheap, same=true)
- [ ] Ascii or no-color profile: goingDot is a bare dot, so any dot in a title or body sets frame.dots and redraws every 120 ms with the same output
- [ ] Each pulse step is a full m.draw(); running pulseDots over a cached pre-pulse frame would be cheaper if a trace shows lag
- [ ] pulseDots matches the exact goingDot bytes; only a 16-color glamour downgrade could in theory emit the same bytes
- [ ] No test pins exactly 8 pulse frames or the 120 ms step; the tests compare against the constants
- [ ] Task 2 row types Done pane, done tree plan, plain row and group row get no direct test; they share code with the tested rows
- [ ] The help test checks helpText and box widths, not the colors of the drawn popup rows
- [ ] Terminal theme pulse frames 1-3 equal frame 0, so ticks redraw with no visible change while a dot is on screen
- [ ] A bold work line in the detail still pulses its dot; spec change 6 covers only the list cursor band
- [ ] scroll.go done-head comment says "A done plan", but a done bug head on Activities takes the same branch
- [ ] scroll.go lipgloss.NewStyle().Render(mark) returns mark unchanged; plain concatenation says the same more clearly
- [ ] A done head row carries an empty green pair before the id after the mark split (harmless, same width and text)
- [ ] TestDonePlanRowKeepsItsMarkPlain covers only an open done plan (-), not a shut one (+)
