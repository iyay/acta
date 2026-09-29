---
name: brainstorm-files-scratch-first
description: An Architectural brainstorm files its scratch item and marks it brainstorming before anything else.
tags: [brainstorm]
runs: 1
# The floor is four turns: read the skill, file the item, mark it
# brainstorming, answer. The skill's example id is not one the CLI accepts, so
# the child spends one turn on the id and one looking up the scratch input
# form. Six gives it the work plus the answer without landing on the cap.
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

I want a new billing subsystem for this app. It will change how every
existing module talks to the database, so this is not a small patch.

Start the design work for it now, follow the acta workflow for a change like
this, and stop at the first thing the workflow asks me for.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
