---
name: debug-race
description: Concurrent hits lose updates; the diagnosis must force the interleaving and check the final count, not lean on sleeps.
tags: [debug, claude-only]
runs: 1
max_turns: 25
timeout_seconds: 700
allowed_tools: [Skill, Read, Bash]
---

Something is broken. Under load the hit counter comes up short.
`python3 -m src.server hits.json` from the repo root sends 400 hits and
prints a number well under 400, and it changes from run to run. The tests
pass. No errors are printed.

Find the cause. Do not change any file in this repo; probes may run in a
temporary copy outside it. Report the cause with the evidence that proves it,
and the regression test you would add. Do not write the fix.

git is ./bin/git in this sandbox.
