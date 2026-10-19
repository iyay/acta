---
name: land-baseline-fail
description: A red full suite in landing keeps the earlier passing baseline and its log, keeps the failure output, and stops before the merge.
tags: [land, baseline]
runs: 1
# The run loads the land skill, reads baseline.md, runs the red suite, keeps
# its output, reports and stops. Thirty turns leave room for the checks and
# one retry of the command without letting one run eat the quota.
max_turns: 30
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is built and reviewed. Its branch, land, sits in the worktree
./repo-land, and the parent branch is main, checked out in ./repo. The plan's
`## Review notes` say the review is CLEAN, and every task is committed.

Land it with the acta land skill (acta:land). The repo has no type checks, no
formatter, no scripts/eval and no wiki pages, so the only gates are the full
suite, `scripts/test --full`, and the acta and git checks the skill names.
This tree may fail its suite; report what the gate says and stop.

The acta CLI is bin/acta and git is bin/git, both in this folder. Run pwd
first and call them by their full paths. The plain names may run a different
program in this sandbox.
