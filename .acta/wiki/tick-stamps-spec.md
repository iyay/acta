---
type: Gotcha
title: Tick stamps the parent file
description: tick writes started and finished into the parent spec or bug; a spec with several plans gets them at land
paths: [internal/cli/tick.go, internal/cli/specdates.go, internal/write/mark.go, internal/write/specdates.go]
---

`acta tick <plan>#task-N --start` and `--all` also write `started:` and `finished:` lines into the frontmatter of the plan's parent, uncommitted. The parent is a spec, or a bug file when the plan sets `parent: bugs/...`. A diff gate that allows only code plus the plan file trips on it.

Fix: never commit the stamp just to pass a gate; revert the parent copy when it blocks, acta writes it again on the next tick. Brief gates should allow the parent file too. At land, commit the stamp before removing the worktree: it is true data and the tree must be clean.

A spec with more than one plan is the exception: tick in a linked worktree never writes its dates (in the main checkout it still does, as no land follows), since two worktrees would stamp it with different times and the second tidy land would clash. Land runs `acta spec-dates <plan hash>` on the parent instead. It sets `started` to the earliest plan start there (a later land may move it earlier) and writes `finished` once every plan is done, then commits.
