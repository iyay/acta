---
parent: specs/2026-10-28-tidy-branch-delete-no-confirm
depth: minimal
id: PLN-0198
created: "2026-10-28 08:28:31"
hash: y25xi9f
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

- [ ] Red: new test asserting the exception text fails because rule 5 has no exception
- [ ] Change: add the exception sentence to `coreRules`, regen `default-rules.md` with `-update`, write the test with exact-line negative match
- [ ] Commit: `fix(hook): exempt post-tidy branch delete from rule 5`

### Task 2: Tidy.md cleanup needs no confirmation

**Files:** Modify: `plugin/skills/land/tidy.md`

**Seam:** tidy.md step 7 prose.

**verify:** Step 7 keeps its proof clause and appends exactly "; per core rule 5 this delete needs no confirmation"; no other step gains an exception.

- [ ] Red: `TestBudgetFiles` fails because tidy.md grew past its cap
- [ ] Change: append the sentence to step 7 (cap bump stays in Task 3)
- [ ] Commit: `docs(land): tidy branch delete needs no confirmation`

### Task 3: Bump size caps to measured sizes

**Files:** Modify: `internal/plugincheck/budget_test.go`

**Seam:** `TestBudgetSessionStart` and `TestBudgetFiles` caps.

**verify:** Both caps equal the new measured byte sizes from the failing test output; no other cap moves.

- [ ] Red: both budget tests fail showing the new measured sizes
- [ ] Change: set `sessionStartCap` and the `skills/land/tidy.md` entry to those sizes
- [ ] Commit: `chore(plugincheck): bump caps for rule 5 exception`
