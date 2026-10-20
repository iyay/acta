---
parent: debt/2026-10-20-close-without-fast-suite
closes: [DBT-0136.01]
depth: minimal
id: PLN-0183
created: "2026-10-20 08:27:35"
hash: kf2mku2
---

# Keep the broad suite out of the dispatch brief gates

**Goal:** Stop the dispatch brief from printing the plan's fast and full test commands, so a dispatched build runs only each ticket's focused tests.
**Spec:** none (Bounded, approved in chat on 2026-10-20)
**Tests:** fast `scripts/test ./internal/cli -run 'TestBuildBrief'`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Keep the brief's `**Tests:**` check: a plan without that line still fails, so slice's contract holds.
- Change only the GATES line and its test. No skill text, no new flag, no change to other brief lines.

## Waves

- Wave 1: Task 1.

### Task 1: Name focused tests in GATES, never the plan test commands

**Files:** Modify `internal/cli/dispatch_brief.go`; update `internal/cli/dispatch_brief_test.go`.
**Seam:** the text `buildBrief` returns.
**verify:** No brief prints the plan's `**Tests:**` text, fast or full command, in any round: first dispatch, fix round, polish. GATES names each ticket's own focused tests and verify line instead; a plan with no `**Tests:**` line still fails to build a brief. List the paths checked: the three round kinds plus the missing-line error.
- [ ] Red: `TestBuildBriefFirstDispatch` demands the new GATES wording and forbids `scripts/test ./internal/cli` and `scripts/test --full`; it fails because the line still carries the plan's tests text.
- [ ] Change: build the GATES line without `marks.tests`, naming each ticket's focused tests and verify line. Keep the missing-`**Tests:**` error.
- [ ] Commit: `fix(dispatch): keep the broad suite out of the brief gates`
