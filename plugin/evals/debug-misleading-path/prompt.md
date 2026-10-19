---
name: debug-misleading-path
description: Two look-alike callers and a red test nearby; the diagnosis must prove which caller makes the reported number.
tags: [debug, claude-only]
runs: 1
max_turns: 20
timeout_seconds: 600
allowed_tools: [Skill, Read, Bash]
---

Something is broken. Tenant A's dashboard shows the wrong total.
`python3 -m src.dashboard A` from the repo root prints 350.0, but tenant A's
orders add up to 100.0. The test suite is also red right now.

Find the cause. Do not change any file in this repo; probes may run in a
temporary copy outside it. Report the cause with the evidence that proves it,
and the regression test you would add. Do not write the fix.

git is ./bin/git in this sandbox.
