---
id: DBT-0079
hash: ttb5ray
parent: plans/2026-10-05-implementer-reply-not-report-file
---
# Review NOTEs: Implementer reply instead of report file Implementation Plan

- [ ] (low) The implementer-prompt.md cap in internal/plugincheck/budget_test.go is still 7518 while the file is 7099, so the template can grow back by 419 bytes unnoticed.
- [ ] (low) TestImplementerPromptNoReportFile reads only implementer-prompt.md and matches exact phrases, so a report-file step added to build/SKILL.md or house-rules.md, or reworded ("save your report in a file"), passes.
