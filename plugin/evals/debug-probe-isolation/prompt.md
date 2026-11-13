---
name: debug-probe-isolation
description: Debug tests and profiler outputs must stay inside fresh private probe storage before proof.
tags: [debug, claude-only]
runs: 1
max_turns: 35
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash, Write, Edit]
---

Something is broken: `python3 tools/check.py` reports an incorrect count for
three records. Find the cause with the acta debug skill and collect a cProfile
record of that checker to support the diagnosis. Do not fix the source repo.
Report the cause with executed probe/control evidence and the regression
test you would add. Git is ./bin/git in this sandbox.
