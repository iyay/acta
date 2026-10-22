---
id: SCR-0078
hash: lykft2o
title: Cheap plan updates and conflict-free planning merges
status: brainstorming
created: "2026-10-22 23:27:15"
schema: "1"
started: "2026-10-22 23:27:18"
finished: "2026-10-22 23:35:15"
---
# Cheap plan updates and conflict-free planning merges

## Words

### 2026-10-22

User ask 2026-10-22: plan and task updates should cost fewer tokens. Agents still hand-edit plans with Read + Edit for `## Fix round <n>`, `## Polish`, `## Review notes`, an unasked `## Verification evidence` section, and the `after:` list. Also: merges of planning files often conflict, even when only a date changed.

Rulings so far (user said "ok semua as long as lebih irit token dan efisien"):
- New command `acta plan add <plan> <section> < stdin`, fixed section list: fix-round, polish, review-notes.
- No `## Verification evidence` section; findings go to `acta state set <plan> findings`.
- fix-round and polish: the command builds the heading, task number, verify line (`--verify`) and commit box; stdin is one box per line.
- `acta set <plan> after <ids>` with the same checks as `acta ready`; empty value clears.
- New writes auto-commit like `acta state set`.

Conflict root cause (repro in a scratch clone): SetField in internal/write/frontmatter.go:163 appends a new key at the end of the frontmatter. Tick in a worktree appends `started:`; `acta set` on main appends `status:` or `type:`. Both insert between `hash:` and `---`, so git marks a conflict on both the plan and the spec.

## Context

## Log

### 2026-10-22

Round 2 rulings: Q1 (B) dates written at land, not by tick in a linked worktree. Q2 plan started = the --start record in internal/write/agents.go, fallback first task commit time. Q3 two plans: plan 1 dates at land (first), plan 2 acta plan add + acta set after + skill text; no wait between them. Tick in the main checkout still stamps.

### 2026-10-22

Design sections 1-2 approved. Refinement: only spec dates move to land (tick in a linked worktree never stamps the spec; SpecDates also covers one-plan specs). Plan dates stay in the worktree, since main never edits a building plan. Q2 dropped. Guard: write commands on main that change a plan refuse when that plan is building in a worktree (goes in plan 2).

### 2026-10-22

Section 3 approved: acta plan add (fix-round, polish, review-notes; prints the new task id), acta set <plan> after with ready checks, skill text moves to the commands, no hand-written plan sections. Spec edits stay with Edit.

## Open questions
