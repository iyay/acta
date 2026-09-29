---
id: SPC-0010
hash: ju877gp
---
# Scratchpad kind and a five-pane sidebar

Status: design approved by the user in chat on 2026-09-28, section by section. Builds on `.acta/specs/2026-09-27-tui-polish-design.md`. This is spec 1 of 3. Spec 2 (skill rules: acta:scratch, brainstorm from a scratch item, one brainstorm per session, default executor in acta:setup) and spec 3 (harness: hook state, plugincheck guards, eval suite) come later and are out of scope here.

## Why

Raw ideas ("catet aja dulu") land in agent memory today, where the board cannot show them and a brainstorm cannot link them. The user wants a Scratchpad kind for them. With one more kind, the two-list sidebar gets crowded, so the user asked for a lazygit-style sidebar of five stacked panes, an Active pane, an expand key, and a counter that shows the selected item, not the scroll line.

## Non-goals

- No skill text changes (spec 2).
- No hooks, plugincheck guards or eval suite (spec 3).
- No newest-first sort, no themes, no drag-select copy, no Jev. These become scratch items (section 5).

## 1. Data

- New kind `KindScratch` in `internal/board/board.go`. Short ID prefix `SCRATCH` in `internal/board/ids.go`.
- Folder `Dirs.Scratch` in `internal/config/config.go`, default `scratch`, overridable in `.acta.yaml` like `Debt`.
- Statuses: `raw`, `brainstorming`, `specced`, `dropped`.
  - `raw`, `brainstorming` and `dropped` are written in the file's frontmatter.
  - `specced` is never written. It is derived (`StatusSource = "derived"`) when a spec names the item with `parent: scratch/<stem>`. A written `dropped` wins over the derived `specced`.
  - `Closed()` treats `specced` and `dropped` as closed.
- A spec's frontmatter `parent` is now read (today only plans and debt read it). A spec whose `parent` target does not exist gets the problem `parent scratch/<stem> not found`, the same way debt does. The scratch item lists the spec in `Children`, and its detail shows the SPEC-n.

## 2. CLI

- `acta scratch new <slug> [--title T] < body.md`
  - Writes `.acta/scratch/YYYY-MM-DD-<slug>.md` with frontmatter `id`, `title` (from the slug when `--title` is missing), `status: raw` and `created`. The body is stdin, kept exactly as given.
  - Follows `NewBug` in `internal/write/ops.go`: slug check, refuse an existing file, commit through `finish()` with the message `acta: new scratch <stem>`, and the item gets its SCRATCH number and hash.
  - Empty stdin fails with `scratch body is empty` and writes nothing. There is no editor fallback, because agents are the main callers.
- `acta scratch add <SCRATCH-n> < text.md`
  - Appends stdin to the end of the body after one blank line and commits `acta: add to scratch <stem>`, following `appendDebt`.
  - An unknown ID, an ID of another kind, or empty stdin fails with a clear error and leaves the file unchanged.
  - Any status can be appended to, `specced` and `dropped` included.
- `acta set SCRATCH-n status <value>` goes through `SetValue`. It accepts `raw`, `brainstorming` and `dropped`. `specced` fails with `specced comes from a spec's parent link`.
- `acta list` and `acta show` show scratch items. The usage lines in `internal/cli/cli.go` name `scratch new` and `scratch add`.

## 3. TUI

- **Sidebar table.** The fixed `pane` enum becomes a table of sidebar panes, each with a title and tabs. Five entries, top to bottom:
  1. `[1] Active`
  2. `[2] Specs ─ Scratchpad`
  3. `[3] Plans ─ Tasks`
  4. `[4] Bugs ─ Debt`
  5. `[5] Done`
  The detail pane sits outside the table, on the right. Scroll offsets are kept per pane, one per table entry plus the detail pane. `[` and `]` switch tabs in panes 2 to 5, as today.
- **Keys.** `1` to `5` focus the sidebar panes. `0` focuses the detail pane (lazygit's main view key; today detail is `3`). `tab` and `shift+tab` cycle all six boxes. The `?` popup lists the new keys.
- **Active pane.** Lists every item that is in progress, across all kinds, with no tabs. In progress means: a plan, spec or task with status `in-progress` or started through `acta tick --start`, a bug with status `fixing`, or a scratch item with status `brainstorming`. Items that are only `open`, `raw` or `todo` stay out. Rows use the SPEC-9 row rules, so they show `<done>/<total> · <agent>`.
- **Kind panes (2 to 4).** Show the open items of the current tab. In-progress rows keep the accent color. The in-progress divider from SPEC-9 is removed.
- **Done pane.** Follows the sidebar pane and tab that had focus last, the way pane [2] follows pane [1] today. For the Scratchpad tab its tabs are `Specced ─ Dropped`. When the last focused pane was Active, Done lists every closed item, in today's order.
- **Heights and expand.** By default the five sidebar panes share the height equally. `z` on the focused sidebar pane expands it: it takes the body height minus 3 lines for each other sidebar pane, and each other pane keeps 3 lines. Pressing `z` again, or moving the focus to another pane, restores equal heights. `z` on the detail pane does nothing. When the terminal is too short to give each other pane 3 lines, the other panes shrink to their title bar only.
- **Counter.** Every sidebar pane writes `<selected> of <total>` at the bottom right of its border, counting items and starting at 1, like lazygit's `8 of 236`. An empty pane shows `0 of 0`. The detail pane shows no counter and keeps its scrollbar. The old line counter `count()` in `internal/tui/scroll.go` is removed.

## 4. Testing

TDD for every task: a failing test first, watched failing, then the code.

- **Board:** scratch statuses parse; `specced` is derived from a spec's `parent`; a written `dropped` beats it; a missing parent gives the spec a problem; `Closed()` covers `specced` and `dropped`; `Dirs.Scratch` can be overridden.
- **Write and CLI:** `scratch new` refuses an empty body, a bad slug and an existing file, and on success commits and prints the ID; `scratch add` refuses an unknown ID, another kind's ID and empty stdin, leaving the file unchanged; `acta set` refuses `specced`; unicode bodies round-trip; tests that run write commands use a temp repo, because write commands commit.
- **TUI:** keys `1` to `5` and `0` focus the right box; `tab` and `shift+tab` cycle six boxes; Active holds in-progress items and never plain open ones; Done follows the last pane, including `Specced ─ Dropped`; `z` expands, keeps 3 lines for the others, restores on a second `z` and on a focus change, collapses the others to title bars on a short terminal, and does nothing on detail; the counter reads `3 of 20` and `0 of 0`, and detail has none; view tests cover a few terminal sizes.
- **Gates:** `gofmt`, `go vet` and `go test ./...` pass.

## 5. Moving the backlog

The last task, after the scratch kind works: each open idea from the agent memory note `tui-backlog-ideas-2026-09-27.md` becomes one scratch item through `acta scratch new`. The items are: newest-first sort, themes, drag-select auto-copy, the Jev spike, spec 2 (skill rules) and spec 3 (harness). Each body holds the idea's notes and rulings from the memory note, word for word. The memory note keeps its rulings and gains a pointer to each SCRATCH-n. This task runs in the worktree, and its commits come from `acta scratch new` itself.
