---
# Passes when the plan names the test file, the table test func test_slugify
# and a row. The three lookaheads may match anywhere in the plan.
type: regex
pattern: '^(?=[\s\S]*tests/test_slug\.py)(?=[\s\S]*\btest_slugify\b)(?=[\s\S]*\brows?\b)'
flags: i
target:
  source: file
  path: ".acta/plans/2026-11-23-slug-punct.md"
---
