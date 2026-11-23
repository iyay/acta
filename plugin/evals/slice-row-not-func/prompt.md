---
# guards: names that test and the row or assert it adds
name: slice-row-not-func
description: Slicing a spec for a unit that already has a table test gives a task that names that test and the row it adds, and no new test func.
tags: [slice]
runs: 1
# The chain: load the slice skill, read the config, the depth file, the spec
# and the test file, write the plan, give it an id, commit it. Eighteen leaves
# room for a git error or two without landing on the cap.
max_turns: 18
timeout_seconds: 600
allowed_tools: []
---

The spec .acta/specs/2026-11-23-slug-punct-design.md is approved. Write the
plan for it and save it as .acta/plans/2026-11-23-slug-punct.md. Commit the
plan on main. Then stop and show me the plan. Do not build anything.

The Edit and Write tools are off here, so write files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
