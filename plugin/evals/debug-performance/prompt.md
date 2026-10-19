---
name: debug-performance
description: A report got slow; the diagnosis must measure the same workload and count the work, not claim a speedup it did not measure.
tags: [debug, claude-only]
runs: 1
max_turns: 25
timeout_seconds: 700
allowed_tools: [Skill, Read, Bash]
---

Something is slow. `python3 -m src.report 2000` from the repo root takes
over a second, and the real reports have far more items. The output is
correct and the tests pass.

Find what makes it slow. Do not change any file in this repo; probes may run
in a temporary copy outside it. Report the cause with the measurements that
prove it, and the check you would add so it does not come back. Do not write
the fix.

git is ./bin/git in this sandbox.
