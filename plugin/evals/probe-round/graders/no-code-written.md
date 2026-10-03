---
type: regex
pattern: "```|\\bfunc \\w+\\(|\\bconst \\w+ =|\\breturn \\w+\\("
match: not_contains
target: last_message
---

Rejects any run that writes code in the round, whether inside a fenced block or
as bare Go or SQL lines. Probe asks questions; the code belongs to the next
step.
