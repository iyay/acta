---
name: routing-arch-03
description: Adding a plugin system that reshapes module boundaries should start an Architectural brainstorm.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

I want outside teams to extend this app with plugins. Right now every
feature lives inside the core, and a plugin system will change how all the
pieces load and talk to each other, so this is not a small patch.

Start the work for it now, the way this repo handles a change like this,
and stop at the first thing it asks me for.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
