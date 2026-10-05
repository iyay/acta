---
id: DBT-0083
hash: o6cwbzx
parent: plans/2026-10-06-rounds-once-and-dispatch-agent
---
# Review NOTEs: Review rounds once per load and dispatch agent name Implementation Plan

- [ ] (low) The "*" default-agent key in .agents.json is skipped only in board/readAgents; any new reader of .agents.json must skip it too, or it shows "*" as a task.
- [ ] (low) RecordAgent now takes the file lock on every plain tick with no name before it finds nothing to write; before, such a tick never took the lock. Same result, slower ticks under contention.
