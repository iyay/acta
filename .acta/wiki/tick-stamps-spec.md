---
type: Gotcha
title: Tick stamps the parent file
description: tick writes started and finished into the plan and its parent; in a worktree a spec gets them at land instead
paths: [internal/cli/tick.go, internal/cli/specdates.go, internal/write/mark.go, internal/write/specdates.go, internal/write/landnotes.go]
---

`acta tick <plan>#task-N --start` and `--all` also write `started:` and `finished:` into the frontmatter of the plan's parent, uncommitted. The parent is a spec, or a bug file when the plan sets `parent: bugs/...`. A diff gate that allows only code plus the plan file trips on it.

Fix: never commit the stamp just to pass a gate; revert the parent copy when it blocks, acta writes it again on the next tick. Brief gates should allow the parent file too. At land, commit the stamp before removing the worktree: it is true data and the tree must be clean.

A spec in a linked worktree is the exception: tick never writes its dates there (in the main checkout it still does, as no land follows). Main can change the spec while the build runs (`acta set <spec> status`), and `SetField` adds a new key at the end of the frontmatter, so a worktree `started:` and a main `status:` would clash at merge. Land writes the dates instead: `acta tidy --plan` into the tidy tip, read from the merged tree, or `acta land-notes` right after a full-mode merge. Both run `write.LandNotes`: `started` is the earliest plan start (a later land may move it earlier), and `finished` comes once every plan is done. A spec of type bug counts: the folder decides. `acta spec-dates` still works for older skill text.

A bug file still gets its dates from tick in a worktree, unless several plans share it.
