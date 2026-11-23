---
id: BUG-0105
hash: eq1ksl3
---
# Eval review-clean-has-evidence is red on main

## Symptom
`scripts/eval --case review-clean-has-evidence` scores 0.33 on main (41905d2) and on branch test-rules: `brief-asks-coverage` (pattern not found in reports/requests.log) and `coverage-recovery` (judge FAIL FAIL FAIL) fail, `verdict` passes. Three runs in a row on 2026-11-24, two on the branch and one on main.

## Root cause
Not known yet. The case checks that the orchestrator sends a report with no `### Coverage` section back to the same reviewer once; the runs do not log that request.

## Repro
From the main checkout: build acta to a temp folder, put it first on PATH, run `scripts/eval --case review-clean-has-evidence --keep-temp`, then read reports/requests.log and out/trace.jsonl in the kept folder.

## Found in
main, land gate of PLN-0252 (test-rules), control run on main.
