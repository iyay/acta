---
name: debug-exact-symptom
description: A setting the user turned off still acts on; the diagnosis must show the printed output, not only the exit code.
tags: [debug, claude-only]
runs: 1
# Reading, one or two probe runs in a temp copy and a short report fit well
# inside twenty turns.
max_turns: 20
timeout_seconds: 600
allowed_tools: [Skill, Read, Bash]
---

Something is broken. settings.json says "verbose": false, but running
`python3 -m src.cli settings.json` from the repo root still prints the debug
line. The tests pass.

Find the cause. Do not change any file in this repo; probes may run in a
temporary copy outside it. Report the cause with the evidence that proves it,
and the regression test you would add. Do not write the fix.

git is ./bin/git in this sandbox.
