---
name: brainstorm-files-scratch-first
description: An Architectural brainstorm files its scratch item and marks it brainstorming before anything else.
tags: [brainstorm]
runs: 1
max_turns: 3
timeout_seconds: 240
allowed_tools: []
---

I want a new billing subsystem for this app. It will change how every
existing module talks to the database, so this is not a small patch.

Start the design work for it now, follow the acta workflow for a change like
this, and stop at the first thing the workflow asks me for.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
