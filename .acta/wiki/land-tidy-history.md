---
type: Decision
title: Land tidies history before moving the parent
description: Tidy mode rebuilds the branch with acta tidy, then fast-forwards or resets the parent; fixed_in is the tidy tip
paths: [plugin/skills/land/, internal/tidy/]
timestamp: 2026-10-07T13:39:26Z
---

With `commit_history: tidy`, land runs `acta tidy` and moves the parent with `merge --ff-only`, or with `reset --keep` when parent chore commits were folded. `full` keeps the old `merge --no-ff`.

Keep rule: a failed tidy stops the land. It never falls back to a plain merge.

Fold rules: the fold point is `@{upstream}`, else `refs/acta/last-land`, else none. Only a run of `chore(...)` commits folds. Pushed commits are never touched.

Why `-D` is safe: the old branch is not an ancestor any more, so `-d` refuses. Tidy already proved the trees match, so `-D` loses nothing. Delete `refs/acta/tidy/<branch>` after.

Known limit: a hash that lands in the last new commit stays old in planning files, since a commit cannot name itself. Read `fixed_in` from `refs/acta/tidy/<branch>` before deleting the ref.
