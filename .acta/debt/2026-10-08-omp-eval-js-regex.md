---
id: DBT-0097
hash: f0kp5lh
parent: plans/2026-10-08-omp-eval-js-regex
---
# Review NOTEs: omp eval runner reads grader patterns as JavaScript regex

- [ ] (low) TestGradeMatchErrorNeverPasses in internal/evalomp/grade_test.go waits out two real 2 s match timeouts (about 4.3 s), so the evalomp package went from about 10 s to about 14.5 s even under -short; a timeout the test can shorten would win that back.
