---
# guards: Add the red test in this order
name: tdd-row-not-func
description: Fixing a bug in a unit that already has a table test with acta:tdd adds a row to that table, seen failing first, and no new test func.
tags: [tdd, build, claude-only]
runs: 1
# The run loads the tdd skill, reads the plan, does one red-green round and
# one commit. Twenty-five turns leave room for a retry without letting one
# run eat the quota.
max_turns: 25
timeout_seconds: 900
# Write and Edit never come from this list; only the run's --allow-tools gives
# them. So the prompt below tells the agent to write files with Bash when they
# are off.
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is approved. Build its one task with the acta tdd skill
(acta:tdd). Work in this folder, on this branch. Do not make a worktree, do
not run a review and do not land.

Write the test first, watch it fail, then write the code and watch it pass,
then commit. Running the acta tick commands is optional.

If the Write and Edit tools are off, write files with a Bash command.
The acta CLI for this repo is ./bin/acta, git is ./bin/git and python is
./bin/python3. Use those three paths. The plain names may run a different
program in this sandbox.

Stop when the task is committed.
