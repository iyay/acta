---
id: BUG-0050
hash: qxcsswn
status: fixed
finished: "2026-10-19 20:46:02"
started: "2026-10-19 20:30:17"
fixed_in: 13565823c944be839d2ffca73d17a4ca8102ba9e
---
# Slice self-review misses required checkboxes in minimal plans

## Symptom
PLN-0176 has no task checkboxes. Running
`acta tick plans/2026-10-19-drop-settings-skills#task-1 --all` exits 1 with
"bad input: the task has no checkboxes". Committing a task with its Task
trailer does not tick plan progress.

## Root cause
The generated plan `.acta/plans/2026-10-19-drop-settings-skills.md:28`
omits the three checkboxes required for every minimal-plan task by
`plugin/skills/slice/SKILL.md:25`: Red, Change and Commit.
`internal/write/tick.go:73` correctly rejects a task with no checkboxes.
The CLI failure is a plan-format error, not a limitation of minimal plans.
The slice workflow has a verification gap: `plugin/skills/slice/SKILL.md:27`
limits minimal-plan self-review to spec coverage and property-shaped verify
lines, without checking the required task checkboxes.

The planning session log shows that the agent loaded the checkbox rule at
20:01:56 +0700 on 2026-10-19, then wrote both split plans at 20:02:36
without checkbox markers, using plain `**Commit:**` fields instead.
The following `acta ready` calls checked dependencies, not task format.
This establishes an instruction-following failure that the slice review
did not catch; the agent's internal reason for omitting boxes is unknown.

The correction belongs in the slice self-review: check that every minimal
task has the required Red, Change and Commit checkboxes before hand-off.
Do not change the CLI to accept tasks without checkboxes.

## Repro
In an isolated clone at `193ac17`, build the CLI with
`go build -o /tmp/acta-tick-repro ./cmd/acta`. Then:

1. Run `/tmp/acta-tick-repro tick plans/2026-10-19-drop-settings-skills#task-1 --all`
   against the original plan. It exits 1 with the error above.
2. Add the required Red, Change and Commit checkboxes to task 1 in the
   isolated copy, without changing `depth: minimal`.
3. Run the same command. It exits 0 and reports task progress `3/3`.

The missing Task trailer warning is separate from checkbox progress and
does not change the successful exit code.

## Found in
Worktree drop-settings-skills, while running acta:build on PLN-0176.
Debug and review reproduced the failure and the checkbox control case in
an isolated clone using a CLI built from source. No CLI fix was needed
for the control case.
The bug was reopened after tracing the planning session and identifying
the missing format check in minimal-plan self-review.
