---
# guards: `subagent_picks` is 3 or more and `subagent_offer` is not `declined`
# Rejects a run that ends without the offer to build parallel plans with
# subagents by default after three picks.
type: regex
pattern: '(3|three) times[\s\S]*default|default[\s\S]*(3|three) times'
flags: i
target: last_message
---
