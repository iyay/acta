---
name: routing-chat-04
description: A plain question about how webhooks retry should get a chat answer, no files or workflow.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: [Skill]
---

How do webhook retries usually work? If the receiver is down for an hour,
what keeps the sender from losing events or hammering it nonstop? Just
explain the common pattern, do not change anything in this repo.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
