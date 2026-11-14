---
id: DBT-0168
hash: geg2bx3
parent: plans/2026-11-12-cold-read-evidence
started: "2026-11-14 06:29:59"
---
# Review NOTEs: Cold read findings carry their evidence

- [ ] (medium) Both new cold read eval prompts ask the run to give every finding with its spec line, which is the very rule the Resolve text now states, so the cases cannot tell the old Resolve from the new one and a later change that lets findings drop would stay green. A case that plants six findings and checks the user sees all of them, or the cut line, would measure it.
- [ ] (medium) cold-read-planted-contradiction passes a reply that quotes both numbers without calling them a contradiction, and fails a right reply that writes 30-day and 7-day. A sound check needs a contradiction word in the pattern or a judge grader, and a new old-arm run.
- [ ] (medium) The cold read brief asks for a Read part with no fixed form, so a prose line or an empty "Read: all of it" passes, and the orchestrator has no sure way to tell when the part is missing.
- [ ] (medium) Six eval graders still say in their comment that input_match is a case-sensitive substring, while plugin/evals/FACTS.md says it is a JavaScript regex over the tool input: polish-full-review (2), review-auto-signal (2), review-clean-has-evidence, review-depth-floor. A later edit that adds a dot or a bracket to one of those patterns would match more than its comment says.
