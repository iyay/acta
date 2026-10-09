---
id: BUG-0038
hash: yrz4tr9
started: "2026-10-09 15:27:55"
fixed_in: 7eb309f
finished: "2026-10-09 15:31:27"
---
# Activity tab keeps showing tasks that are already done

## Symptom
On the Activity tab, a plan and its task stay listed as under way after the work is finished. Seen in be-pmis: PLN-0017 (status done, tasks 1/1) and PLN-0018 both stay on Activity, while their tasks PLN-0017.01 and PLN-0018.01 show a check mark and 3/3.

## Root cause
internal/tui/model.go:408: inProgress returns true as soon as the task has Started, before it looks at Status. The started record from `acta tick --start` never goes away, so a task that reached done (or fixed) still counts as under way. activityRows (internal/tui/sidebar.go:195) then keeps the task and its plan head on Activity.

## Repro
1. In any repo, run `acta tick --start` on a task.
2. Tick every box of that task so its status becomes done.
3. Open the acta TUI and go to Activity: the task and its plan are still listed.

## Found in
main at bed0d1d (v0.1.45), found by reading the code after the user reported be-pmis.
