---
name: polish-light-review
description: A polish commit that only changes comments gets the light review, so no reviewer subagent runs and the reply gives a verdict.
tags: [review]
runs: 1
# The run loads the review skill, reads the plan and the diff, and then either
# decides alone or waits on reviewer subagents. The cap leaves room for that
# without letting one run eat the quota.
max_turns: 14
timeout_seconds: 420
allowed_tools: [Skill, Read, Agent]
---

Plan PLN-0001 is built on this branch, ratio-helper, and the parent branch is
main. Round 1 of the review was CLEAN. The last commit, `polish: review notes
for PLN-0001`, is the polish commit for that round. This folder is the
worktree.

Review that polish commit with the acta review skill (acta:review). Do not
land. Change no file, except you may add the one `## Review notes` line to
the plan. The repo has no tests and no formatter. Say
which review you ran and the verdict, CLEAN or BLOCKED, then stop.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
