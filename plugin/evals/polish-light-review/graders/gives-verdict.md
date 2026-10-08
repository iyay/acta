---
# Rejects a run that stops with no verdict, or that never finishes the review.
# The light review ends in CLEAN or BLOCKED, said in the last reply.
type: regex
pattern: "CLEAN|BLOCKED"
target: last_message
---
