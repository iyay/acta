---
# guards: test it once, at the layer that owns it
name: lean-tests
description: Building a two-layer plan with acta:tdd gives each case one test at the layer that owns it, and the upper layer tests wiring only.
tags: [build, tdd]
runs: 1
# The run loads the tdd skill, reads the plan, and does two red-green rounds
# with two commits. Thirty turns leave room for that and for a retry or two,
# without letting one run eat the quota.
max_turns: 30
timeout_seconds: 900
# Write and Edit are listed so the case passes when a run grants them. The land
# gate grants only Bash and Read, so the prompt below also says Bash can write.
allowed_tools: [Skill, Read, Write, Edit, Bash]
---

Plan PLN-0001 is approved. Build both of its tasks, in order, with the acta
tdd skill (acta:tdd). Work in this folder, on this branch. Do not make a
worktree, do not run a review and do not land.

For each task, write the test first, watch it fail, then write the code and
watch it pass, then commit. Running the acta tick commands is optional.

If the Write and Edit tools are off, write files with a Bash command.
The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.

Stop when both tasks are committed.
