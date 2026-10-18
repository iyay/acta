---
id: SCR-0074
hash: ilg5eeb
title: Split one spec into several parallel plans
status: brainstorming
created: "2026-10-18 18:44:45"
schema: "1"
started: "2026-10-18 18:44:47"
finished: "2026-10-18 20:15:10"
---
# Split one spec into several parallel plans

## Words

### 2026-10-18

Let one spec split into several plans, so several sessions can build the same spec at once. Useful when the build executor is inline: each session runs one plan in its own worktree. User wants it as an option in acta setup.

Today: plans link to a spec through parent:, and nothing stops two plans per spec, but slice never offers a split, there is no rule for which plans can run together, no file-overlap check, and spec status may not know about partial progress.

## Context

## Log

### 2026-10-18

2026-10-18: user split the inline-default ask out; it ships first as its own Bounded spec.

### 2026-10-18

2026-10-18: fact found: board already sums tasks over every plan with the same parent (parentStatus in internal/board/board.go), so a spec with 3 plans shows done only when all their boxes are ticked. Inline default landed as f64e67c.

### 2026-10-18

2026-10-18 round 1 answers (user took every pick): Q1 setting plan_split in acta setup, default off. Q2 slice cuts by non-overlapping file areas; a plan that needs another writes after: PLN-xxxx. Q3 at most 4 plans per spec; no split when one plan is enough. Q4 slice prints one prompt line per plan (build PLN-xxxx); plans with after: are marked wait. Q5 each plan lands on its own via tidy; build refuses a plan whose after: has not landed.

### 2026-10-18

2026-10-18: design approved in three sections (setting plan_split, slice split with after:, acta ready check + board waiting mark).

### 2026-10-18

2026-10-18: outside review found 3 gaps, all fixed in spec: split plans clash on spec started/finished (land now writes them); ready reads git show <parent ref>, not the board; after: rejects non-plan ids, self-reference and cycles with exit 2.

### 2026-10-18

2026-10-18: second outside review, 3 more gaps fixed: ready --head checks the landing commit is an ancestor of a reused worktree; spec started may move earlier on a later land; land commits the spec dates before last-land and checks the main checkout is clean.

## Open questions
