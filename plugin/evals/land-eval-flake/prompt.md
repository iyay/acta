---
name: land-eval-flake
description: The eval gate of a land shows one red case that passes when it runs again alone, and the land records it with acta flake record before it merges, without running the whole eval set again.
tags: [land, flake]
runs: 1
# The run loads the land skill, runs the suite and the eval set, runs the red
# case again, records the flake, merges, cleans up and reports. Fifty turns
# leave room for the extra eval runs without letting one run eat the quota.
max_turns: 50
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is built and reviewed. Its branch, land, sits in the worktree
./repo-land, and the parent branch is main, checked out in ./repo. The plan's
`## Review notes` say the review is CLEAN, and every task is committed.

Land it with the acta land skill (acta:land). The repo has no type checks, no
formatter and no wiki pages. Its gates are the full suite,
`scripts/test --full`, the eval gate, `scripts/eval`, and the acta and git
checks the skill names. Then write the landing report and stop.

The acta CLI is bin/acta and git is bin/git, both in this folder. Run pwd
first and call them by their full paths. The plain names may run a different
program in this sandbox.
