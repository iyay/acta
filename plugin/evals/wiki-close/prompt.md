---
name: wiki-close
description: Closing a built plan re-checks the wiki page that covers a file the branch changed, and bumps its timestamp, before review.
tags: [wiki, build]
runs: 1
# The close is a chain: load the build skill, read its wiki steps, run the
# check, read the page and the diff, fix the page, run the check again, run
# the progress check. The commit may come back with a git error in the sandbox
# and cost a turn or two. Sixteen leaves room without landing on the cap.
# The eval run grants Bash and no edit tool. A child that finds Edit off gives
# up, so the prompt says to change files with Bash.
max_turns: 16
timeout_seconds: 480
allowed_tools: []
---

Plan PLN-0001 is built. Every task is committed on this branch, parse-note,
and every box is ticked. The parent branch is main. This folder is the
worktree.

Do the Close steps of acta:build, in order, up to the review. Do not start the
review. The repo has no tests, so there is nothing to run for that step. Stop
when you reach the review.

The Edit and Write tools are off here, so change files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
