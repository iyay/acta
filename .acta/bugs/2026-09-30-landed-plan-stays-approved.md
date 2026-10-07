---
id: BUG-0013
hash: ha43l97
fixed_in: f46ab4f
finished: "2026-10-01 18:25:09"
---
# A landed plan stays in the Open pane as approved

## Symptom
A plan whose tasks are all ticked and whose branch has landed still shows as `approved` in `acta list` and in the TUI Open pane. On 2026-09-30 three landed plans hung there: PLN-0031 (9/9, landed dc983b7), PLN-0032 (7/7, landed 6ea280f) and PLN-0042 (7/7, landed f4e5f48). Each needed `acta set plans/<stem> status done` by hand (ad2b3a4, ad2b3a4, ad2b3a4).

## Root cause
The plan was approved with `acta set plans/<stem> status approved`, which writes `status: approved` into the plan frontmatter (70f89de, 4124585, 9e4ca87). In internal/board/board.go the status switch (around line 600-620) derives a status only when the written status is empty; for a plan with a written status the `default:` branch keeps the written word, so ticked boxes never turn it into `done`. Specs already have the opposite rule (the `KindStory && it.plans > 0` case, SPEC-6). `acta:land` does not clear the written status either.

## Repro
In a scratch repo: write a plan with one task, run `acta set plans/<stem> status approved`, tick the task with `acta tick plans/<stem>#task-1 --all`, then `acta list --json`: the plan reports `status: approved`, `status_source: frontmatter`, progress 1/1.

## Found in
main ad2b3a4, while the user asked why so many plans hang in the Open pane; found by reading the board and git log.
