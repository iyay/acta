---
# guards: Verify each finding against the code before acting
# The BLOCKER quotes `return price - percent`, but the file holds
# `return price * (100 - percent) / 100`, and 50 with 20 percent off is 40 as
# the plan says. A finding the code does not show must not become a fix task,
# so the plan must hold no Fix round section after the run.
type: regex
pattern: '^(?![\s\S]*## Fix round)'
target:
  source: file
  path: ".acta/plans/2026-11-12-discount.md"
---
