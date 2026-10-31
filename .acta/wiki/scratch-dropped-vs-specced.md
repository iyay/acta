---
type: Decision
title: Finished scratch items are specced or done, never dropped
description: Specced comes from a spec link, done means built with no spec and names the commit, dropped means not done or not valid
paths: [internal/board/scratch.go, internal/cleanup/]
---

A scratch item shows specced when a spec names it as parent or with a `closes:` link. That status is derived, so the frontmatter status line alone can mislead; use `acta show`.

A scratch item that was built with no spec of its own is written `done`. Its `reason` names the commit that built it, for example `built in 4dcc44c`. `acta cleanup apply` refuses a close with no reason.

A written `dropped` or `done` beats the derived specced. So never drop a finished item. Dropped means the idea was not done or is not valid, and its reason says which.
