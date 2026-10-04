---
id: SCR-0041
hash: utm7npc
title: Live work state for hand-offs between agents
status: raw
created: "2026-10-04 18:46:46"
schema: "1"
---
# Live work state for hand-offs between agents

## Words

### 2026-10-04

Split from SCR-0040 on 2026-10-04 at the user's call.

User's words (2026-09-27, chat in Indonesian, put into English): share live work state and hand-offs between agents and harnesses (Claude Code, omp, Codex). That means what an agent is doing and how far it got, so the next agent goes on without a long brief. The user called this core, next to agent memory.

Ideas raised on 2026-10-04, not decided:

- One state file per running plan, rewritten in place and never appended. Fixed sections, like the STATE block in the Context Language Models paper: status, agents running, key files, findings, next.
- .acta/state/ is taken by runtime files (sessions.json, gitignored).
- Cheaper option: a ## State section inside the plan file, next to the task ticks that already track progress.

## Context

## Log

## Open questions
