---
type: Decision
title: Land tidies history before moving the parent
description: Tidy mode rebuilds the branch with acta tidy --plan, then moves the parent; bookkeeping rides the tidy tip
paths: [plugin/skills/land/, internal/tidy/]
---

See `plugin/skills/land/tidy.md`.

Tidy mode runs `acta tidy --plan <plan hash>`, then `merge --ff-only` or `reset --keep` for folded parent chores. `full` uses `merge --no-ff`, then `acta land-notes`. A failed tidy stops land; never fall back to a plain merge.

Temporary gate Git commands use the saved absolute owning checkout, never the detached path.

Fold point: `@{upstream}`, else `refs/acta/last-land`, else none. Only a run of `chore(...)` commits folds; pushed commits never do. Before the review marker only `chore`, `polish` and `wiki` fold by type; a `docs` commit outside the planning root stays. If all fold, the subject is the newest commit touching real files, else the newest of no fold type, else the newest non-chore, else the newest, so planning never hides real work (BUG-0043, BUG-0051, BUG-0055).

`--plan` writes `fixed_in: PLN-<hash>`, debt ticks and spec dates into the last tidy commit, read from the merged tree, so nothing lands on the parent after.

`-D` is safe: tidy proved the tip equals the parent plus branch merge. Delete `refs/acta/tidy/<branch>` after.

A fold rewrites planning commits other worktrees grew from (PLN-0247, PLN-0246); their next tidy exits 3 with `clashes with main`. Move such a branch first: `git rebase --onto main <old base> <branch>` in its worktree, resolve by intent, tidy again, run the full suite on the new tip.

Tidy moves each folded `Task:` trailer into the kept commit's last paragraph, once (BUG-0051). Wiki checks ride in `.acta/wiki/.checked`, so a folded re-check stays a check on the kept commit.

Cleanup: [typed approval](gate-output-cleanup.md).
