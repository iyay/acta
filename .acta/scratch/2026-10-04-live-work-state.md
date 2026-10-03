---
id: SCR-0041
hash: utm7npc
title: Live work state for hand-offs between agents
status: brainstorming
created: "2026-10-04 18:46:46"
schema: "1"
started: "2026-10-05 07:49:51"
finished: "2026-10-05 08:14:43"
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

### 2026-10-05

Q1 main scenario: user picked (1) context runs out or session dies mid-plan; a fresh session in any harness continues without the user re-explaining. Hand-off to omp and parallel-agent visibility come later.

### 2026-10-05

Q2 storage: user picked a ## State section inside the plan file, committed with the ticks, travels with the worktree, no new kind. Work without a plan (debug, shape) is out of scope for now.

### 2026-10-05

Q3 writer: user picked mixed. Mechanical facts (current task, last commit, worktree, review round) come from acta; thinking parts (findings, next, open rulings) are written by the agent through an acta command that replaces the section in place and commits.

### 2026-10-05

Q4 reading: user picked a SessionStart injection: a short summary per running plan (build started, not landed): current task, last commit, worktree, plus next from ## State; acta state <plan> prints the full view. Mechanical facts computed at read time, not written (proposed; user answered 1 without objecting, confirm in design).

### 2026-10-05

Section 1 approved: plan file gets ## State with ### Next, ### Findings, ### Open rulings; agent writes via acta state set <plan> next|findings|rulings (stdin body, replaces that subsection, commits 'acta: state PLN-x') in the build worktree; mechanical facts never written, computed at read time (confirmed); cap about 10 lines per subsection, command refuses more; State stays after land, acta:land moves lasting findings to wiki or debt.

### 2026-10-05

Section 2 approved: running plan = started, not finished, has a worktree (internal/trees); acta state <plan> prints current task, last commit, worktree, last review round, then the three subsections; acta state alone lists one line per running plan; SessionStart injects max 3 lines per plan and max 3 plans for this repo and its worktrees, plus one rule line to write next before stopping; silent on failure.

### 2026-10-05

Section 3 approved: build updates next after each task commit, findings and rulings as they come (replaces the private .acta/handoff rule); land moves lasting findings to wiki or debt before merge; omp extension uses the same session-start hook; dispatch brief adds one line to read acta state first; Go tests for set/view/hook plus one eval (fresh session resumes from Next without asking); version bump to 0.1.6 last.

## Open questions
