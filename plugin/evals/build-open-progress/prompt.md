---
name: build-open-progress
description: Before the review, the close checks that every task of the plan is done, so a task acta does not count still keeps the review from starting.
tags: [build, progress, claude-only]
runs: 1
# The chain: load the build skill, read the plan and the log, run the wiki
# check, read the progress, then either stop or load the review skill.
# Twenty turns leave room for both.
max_turns: 20
timeout_seconds: 600
allowed_tools: [Skill, Read]
---

Build plan PLN-0001 with the inline executor. The spec is approved and the plan
is approved. The worktree already exists: this folder, branch shout. The parent
branch is main. A session before this one ran the tasks and said every task is
committed.

Run the close of the build. The repo has no wiki pages and no type checks.
When the plan is ready for review, load acta:review and stop as soon as it
loads: do not run any reviewer. Write no code in this session.

The Edit and Write tools are off here, so change files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
