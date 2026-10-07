---
id: SCR-0054
hash: byb1bux
title: Tidy commit history as acta's commit rule
status: brainstorming
created: "2026-10-07 19:23:36"
schema: "1"
started: "2026-10-07 19:23:41"
finished: "2026-10-07 19:26:44"
---
# Tidy commit history as acta's commit rule

## Words

### 2026-10-07

Make a tidy commit history acta's official rule for every repo that uses acta.

Target shape (tried by hand on this repo, 1984 commits down to 452):
- one commit per code task, with the task's own message (feat/fix/...)
- commits that only touch planning files, chore, polish, wiki and review fix-ups fold into the task they belong to
- no merge commits; land goes in linear
- real dates kept, final tree identical to the unsquashed branch

Leaning: tidy the branch at land time, not change auto-commit at the source.

## Context

## Log

### 2026-10-07

Round 1 (all recommended):
- Q1: setting in acta config, default tidy (tidy | full).
- Q2: tidy at acta:land, before the branch reaches its parent; auto-commit stays as is.
- Q3: linear, no merge commit; fixed_in points at the last task commit.
- Q4: chore commits on main stay; at the next land, unpushed chore commits fold into the first commit that lands.
- Q5: land remaps old hashes in the branch's planning files to the new ones, in the last commit.

### 2026-10-07

Round 2 (all recommended):
- Q6: no upstream: fold only chore commits after the last land; land records its tip in refs/acta/last-land. With upstream: only after @{upstream}.
- Q7: new Go subcommand acta tidy <base> <branch>; land skill calls it.
- Q8: review fix-ups fold into the last task commit.
- Q9: subject, body, trailers and author stay as written.
- Q10: a clashing replay folds into the commit before it; a final tree that differs stops land and leaves the branch untouched; never fall back to a plain merge.

## Open questions
