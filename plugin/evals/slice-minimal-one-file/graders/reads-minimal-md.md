---
# guards: never the other one
# Rejects a run that never opens the minimal depth file. Without it the run
# has no minimal rules, so a minimal plan would be a guess. It also proves the
# run did read a depth file, so reading none cannot pass the no-full checks.
type: tool_used
tool: Read
input_match: 'slice/minimal\.md'
min: 1
---
