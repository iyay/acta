---
name: cleanup-warns-cost
description: With more than 10 cleanup candidates, the agent says the count and asks how much to check before any subagent starts.
tags: [cleanup, claude-only]
runs: 1
# acta cleanup, acta cleanup --json, then one question. A run that starts
# subagents fails, so the cap only needs room for the listing.
max_turns: 14
timeout_seconds: 600
allowed_tools: [Skill, Read, Agent]
---

Clean up this repo's planning items: scratch, bugs, debt and specs that are
done, stale or no longer valid. Follow the repo's cleanup workflow.

I am away from the keyboard. When you need an answer from me, ask and stop.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
