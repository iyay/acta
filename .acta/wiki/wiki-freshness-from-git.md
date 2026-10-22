---
type: Decision
title: Wiki freshness comes from git ancestry
description: A page is stale when a commit under its paths is not behind a check commit; checks are lines added to wiki/.checked
paths: [internal/wiki/, internal/cli/wiki.go, internal/gitc/gitc.go, internal/doctor/]
---
Freshness is a review reminder. Stale warnings alone exit 0; format, path,
loading, range and history errors still block. Read affected notes and relevant
diffs, then fix incorrect facts. An unchanged note needs no edit or confirmation.

A check commit changes the page file or adds its `<page id> <UTC time>` line to
`.checked`. A page is stale when a reachable path-changing commit is not an
ancestor of any check commit. A range selects pages; ancestry still uses HEAD.
Uncommitted page edits count as checked. Read the full page and relevant code
before any edit, including paths alone; disappearing warnings do not prove review.

After a real review, an agent or maintainer may run `acta wiki ok <id>...` to
record it. The command rewrites `.checked` and commits the change. Without it,
warnings can persist. Times only change the line; freshness never reads them.

`merge=union` keeps parallel checks conflict-free. A rewrite sorts and deduplicates
lines and drops removed pages. Removed or merely moved lines are not checks.
Never clear the file at land: tidy could fold away the check. This record replaced
empty commits and leaked `Wiki-Checked` trailers. `acta doctor --fix` repairs it.

History is read once per run. `--no-merges --no-renames` preserves old paths for
moves; conflict fixes made inside merge commits are not seen.
