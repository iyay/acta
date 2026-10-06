---
id: DBT-0009
hash: pximcqy
parent: plans/2026-09-27-tui-polish
---
# Review NOTEs: TUI Polish Implementation Plan

- [x] A selected in-progress row loses its accent color (selected foreground overrides it, scroll.go:160, view.go:55); spec 1 says the accent still shows; user ruling said bright text; ask the user which.
- [x] At about 4 rows tall split can give a pane height 1 and paneView draws no border for h<2 (view.go:194); frame still fits.
- [ ] (low) Selecting the debt file row itself (KindDebt) shows the checklist list and then the full body, so the checklist can appear twice.
- [ ] (low) Wheel-scrolling a list moves only the offset; the selection can go off screen and the next j/k snaps back.
- [x] No revert-red check was run for PLAN-16's tests. (stale)
- [x] AUTHOR lookup runs one git log per .acta file on every board load: 64 git calls vs 30, list --all 0.96s to 2.0s, and the TUI reloads on each watch event; batch it in one git log --name-only or load it lazily for the detail pane (closed.go:56-79, gitc.go:119).
- [ ] (low) gitc.Author has no --follow, so a renamed file falls back to the current user's name (gitc.go:123).
- [x] tui tests went from 24s to 62s: TestViewNeverOverflowsAnyWindow 29s, TestNoRoundedCorners 11s, TestViewFitsEveryTerminalSize 7s; trim the size sweeps.
- [ ] (low) model.go is 790 lines, 10 short of the 800 limit; model_test.go 1801, view_test.go 1021, scroll_test.go 855 lines.
- [x] The omp recipient ran git reset mid-build and squashed its wip commit, dropping orchestrator commit e2591d2 from history (content survived in 9fcf007).
