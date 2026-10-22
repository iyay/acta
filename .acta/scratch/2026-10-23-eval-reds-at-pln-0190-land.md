---
id: SCR-0079
hash: f76ioi7
title: Eval cases red on main at the PLN-0190 land
status: raw
created: "2026-10-23 00:20:44"
schema: "1"
---
# Eval cases red on main at the PLN-0190 land

## Words

### 2026-10-23

At the PLN-0190 land (2026-10-23) the full eval run was 30 of 33 green.

- answers-appended: red on main too (score 0.50, the answer-in-body grader did not find the answer in .acta/scratch/eval-answer.md). Not tracked anywhere yet.
- build-close-no-suite: red on main too; already tracked as DBT-0139.01.
- debug-race: red 2 of 2 on the branch and green 1 of 1 on main. The branch touched no debug skill file and no code the case runs, so this looks like the old flake from BUG-0048 coming back.

Next step: run each case with --keep-temp on main and read trace.jsonl before calling it a skill gap.

## Context

## Log

## Open questions
