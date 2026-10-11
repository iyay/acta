---
type: Decision
title: Wiki freshness comes from git ancestry
description: A page is stale when a commit under its paths is not behind a check commit; acta wiki ok re-checks with no page edit
paths: [internal/wiki/, internal/cli/wiki.go]
---

A check commit of a page is a commit that changed the page file, or one with a `Wiki-Checked: <page id>` trailer. A page is stale when a commit reachable from HEAD changed a file under its `paths` and is not an ancestor of any check commit. An uncommitted or untracked page file counts as checked.

`acta wiki ok <id>...` makes an empty `chore(wiki): re-check` commit with one trailer per page. `acta tidy` moves those trailers onto the commit they fold into.

Why no `timestamp` field: every re-check bumped the same line with a new value, so every merge of two branches conflicted on it. An empty commit has no diff, so it cannot conflict. Dates were also wrong: a branch with old commit dates passed against a newer timestamp it never saw.

The git queries use `--full-history --no-merges`. Without full history git drops side commits with the same content. Without no-merges a merge commit counts as a path change forever. So a conflict fix made only inside a merge commit is not seen.

An old `timestamp` line in a page is ignored, never an error.
