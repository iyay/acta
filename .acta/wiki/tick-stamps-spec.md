---
type: Gotcha
title: Tick stamps the parent file
description: tick --start and --all write started and finished into the plan's parent spec or bug, left uncommitted on the branch
paths: [internal/cli/]
timestamp: 2026-10-09T14:34:45Z
---

`acta tick <plan>#task-N --start` and `--all` also write `started:` and `finished:` lines into the frontmatter of the plan's parent, uncommitted. The parent is a spec, or a bug file when the plan sets `parent: bugs/...`. A diff gate that allows only code plus the plan file trips on it.

Fix: never commit the stamp just to pass a gate; revert the parent copy when it blocks, acta writes it again on the next tick. Brief gates should allow the parent file too. At land, commit the stamp before removing the worktree: it is true data and the tree must be clean.
