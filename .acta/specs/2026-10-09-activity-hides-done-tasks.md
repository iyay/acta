---
parent: bugs/2026-10-09-activity-shows-done-tasks
id: SPC-0116
created: "2026-10-09 15:11:06"
hash: mhzoii2
---
# Activity hides tasks that are done

Status: Bounded, approved by the user in chat on 2026-10-09.

Why: `inProgress` in `internal/tui/model.go:404` returns true as soon as a task has `Started`, before it looks at the status. The started record from `acta tick --start` stays forever, so a task at done still counts as under way. `activityRows` in `internal/tui/sidebar.go:195` then keeps the task and its plan on Activity. Seen in be-pmis with PLN-0017 and PLN-0018.

Design:
- `inProgress` returns false first when the status is `done` or `fixed`, then checks `Started` as before.
- The fix sits in `inProgress`, not in `activityRows`, because detail, scroll and model call it too. They all stop treating a finished task as under way.
- The started record on disk does not change. Only how the TUI reads it changes.
- A plan head on Activity already shows only while one of its tasks is under way, so it goes away by itself.
- Only `internal/tui/model.go` and one test file in `internal/tui` change. No version bump.
- Test: a task that has `Started` and status `done` is not in `activityRows`, and its plan head is gone. The test fails before the change and passes after it.
