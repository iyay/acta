---
id: SPEC-4
hash: ttye
---
# pmb: TUI and CLI over the pm-board file contract

Date: 2026-09-26. Sub-project 1 of pm-board.
Contract: `.pm/specs/2026-09-26-file-contract-design.md`. This spec does not restate it;
every rule about folders, IDs, types, frontmatter, derived status, bug files, writes and
auto-commit comes from there.

## Goal

One binary, `pmb`, run from inside a repo:

- with no arguments it opens a lazygit-style TUI that shows every story, task and bug;
- with a subcommand it does the same reads and writes for agents and scripts.

Both paths share one implementation of the contract.

Not goals: multi-repo views, a board or kanban view, deleting items, moving legacy files,
themes, a server.

## 1. Structure

```
cmd/pmb/          main: no args = TUI, otherwise a subcommand
internal/config   find the root and read .pm.yaml
internal/board    scan, parse, link plans, derive status -> Board
internal/write    frontmatter edits and new bug files
internal/gitc     auto-commit rules
internal/tui      Bubble Tea model and views
```

- `board` never touches the screen or git. Input is a root folder plus legacy folders;
  output is a `Board` value holding stories, bugs, tasks, their links and statuses.
  The TUI and the CLI both read from it.
- `gitc` runs the `git` binary. It does not use go-git, so repo config and hooks behave
  exactly as they do for the user. (go-git failed on `extensions.worktreeconfig` in a
  real repo.)
- `write` edits frontmatter through `yaml.v3` nodes so unknown fields and field order
  survive.
- Subcommands are parsed with the standard `flag` package, with `ContinueOnError` so bad
  flags exit 1 rather than the package's default 2 (2 is taken, see the CLI exit codes).
- Markdown is read by scanning lines. The only things read are the first `# ` heading,
  `### Task N` headings, `- [ ]` / `- [x]` checkboxes and the `**Spec:**` line.

Dependencies: `bubbletea`, `lipgloss`, `glamour`, `yaml.v3`, `fsnotify`. Nothing else.

### Root lookup

`--root <dir>` flag, then `PM_ROOT`, then `root:` in `.pm.yaml` at the repo root, then
`.pm/`. The repo root is found with `git rev-parse --show-toplevel`; outside a git repo it
is the current directory.

## 2. CLI

| Command | Does |
|---|---|
| `pmb` | open the TUI |
| `pmb list [--type story\|task\|bug] [--all] [--json]` | the same list the TUI shows; active items only unless `--all` |
| `pmb show <id> [--json]` | one item with its children and derived status |
| `pmb set <id> status\|type <value>` | same as the TUI's `s` / `t`, with auto-commit |
| `pmb bug new <slug> [--ref X] [--title T]` | new bug file; body read from stdin; no editor |

Every command accepts `--root`.

Exit codes:

| Code | Meaning |
|---|---|
| 0 | done; for writes, the file is written and committed (or auto-commit is off) |
| 1 | bad input: unknown id, unknown status or type value, legacy file, bad flags |
| 2 | the file was written but the commit was skipped; stderr says why |
| 3 | anything else failed (I/O, git error) |

`--json` output is a stable shape: `id`, `type`, `title`, `status`, `status_source`
(`derived` or `frontmatter`), `ref`, `parent`, `children` (ids), `progress`
(`{done, total}`), `path`, `legacy`, `problems` (list of strings).

`pmb bug new` without stdin (a terminal) opens `$EDITOR` on the template, the same as `n`.
With stdin, the body becomes everything below the `# ` title; if stdin has no `## Symptom`
section, the command fails with exit 1 and writes nothing.

## 3. TUI

### Layout

