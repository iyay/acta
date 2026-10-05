---
id: BUG-0032
hash: aer6v2y
fixed_in: e0c24ec
finished: "2026-10-06 06:19:06"
---
# Switching to the Plans tab lags when a plan is in review

## Symptom
With a plan in derived status review selected (unmerged worktree, every task ticked), switching to the Plans tab from another tab, or moving onto that plan, freezes the TUI for a few seconds.

## Root cause
The detail ROUND line runs a full worktree scan on every render. `internal/tui/detail.go:460` (`roundText`) calls `rounds(m.cfg)`, which is `board.PlanStates` (`internal/tui/model.go:66`). `planStates` (`internal/board/state.go:54`) runs `git worktree list`, loads a whole board for every worktree and runs a git diff per branch (`branchFiles`). Measured 1.28-1.31 s per call on this repo with one worktree. The detail is built at least twice per frame (`view.go:373` through `detailScroll`, `view.go:442` through `detailLines`), plus `scroll.go:70` and `scroll.go:85`, so one frame costs 2.5 s or more. Only plans in review reach the call, because `roundText` returns early for anything else.

## Repro
1. Have a worktree whose plan has every task ticked and is not merged (status review).
2. Run `acta` (TUI), open another tab, then switch to Plans with that plan selected.
3. The switch takes about 2-3 s. Timing probe: `board.PlanStates(cfg)` takes about 1.3 s per call.

## Found in
main at 87c911c, while PLN-0097 waited in review; found by acta:debug with a timing probe.
