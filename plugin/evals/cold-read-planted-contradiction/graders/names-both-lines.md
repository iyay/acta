---
# guards: contradictions
# The Storage section keeps a finished archive for 30 days and the Cleanup
# section deletes it after 7. A cold read that found this names both numbers
# in what it tells the user. A reply that only says the spec looks fine, or
# names one of the two, fails. The two lookaheads take the numbers in any
# order.
type: regex
pattern: '^(?=[\s\S]*\b30 days\b)(?=[\s\S]*\b7 days\b)'
target: last_message
---
