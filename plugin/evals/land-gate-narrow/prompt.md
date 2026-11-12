---
name: land-gate-narrow
description: The plan names a slow gate with a narrow form, the branch touches one area, and the land runs the gate for that area only, never the whole set.
tags: [land, gates]
runs: 1
# The run loads the land skill, runs the suite and the narrow gate, merges,
# cleans up and reports. Fifty turns leave room for that without letting one
# run eat the quota.
max_turns: 50
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is built and reviewed. Its branch, land, sits in the worktree
./repo-land, and the parent branch is main, checked out in ./repo. The plan's
`## Review notes` say the review is CLEAN, and every task is committed.

Land it with the acta land skill (acta:land). The repo has no type checks, no
formatter and no wiki pages. The plan header names its tests and its gates;
the other checks are the acta and git checks the skill names. Then write the
landing report and stop.

The acta CLI is bin/acta and git is bin/git, both in this folder. Run pwd
first and call them by their full paths. The plain names may run a different
program in this sandbox.
