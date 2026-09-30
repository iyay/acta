---
parent: bugs/2026-09-30-landed-plan-stays-approved
created: "2026-10-01 16:20:02"
id: SPC-0063
hash: xaknfg0
started: "2026-10-01 16:23:35"
finished: "2026-10-01 16:25:28"
---
# A plan's written approved or in-progress follows its task boxes

Status: design approved by the user in chat on 2026-10-01. Bounded (one switch in `internal/board/board.go`, plus its test).

## Why

A plan gets `status: approved` written into its frontmatter when the user approves it. The status switch in `internal/board/board.go` (around line 615-642) keeps any written status in its `default:` branch, so the plan stays `approved` after every task is ticked and the branch has landed. It hangs in the Open pane until someone runs `acta set plans/<stem> status done` by hand. Three plans needed that on 2026-09-30 (BUG-0013).

## Design

1. A new case in that switch, for `KindPlan` only: when the written status is `approved` or `in-progress`, the board ignores it and works the status out from the task boxes with `parentStatus`. No box started gives `approved`, some started gives `in-progress`, all ticked gives `done`. `StatusSource` is `derived`.
2. A written `draft`, `done` or `dropped` still wins. Each one is a choice a person made: not approved yet, closed by hand, or given up. The three plans set to `done` by hand stay as they are.
3. No "written status ignored" problem is added. Every approved plan carries `approved` in its file, so the warning would show on all of them.
4. No other code reads a plan's written `approved`. The only other `"approved"` is a derived value in `internal/board/closes.go:108`. So no build gate changes.
5. Specs, bugs, tasks and other kinds behave as today.

## Testing

A board test in `internal/board`, red before the code changes:
- plan written `approved`, 1 of 1 task ticked: status `done`, source `derived`
- plan written `approved`, 1 of 2 tasks ticked: status `in-progress`
- plan written `in-progress`, no task ticked: status `approved`
- plan written `draft`, and plan written `dropped`, every task ticked: status stays as written, source `frontmatter`

## Out of scope

Changing `acta:land` to clear the written status. Removing the written status from old plans.
