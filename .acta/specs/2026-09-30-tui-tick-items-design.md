---
created: "2026-09-30"
parent: scratch/2026-09-29-remove-drop-wontfix-all-kinds
id: SPC-0044
hash: boahydc
started: "2026-09-30"
---
# Mark tasks and debt lines done from the TUI, and a clearer key help

Status: design approved by the user in chat on 2026-09-30. Bounded: the tick flow (`internal/write/tick.go`, `internal/cli/tick.go`) and the TUI status popup already exist.

## Why

1. SCR-0018 asked for drop, wont-fix and remove on every kind. Most of it is already there. The `s` popup (`internal/tui/model.go:746`) lists every status of the kind, `dropped` and `wontfix` included, and `acta set <id> status` does the same from the CLI. The one gap is tasks and debt lines: the popup refuses a task (`model.go:737`), so the TUI cannot mark one done.
2. The `?` help (`internal/tui/view.go:59`) packs several keys on one line, like `t s n  set a value, new bug`. The user cannot tell what each key does.

## Design

1. **`+` marks the item done.** On a task row it ticks every box of the task, the same as `acta tick <id> --all`. On a debt line it ticks its one box, so the line is done.
2. **`-` puts the item back to open.** On a task row it clears every box of the task. On a debt line it clears its box. This works from done and from wontfix.
3. **CLI `acta tick <id> --undo`** does what `-` does. It is one more action flag, so it cannot be mixed with `--step`, `--all`, `--start` or `--wontfix`.
4. **One shared path.** The work `internal/cli/tick.go` does after a tick (the plan and spec dates in `markTaskDates`, the agent record) moves into `internal/write`, so the CLI and the TUI give the same file and the same dates. The TUI commits the change the same way its `s` popup does.
5. **Same guards as the `s` popup.** A row shown from a worktree, a legacy row, or a row that is not a task or a debt line gets a status line that says why, and nothing is written.
6. **Help, one key per line.** Every line of `helpLines` names one key, or one pair that are opposites (`j k`, `[ ]`), and says what it does in plain words. The new lines include:
   - `s   set the status; dropped and wontfix close an item`
   - `t   set the type: spec or bug`
   - `n   new bug`
   - `+   mark the task or debt line done`
   - `-   put the task or debt line back to open`

## Out of scope

- **Wontfix for a task.** The CLI refuses it on purpose (`internal/cli/tick.go:78`). Allowing it would change how progress, the dispatch gate and review count a task. A task that will not be done is edited out of its plan.
- **A remove or drop command.** `s` and `acta set ... status` already close any item, and the file stays in the repo as history.
- **Ticking one box from the TUI.** `acta tick <id> --step N` still does that.

## Testing

A failing test comes first for each part: `--undo` on a task and on a debt line in a temp repo; `+` and `-` in the TUI model change the file and leave the same dates the CLI leaves; each guard writes nothing; the help text has one key per line and a line for `+` and `-`. If the longer help no longer fits a small screen, the plan checks that it scrolls or fits.
