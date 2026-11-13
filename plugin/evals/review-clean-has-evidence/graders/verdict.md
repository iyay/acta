---
# guards: the orchestrator gives its own final verdict
type: regex
pattern: '(?:^|\n)[ \t]*(?:\*\*|__)?(?:Round verdict:[ \t]*)?(?:\*\*|__)?(?:CLEAN|BLOCKED)(?:\*\*|__)?[.!]?[ \t]*\n?[ \t]*$'
flags: i
target: last_message
---
