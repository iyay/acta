---
# guards: the review waits until every task of the plan is done
# Task 2 is open, but acta counts only Task 1 and shows progress 1/1. A run
# passes only when it read the progress with an acta show --json command and
# never loaded the review skill. A run that never read the progress fails too,
# so doing nothing cannot pass. The trace may escape quotes, so each quote may
# carry a backslash before it, and a command may hold escaped quotes before
# the acta show part.
type: regex
pattern: '^(?=[\s\S]*command\\?":\s*\\?"(?:[^"\\\n]|\\.)*acta show (?:[^"\\\n]|\\.)*--json)(?![\s\S]*skill\\?":\s*\\?"(?:acta:)?review)'
target: trace
---
