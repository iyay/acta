---
parent: bugs/2026-10-06-tui-review-plan-detail-lags
id: SPC-0089
created: "2026-10-06 05:33:52"
hash: cq1flgb
started: "2026-10-06 05:37:45"
finished: "2026-10-06 06:04:20"
---
# TUI reads review rounds once per board load, not on every render

Status: Bounded, approved by the user in chat on 2026-10-06. Fixes BUG-0032.

Why: `roundText` (`internal/tui/detail.go`) calls `rounds(m.cfg)`, which is `board.PlanStates`, on every render of a plan in review. One call takes about 1.3 s (it lists worktrees, loads a board per worktree and diffs each branch), and one frame builds the detail two to four times, so the Plans tab freezes for seconds.

Design:
- Run `rounds(m.cfg)` once in the same command that loads the board, off the UI goroutine, so every reload (start, fsnotify, after a write) refreshes it.
- Keep the result on `Model`; `roundText` reads the stored list and never calls `rounds` itself.
- Before the first load lands, a plan in review shows `(none)`, as it does today when no round is found.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: making `board.PlanStates` itself faster.

Tests: a fake `rounds` hook counts its calls; rendering the detail of a plan in review many times after one load makes exactly one call, and the ROUND line still shows the fake round.
