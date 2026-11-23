---
id: SCR-0098
hash: tlu4n0a
title: Fewer planning commits
status: brainstorming
created: "2026-11-23 20:20:00"
schema: "1"
started: "2026-11-23 20:20:04"
finished: "2026-11-23 20:24:12"
---
# Fewer planning commits

## Words

### 2026-11-23

User (2026-11-23): "too many commits, especially commits that touch only 1 file. That should never happen." Seen in every repo that uses acta, not only this one. Picked A+B.

A. Commit per step, not per write: acta writes stage only; one commit at the end of shape, slice, land, bug close. scratch add does not commit.
B. Land bookkeeping (fixed_in, debt tick, spec dates, scratch done) is written before tidy so it folds into the land.

Evidence, last 30 days: acta 348 commits, 39% one-file, 48% planning; walkie 2633, 45% one-file; shopify-app-admin-dashboard 124, 54% one-file; clever-platform 325 commits in 4 days, 36% one-file, 40% planning, 71 bug files.

## Context

## Log

### 2026-11-23

Round 1 (2026-11-23): Q5 history: rewrite clever-platform history only; other repos keep theirs. Q1-Q4 pending.

### 2026-11-23

Round 1 answers (2026-11-23), all as recommended:
Q1 target: no stream of small planning commits; at most one planning commit per step; a lone note may stay one commit.
Q2 mechanism: keep auto-commit, widen the fold to every planning kind while the planning HEAD is unpushed.
Q3 pushed commits are never amended.
Q4 fixed_in records the plan id; land bookkeeping is written in the worktree before tidy.
Q5 rewrite clever-platform history only (separate job after this spec; user runs the force push).

### 2026-11-23

Spec 1 section 1 (wider fold) approved 2026-11-23.

### 2026-11-23

Spec 1 section 2 (land bookkeeping in the worktree before tidy, fixed_in PLN-<hash>) approved 2026-11-23.

### 2026-11-23

Spec 1 section 3 (multi-path fold, edge cases) approved 2026-11-23. Writing spec 1.

### 2026-11-23

Cold read of SPC-0217 (2026-11-23): user agreed with all 5 findings (tidy does not resolve clashes; spec-dates in a worktree reads stale sibling plans; no parent status write exists; acta tick never commits; fixed_in before merge marks the bug fixed early). Agreed fix: acta tidy writes the land bookkeeping (fixed_in PLN-<hash>, debt ticks, spec dates) as one chore(acta) commit at the tidy tip, computed from the parent side; full mode makes one bookkeeping commit after the merge. Step 4 (parent status) dropped. Asked the reviewer for its 4 cut findings.

### 2026-11-23

Spec 1 rewritten after the cold read (2026-11-23), findings 1-9 applied: tidy --plan writes fixed_in PLN-<hash>, debt ticks and spec dates into the last tidy commit (no new commit); full mode uses a new acta land-notes command, one commit; fold widening restated (chore( prefix and same-kind checks dropped when all paths sit under the root); cleanup apply.go unchanged; land evals get graders for the bookkeeping.

### 2026-11-23

clever-platform history (2026-11-23): branch history-tidy built from 5fa8800, each run of .acta-only commits folded into one; 325 -> 207 commits, tree identical (31e0b09), 8 one-file commits left. main untouched; the user runs the force push.

### 2026-11-23

clever-platform force push done 2026-11-23 on the user order: origin/main 0e10164 -> 45a5705, 207 commits, tree 31e0b09. Old main kept locally as main-pre-fold (5fa8800).

## Open questions
