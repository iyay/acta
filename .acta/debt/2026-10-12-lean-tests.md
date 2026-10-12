---
id: DBT-0111
hash: quafh1k
parent: plans/2026-10-12-lean-tests
---
# lean-tests follow-ups

- [ ] (medium) lean-tests eval: graders/no-extra-test-files.md uses `target: files`, and FACTS.md does not say that target sees files made with Bash; the run writes its tests with Bash, so the grader may pass every run. A `trace` pattern must accept a bare `test_age.py` (as in `ls tests` output) and must not match inside longer names like `latest_version.py`.
- [ ] (medium) lean-tests eval: graders/cli-wiring-only.md lists "the negative number rule" as a FAIL example, so a correct run whose one error test is `main(["-1"])` can fail the land gate eval.
- [ ] (low) internal/plugincheck/evals_test.go skips lean-tests/no-extra-test-files.md by name; phrase the rule as "a non-judgement case has no llm grader, and a judgement case has at least one".
- [ ] (low) internal/plugincheck/skill_slice_test.go pins "a layer above is wiring only", which matches two places in slice/SKILL.md, so reverting one copy stays green.
- [ ] (low) graders/age-three-cases.md has one comment line about 92 characters wide; the others wrap at about 78.
- [ ] (low) The removed phrases ("Every new function/method has a test" and the others) are pinned only in the file each came from, so they could come back in another skill unseen.
- [ ] (low) build/SKILL.md "Handle the report" gives the orchestrator no rule for the implementer's "Cases the plan did not name" lines.
- [ ] (low) Nothing pins the lean-tests grader files: deleting one, or reverting its text, keeps every Go test green.
- [ ] (low) Review deep lens still hunts adversarial inputs at trust boundaries, while tdd now writes only the ones the plan names; review should call a missing one a plan gap, not a code bug.
