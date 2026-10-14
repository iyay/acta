---
type: Decision
title: Wiki freshness comes from git ancestry
description: A page is stale when a commit under its paths is not behind a check commit; checks are lines added to wiki/.checked
paths: [internal/wiki/, internal/cli/wiki.go, internal/gitc/gitc.go, internal/doctor/doctor.go]
---

A check commit of a page is a commit that changed the page file, or one whose diff adds a line `<page id> <date>` to `.acta/wiki/.checked`. A page is stale when a commit reachable from HEAD changed a file under its `paths` and is not an ancestor of any check commit. An uncommitted page file counts as checked.

`acta wiki ok <id>...` rewrites `.checked` with today on those pages and commits it as `chore(wiki): re-check <ids>`. Every write sorts it, keeps one line per page and drops gone pages. A commit that only removes a line is never a check.

Why not empty commits with `Wiki-Checked` trailers: they piled up, and tidy pushed the trailers into kept messages.

Why one line per page with a date: the date makes the line change on each check, so the commit has a diff. Freshness never reads it.

Why `merge=union` in `.gitattributes`: two branches that re-check one page both change its line. Union keeps both lines with no conflict; the next rewrite drops the extra one. `acta doctor --fix` adds the line.

Why the file is never cleared at land: tidy would fold the add and the clear into one commit with no diff, and the check would be lost.

The git queries use `--full-history --no-merges`: without the first git drops side commits with the same content; without the second a merge counts as a path change forever. A conflict fix made only inside a merge is not seen.
