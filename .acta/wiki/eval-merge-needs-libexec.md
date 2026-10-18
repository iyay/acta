---
type: Gotcha
title: Eval land cases need a libexec git
description: In the eval sandbox git merge dies with "stash failed"; link libexec/git-core/git to the copied bin/git
paths: [plugin/evals/]
---

A scaffold copies git into `bin/git`. Read-only git and commits work, but `git merge` starts a second git (`git stash create`, then `git maintenance`) from its exec path, `bin/../libexec/git-core`. That folder is missing, so git falls back to the `/usr/bin/git` shim, the sandbox blocks the shim's cache write, and the merge fails with `fatal: stash failed`.

Fix: in the scaffold, `mkdir -p libexec/git-core` and `ln -s ../../bin/git libexec/git-core/git`. See `plugin/evals/review-understanding-report/scaffold.sh`.

Two more traps in a case that lands a worktree:
- acta writes an untracked `.acta/.gitignore` (`state/`) on its first run. Left untracked, it blocks `git worktree remove` and the agent reaches for `--force`. Commit it on main in the scaffold.
- A plan `hash:` must be 7 characters; a shorter one makes `acta id` print `bad hash` into the report.

`git branch -d` still prints `could not write config file .git/config` in the sandbox. The branch is deleted; ignore it.
