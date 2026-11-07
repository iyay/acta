---
type: Gotcha
title: omp report file blocks worktree remove
description: omp subagents leave an untracked report in .acta/reports, and worktree remove refuses untracked files
paths: [plugin/omp/]
---

omp task subagents write an untracked report under `<worktree>/.acta/reports/`. `git worktree remove` refuses untracked files, and force-deleting is banned.

Fix at land: read the report's concerns first, move the file to the session scratchpad, remove the now-empty reports dir with `rmdir`, then remove the worktree. Moving acta's own report out is not a delete, so it stays allowed. Any other untracked file follows the land rule: show it and let the user choose; never delete it first.
