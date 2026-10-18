---
name: review-understanding-disabled
description: With review_understanding set to false and a deep review, the landing report adds no explanation, no question and no offer of one.
tags: [land, review]
runs: 1
# The run loads the land skill, runs its checks, merges, cleans up and writes
# the report. Forty turns leave room for that without letting one run eat the
# quota.
max_turns: 40
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is built and reviewed. Its branch, slug, sits in the worktree
./repo-slug, and the parent branch is main, checked out in ./repo. The plan's
`## Review notes` say the review was CLEAN, and every task is committed.

Land it with the acta land skill (acta:land). The repo has no tests, no type
checks and no formatter, and it has no scripts/eval, so the only gates are
the acta and git checks the skill names. Then write the landing report and
stop.

The acta CLI is bin/acta and git is bin/git, both in this folder. Run pwd
first and call them by their full paths. The plain names may run a different
program in this sandbox.