```
 pmb · be-pmis · active                                    ⟳ live
┌─[1] Stories 12─[2] Tasks 31─[3] Bugs 3─┬─ STORY · New-261 ──────────────────┐
│ ◐ New-261 Post TOP per type   3/4      │ Post TOP per material type         │
│ ○ New-262 Opex TOP column     draft    │ in-progress · 3/4 tasks · 2 bugs   │
│ ● New-240 Evidence revoke doc  done    │ spec  specs/2026-09-25-cac-post... │
│ ▸ untyped (403)                        │                                    │
│                                        │ Tasks                              │
│                                        │  ✓ 1 ATP save carries type    5/5  │
│                                        │  ◐ 3 Post reads own type row  2/5  │
│                                        │ ─────────────────────────────────  │
│                                        │ (file body rendered with glamour)  │
├────────────────────────────────────────┴────────────────────────────────────┤
│ pm: bugs/…-atp-print status fixed ✓ committed                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

- Header: repo name, filter (`active` or `all`), refresh mode (`live` or `manual`).
- Tabs show a count of the items the current filter lets through.
- List rows: status icon, `ref` if set, title, and a short status or progress.
- Detail pane: a summary block (status, progress, parent, children, path), then the file
  body rendered with glamour. A task shows only its own section of the plan.
- Status bar: the result of the last action.

Icons: `○` not started, `◐` in progress, `●` done, `!` has a problem.
Colours follow the terminal theme. There is no theme config.

### Lists

- Default filter is active items: everything not `done`, `fixed`, `dropped` or `wontfix`.
- `a` switches between active and all.
- Newest first, by the date in the file name. Tasks sort by their plan's date, then task
  number.
- Legacy items sit in a folded `untyped` group at the bottom of the Stories tab. Their
  tasks show only in that story's detail pane, never in the Tasks tab: old plans are full
  of boxes nobody ticked, and would bury the real work under thousands of `todo` rows.
  `pmb list --type task` leaves them out the same way; `--all` includes them.
- The Tasks tab shows each task's parent after its title.
- `/` searches title, `ref` and slug across every item, whatever the filter. `esc` clears.

### Keys

| Key | Does |
|---|---|
| `1` `2` `3`, `tab` | switch tab |
| `j` `k`, arrows, `g` `G` | move in the list, jump to top or bottom |
| `ctrl-d` `ctrl-u` | scroll the detail pane |
| `enter` | open the file in `$EDITOR`; a task opens at its heading line |
| `/` | search |
| `a` | active / all |
| `t` | pick a type (popup) |
| `s` | pick a status (popup, only values valid for that type) |
| `n` | new bug in `$EDITOR` |
| `r` | reload (always works; the only way in manual mode) |
| `?` | key help |
| `q` | quit |

`enter` on a folded group opens or closes it.

### Live reload

The root folder and legacy folders are watched with fsnotify. Events are gathered for
200 ms, then the board is rescanned. The cursor stays on the same item by ID; if that item
is gone, it moves to the nearest row.

## 4. Errors

- A broken file does not stop the board. Bad frontmatter YAML, a `parent` that points
  nowhere, a `**Spec:**` path that matches no spec: the item still shows, marked `!`,
  and the detail pane lists the problem. `--json` puts it in `problems`.
- A status or type value outside the contract's list is shown as written, marked `!`.
  `pmb set` refuses to write one (exit 1).
- No root folder and no legacy folder: the TUI opens with an empty list and says the repo
  has no `.pm/` yet. `pmb bug new` creates the folder.
- Not a git repo: reads work. Writes still write; auto-commit is skipped with a warning
  (TUI) or exit 2 (CLI).
- fsnotify fails to start (for example, the watch limit): the TUI falls back to manual
  mode and the header says `manual`.
- `$EDITOR` unset: fall back to `vi`.

## 5. Testing

TDD. `go test ./...` is the whole gate.

| Package | How it is tested |
|---|---|
| `config` | table tests for the lookup order, with a temp repo |
| `board` | fixture folders under `testdata/`, one test per contract rule: each derived status row, the three plan-link rules, legacy folders, broken files |
| `write` | golden files: body and field order unchanged after an edit; new block added to a file with none |
| `gitc` | a real git repo in `t.TempDir()`: commits only the given path, skips on a dirty file, mid-merge, detached HEAD; a failing hook keeps the change |
| CLI | build the binary, run it on fixtures, check `--json` output and exit codes |
| `tui` | model-level tests (send a key, check the state); no screen snapshots |

Manual check on real data: run `pmb` inside be-pmis; the 124 specs and 403 plans under
`docs/superpowers/` show in the `untyped` group, and `pmb list --all --json | jq length`
returns a count with no crash.
