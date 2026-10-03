---
name: probe-round
description: Grill me on a short design produces one round of at most five questions, each with a recommended answer, and no code.
tags: [shape]
runs: 1
# The floor is three turns: read the shape skill, read probe, ask the round.
# Shape calls this Architectural because there is no existing flow, so step 0
# runs acta scratch new and acta set before any question is asked, and the
# acta commands come back with a git error in the eval sandbox. That setup
# and that error need more than one spare turn, so seven.
max_turns: 7
timeout_seconds: 240
allowed_tools: []
---

Design: a rate limiter for one API. Every caller carries a key. The count
lives in Postgres so a restart does not hand a caller a fresh budget. Over the
limit, requests queue instead of failing, and the user sees their position.

Grill me on this. Follow the acta workflow, and stop after your first round of
questions. Do not write code.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
