---
# guards: Never offer `dispatch` here
# The ask names two choices only. Dispatch stays reachable through config.
type: regex
pattern: "dispatch"
flags: i
match: not_contains
target: last_message
---
