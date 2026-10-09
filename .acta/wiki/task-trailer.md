---
type: Decision
title: Task trailer on code commits
description: "Each code commit of a plan task ends with a Task trailer, PLN-<hash>#<n>; acta commits and the Commits screen read it"
paths: [internal/commits/, plugin/skills/build/, internal/cli/tick.go]
timestamp: 2026-10-09T14:25:00Z
---

Every code commit of a plan task ends with one line: `Task: PLN-<hash>#<n>`. The `<hash>` is the 7 character hash of the plan. The `<n>` is the task number with no leading zero.

Why the hash and not the short id: `acta id --fix-duplicates` can renumber short ids after a land. A hash never changes.

Why a trailer and not a stored sha: tidy rewrites every sha at land, but it keeps the commit messages.

A fix-round or polish commit carries one trailer per task it fixes. A planning `chore(` commit carries none.

`acta tick` only warns when HEAD lacks the trailer. It never fails.

`acta commits <plan> [task]` and the TUI Commits screen (`d`) read these trailers.
