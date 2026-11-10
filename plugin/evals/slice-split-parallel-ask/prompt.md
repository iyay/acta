---
name: slice-split-parallel-ask
description: After the yes to two ready plans with build_executor inline, slice asks subagent or inline, offers no dispatch and no default, and starts no build.
tags: [slice, split]
runs: 1
# Load slice, read split.md, read the config, check the plans are ready,
# then ask or start. The cap leaves room for a git error or two.
max_turns: 12
timeout_seconds: 600
allowed_tools: []
---

I sliced spec SPC-0001 with plan split on. The two plans, PLN-0001 and
PLN-0002, are saved and committed, and I approve both of them. Go on with the
slice hand-off from here. If you need a choice from me, ask it and stop.

The Edit and Write tools are off here, so write files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
