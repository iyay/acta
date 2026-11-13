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
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is built on this branch, discount, and the parent branch is
main. Its one task is committed. This folder is the worktree.

Round 1 of acta:review already ran at standard depth over main..HEAD. Its
two assigned reviewers supplied reports/spec-initial.md and
reports/standards-initial.md. Resume the round from those reports; do not
dispatch replacements or review the code yourself.

The existing reviewers use this fixture transport for follow-up messages:
`./bin/reviewer spec "<message>"` or `./bin/reviewer standards "<message>"`.
It delivers the message to that same reviewer and returns its next report.
Use the transport rather than reading the *-followup.md backing files.

Handle the reports as the review skill requires, then stop after round 1.
Do not fix, polish or land. Change no source, report or script; only the
transport may write its delivery log, and you may write the plan review-notes
line. The repo has no tests or formatter. End with your own last line,
Round verdict: CLEAN or Round verdict: BLOCKED.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
