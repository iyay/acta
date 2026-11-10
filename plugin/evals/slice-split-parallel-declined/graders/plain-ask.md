---
# guards: acta config set --subagent-offer declined
# Rejects a run that ends without the plain subagent or inline question.
type: regex
pattern: 'subagent[\s\S]*inline|inline[\s\S]*subagent'
flags: i
target: last_message
---
