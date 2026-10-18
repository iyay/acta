---
name: review-understanding-followup
description: A short, differently worded answer to the landing question that misses the cap gets the real cause from the source, no grade on the wording, and at most one optional new scenario.
tags: [land, review]
runs: 1
# The run reads the source and answers once. Twelve turns leave room for the
# skill and a file read without letting one run eat the quota.
max_turns: 12
timeout_seconds: 420
allowed_tools: [Skill, Read, Bash]
---

Plan PLN-0001 is already landed on main in this folder; src/retry.py is the
code that landed. Your landing report, written with the acta land skill
(acta:land) because review_understanding is on, ended with this optional
question:

> The job fails for the sixth time in a row, so the next call is
> backoff(6). How long does it wait, and why?

My answer: about a minute, it keeps doubling.

Reply to my answer the way the acta land skill says, then stop.

The acta CLI is ./bin/acta, and git is ./bin/git. Use those two paths. The
plain names may run a different program in this sandbox.
