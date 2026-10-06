---
id: DBT-0050
hash: l2oztgv
parent: plans/2026-09-30-tui-date-times-and-toasts
---
# Review NOTEs: Times in the Detail Dates and Self-Hiding Toasts Implementation Plan

- [ ] (low) Setting the same status text again within toastFor starts no new timer, so the first timer clears the second toast early.
- [ ] (low) A hand-written unquoted created/started/finished value comes back from yaml as a time and shows as the day only.
- [ ] (low) isDate in internal/board/board.go writes out "2006-01-02 15:04:05" again instead of sharing write.stampLayout; the two can drift.
- [ ] (low) isDate accepts fractional seconds like "2026-09-30 16:14:05.5", because time.Parse allows them; the spec says exactly two shapes.
- [ ] (low) The eval fixture plugin/evals/answers-appended/scaffold.sh still writes a day-only created; still valid.
- [ ] (low) Watch and reload errors now hide after toastFor like every toast; manual mode still shows in the mode word.
- [ ] (low) No test checks that copyID starts no timer of its own; putting clearStatusAfter back there keeps the suite green.
- [ ] (low) No test fails when the after == before guard in Update is removed; the only effect would be extra timers.
- [ ] (low) Init now starts a timer for a status set before the first frame (the theme error); outside the plan task list but covered by a test.
