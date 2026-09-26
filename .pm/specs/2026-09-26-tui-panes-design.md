---
status: done
id: SPEC-5
hash: jtlx
---
# TUI: three panes, lazygit-style tabs, agent tags

Status: approved by the user on 2026-09-26. Builds on `.pm/specs/2026-09-26-tui-design.md` (the current TUI) and `.pm/specs/2026-09-26-short-ids-design.md` (short IDs, built first).

## Why

The user tried the TUI and wrote 13 notes (2026-09-26). The sidebar is too wide and too tight, finished work is hidden behind a toggle, the tab row wastes a line, focus is invisible, the detail pane cannot scroll, help is a line of text, and nothing says which agent is on a task. This spec turns those notes into one layout.

## Goals

- Three panes with visible focus: `[1]` open items, `[2]` finished items, `[3]` detail.
- Tabs live in the pane border title, lazygit style.
- Specs and plans are separate tabs.
- Rows breathe (two lines each) and show who is working.
- One status line at the bottom right, with the time.
- `?` opens a help popup.

## Non-goals

- No WebUI now. Everything the TUI shows comes from files and `pmb`, so a WebUI can read the same data later.
- No change to how items are edited (`enter`, `t`, `s`, `n` keep working).
- No herdr dependency. herdr may fill agent data later; the TUI never calls it.

## 1. Layout

```
╭─[1]─Specs ─ Plans ─ Tasks ─ Bugs──╮╭─[3]─Detail────────────────────────────╮
│ PLAN-12  Short IDs                ││ ID        : PLAN-12 · plans/2026-...  │
│   in-progress · 2/4 · omp         ││ PLAN      : Short IDs                 │
│                                   ││ STATUS    : in-progress               │
│ PLAN-13  TUI panes                ││ WORKTREE  : short-ids                 │
│   approved · 0/6                  ││ FILE      : ../pm-board-short-ids/... │
╰───────────────────────────────────╯│ TASKS     : 2/4                       │
╭─[2]─Done ─ Dropped────────────────╮│                                       │
│ PLAN-11  Dispatch new/goal        ││ (body, scrolls when focused)          │
│   done · 1/1                      ││                                       │
╰───────────────────────────────────╯╰───────────────────────────────────────╯
 tab pane · ] [ tab · 1 2 3 · j k · enter edit · ? help      pmb · pm-board · live · 20:46
```

- Left column: pane `[1]` on top, pane `[2]` below. `[1]` gets about two thirds of the height, `[2]` one third, each at least 5 rows.
- Left column width: 30% of the screen, at least 28 and at most 48 columns. Pane `[3]` takes the rest.
- No separate tab row and no header line at the top. The panes start on the first line.
- Below 60 columns wide, only the focused pane shows, full width, so a narrow terminal still works.

## 2. Tabs

- Pane `[1]` title: `[1]─Specs ─ Plans ─ Tasks ─ Bugs`. The active tab is bold in the accent color, the others dim.
- Pane `[2]` title follows pane `[1]`'s tab: `Done ─ Dropped` for Specs, Plans and Tasks (Tasks has only Done, so its title is `Done`), `Fixed ─ Wontfix` for Bugs.
- Pane `[3]` title: `[3]─Detail`.
- What each list holds:

| Tab | Pane [1] (open) | Pane [2] tab 1 | Pane [2] tab 2 |
|---|---|---|---|
| Specs | spec files, status draft / approved / in-progress | done | dropped |
| Plans | every plan file, status not closed | done | dropped |
| Tasks | tasks todo / doing | done | — |
| Bugs | open / fixing | fixed | wontfix |

- A plan is its own row in Plans whether or not it has a spec. Its detail shows `SPEC : SPEC-4 · <title>` when it has one. This needs the board to keep plans as items (today a plan with a spec is folded into the spec); that change is part of this work.
- Pane `[2]` sorts newest finished first: by the date of the last commit that touched the file, falling back to the file-name date.

## 3. Keys

| Key | Does |
|---|---|
| `tab` / `shift+tab` | focus next / previous pane: `[1]` → `[2]` → `[3]` |
| `1` `2` `3` | focus that pane |
| `]` `[` | next / previous tab inside the focused pane (`[1]`: Specs/Plans/Tasks/Bugs; `[2]`: its two tabs; `[3]`: nothing) |
| `j` `k`, `g` `G`, `ctrl+d` `ctrl+u` | move in `[1]`/`[2]`; scroll in `[3]` |
| `enter`, `t`, `s`, `n`, `/`, `r`, `q` | as today |
| `?` | help popup; `?` or `esc` closes it |

- `a` (show all) goes away: finished items live in pane `[2]` now.
- Plain `]` and `[`, not `ctrl` or `cmd`: terminals send `ctrl+[` as Esc and do not pass `cmd` keys to the app.
- Selecting a row in `[1]` or `[2]` shows it in `[3]`.

### Mouse

