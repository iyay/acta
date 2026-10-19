---
name: build-inline-flow
description: Building a one-task plan inline writes the code in this session, with no worker hand-off and no install for a repo that needs none, and closes the task's ticks.
tags: [build, inline]
runs: 1
# The chain: load the build skill, read the plan, tick the start, write the
# failing test, run it, write the code, run it, tick, commit. Twenty leaves
# room for a git error or two without landing on the cap.
max_turns: 20
timeout_seconds: 600
allowed_tools: []
---

Build plan PLN-0001 with the inline executor. The spec is approved and the plan
is approved. The worktree already exists: this folder, branch slug. The parent
branch is main.

Do Task 1 of the plan and commit it. Then stop. Do not run the close steps and
do not start a review.

The Edit and Write tools are off here, so change files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
