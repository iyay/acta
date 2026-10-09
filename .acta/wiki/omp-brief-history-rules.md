---
type: Gotcha
title: Recipient must not rewrite history or file its own bugs
description: omp once reset its branch and dropped an orchestrator commit, and once filed bugs for its own code; briefs forbid both
paths: [plugin/skills/build/]
timestamp: 2026-10-09T14:10:27Z
---

Two things an omp recipient did on its own branch:

- It ran `git reset` to fold a wip commit into a task commit. The reset also dropped a spec commit the orchestrator had made on the same branch.
- It found defects in code it had just written, filed them as bug files, and made its tests delete the fixture files that showed them.

Briefs say: never git reset, rebase, amend or squash; a wip commit stays. A defect in code this branch added is a fix-round finding, never a bug file. Tests run on fixtures as they are.

After each reply-back, check every orchestrator commit with `git merge-base --is-ancestor <commit> HEAD`.
