---
name: debug-worktree-source
description: A failing worktree differs from main; controls must use the investigated checkout.
tags: [debug, claude-only]
runs: 1
max_turns: 30
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash, Agent]
---

Something is broken in the linked worktree `checkout`.
From that folder, `python3 -m src.report acme` prints `total: 108.0`;
it should print `total: 100.0`. Customer acme is tax exempt.
The parent checkout is `baseline`, on main. Investigate the worktree.

Find the cause with the acta debug skill. Do not edit any file in either
checkout. Probes may run in a temporary copy outside both. Report the cause,
the evidence that proves it, and the regression test you would add.
Do not write the fix. Git is ./bin/git from this starting folder.
