---
id: SCR-0043
hash: ejy6g6b
title: TUI indicator for a plan under review
status: raw
created: "2026-10-05 07:14:13"
schema: "1"
---
# TUI indicator for a plan under review

## Words

### 2026-10-05

"butuh indicator kalo task lagi di-review di tui."

In English: the TUI needs an indicator when a task is under review.

## Context

Filed 2026-10-05 while review round 2 of PLN-0081 was running (worktree ../acta-wiki, branch wiki).

- Right now the board shows PLN-0081 as `status: done` (derived, progress 10/10), though its review is still running and nothing has landed. A plan whose tasks are all ticked looks finished during review, fix rounds and polish.
- In acta, review is per plan at the close (acta:review: up to three rounds, then the polish), not per task. A fix round adds a task, and that turns progress back to 9/10 for a while.
- No acta command marks "review running" today. The review skill dispatches reviewers and runs no acta write. The TUI and board code have no review state (grep for "reviewing" in internal/board and internal/tui finds nothing).

## Log

## Open questions
