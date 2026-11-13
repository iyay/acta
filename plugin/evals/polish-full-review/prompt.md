---
name: polish-full-review
description: At standard depth a logic polish gets separate Spec and Standards reviewers.
tags: [review]
runs: 1
# The run loads the review skill, reads the plan and the diff, and then either
# decides alone or waits on reviewer subagents. The cap leaves room for that
# without letting one run eat the quota.
max_turns: 20
timeout_seconds: 900
allowed_tools: [Skill, Read, Agent]
---

Plan PLN-0001 is built on this branch, ratio-helper, and the parent branch is
main. Round 1 of the review was CLEAN. The last commit, `polish: review notes
for PLN-0001`, is the polish commit for that round. This folder is the
worktree.

Review that polish commit with the acta review skill (acta:review) at standard depth. Do not
land. Change no file, except you may add the one `## Review notes` line to
the plan. The repo has no tests and no formatter. Say
which review you ran and the verdict, CLEAN or BLOCKED, then stop.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.

For trace evidence, start each reviewer Agent prompt with
`ACTA_REVIEW_AXIS=BOTH`, `ACTA_REVIEW_AXIS=SPEC`, or
`ACTA_REVIEW_AXIS=STANDARDS`, then a newline and its full normal review brief.
Choose the declaration that matches the work actually assigned to that
reviewer. Helpers receive no reviewer declaration. Keep these declarations
out of helper prompts and quote them nowhere in Agent inputs except as
the actual reviewer prompt prefix. A resumed reviewer keeps its identity.
