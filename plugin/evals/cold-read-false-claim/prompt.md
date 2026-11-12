---
name: cold-read-false-claim
description: A spec states a fact about a file in the repo that the file does not hold, and the cold read reports the wrong claim with what the file really says.
tags: [shape, cold-read, claude-only]
runs: 1
max_turns: 18
timeout_seconds: 900
allowed_tools: [Skill, Read, Agent]
---
Resume acta:shape for the written, committed, self-reviewed spec at .acta/specs/size-design.md. You already offered a cold read and I accept. Run that pass now. I have not approved spec edits or planning. Give me every finding the reviewer returned, each with the spec line it is about, then stop when you need my decision.
