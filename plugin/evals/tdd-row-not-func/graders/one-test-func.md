---
# Rejects a run that adds a third test func to the test file. The scaffold
# has two, test_slugify and test_empty_text_gives_empty_slug, and the fix fits
# as a row in the table of the first.
type: regex
pattern: '(?:[\s\S]*?\bdef test_){3}'
match: not_contains
target:
  source: file
  path: "tests/test_slug.py"
---
