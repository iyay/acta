---
type: Gotcha
title: Review from the merge base, not the parent tip
description: A parent branch that moved after the worktree was made makes the two-dot diff show deletions the branch never made
paths: [plugin/skills/review/, plugin/skills/build/]
---

The review brief takes `BASE_SHA=$(git rev-parse <parent>)`. When another session lands on the parent while the build runs, that tip is not the branch point, so `git diff <parent>..HEAD` shows the parent's newer commits as deletions the branch never made. Both reviewers then spend time proving nothing was reverted.

Read the base as `git merge-base <parent> HEAD`, and put that sha in the brief. `git log --oneline <base>..HEAD` then lists only the branch's own commits.

Land is unaffected: `acta tidy` and `merge --no-ff` merge the branch, they never diff it against the parent tip.
