---
name: brainstorm-related-topic
description: A related second topic stays in this session, gets its own item and reuses only the earlier decisions that still apply.
tags: [brainstorm]
runs: 1
max_turns: 8
timeout_seconds: 240
allowed_tools: []
---

Earlier in this session we shaped the billing page. Its scratch item is
SCR-0001, and its Log holds what we decided.

Now, in this same session, I want refunds: a customer can get money back for
an invoice they paid. It changes how invoices and payments work, so it is a big
change, and it builds on the billing work.

Follow the acta workflow for a change like this. Before your first question,
tell me which earlier billing decisions carry over to refunds. Then stop at the
first thing the workflow asks me for.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
