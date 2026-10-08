---
# Rejects a run that rewrote or deleted the shared repo file. The value must
# still be on its own line, however the agent edited around it.
type: regex
pattern: "^build_executor: dispatch$"
flags: m
target:
  source: file
  path: ".acta.yaml"
---
