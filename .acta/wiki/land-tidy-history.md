---
type: Decision
title: Land tidies history before moving the parent
description: Tidy mode rebuilds the branch with acta tidy, then fast-forwards or resets the parent; fixed_in is the tidy tip
paths: [plugin/skills/land/, internal/tidy/]
---

The tidy path lives in `plugin/skills/land/tidy.md`; `SKILL.md` only points to it.

With `commit_history: tidy`, land runs `acta tidy` and moves the parent with `merge --ff-only`, or `reset --keep` when parent chore commits were folded. `full` keeps `merge --no-ff`.

Keep rule: a failed tidy stops the land and never falls back to a plain merge.

Fold point: `@{upstream}`, else `refs/acta/last-land`, else none. Only a run of `chore(...)` commits folds; pushed commits are never touched. Before the review marker only `chore`, `polish` and `wiki` fold by type; a `docs` commit outside the planning root stays. When every commit folds, the subject comes from the newest commit that changed real files, else the newest non-chore, else the newest commit, so a closing planning commit cannot hide the real change (BUG-0043, BUG-0051).

`-D` is safe: the old branch is no ancestor, and tidy proved the tip equals the parent plus branch merge, so it loses nothing. Delete `refs/acta/tidy/<branch>` after.

Known limit: a hash in the last new commit stays old in planning files, since a commit cannot name itself. Read `fixed_in` from `refs/acta/tidy/<branch>` before deleting the ref.

After the parent moves, land runs `acta spec-dates <plan hash>` and only then moves `refs/acta/last-land`, so a split spec never leaves the parent dirty.

Tidy keeps the kept commit's message and moves each folded `Task:` trailer into its last paragraph, once (BUG-0051); git and `internal/commits` read trailers only there. Wiki checks ride in `.acta/wiki/.checked`, so a folded re-check stays a check on the kept commit.
