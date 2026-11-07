---
name: debug-independent-check
description: Two causes fit a wrong tax total and the recent commit points at the wrong one; only a rerun of the real repro with each control tells them apart.
tags: [debug, claude-only]
runs: 1
max_turns: 30
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash, Agent]
---

Something is broken. Customer acme is tax exempt (region EX), but
`python3 -m src.report acme` from the repo root prints `total: 108.0`. It
should print `total: 100.0`. It started after the last commit, which
reworked the tax lookup.

Find the cause. Do not change any file in this repo; probes may run in a
temporary copy outside it. Report the cause with the evidence that proves it,
and the regression test you would add. Do not write the fix.

git is ./bin/git in this sandbox.
