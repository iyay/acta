---
parent: specs/2026-10-18-setup-review-understanding-design
depth: minimal
id: PLN-0161
created: "2026-10-18 21:21:19"
hash: hcckefi
---
# Review understanding setup plan

**Goal:** Let setup configure the independent review understanding boolean immediately after review depth.
**Spec:** `.acta/specs/2026-10-18-setup-review-understanding-design.md`
**Tests:** `scripts/test` (fast), `scripts/test --full` (full, land only); task loop: `scripts/test ./internal/setup/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Use the existing wizard controls and config key; no new flag or dependency. Build inline in an isolated worktree, write repo content in English, and never push.
- Default off, retain saved booleans, save the selected boolean explicitly, and keep review depth independent. The optional scenario question follows successful landing and never blocks it.

## Waves

- Wave 1: Task 1.

### Task 1: Add the review understanding wizard question

**Files:** Modify `internal/setup/form.go`, `internal/setup/form_test.go`, and `internal/setup/form_text_test.go` if its option-copy assertions need updating.
**Seam:** `setup.FormDefaults`, `setup.FormGroups`, `setup.AnswersUser`, and `setup.AnsweredLines`; use the existing form test helpers to exercise selection.
**verify:** Every setup path shows Review understanding immediately after Review depth and before optional plugin install; unset/true/false defaults select off/on/off, changing either selection saves that explicit boolean independently of review depth, the summary reports it, and question counts stay correct with and without plugin install. Enumerate these paths in the tests; copy explains the optional question after successful landing.
- [ ] Red: Replace the no-question preservation test with defaults, selection, save, summary and independence checks, update order/count expectations, then observe `scripts/test ./internal/setup/` fail because the question is missing.
- [ ] Change: Add the boolean default and existing-style off/on control, save its answer, and pass `scripts/test ./internal/setup/`.
- [ ] Commit: `feat(setup): add review understanding question`.
