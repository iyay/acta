---
name: routing-scratch-01
description: A passing remark about caching the reports API should file a scratch item, not become a task.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

The test suite in this repo fails on one case. Fix that one case.

While you are in there you will see every call to the reports API
downloads the same file again. A small cache in front of it would fix that
for good. I do not want it now, and I do not want it in CLAUDE.md or any
notes file. Keep it where this repo keeps things like that, say in one line
what you kept, then stop.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
