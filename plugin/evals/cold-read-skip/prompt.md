---
name: cold-read-skip
description: Cold read skip preserves the user decision and spec approval gates.
tags: [shape, cold-read]
runs: 1
max_turns: 8
timeout_seconds: 900
allowed_tools: [Skill, Read]
---
Resume acta:shape for .acta/specs/export-design.md after self-review. You already offered a cold read and I decline it. I have not approved the current spec for planning. Do the next step and stop.
