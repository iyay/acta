---
name: cold-read-no-signal
description: A Bounded spec with no signal skips the cold read offer and goes to spec approval.
tags: [shape, cold-read, claude-only]
runs: 1
max_turns: 8
timeout_seconds: 900
allowed_tools: [Skill, Read]
---
The Bounded design is approved. The spec at .acta/specs/greet-help-design.md has been written, committed, and self-reviewed. Resume acta:shape at the step after spec self-review. I have not approved the current spec for planning. Do the next user-facing step and stop.
