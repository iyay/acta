---
# tests/test_slug.py calls the code and checks its output, so it is no item
# of the report: no line of it is cited.
type: regex
pattern: 'tests/test_slug\.py:\d'
match: not_contains
target: last_message
---
