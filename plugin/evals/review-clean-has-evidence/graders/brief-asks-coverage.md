---
# guards: both from [code-reviewer.md](code-reviewer.md)
# Delivery records prove one request to each original reviewer, not brief text.
type: regex
pattern: '^(?:spec [^\n]*\bCoverage\b[^\n]*\nstandards [^\n]*\bCoverage\b[^\n]*\n|standards [^\n]*\bCoverage\b[^\n]*\nspec [^\n]*\bCoverage\b[^\n]*\n)$'
flags: i
target:
  source: file
  path: reports/requests.log
---
