---
name: routing-chat-02
description: A plain question about Go slices should get a chat answer, no files or workflow.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

Quick Go question. When I append to a slice inside a function, the caller
sometimes sees the new item and sometimes does not. What decides that, and
what is the safe pattern? Just explain, do not change anything in this
repo.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
