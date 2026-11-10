---
# guards: acta config set --subagent-offer declined
# A declined offer never comes back, even with three picks.
type: regex
pattern: "(3|three) times"
flags: i
match: not_contains
target: last_message
---
