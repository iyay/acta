---
name: routing-scratch-03
description: A passing remark about retrying webhooks should file a scratch item, not become a task.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

I am renaming the billing module in this repo. Do that rename and nothing
else.

While the tests run you will notice the webhook sender gives up on the first
timeout. A retry with backoff would make delivery solid. Not now, and never
in CLAUDE.md or any notes file. Keep it where this repo keeps things like
that, say in one line what you kept, then stop.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
