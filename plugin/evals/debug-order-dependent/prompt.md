---
name: debug-order-dependent
description: A test fails only after another test runs; the diagnosis must name the polluting test and keep the failing order.
tags: [debug, claude-only]
runs: 1
max_turns: 20
timeout_seconds: 600
allowed_tools: [Skill, Read, Bash]
---

Something is broken. `python3 -m unittest discover -s tests -t .` from the
repo root fails in tests/test_c_fmt.py. Running
`python3 -m unittest tests.test_c_fmt` on its own passes.

Find the cause. Do not change any file in this repo; probes may run in a
temporary copy outside it. Report the cause with the evidence that proves it,
and the regression check you would add. Do not write the fix.

git is ./bin/git in this sandbox.
