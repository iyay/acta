---
name: task-trailer
description: Building a one-task plan inline ends the code commit with the Task trailer line that names the plan hash and the task number.
tags: [build, commit]
runs: 1
# The chain: load the build skill, read the plan, tick the start, write the
# failing test, run it, write the code, run it, tick, commit. The commit may
# come back with a git error in the sandbox and cost a turn or two. Twenty
# leaves room without landing on the cap.
max_turns: 20
timeout_seconds: 600
allowed_tools: []
---

Build plan PLN-0001 with the inline executor. The spec is approved and the plan
is approved. The worktree already exists: this folder, branch shout. The parent
branch is main.

Do Task 1 of the plan and commit it. Then stop. Do not run the close steps and
do not start a review.

The Edit and Write tools are off here, so change files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
