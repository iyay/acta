---
id: DBT-0092
hash: ub7frls
parent: plans/2026-10-07-eval-skip-routing
---
# Review NOTEs: Eval skips routing Implementation Plan

- [ ] (low) internal/evalomp/case_test.go loads only plugin/evals, so nothing proves plugin/evals-routing loads under the omp loader; a case shape omp rejects there would only show in a paid omp run.
