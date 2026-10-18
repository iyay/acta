---
id: SCR-0076
hash: vjukzmc
title: Land runs force and rm -rf the skill never asks for
status: raw
created: "2026-10-19 06:50:17"
schema: "1"
---
# Land runs force and rm -rf the skill never asks for

## Words

### 2026-10-19

In eval runs on 2026-10-19 the land step did things the land skill never names:
- review-understanding-report: `git commit --amend` on the merge commit.
- state-resume (trace e-nMn6yY, call 26 and 28): `rm -rf src/__pycache__` and `git worktree remove --force wt`.

The skill says `git worktree remove <path>` and never delete a worktree with uncommitted work. Find why agents reach for force and amend here, and whether the skill should ban them by name.

## Context

## Log

## Open questions
