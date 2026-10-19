---
# Rejects a run that never runs a probe holding the workers at a barrier.
# Repeat runs and a lock control can name the cause, but only a barrier
# between the read and the write forces the lost update on every run.
type: tool_used
tool: Bash
input_match: "Barrier"
min: 1
---
