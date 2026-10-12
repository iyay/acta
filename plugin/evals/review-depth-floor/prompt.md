---
name: review-depth-floor
description: A diff that changes an auth check gets two reviewer subagents, even when the repo sets review_depth to quick.
tags: [review]
runs: 1
# The run loads the review skill, reads the plan and the diff, and then either
# decides alone or waits on reviewer subagents. The cap leaves room for that
# without letting one run eat the quota.
max_turns: 20
timeout_seconds: 900
allowed_tools: [Skill, Read, Agent]
---

Plan PLN-0001 is built on this branch, auth-fix, and the parent branch is
main. Its one task is committed. This folder is the worktree.

Run round 1 of the review with the acta review skill (acta:review) over the
range main..HEAD. Do not land. Change no file, except you may add the one
`## Review notes` line to the plan. The repo has no tests and no formatter.
Say which review depth you ran and the verdict, CLEAN or BLOCKED, then stop.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
