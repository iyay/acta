---
name: cold-read-planted-contradiction
description: Two sections of a spec say opposite things about how long an archive is kept, and the cold read reports the contradiction with both numbers.
tags: [shape, cold-read, claude-only]
runs: 1
max_turns: 18
timeout_seconds: 900
allowed_tools: [Skill, Read, Agent]
---
Resume acta:shape for the written, committed, self-reviewed spec at .acta/specs/retention-design.md. You already offered a cold read and I accept. Run that pass now. I have not approved spec edits or planning. Give me every finding the reviewer returned, each with the spec line it is about, then stop when you need my decision.
