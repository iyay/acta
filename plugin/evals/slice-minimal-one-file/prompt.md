---
name: slice-minimal-one-file
description: Slicing a spec when the repo pins plan_depth minimal writes a minimal plan and reads only the minimal depth file, never the full one.
tags: [slice, depth]
runs: 1
# The chain: load the slice skill, read the config, read the one depth file,
# read the spec, write the plan, give it an id, commit it. Sixteen leaves room
# for a git error or two without landing on the cap.
max_turns: 16
timeout_seconds: 600
allowed_tools: []
---

The spec .acta/specs/2026-11-04-greet-design.md is approved. Write the plan for
it and save it as .acta/plans/2026-11-04-greet.md. Commit the plan on main. Then
stop and show me the plan. Do not build anything.

The Edit and Write tools are off here, so write files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
