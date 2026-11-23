---
# Rejects a run that puts the new test in a new file, which is a new test func
# by another road. Each line of the list is one path. The pattern finds a path
# that is a test file (test_*.py or *_test.py) but not tests/test_slug.py.
type: regex
target: files
match: not_contains
pattern: '^(?!(?:.*/)?tests/test_slug\.py$)(?:.*/)?(?:test_[^/\n]*|[^/\n]*_test)\.py$'
flags: m
---
