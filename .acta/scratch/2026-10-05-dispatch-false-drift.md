---
id: SCR-0045
hash: rumqw5s
title: Dispatch checkpoint reports drift while omp is still reading the brief
status: raw
created: "2026-10-05 13:54:44"
schema: "1"
---
# Dispatch checkpoint reports drift while omp is still reading the brief

## Words

### 2026-10-05

User, 2026-10-05 (chat in Indonesian, put into English): note it.

Seen three times on 2026-10-05: `acta dispatch send` exits 4 with `checkpoint: drift: missing <task ids>` (slugs live-state fix-1, doctor-executor, polish-task). Each time the pane showed omp had only just read the brief and its todo list was a generic one (Scope / Build / Close), not yet the plan tasks. The work then went on fine.

So the one pane read about 20 seconds after send comes too early, and exit 4 asks the orchestrator to judge real drift by hand every time.

Ideas, not decided: read again after a longer wait or until the todo list names task ids; treat a todo list with no task ids as `unconfirmed`, not drift; only call it drift when the list names work outside the plan.

## Context

## Log

## Open questions
