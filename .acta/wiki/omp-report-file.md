---
type: Gotcha
title: omp report file blocks worktree remove
description: omp subagents leave an untracked report in .acta/reports, and worktree remove refuses untracked files
paths: [plugin/omp/]
timestamp: 2026-10-05T15:28:00Z
---

omp task subagents write an untracked report under `<worktree>/.acta/reports/`. `git worktree remove` refuses untracked files, and force-deleting is banned.

Fix at land: read the report's concerns first, move the file to the session scratchpad, remove the now-empty reports dir, then remove the worktree.
