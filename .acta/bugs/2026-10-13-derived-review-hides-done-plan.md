---
id: BUG-0044
hash: p1zabpo
priority: medium
---
# Plan on a non-checked-out branch stays review after status done

## Symptom
A plan file that lives only on another local branch (not the checked-out one) shows `review (derived)` on the board even after `acta set <id> status done` was committed on that branch. The written `status: done` is never read.

## Expected
A plan whose frontmatter says `done` shows `done`, wherever the file lives.

## Repro
1. Checkout `hotfix/2.0.9`. PLN-0017 lives only on `aj/development` (`.acta/plans/2026-10-09-fe-prod-fusion-visible.md`), all task boxes ticked.
2. In a temp worktree on `aj/development`: `acta set PLN-0017 status done` (commit 7a5eac353).
3. `git show aj/development:.acta/plans/2026-10-09-fe-prod-fusion-visible.md` prints `status: done`.
4. Back on `hotfix/2.0.9`: `acta show PLN-0017` prints `status: review (derived)`.
Same for PLN-0018, 0020, 0025, 0027, 0028, 0029. Plans whose file is on the checked-out branch flip to `done (frontmatter)` fine.

## Root cause
`internal/board/board.go` `derive()`: any plan with `Worktree != ""` and every child task done goes into the `review` map (around line 596-609). The status switch then hits `case it.Kind == KindPlan && review[it.ID]` (line 662) before any case that reads `it.fmStatus`, so a written `done` loses. A plan read from a non-checked-out branch has `Worktree` set to that branch name, so it is treated as "ticked but not merged" even when that branch is the parent and the work landed long ago.

## Fix idea
Skip the review rule when the frontmatter status is `done` (or any closed status): `if it.fmStatus == "done" { continue }` before adding to `review`, or move the review case below a frontmatter-done case.

## Found in
be-pmis, 2026-10-13, acta v0.1.45, while closing plans that land never set to done. Related: SCR-0064 (land does not write status done).
