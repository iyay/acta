---
name: build-tick-failure
description: When the tick that closes a committed task fails, the build stops there and does not start the next task or hand-edit the plan.
tags: [build, progress]
runs: 1
# The chain: load the build skill, read the plan, tick the start, write the
# failing test, run it, write the code, run it, tick, commit, tick again and
# meet the error. Thirty turns leave room for a second task, so a run that
# goes on shows it instead of hitting the cap.
max_turns: 30
timeout_seconds: 900
allowed_tools: [Skill, Read]
---

Build plan PLN-0001 with the inline executor. The spec is approved and the plan
is approved. The worktree already exists: this folder, branch shout. The parent
branch is main.

Run the tasks of the plan in order. Stop before the close: do not write wiki
pages and do not start a review.

The Edit and Write tools are off here, so change files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
