---
# guards: `parallel_plans: subagent` or `build_executor: subagent`
# With the setting on, no subagent or inline question is asked.
type: regex
pattern: "\\bor inline\\b"
flags: i
match: not_contains
target: last_message
---
