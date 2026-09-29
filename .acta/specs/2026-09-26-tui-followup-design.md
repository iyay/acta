---
id: SPC-0007
hash: fed3wwe
---
# TUI follow-up: status line, started tasks, in-progress split

Status: approved by the user on 2026-09-26 (notes 15 to 19, given while PLAN-12 was being built). Built right after the TUI panes work (`.pm/specs/2026-09-26-tui-panes-design.md`, PLAN-12) lands, and before the review debt work.

## Why

While PLAN-12 was being built the user asked for five more changes. They were not added to PLAN-12 because its plan file was being ticked by the recipient at the time. They live here so they show on the board.

## 1. Bottom line

- Left side: only `? help`. The full key list lives in the `?` popup. Search box and status messages still take the left side while active, as today.
- Right side, lazygit style: `<project> · live · <YYYY-MM-DD HH:MM> | Donate  Feedback  <version>`.
  - No `pmb` word.
  - `live` becomes `paused` when file watching is off.
  - The date and time update every minute.
  - `Donate` links to the URL in config key `links.donate` (the user will use Ko-fi). With no URL set, `Donate` is not shown.
  - `Feedback` links to config key `links.feedback`, default `https://github.com/iyay/acta/issues`.
  - `<version>` comes from the Go build info (`v0.3.0` for a tagged `go install`), `dev` for a local build.
  - A click on `Donate` or `Feedback` opens the URL in the default browser (`open` on macOS, `xdg-open` on Linux). The text is also written as a terminal hyperlink (OSC 8), so cmd+click works in terminals that support it.
- Narrow window: the right side drops parts from the left end first (project, then status), and always keeps the date and time.

## 2. `pmb tick <id> --start`

- Ticks no box. Writes a `started` record for the task to `<root>/.agents.json`: `{"<task id>": {"agent": "<name>", "at": "...", "started": true}}`, with the agent from `--agent` or the environment, like other ticks.
- The board shows a task with a `started` record and no ticked box as `doing`, and its plan or spec as in progress.
- `--start` cannot be combined with `--step` or `--all` (exit 1).
- A later normal tick keeps the record; when the task is done the agent no longer shows (as in PLAN-12).
- The build skill, the implementer prompt and the dispatch delivery tell each implementer to run `pmb tick <id> --start --agent <name>` as the very first action of a task, before writing the failing test.

## 3. In progress and not started in pane [1]

- In every tab, items in progress come first, then items not started.
  - In progress: `doing`, `in-progress`, `fixing`, or a task with a `started` record.
  - Not started: `todo`, `draft`, `approved`, `open`.
- No section labels. In-progress rows use the accent text color; not-started rows use normal text.
- One dim divider line sits between the two groups. It is hidden when either group is empty.
- `j`/`k`, `g`/`G` and mouse clicks move over both groups; the divider cannot be selected.

## 4. Enter opens the detail, `e` edits

- `enter` on a row in pane `[1]` or `[2]` moves focus to pane `[3]` (detail) with that item, so it can be read and scrolled. It no longer opens the editor.
- `enter` on the legacy group row still opens or closes the group.
- `e` opens the selected item in the editor (what `enter` did before), from pane `[1]`, `[2]` or `[3]`. Items not on disk still get the "not checked out" message.
- `esc` in pane `[3]` returns focus to the list pane the item came from.
- The `?` popup lists `enter`, `e` and `esc` with their new meaning.

## Testing

- Bottom line: left shows only `? help`; right shows project, `live`/`paused`, date and time from a fixed clock, `Feedback`, version; `Donate` only when `links.donate` is set; narrow window keeps the date and time; clicking `Donate`/`Feedback` calls the opener with the right URL (opener injected in tests).
- Version: build info with a tag gives the tag; no build info gives `dev`.
- `--start`: writes a started record and ticks no box; the task shows `doing`; combined with `--step` or `--all` exits 1; plugin text in build, implementer prompt and dispatch names `--start` as the first action, guarded per file.
- Pane [1]: in-progress rows come first in the accent color, a divider follows, not-started rows after; no divider when one group is empty; the cursor skips the divider.
- Keys: `enter` on a row focuses `[3]` with that item and opens no editor; `enter` on the group row still toggles it; `e` opens the editor from each pane; `esc` in `[3]` returns to the list pane it came from; the help popup shows the new meanings.
