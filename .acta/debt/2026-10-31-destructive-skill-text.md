---
id: DBT-0158
hash: znagyz9
parent: plans/2026-10-31-destructive-skill-text
---
# Review NOTEs: Test output and cleanup skill text implementation plan

- [ ] (medium) With the subagent executor the implementer commits before the orchestrator sees untracked test output, and plugin/skills/build/implementer-prompt.md has no line telling it to report such files back, so the build-time .gitignore offer never happens; land step 1 catches them later.
- [ ] (medium) The state-resume grader for the test folder has never gone red in a real run: in the one run the agent ran no tests and had no reason to clean, so the eval does not yet prove the skill text changes behavior.
