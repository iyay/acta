---
# guards: Never offer `dispatch` here
# With no picks yet, the run must not offer to make subagent the default.
type: regex
pattern: "(3|three) times"
flags: i
match: not_contains
target: last_message
---
