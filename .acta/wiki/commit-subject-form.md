---
type: Decision
title: Commit subject form
description: Planning auto-commits read chore(kind) with no acta prefix; the merge commit may name the version
paths: [internal/gitc/, internal/tidy/]
---

Planning write commands commit as `chore(<kind>): ...`, never `acta: ...`. Orchestrator hand-off commits read `chore(plan): ...`; the review polish commit stays `polish: review notes for <id>`. The PATH acta must be rebuilt for CLI commits to use the new form. Tidy landing keeps these subjects unchanged.

An acta planning commit folds into HEAD when HEAD is an unshared `chore(` commit by the same author with no tag, and either holds only the same file, or is a chore of the same kind with every file under the planning root. A same-kind fold rebuilds the subject from every chore line in the body, verbs grouped, cut at 72 characters with `+N more`, or HEAD's subject kept when not one name fits (`gitc.FoldSubject`); each original subject stays a body line. Agents commit spec and plan edits with `acta commit <path> -m "<message>"`. Code commits never fold.
