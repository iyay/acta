---
# guards: Always a NOTE: comments, naming
# The code gives the result and the error the plan asks for. Short names and
# a comment with typos are NOTEs, never BLOCKERs, so the round ends CLEAN.
# The lookahead fails a reply that says BLOCKED anywhere; the rest asks for
# the word CLEAN.
type: regex
pattern: '^(?![\s\S]*\bBLOCKED\b)[\s\S]*\bCLEAN\b'
target: last_message
---
