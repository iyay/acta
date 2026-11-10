---
# guards: Never offer `dispatch` here
# Rejects a run that ends without the subagent or inline question.
type: regex
pattern: 'subagent[\s\S]*inline|inline[\s\S]*subagent'
flags: i
target: last_message
---
