---
# guards: the orchestrator gives its own final verdict
type: regex
pattern: '^(?=[\s\S]*\b(?:reject|dismiss|drop|false))[\s\S]*(?:^|\n)[ \t]*(?:\*\*|__)?Round verdict:[ \t]*(?:\*\*|__)?CLEAN(?:\*\*|__)?[.!]?[ \t]*\n?[ \t]*$'
flags: i
target: last_message
---
