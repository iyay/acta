---
parent: specs/2026-10-22-style-no-rule-cite-design
depth: minimal
id: PLN-0186
created: "2026-10-22 18:46:08"
hash: yxohzym
started: "2026-10-22 18:47:07"
finished: "2026-10-22 18:47:40"
---

# Style bans citing rules and skill paths

**Goal:** The acta output style tells the model never to cite a rule, skill or file path to justify a step unless asked why, in Claude Code and in Codex.
**Spec:** .acta/specs/2026-10-22-style-no-rule-cite-design.md
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The new style line is exactly `- Never cite a rule, skill or file path to justify a step unless asked why.` under `## Every reply`.
- The style body word cap stays 300. Every other check in `internal/plugincheck` keeps passing.

## Waves

- Wave 1: Task 1.

### Task 1: Add the no-cite line to the style

**Files:** Modify `plugin/output-styles/acta.md`, `internal/plugincheck/budget_test.go`; test in `internal/plugincheck/style_test.go`.
**Seam:** `styleParts` over `plugin/output-styles/acta.md`, checked by a new test in `style_test.go`.
**verify:** The style body holds the exact no-cite line, inside the `## Every reply` block (after that heading and before the ADHD heading), so both Claude Code and the Codex session note (`plugin.OutputStyle()`) carry it. The body stays at or under 300 words, the byte cap of `output-styles/acta.md` equals the new file size, and no other style line changes. List each check run.
- [x] Red: a new test in `style_test.go` requires the exact line inside the `## Every reply` block; run `scripts/test ./internal/plugincheck -run TestStyle` and see it fail because the line is missing.
- [x] Change: add the line as the last bullet of `## Every reply`; set the `output-styles/acta.md` cap in `budget_test.go` to the new byte size; run `scripts/test ./internal/plugincheck` and see it pass.
- [x] Commit: `feat(style): never cite rules or skill paths to justify a step`

## State

### Next

Task 1 committed. Round 1 review (spec + standards reviewers) running over 0c51151..24cd020; then polish if [fix] NOTEs, then acta:land.

## Review notes

Deep review, full tier (style text agents read), two reviewers; round 1 CLEAN, no BLOCKER, no fix or debt NOTEs, no fact-check needed. Notes left: the new line does not clash with the error-quote line or the slice verify-line rule; the test also fails if the line moves under the ADHD block.