- Clicking a row in pane `[1]` or `[2]` focuses that pane, selects the row and shows it in `[3]`. A click anywhere on the row's two lines counts.
- Clicking a tab name in a pane title switches to that tab.
- Clicking inside a pane focuses it.
- The mouse wheel moves the selection in `[1]`/`[2]` and scrolls `[3]`, in the pane under the pointer.
- Mouse and keys always agree: after a click, `j`/`k` move from the clicked row.
- Mouse support is on by default. Holding `shift` (or `option` on macOS, per terminal) while dragging still selects text for copy, since the terminal keeps that for itself.

## 4. Look

- Focused pane: border in the accent color. Other panes: dim border.
- Open items (pane `[1]`): title in the accent text color. Finished items (pane `[2]`): normal text, dim meta line.
- The selected row: reverse or highlighted background, in every pane, so focus and selection are two different signals.
- Each row is two lines plus a blank line between rows:
  - line 1: short ID (or the path when there is none), two spaces, title, clamped to the pane width with `…`.
  - line 2, dim, indented two spaces: status · progress · agent, for example `in-progress · 2/4 · omp`. Agent shows only when there is one (section 6).
- Colors come from one small palette in the TUI code, readable in dark and light terminals.

## 5. Detail pane `[3]`

Header, labels upper case padded to one width, colons in one column, ID first:

```
ID        : PLAN-12 · PLAN-k3f2 · plans/2026-09-26-short-ids
PLAN      : Short IDs
STATUS    : in-progress
SPEC      : SPEC-4 · Short IDs for specs, plans, tasks and bugs
WORKTREE  : short-ids
AGENT     : omp
FILE      : ../pm-board-short-ids/.pm/plans/2026-09-26-short-ids.md
TASKS     : 2/4
```

- The second label names the kind: `SPEC`, `PLAN`, `TASK` or `BUG`.
- A line whose value is empty is left out (no `SPEC` line for a plan with no spec, no `AGENT` line when nobody works on it).
- Problems (`! ...`) follow the header, then the body rendered as today.
- When `[3]` has focus, `j k g G ctrl+d ctrl+u` scroll the whole pane; a `n/m` line count shows in the bottom border.

## 6. Agent on a task

- `pmb tick` records who ticked. It reads the agent from the environment: `AI_AGENT` (Claude Code sets it, for example `claude-code_2-1-283_agent`, which becomes `claude`), and any other known variable a later check finds for omp. A new flag `pmb tick --agent <name>` overrides it. No value means nothing is recorded.
- The record lives in `<root>/.agents.json` in the worktree where the tick ran: `{"plans/2026-09-26-short-ids#task-3": {"agent": "omp", "at": "2026-09-26T20:46:00+07:00"}}`. It is git-ignored, so it never lands in a commit:
  - `pmb hook session-start` (run by the plugin at the start of every session) adds `.agents.json` to `<root>/.gitignore` when the line is missing, creating the file if needed. It runs only when the root folder exists and the repo uses git, never commits, and stays silent on any error so the session is never blocked. The change shows in `git status` until the user commits it.
  - `pmb tick` does the same check before it writes the record, for repos used without the plugin session.
- The board reads that file from the main checkout and from every other worktree on disk. A branch read from git has no agent data.
- A task shows its agent while it is not done. A plan or spec row shows the agents of its open tasks, joined with `,` when more than one.
- The dispatch skill and the implementer prompt add `--agent omp` to their tick commands when the harness sets nothing, so omp work shows up even before an env variable is found.

## 7. Status line

- Bottom line, right side: `pmb · <repo> · live · HH:MM`. `live` becomes `paused` when file watching is off, as today. The clock updates every minute.
- Left side: short key hints, or the search box while searching, or the last status message.

## 8. Help popup

- `?` opens a centered box over the panes, with a border and the title `Keys`, listing section 3 grouped by pane. `?` or `esc` closes it. Other keys do nothing while it is open.

## Testing

- Layout: at 120×40 the left column is between 28 and 48 wide and panes `[1]`, `[2]`, `[3]` all render with their numbered titles; below 60 columns only the focused pane renders.
- Tabs: `]`/`[` cycle the focused pane's tabs and wrap; pane `[2]`'s tabs follow pane `[1]`'s tab (Bugs gives Fixed/Wontfix, Tasks gives only Done).
- Lists: each tab puts the right statuses in `[1]` and `[2]` (table in section 2); a plan with a spec is its own row in Plans.
- Focus: `tab`, `shift+tab`, `1 2 3` move focus; the focused border uses the accent color.
- Rows: two lines, clamped with `…`, meta line holds the agent when present.
- Detail: header labels align, empty lines are left out, ID line comes first; `[3]` scrolls when focused.
- Agent: `pmb hook session-start` adds `.agents.json` to `<root>/.gitignore` once (second run adds nothing, no commit, errors stay silent, no root folder means no file); `pmb tick` with `AI_AGENT` set records `claude`; `--agent omp` records `omp`; nothing set records nothing; `.agents.json` is git-ignored; the board reads records from another worktree; a done task shows no agent.
- Status line shows `HH:MM` and `live`/`paused`; help popup opens and closes, and swallows other keys.
- Mouse: a click on a row focuses its pane, selects it and updates `[3]`; a click on a tab name switches tab; the wheel moves or scrolls the pane under the pointer; `j`/`k` continue from the clicked row.
- `a` no longer toggles anything.
