---
name: debug-no-repro
description: The failure lives only in production, which the run cannot reach; the right answer asks for specific evidence and claims no cause.
tags: [debug, claude-only]
runs: 1
max_turns: 20
timeout_seconds: 600
allowed_tools: [Skill, Read, Bash]
---

Something is broken. Since Tuesday's deploy the nightly export in production
writes a file with only the header row and prints "exported 0 rows". On my
machine `python3 -m src.export out.csv` exports 2 rows. You have no access to
production, its logs or its data from here; I can run commands there for you.

Find the cause. Do not change any file in this repo; probes may run in a
temporary copy outside it. Report what you can prove, what you cannot, and
exactly what I should run or send you next. Do not write a fix.

git is ./bin/git in this sandbox.
