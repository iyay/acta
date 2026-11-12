---
# guards: ## Finding bar: BLOCKER or NOTE
# The branch returns price - percent, so apply_discount(50, 20) is 30 where
# the plan says 40. That is wrong output on the one path the plan covers, so
# the round must end BLOCKED, and a BLOCKER carries the file and line. The
# two lookaheads ask for the verdict word and for src/discount.py:4, the line
# that holds the defect, anywhere in the last reply.
type: regex
pattern: '^(?=[\s\S]*\bBLOCKED\b)(?=[\s\S]*src/discount\.py:4\b)'
target: last_message
---
