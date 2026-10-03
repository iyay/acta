---
id: DBT-0069
hash: vf8q8mw
parent: plans/2026-10-03-rename-shape-slice
---
# Review NOTEs: Rename to Shape and Slice, Version Policy, Token Budgets Implementation Plan

- [ ] (low) internal/plugincheck/budget_test.go counts raw bytes, so a checkout with core.autocrlf=true (the Git for Windows default) turns every cap red (shape/SKILL.md reads 17743 bytes against its 17458 cap); TestDefaultRulesFile makes the same LF assumption. A .gitattributes with eol=lf would cover both.
- [ ] (low) Each skill now has two size caps, SkillRule.MaxLines in lines and fileCaps in bytes in internal/plugincheck/budget_test.go, so every lean-down commit has to lower both numbers. Drop MaxLines so one number rules.
