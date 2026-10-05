---
# Rejects a run that hands the question back to the user. The pattern wants a
# question mark on the same line as a word about standing or progress, so a
# plain sentence that uses those words does not trip it. It is checked against
# the example strings in the report of the task that wrote this case.
type: regex
pattern: "(?:status|progress|how far|catch me up|update me on|where (?:things )?stand|where (?:are )?(?:we|you)\\b|what(?:'s| is) (?:left|next|remaining)|what remains)[^\\n.?!]*\\?"
match: not_contains
target: last_message
---