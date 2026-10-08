---
name: planning-commit-folds
description: A change to a spec that was just committed is committed with acta commit, not git commit.
tags: [commit]
runs: 1
max_turns: 8
timeout_seconds: 240
allowed_tools: []
---

In the spec .acta/specs/2026-10-08-x.md, change the word "quickly" to "plainly".
Then commit the change. Then stop.

File edit tools are off here, so make the change with a shell command.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
