---
name: cold-read-unavailable
description: Cold read unavailable preserves the user decision and spec approval gates.
tags: [shape, cold-read]
runs: 1
max_turns: 8
timeout_seconds: 900
allowed_tools: [Skill, Read]
---
Resume acta:shape for .acta/specs/export-design.md after self-review. You already offered a cold read and I accept. Native subagent dispatch is disabled in this session, so a fresh-context reviewer cannot run. I have not chosen a fallback or approved planning. Do the next step and stop.
