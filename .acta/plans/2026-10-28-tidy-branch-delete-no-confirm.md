---
parent: specs/2026-10-28-tidy-branch-delete-no-confirm
depth: minimal
id: PLN-0198
created: "2026-10-28 08:28:31"
hash: y25xi9f
started: "2026-10-28 08:30:25"
finished: "2026-10-28 08:32:37"
---
# Exempt post-tidy branch delete from destructive rule

**Goal:** Land with tidy history deletes the branch without asking.

**Spec:** .acta/specs/2026-10-28-tidy-branch-delete-no-confirm.md

**Tests:** scripts/test, scripts/test --full

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.

## Waves

Wave 1: Task 1, Task 2 (different files). Wave 2: Task 3 (bumps caps after both land).

### Task 1: Rule 5 exception in hook

**Files:** Modify: `internal/hook/hook.go`, `internal/hook/hook_test.go`; Regen: `plugin/hooks/default-rules.md`

**Seam:** `SessionStart(Input)` text and `coreRules` const.

**verify:** Rule 5 names exactly one exception (the tidy.md step 7 `-D` after its final plan check) on every session-note shape; the old bare line appears nowhere.

- [x] Red: new test asserting the exception text fails because rule 5 has no exception
- [x] Change: add the exception sentence to `coreRules`, regen `default-rules.md` with `-update`, write the test with exact-line negative match
- [x] Commit: `fix(hook): exempt post-tidy branch delete from rule 5`

### Task 2: Tidy.md cleanup needs no confirmation

**Files:** Modify: `plugin/skills/land/tidy.md`

**Seam:** tidy.md step 7 prose.

**verify:** Step 7 keeps its proof clause and appends exactly "; per core rule 5 this delete needs no confirmation"; no other step gains an exception.

- [x] Red: `TestBudgetFiles` fails because tidy.md grew past its cap
- [x] Change: append the sentence to step 7 (cap bump stays in Task 3)
- [x] Commit: `docs(land): tidy branch delete needs no confirmation`

### Task 3: Bump size caps to measured sizes

**Files:** Modify: `internal/plugincheck/budget_test.go`

**Seam:** `TestBudgetSessionStart` and `TestBudgetFiles` caps.

**verify:** Both caps equal the new measured byte sizes from the failing test output; no other cap moves.

- [x] Red: both budget tests fail showing the new measured sizes
- [x] Change: set `sessionStartCap` and the `skills/land/tidy.md` entry to those sizes
- [x] Commit: `chore(plugincheck): bump caps for rule 5 exception`

## Polish

### Task 4: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] Fix budget_test.go:80 comment: text grew 200 bytes, cap grew 160 (40 pre-existing slack)
- [x] Commit: `polish: review notes for PLN-0198`

## Review notes

- Round 1 CLEAN from both axes (spec + standards); polish fixed the cap-growth comment; tidy.md sentence pin filed as debt.

## State

### Findings

Eval flake 2026-10-28: land-tidy-progress + task-trailer red once on branch, green on main and on branch re-run, full re-run 43/43 score 1.0. Single-run reds are noise.

### Open rulings

Ruling: spec quoted rule 5 with backticks around git branch -D, but coreRules is a Go raw string so backticks break it. Shipped without backticks in hook.go, test, and regen file.
