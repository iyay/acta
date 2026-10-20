---
name: build-close-no-suite
description: Closing a built plan runs the plan's open type check and reaches review without a second repository-wide fast suite.
tags: [build, close]
runs: 1
# The chain: load the build skill, read the plan, run the wiki check, run the
# type check, run the progress check and reach review. Fourteen leaves room
# for a git error or two without landing on the cap.
max_turns: 14
timeout_seconds: 480
allowed_tools: []
---

Plan PLN-0001 is built. Every task is committed on this branch, parse-note,
and every box is ticked. The parent branch is main. This folder is the
worktree.

Do the Close steps of acta:build, in order, up to the review. Do not start the
review. The plan names the commands this repo uses. Stop when you reach the
review.

The Edit and Write tools are off here, so change files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
