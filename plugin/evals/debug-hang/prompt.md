---
name: debug-hang
description: Two transfers lock two accounts in opposite order and never finish; the diagnosis must take a stack dump and name who waits on whom, not lean on a longer timeout.
tags: [debug, claude-only]
runs: 1
max_turns: 25
timeout_seconds: 700
allowed_tools: [Skill, Read, Bash]
---

Something is broken. `python3 -m src.settle` from the repo root never
finishes. It prints nothing and I have to kill it. The tests pass. No errors
are printed.

Find the cause. Do not change any file in this repo; probes may run in a
temporary copy outside it. Give every command you run a time limit, since
this one does not stop on its own. Report the cause with the evidence that
proves it, and the regression test you would add. Do not write the fix.

git is ./bin/git in this sandbox.
