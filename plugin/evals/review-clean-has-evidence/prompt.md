---
name: review-clean-has-evidence
description: A review of a correct diff asks each reviewer for a Coverage section, so a CLEAN verdict comes with what was read and run.
tags: [review]
runs: 1
# The run loads the review skill, reads the plan and the diff, and waits on
# reviewer subagents. The cap leaves room for that without letting one run
# eat the quota.
max_turns: 20
timeout_seconds: 900
allowed_tools: [Skill, Read, Agent]
---

Plan PLN-0001 is built on this branch, discount, and the parent branch is
main. Its one task is committed. This folder is the worktree.

Run round 1 of the review with the acta review skill (acta:review) over the
range main..HEAD, at the standard depth. Run round 1 only: do not fix
anything, do not polish, do not land. Change no file, except you may add the
one `## Review notes` line to the plan. The repo has no tests and no
formatter. Say the verdict, CLEAN or BLOCKED, then
stop.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
