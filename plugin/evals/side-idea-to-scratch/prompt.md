---
name: side-idea-to-scratch
description: A side idea from the user during a build is filed as a scratch item, not as memory or a new task.
tags: [scratch]
runs: 1
# The prompt hands over two jobs, so the child reads the debug skill and the
# scratch skill before it files. Three turns is the cap for one job; the
# control case that scores green needs only that.
max_turns: 5
timeout_seconds: 240
allowed_tools: []
---

The test suite in this repo fails on one case. Fix that one case.

While you are in there you can see that every call to the reports API
downloads the same file again. Keep that as an idea in this repo. Do not fix
it now, and never write it into CLAUDE.md or any other notes file. Say in one
line what you filed, then stop.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
