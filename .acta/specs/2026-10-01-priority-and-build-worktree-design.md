---
parent: scratch/2026-09-29-bug-debt-severity
closes: [SCR-0023]
id: SPC-0055
created: "2026-10-01 05:15:46"
hash: g3ul97k
started: "2026-10-01 05:27:35"
finished: "2026-10-01 05:52:33"
---
# Priority for bugs and debt items, and build always uses git worktree add

Status: design approved by the user in chat on 2026-10-01. Architectural for the priority part (file format, CLI and TUI). The build worktree part (SCR-0023) is Bounded and rides along, by the user's ruling.

## Why

- Bugs and debt items have no order of importance. With 23 bugs and 56 debt files, the user cannot see which ones to pick first (SCR-0011).
- `acta:build` tells the agent to prefer a native worktree tool. In Claude Code that is `EnterWorktree`. It puts the worktree inside the repo (`.claude/worktrees/`), which breaks the `../<repo>-<slug>` rule. It also starts from `origin/<default-branch>`. The user pushes by hand, so a spec and plan just committed on local main are often not on origin, and the worktree starts without the plan (SCR-0023).

## Part 1: build worktree

- In `plugin/skills/build/SKILL.md`, Step 1: remove "1a. Native Worktree Tools (preferred)".
- `git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"` becomes the only way. Drop the "Fallback" label and the "Only use this if Step 1a..." lines.
- Add one line on why: a native tool puts the worktree inside the repo and starts from origin, so it misses the spec and plan on local main.
- The directory order and the sandbox fallback stay as they are.

## Part 2: priority

### One field, three levels

- The field is `priority`. Levels: `high`, `medium`, `low`. No severity field.
- It is optional everywhere. An item without it is unset. Old files are not migrated.

### File format

- Bug: frontmatter line `priority: high|medium|low`.
- Debt item: a tag right after the checkbox: `- [ ] (high) text`, `(medium)` or `(low)`.
- The board reads both into a new `Item.Priority` field (`internal/board`). A debt item's `Title` does not include the tag.
- A bad bug value (for example `priority: urgent`) adds a line to the item's `Problems` and the bug counts as unset.
- A bad debt tag (for example `(hgh)`) is not a tag. It stays part of the text, because a note may really start with brackets.

### CLI

- `acta bug new <slug> --priority high|medium|low`. Optional. A bad value is refused and no file is written.
- `acta debt new <plan id>`: a stdin line that starts with `(high) `, `(medium) ` or `(low) ` keeps its tag. Other lines are stored as today.
- `acta set <id> priority high|medium|low|none`. The id is a bug or a debt item (for example `DBT-0055.01`). `none` removes the field or the tag. It auto-commits like every other `set`.
  - Bugs go through the new `priority` case in `SetValue` (`internal/write/ops.go`).
  - Debt items change their own line in the debt file, through a small helper next to `MarkItem` in `internal/write/mark.go`.
  - Any other kind is refused with `priority is only for bugs and debt items`.
- The `acta set` usage line names `priority`.

### Skills

- `plugin/skills/review/SKILL.md` (the line about `acta debt new`): a reviewer may start a NOTE with `(high)`, `(medium)` or `(low)`. No tag is fine.
- `plugin/skills/bug/SKILL.md`: names `--priority` as an optional flag.

### TUI

- Sort: a new `byPriority(items)` in `internal/tui/order.go` runs after `ordered()`. It is a stable sort: high, then medium, then low, then unset. Inside each group the id order stays, and `o` still flips old and new. It is used only in `openRows`, for the Bugs and Debts tabs. The Done pane (Fixed, Wontfix, Done) keeps plain id order.
- Row tag: in `rowText` (`internal/tui/scroll.go`) an `H`, `M` or `L` right after the id, before the title, for example `BUG-0023 H Scroll resets…`. No tag when unset. Colors come from the ANSI slots that already exist: H bright red, M yellow, L dim. No new theme field.
- Detail pane: a meta line `PRIORITY high` next to STATUS. No line when unset.
- Search and the other tabs do not change.

## Testing

Every test is written first and seen failing.

- `internal/plugincheck`: fails when `plugin/skills/build/SKILL.md` names `EnterWorktree` or "native worktree".
- `internal/board`: reads `priority:` on a bug and `(high)` on a debt line; the tag is gone from the Title; a bad bug value lands in Problems and counts as unset; `(hgh)` stays text.
- `internal/write`: `bug new --priority` with a good and a bad value; `debt new` keeps the tag; `set priority` on a bug, on a debt item, and with `none`; other kinds refused.
- `internal/tui`: sort order in Bugs and Debts with mixed levels and both `o` directions; Done pane order unchanged; the row tag; the PRIORITY meta line.

## Out of scope

- Priority on specs, plans, tasks or scratch items.
- A TUI key to set priority. `acta set` does it.
- Filling in priority for existing bugs and debt.
