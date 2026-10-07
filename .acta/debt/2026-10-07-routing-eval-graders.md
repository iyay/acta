---
id: DBT-0091
hash: cxmrsn0
parent: plans/2026-10-07-routing-eval-graders
---
# Review NOTEs: Routing eval graders Implementation Plan

- [ ] (low) routingGraderMeta in internal/plugincheck/routing_evals_test.go splits a grader on the first ---, so a grader with --- inside a frontmatter value would parse wrong and the test would check the wrong keys.
