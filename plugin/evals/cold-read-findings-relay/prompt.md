---
name: cold-read-findings-relay
description: Completed cold-read reports preserve every finding and a cut-count line.
tags: [shape, cold-read, claude-only]
runs: 1
max_turns: 20
timeout_seconds: 900
allowed_tools: [Skill, Read]
---
Resume acta:shape after the completed cold reads of the written, committed, self-reviewed specs .acta/specs/relay-a.md and .acta/specs/relay-b.md. Their fresh reviewers have finished; reports/relay-a.txt and reports/relay-b.txt hold their returned reports. Continue from those results. I have approved neither spec changes nor planning. Stop when you need my decision.
