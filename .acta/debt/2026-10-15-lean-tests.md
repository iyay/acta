---
id: DBT-0117
hash: g7e1bme
parent: plans/2026-10-15-lean-tests
---
# Review NOTEs: Lean tests Implementation Plan

- [-] (low) slice self-reviews (minimal and full) do not check that every task names a Seam, so a plan without one passes self-review.
- [-] (low) tdd/SKILL.md:215 says never private code but drops the spec's 'or internal state'; a test can go through the seam and assert on unexported fields.
