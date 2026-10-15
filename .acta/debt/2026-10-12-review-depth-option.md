---
id: DBT-0110
hash: gp541cz
parent: plans/2026-10-12-review-depth-option
---
# Review NOTEs: Review depth option

- [ ] (medium) plugin/skills/review/depth.md: "A light-tier diff keeps the inline light review" is not limited to round 1 and polish; an agent in a light-tier fix round could skip the two reviewers the fix-round floor needs. Say "(round 1 and polish only)" in the quick, standard and deep bullets.
- [ ] (medium) plugin/skills/review/depth.md deep bullet: "It runs on the reviewer model" and "With no BLOCKER it does not run" now follow "The orchestrator numbers the BLOCKERs", so "It" can read as the orchestrator. Name the fact-check; say "A BLOCKER the fact-check moves becomes a NOTE".
- [-] (low) plugin/skills/setup/SKILL.md description lists plan depth but not review depth, so an agent asked to change review depth may not load the setup skill.
- [-] (low) internal/config/repo_user.go: a non-string review_depth or plan_depth in .acta.yaml (for example 3) is skipped silently instead of refused, so a typo there leaves the user on the default with no message.
- [-] (low) internal/setup/form.go Ask: no test proves ReviewDepth (or PlanDepth) is copied into Answers.User; dropping that line keeps every test green. Needs a test seam for the interactive form.
