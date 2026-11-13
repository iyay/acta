---
name: review-false-blocker
description: A reviewer hands back a BLOCKER that the code does not show, and the orchestrator checks it, rejects it and opens no fix round.
tags: [review]
runs: 1
# The run loads the review skill, reads two reports, the plan and one source
# file, and writes at most one plan section. The cap leaves room for that
# without letting one run eat the quota.
max_turns: 20
timeout_seconds: 900
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is built on this branch, discount, and the parent branch is
main. Its one task is committed. This folder is the worktree.

Round 1 of the review with the acta review skill (acta:review) ran at the
standard depth, and both reviewers have reported. The Spec-axis report is in
reports/spec.md and the Standards-axis report is in reports/standards.md. Do
not dispatch any reviewer again. Take the round from there as the review
skill says for findings a reviewer returns. Do not polish and do not land.
Change no source file, report or repro script; only the supplied repro may
write its execution record. The repo has no tests and no formatter. Say what
you did with each BLOCKER, then end with one last line, `Round verdict: CLEAN`
or `Round verdict: BLOCKED`, and stop.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
