---
# guards: Verify each finding against the code before acting
# With its one BLOCKER rejected the round is CLEAN, and the last reply says so.
type: regex
pattern: 'CLEAN'
target: last_message
---
