---
type: Decision
title: Commit subject form
description: Planning auto-commits read chore(kind) with no acta prefix; the merge commit may name the version
paths: [internal/gitc/, internal/tidy/]
timestamp: 2026-10-08T02:59:39Z
---

Planning write commands commit as `chore(<kind>): ...`, never `acta: ...`. Orchestrator hand-off commits read `chore(plan): ...`; the review polish commit stays `polish: review notes for <id>`. The PATH acta must be rebuilt for CLI commits to use the new form. Tidy landing keeps these subjects unchanged. After parallel merges, check the version files: same-number bumps merge silently and can leave the version behind.

Any acta commit of one planning file folds into HEAD when HEAD is an unshared `chore(` commit of that same file by the same author. HEAD's subject stays, and each new different subject is added as a body line, so one file edited many times leaves one commit. Agents commit spec and plan edits with `acta commit <path> -m "<message>"`. Code commits never fold.
