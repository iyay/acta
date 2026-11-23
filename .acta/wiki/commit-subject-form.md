---
type: Decision
title: Commit subject form
description: Planning auto-commits read chore(kind) with no acta prefix; the merge commit may name the version
paths: [internal/gitc/, internal/tidy/]
---

Planning writes commit as `chore(<kind>): ...`, never `acta: ...`; agents commit spec and plan edits with `acta commit <path> -m "<message>"`. The review polish commit stays `polish: review notes for <id>`. Tidy landing keeps these subjects. Rebuild the PATH acta for CLI commits to use a new form.

A planning write folds into HEAD when HEAD is unpushed, untagged, not a merge, by the same author, has no `Task:` trailer, no other worktree stands on it, and every file of HEAD and of the write sits under the planning root (`gitc.canFold`). A root that is the repo itself never folds this way. Outside the root only the old rules fold: HEAD holds only the same file, or `acta id` folds into a non-`chore(` HEAD that added exactly the files it writes. Code commits never fold.

Subject of a fold:
- one kind: rebuilt from every chore line in the body, cut at 72 characters with `+N more` (`gitc.FoldSubject`);
- mixed kinds or no kind, on a chore HEAD: `chore(acta): <kinds> of <stems>`, kinds from the subject lines then the folders;
- a hand-written HEAD (`docs(spec): ...`, `spec: b`) keeps its subject, even when a new file comes in; the write goes in as a body line.

Every earlier subject stays a body line. On a chore subject, a fold that adds a file the subject does not name rewrites it to name every new file, old names cut first (BUG-0045).
