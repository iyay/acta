---
type: Decision
title: Finished scratch items are specced, never dropped
description: Dropped means not done or not valid; a written dropped beats the derived specced status
paths: [internal/board/]
timestamp: 2026-10-09T14:15:13Z
---

A scratch item shows specced when a spec names it as parent or with a `closes:` link. That status is derived, so the frontmatter status line alone can mislead; use `acta show`.

A written `dropped` beats the derived specced. So never drop a finished item. Dropped means the idea was not done or is not valid.
