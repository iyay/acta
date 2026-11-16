---
# guards: follow its references
# Check the actual reviewer prompt for all named reading-evidence fields.
type: tool_used
tool: Agent
input_match: '^(?=[\s\S]*### Read)(?=[\s\S]*Spec sections:)(?=[\s\S]*Reference paths:)(?=[\s\S]*\bnone\b)'
min: 1
---
