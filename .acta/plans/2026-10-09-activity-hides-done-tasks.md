---
parent: bugs/2026-10-09-activity-shows-done-tasks
depth: minimal
id: PLN-0125
created: "2026-10-09 15:27:09"
hash: k43mlct
started: "2026-10-09 15:27:55"
finished: "2026-10-09 15:28:55"
---
# Activity hides tasks that are done

**Goal:** A task that is done or fixed never counts as under way, even when it has a started record.

**Spec:** `.acta/specs/2026-10-09-activity-hides-done-tasks.md`

**Tests:** fast `scripts/test ./internal/tui`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Only `internal/tui/model.go` and `internal/tui/sidebar_test.go` change. The started record on disk does not change. No version bump.

## Waves

- Wave 1: Task 1

### Task 1: inProgress ignores Started once a task is done

**Files:**
- Modify: `internal/tui/model.go` (`inProgress`)
- Test: `internal/tui/sidebar_test.go`

**verify:** No item with status `done` or `fixed` is ever reported as under way, whatever its started record says. List every caller of `inProgress` and say what each now shows for such an item.

- [x] **Step 1: Write the failing test** `TestActivityDropsStartedDoneTask` in `internal/tui/sidebar_test.go`: build a model from the fixture, pick a task item, set `Started = true` and `Status = "done"`, then check that `inProgress` returns false, that its id is not in `activityTaskIDs(m)`, and that its plan id is not a depth-0 row of `m.rowsOf(paneList)` when no other task of that plan is under way. Repeat once with `Status = "fixed"`.
- [x] **Step 2: Run it and watch it fail:** `scripts/test ./internal/tui -run TestActivityDropsStartedDoneTask`
- [x] **Step 3: Minimal code:** at the top of `inProgress`, after the nil check, return false when `it.Status` is `done` or `fixed`.
- [x] **Step 4: Run it and watch it pass:** `scripts/test ./internal/tui`
- [x] **Step 5: Commit** with gofmt and `go vet ./internal/tui` clean: `fix(tui): activity hides tasks that are done`
