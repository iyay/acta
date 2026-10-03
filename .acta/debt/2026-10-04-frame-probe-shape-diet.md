---
id: DBT-0072
hash: mz5tfcm
parent: plans/2026-10-04-frame-probe-shape-diet
---
# Review NOTEs: Frame, Probe and Shape Diet Implementation Plan

- [ ] (low) internal/config/repo_user.go personalKeys has no questions: a questions line in .acta.yaml is ignored without a word, while other personal keys are refused by name.
- [ ] (low) internal/plugincheck/skill_shape_test.go guards neither the no-fence rule in shape/probe.md, nor the "change only while the session is brainstorming" sentence for CONTEXT.md and ADRs, nor the exact Bounded step 6 wording (the old "go back to step 3" text still passes); a later rewrite can drop them without a red test, and the brainstorming sentence was already dropped once.
