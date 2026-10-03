---
id: DBT-0070
hash: n4wuwrj
parent: plans/2026-10-03-output-style-and-lean
---
# Review NOTEs: Output Style, Lean Coding Guide, Session Start Diet Implementation Plan

- [ ] (medium) plugin/evals/style-short-answer has no proof it fails without the acta style: a plain model may already answer the git question with no opener, so a broken style could still pass. Run claude plugin eval --ablation with-without --case style-short-answer once, keep the replies, record the result in plugin/evals/FACTS.md, and make the prompt harder if both arms pass.
- [ ] (low) internal/plugincheck/skill_lean_test.go: nothing keeps ponytail's wording out of plugin/skills/lean/SKILL.md. MustNot is case-sensitive, so a "## Intensity" heading or "Ultra" passes, and the reworded sentences ("A report names a symptom", "say so in one line", "every file the change touches") are not pinned out, so they can come back with every test green.
