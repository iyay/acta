---
type: Decision
title: Wiki freshness comes from git ancestry
description: A page is stale when a commit under its paths is not behind a check commit; checks are lines added to wiki/.checked
paths: [internal/wiki/, internal/cli/wiki.go, internal/gitc/gitc.go, internal/doctor/]
---

A check commit of a page is a commit that changed the page file, or one whose diff adds a line `<page id> <UTC time>` to `.acta/wiki/.checked`. A page is stale when a commit reachable from HEAD changed a file under its `paths` and is not an ancestor of any check commit. An uncommitted page file counts as checked.

`acta wiki ok <id>...` rewrites `.checked` with the current time on those pages and commits it as `chore(wiki): re-check <ids>`. Every write sorts it, keeps one line per page and drops gone pages. A removed or merely moved line is no check.

Why not empty commits with `Wiki-Checked` trailers: they piled up and leaked into kept messages.

Why a time to the second: a second check on one day must still change the line, or there is no commit. Freshness never reads it.

Why `merge=union` in `.gitattributes`: two branches that re-check one page both change its line. Union keeps both lines with no conflict; the next rewrite drops the extra one. `acta doctor --fix` adds the line.

Why the file is never cleared at land: tidy would fold the add and the clear into one commit with no diff, and the check would be lost.

Check reads git once per run (`gitc.ReadWikiHistory`), matching paths and ancestry in Go: per-page queries took minutes on 20k commits. The log uses `--no-merges --no-renames`: a merge is no path change, a move keeps its old path. A conflict fix inside a merge is not seen.
