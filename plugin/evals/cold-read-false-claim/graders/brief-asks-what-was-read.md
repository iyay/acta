---
# guards: follow its references
# A reply that says "no findings" must show what the reviewer read, so the
# brief the orchestrator sends asks for the references it opened. Some Agent
# call must carry those words in its prompt. The match is a case-sensitive
# substring.
type: tool_used
tool: Agent
input_match: "the references you opened"
min: 1
---
