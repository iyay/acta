---
name: frame-no-brainstorming-status
description: Frame files the idea under the name given and asks the goal question first, without spending the session brainstorm on status brainstorming.
tags: [frame]
runs: 1
# The floor is three turns: read the frame skill, read the item, ask the goal
# question. Frame files the item with acta scratch new before it asks anything,
# and the acta commands come back with a git error in the eval sandbox, so the
# setup and the error need more than one spare turn. Seven.
max_turns: 7
timeout_seconds: 240
allowed_tools: []
---

I have an idea for a product. A bot that reads the standup notes people
already paste into the team chat and posts one list of what actually got
decided and what is still open. Everyone I mention it to says a summary bot
already exists, which is why I want to think it through before I build it.

It is already filed in this repo as SCR-0001, in .acta/scratch/eval-frame.md.

Follow the acta workflow for it, and stop at the first thing the workflow asks
me for.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
