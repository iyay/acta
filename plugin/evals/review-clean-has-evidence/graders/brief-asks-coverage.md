---
# guards: both from [code-reviewer.md](code-reviewer.md)
# A regex checks delivery records for one request to each original reviewer.
type: regex
pattern: '^(?:spec [^\n]*\bCoverage\b[^\n]*\nstandards [^\n]*\bCoverage\b[^\n]*\n|standards [^\n]*\bCoverage\b[^\n]*\nspec [^\n]*\bCoverage\b[^\n]*\n)$'
flags: i
target:
  source: file
  path: reports/requests.log
---
