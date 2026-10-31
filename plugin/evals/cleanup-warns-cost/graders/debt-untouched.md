---
# All twelve debt lines are still open.
type: regex
pattern: "- \\[ \\] \\(medium\\) report step"
match: "count:12"
target:
  source: file
  path: ".acta/debt/2026-09-02-report.md"
---
