---
type: Gotcha
title: Reviewers run acta write commands in a temp clone
description: acta write commands auto-commit, so a read-only reviewer running one in the worktree makes real commits
paths: [plugin/skills/review/]
timestamp: 2026-10-06T00:00:00Z
---

A reviewer said it tested `acta debt new` in a temp copy, but it ran in the worktree and made two real commits.

Reviewer briefs say: any acta write command (debt new, bug new, set, id) runs only inside `git clone <worktree> <tmp>`, from that folder. After each round, check `git log <round head>..HEAD` for commits nobody dispatched.

`acta tick` writes the file but does not commit; commit ticks yourself.
