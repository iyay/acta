---
id: SCR-0084
hash: h5fd7xz
title: Usage per plan for every harness
status: raw
created: "2026-10-31 17:28:57"
schema: "1"
---
# Usage per plan for every harness

## Words

### 2026-10-31

User asked (2026-10-31, WhatsApp): "analisa dan pelajari codebase terakhir. menurut lo ada missing features yang ideal acta punya?" then "catet aja semua ke scratch".

Agent proposal the user filed: Usage per plan. acta usage reads only Codex; there are no tokens or time per plan for Claude Code and omp.

## Context

Found from `acta help usage`: it reads $CODEX_HOME/sessions only (SPC-0182, internal/usage). Claude Code keeps sessions under ~/.claude/projects; omp has its own store. Tying usage to a plan could use the plan's worktree path or its Task trailers (`acta commits`).

## Log

## Open questions
