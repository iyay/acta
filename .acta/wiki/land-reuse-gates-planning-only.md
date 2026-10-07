---
type: Decision
title: Reuse gates when only planning files differ
description: After a merge where only planning-root files changed, reuse the branch gates instead of re-running
paths: [plugin/skills/land/]
timestamp: 2026-10-07T13:56:05Z
---

At land step 6: when the branch tree equals the merged tree, or `git diff --name-only <branch> HEAD` lists only planning-root files (the folder above `plans/` in `acta show <id> --path`), reuse the branch gates. Any other path, including `plugin/*.md`, re-runs the gates.

Tidy landing reuses the branch gates only when the tidy tip differs from the branch in planning-root paths alone. When the parent moved, other paths differ and the gates run on the tidy tip first.

Why: full re-runs after merges where main only gained spec or plan commits waste a whole gate for no new code.
