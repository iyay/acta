---
name: routing-arch-01
description: Rewriting auth so every module talks to it differently should start an Architectural brainstorm.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

I want a new login system for this app. Every module checks users its own
way today, and the new one will change how all of them talk to the session
store, so this is not a small patch.

Start the work for it now, the way this repo handles a change like this,
and stop at the first thing it asks me for.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
