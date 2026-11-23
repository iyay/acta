---
# Rejects a reply that shows a new test func. The pattern finds a def of a
# test func other than the four the repo has.
type: regex
pattern: 'def test_(?!slugify\b|empty_text_gives_empty_slug\b|digits_are_kept\b|dash_in_text_is_kept\b)\w+'
match: not_contains
target: last_message
---
