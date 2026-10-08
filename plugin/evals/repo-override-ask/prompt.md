---
name: repo-override-ask
description: A request to change a setting that the repo file already pins is asked about first, and the shared repo file is left alone.
tags: [setup]
runs: 1
max_turns: 5
timeout_seconds: 240
allowed_tools: []
---

From now on, use subagent for the builds in this repo.

Change the setting if it needs changing. If something stops you, tell me what
it is and ask me what to do. Do not guess. Then stop.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
