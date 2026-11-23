---
# Passes when the test file holds a new row in the table: a pair whose input
# has a run of two or more spaces between words, such as
# ("Hello   World", "hello-world"). An extra assert such as
# assertEqual(slugify("a  b"), "a-b") does not match, since a row is a pair.
type: regex
pattern: '\(\s*["''][^"''\n]*\S {2,}\S[^"''\n]*["'']\s*,\s*["'']'
target:
  source: file
  path: "tests/test_slug.py"
---
