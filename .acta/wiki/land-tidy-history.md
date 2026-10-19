---
type: Decision
title: Land tidies history before moving the parent
description: Tidy mode rebuilds the branch with acta tidy, then fast-forwards or resets the parent; fixed_in is the tidy tip
paths: [plugin/skills/land/, internal/tidy/]
---

The tidy path lives in `plugin/skills/land/tidy.md`; `SKILL.md` only points to it, so it stays under its byte cap.

With `commit_history: tidy`, land runs `acta tidy` and moves the parent with `merge --ff-only`, or with `reset --keep` when parent chore commits were folded. `full` keeps the old `merge --no-ff`.

Keep rule: a failed tidy stops the land. It never falls back to a plain merge.

Fold rules: the fold point is `@{upstream}`, else `refs/acta/last-land`, else none. Only a run of `chore(...)` commits folds. Pushed commits are never touched. When no branch commit is kept, the one commit takes the message of the newest commit that changed real files, else the newest that is not a `chore`, so a closing planning commit cannot hide the real change (BUG-0043, BUG-0051).

Why `-D` is safe: the old branch is not an ancestor any more, so `-d` refuses. Tidy proved the tip equals the merge of parent and branch, so `-D` loses nothing. Delete `refs/acta/tidy/<branch>` after.

Known limit: a hash that lands in the last new commit stays old in planning files, since a commit cannot name itself. Read `fixed_in` from `refs/acta/tidy/<branch>` before deleting the ref.

After the parent moves, land runs `acta spec-dates <plan hash>` and only then moves `refs/acta/last-land`, so a split spec never leaves the parent dirty for the next land.

Tidy keeps the kept commit's message, and moves every `Task:` trailer of a folded commit into its last paragraph, once (BUG-0051). Git and `internal/commits` read trailers only there. A `docs` commit that changes files outside the planning root stays a commit; before the review marker only `chore`, `polish` and `wiki` fold by type. When every commit folds, the subject comes from the newest commit that changed real files. Wiki checks ride in `.acta/wiki/.checked`, so a folded re-check stays a check on the kept commit.
