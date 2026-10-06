---
type: Gotcha
title: Tick stamps the parent spec
description: tick --start and --all write started and finished into the parent spec, left uncommitted on the branch
paths: [internal/cli/]
timestamp: 2026-10-06T01:42:42Z
---

`acta tick <plan>#task-N --start` and `--all` also write `started:` and `finished:` lines into the parent spec frontmatter, uncommitted. A diff gate that allows only code plus the plan file trips on it.

Fix: never commit the stamp just to pass a gate; revert the spec copy when it blocks, acta writes it again on the next tick. Brief gates should allow the parent spec too. At land, commit the stamp before removing the worktree: it is true data and the tree must be clean.
