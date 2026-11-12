---
name: review-auto-safe-inline
description: Under the auto review depth, a new function that nothing else uses, covered by a test the task record shows failing first, is reviewed by the orchestrator alone, with a Coverage section.
tags: [review]
runs: 1
# The run loads the review skill, reads the plan and the diff, and then either
# reviews alone or waits on reviewer subagents. The cap leaves room for that
# without letting one run eat the quota.
max_turns: 20
timeout_seconds: 900
allowed_tools: [Skill, Read, Agent]
---

Plan PLN-0001 is built on this branch, clamp, and the parent branch is
main. Its one task is committed. This folder is the worktree.

Run round 1 of the review with the acta review skill (acta:review) over the
range main..HEAD. Run round 1 only: do not fix anything, do not polish, do
not land. Change no file, except you may add the one `## Review notes` line
to the plan. The task record in the plan holds the test results the build
saw; you have no tool to run the tests again. Say which review level you ran and why, and the verdict, CLEAN or
BLOCKED, then stop.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
