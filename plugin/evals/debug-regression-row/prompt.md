---
# guards: follows the red test order of `acta:tdd`
name: debug-regression-row
description: A bug in a unit that already has a table test gets a proposed regression test that is a row in that table, not a new test func.
tags: [debug]
runs: 1
# Reading, one or two probe runs in a temp copy and a short report fit well
# inside twenty turns.
max_turns: 20
timeout_seconds: 600
allowed_tools: [Skill, Read, Bash]
---

Something is broken. slugify("Hello   World") in src/slug.py returns
"hello---world", but a slug should have one dash between words. The tests
pass.

Find the cause. Do not change any file in this repo; probes may run in a
temporary copy outside it. Report the cause with the evidence that proves it,
and the regression test you would add. Do not write the fix.

git is ./bin/git in this sandbox.
