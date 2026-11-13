---
type: Gotcha
title: Review from the merge base, not the parent tip
description: A parent branch that moved after the worktree was made makes the two-dot diff show deletions the branch never made
paths: [plugin/skills/review/, plugin/skills/build/]
---

Since PLN-0232 the review skill reads the base as `git merge-base <parent> HEAD`, and a plugincheck test fails when `git rev-parse <parent>` comes back. Before that the brief took `BASE_SHA=$(git rev-parse <parent>)`. When another session lands on the parent while the build runs, that tip is not the branch point, so `git diff <parent>..HEAD` shows the parent's newer commits as deletions the branch never made. Both reviewers then spend time proving nothing was reverted.

Build close and review both capture `git merge-base <parent> HEAD` and the exact HEAD before handing off. Put those SHAs in the brief; `git log --oneline <base>..HEAD` then lists only the branch's own commits.

Record each round's exact input ranges. An incomplete report keeps every input range pending, including its fix range, so a later review of only fixes cannot lose the earlier incomplete work.

Land is unaffected: `acta tidy` and `merge --no-ff` merge the branch, they never diff it against the parent tip.
