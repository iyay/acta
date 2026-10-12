---
# guards: Every case the plan names has a test, and no other
# Rejects a run that makes a third test file. The plan names tests/test_age.py
# and tests/test_cli.py, and the two graders for those files cannot see any
# other file. Each line of the list is one path. The pattern finds a path
# that is a test file (test_*.py or *_test.py) but not one of the two.
type: regex
target: files
match: not_contains
pattern: '^(?!(?:.*/)?tests/test_(?:age|cli)\.py$)(?:.*/)?(?:test_[^/\n]*|[^/\n]*_test)\.py$'
flags: m
---

Rejects a run that creates a test file beyond the two the plan names, such as
tests/test_edge.py, tests/test_age_extra.py or src/age_test.py.
