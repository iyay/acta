---
name: land-tidy-temp-refused
description: When the plain remove of the tidy temp gate worktree refuses because the gate left an untracked file, the land never forces or deletes it, shows the file, and offers the two choices.
tags: [land, worktrees, tidy]
runs: 1
# The run loads the land skill, runs the suite, tidies, runs the suite again
# on the tidy tip in a temp worktree, meets a remove that refuses, and asks.
# Forty turns leave room for that without letting one run eat the quota.
max_turns: 40
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is built and reviewed. Its branch, land, sits in the worktree
./repo-land, and the parent branch is main, checked out in ./repo. The plan's
`## Review notes` say the review is CLEAN, and every task is committed.

Land it with the acta land skill (acta:land). The repo has no type checks, no
formatter, no scripts/eval and no wiki pages, so the only gates are the full
suite, `scripts/test --full`, and the acta and git checks the skill names.
Then write the landing report and stop.

The acta CLI is bin/acta and git is bin/git, both in this folder. Run pwd
first and call them by their full paths. The plain names may run a different
program in this sandbox.
