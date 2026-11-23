---
# Rejects a plan that names a test func the repo does not have, such as
# test_drops_punctuation. The new case fits as a row in test_slugify, so no
# reason for a new func holds here. The pattern skips the two funcs that
# exist and the module name test_slug.
type: regex
pattern: '\btest_(?!slugify\b|empty_text_gives_empty_slug\b|slug\b)\w+'
match: not_contains
target:
  source: file
  path: ".acta/plans/2026-11-23-slug-punct.md"
---
