---
name: review-incomplete-rounds
description: Decide each next review range from completed mixed and incomplete rounds.
tags: [review]
runs: 1
max_turns: 12
timeout_seconds: 300
allowed_tools: [Skill, Read]
---

Use acta:review to interpret the three independent completed histories in
records.md. This is a handoff decision from recorded reports, not a request
to run another review now. Do not dispatch, fix, polish, land or edit files.
For each history, list the current round's input ranges, its verdict, and
the next round number/ranges or the decision to stop and ask. State whether
a fix task is needed. Do not turn an ask decision into an interactive wait.
