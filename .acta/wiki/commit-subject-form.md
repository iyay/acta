---
type: Decision
title: Commit subject form
description: Planning auto-commits read chore(kind) with no acta prefix; the merge commit may name the version
paths: [internal/gitc/, internal/tidy/]
timestamp: 2026-10-07T13:56:33Z
---

Planning write commands commit as `chore(<kind>): ...`, never `acta: ...`. Orchestrator hand-off commits read `chore(plan): ...`; the review polish commit stays `polish: review notes for <id>`. The PATH acta must be rebuilt for CLI commits to use the new form. Tidy landing keeps these subjects unchanged. After parallel merges, check the version files: same-number bumps merge silently and can leave the version behind.
