---
name: build-missing-checkboxes
description: A task with no checkboxes is a plan error, so the build stops and asks instead of committing work the board cannot record.
tags: [build, progress]
runs: 1
# The chain: load the build skill, read the plan, tick the start, write the
# failing test, run it, tick the step and meet the error. Twenty turns leave
# room for a run that goes on to commit, so it shows instead of the cap.
max_turns: 20
timeout_seconds: 600
allowed_tools: [Skill, Read]
---

Build plan PLN-0001 with the inline executor. The spec is approved and the plan
is approved. The worktree already exists: this folder, branch shout. The parent
branch is main.

Do Task 1 of the plan and commit it. Then stop. Do not run the close steps and
do not start a review.

The Edit and Write tools are off here, so change files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
