---
id: SCR-0033
hash: gauqdwi
title: Config for how deep a plan goes
status: brainstorming
created: "2026-10-01 05:15:07"
schema: "1"
started: "2026-10-01 05:16:08"
---
# Config for how deep a plan goes

## Words

### 2026-10-01

User 2026-10-01: "menarik.. kita kayaknya perlu config buat nentuin planing, mau seberapa dalem planing-nya. minimalm hanya instruksi kecil kayak barusan (harusnya lebih Concise), atau default which termasuk nulis real code di step planing."

## Context

Came up while building SPC-0054. The user said "skip plan, langsung build", but acta:build refuses to run without a plan file (tick, reply-back and dispatch read it). So the orchestrator wrote PLN-0063 (.acta/plans/2026-10-01-show-path-lookup.md): 2 tasks, verify lines and short steps, no real code. PLN-0062 is an example of the default depth, with full test and implementation code in its steps.

Setting would live next to build_executor in ~/.acta/config.yaml (internal/config/user.go), read by acta:plan. Two levels named so far: minimal (short instructions, tighter than PLN-0063) and default (real code in the steps).

## Log

### 2026-10-01

Q1 levels: user picked (a) two levels, `minimal` (tasks, files, verify line, one-sentence steps, no code) and `full` (today: real code in steps).

### 2026-10-01

Q2 default: user picked `full` when unset (same as today). Per-run override `/plan minimal|full`, like `/build <executor>`, for that run only, never saved: yes.

### 2026-10-01

Q3 approval: user picked (b) for `minimal`: plan is written, committed, and build starts with no plan yes. The spec yes still gates. `full` keeps the plan yes.

## Open questions
