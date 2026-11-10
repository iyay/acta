---
name: slice-split-parallel-set
description: With parallel_plans subagent, slice asks nothing and opens the parallel build file for the two ready plans.
tags: [slice, split]
runs: 1
# Load slice, read split.md, read the config, check the plans are ready,
# then ask or start. The cap leaves room for a git error or two.
max_turns: 14
timeout_seconds: 600
allowed_tools: []
---

I sliced spec SPC-0001 with plan split on. The two plans, PLN-0001 and
PLN-0002, are saved and committed, and I approve both of them. Go on with the
slice hand-off from here. If you need a choice from me, ask it and stop.

The Edit and Write tools are off here, so write files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
