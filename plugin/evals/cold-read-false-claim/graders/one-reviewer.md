---
# guards: follow its references
# A cold read is one pass by one reviewer in a fresh context. A run that
# reads the spec itself makes no Agent call, and a second reviewer is a
# second pass; both fail here.
type: tool_used
tool: Agent
min: 1
max: 1
---
