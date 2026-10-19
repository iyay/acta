---
# guards: **Start the first task.**
# Rejects a run that runs the project's suite before its first task. The
# scaffold saved a stale result, and neither the stale record nor its absence
# makes setup run a suite.
type: tool_used
tool: Bash
input_match: '"command":"((?:[^"\\]|\\.)*[;&|](?:\s|\\[nrt]|\\\\)*)?(\./)?scripts/test(\s|"|$)'
min: 0
max: 0
---
