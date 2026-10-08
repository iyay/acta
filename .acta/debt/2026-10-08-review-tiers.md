---
id: DBT-0100
hash: dri7bqm
parent: plans/2026-10-08-review-tiers
---
# Review NOTEs: Review depth follows what the diff changes

- [ ] (low) polish-full-review grader matches reviewer briefs with input_match Axis; FACTS.md does not say whether that match is case-sensitive or a substring, so an orchestrator that rewrites the brief as "spec axis" could fail a run that did dispatch two reviewers. Check the match rule in a kept-temp run and note it in FACTS.md.
- [ ] (low) build/dispatch.md shrank from 6477 to 6445 bytes but its byte cap in budget_test.go stays at 6477, so the saving is not locked in.
- [ ] (medium) polish-full-review graders count Spec and Standards with min 1 each, so one Agent call naming both words, or a Standards reviewer whose brief pastes the plan text with "**Spec:** none", passes both. Add a third tool_used grader on Agent with min: 2 and no input_match.
- [ ] (low) review SKILL.md: "Reading the diff yourself never replaces the two reviewers" names no tier; read alone it could clash with the light review. Start it with "In a full review,".
