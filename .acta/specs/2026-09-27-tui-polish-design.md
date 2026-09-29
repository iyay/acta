---
id: SPC-0009
hash: w1o6wzq
---
# TUI polish: one-line rows, task dots, per-pane scroll, no broken frames

Status: design approved by the user in chat on 2026-09-27, section by section. Builds on `.acta/specs/2026-09-26-tui-panes-design.md`, `.acta/specs/2026-09-26-tui-followup-design.md` and `.acta/specs/2026-09-26-review-debt-design.md`.

## Why

After the Debt tab landed, the user listed eleven problems with the TUI (2026-09-27): rows are two lines tall and carry noise, the selected row uses a background, the screen breaks on resize and scroll, there is one scroll for everything, the in-progress divider looks wrong, task lists vanished from the detail pane, the debt detail is empty, corners are rounded, and the help popup covers a full-width band. The user also asked for an AUTHOR field.

## Non-goals

- No new keys, no new tabs, no new data files.
- No change to how status is derived. The only status change is the word: a task in progress reads `in-progress`, not `doing` (user ruling 2026-09-27, section 5).
- No width or tab-title changes beyond what is written here.

## 1. List panes (pane [1] and [2], every tab)

- One row per item, one line tall. No blank line between rows.
- A row shows the short ID and the title.
- A row whose item is in progress (status doing, in-progress or fixing, or started through `acta tick --start`) shows the title in the accent color and, at the right end of the row, `<done>/<total> · <agent>`. The agent part is left out when no agent is known. Rows not in progress show no progress, no agent and no status word.
- The selected row has a subtle dark background across the full row width (not reverse video; `lipgloss.AdaptiveColor{Light: "254", Dark: "236"}`) and bright text (user ruling 2026-09-27). Every row that is not selected is dimmed. The accent color of an in-progress row still shows, dimmed when not selected.
- The divider between in-progress and not-started items is one full-width line of `─` across the pane, drawn the same way as the rule under the detail header. It hides when either group is empty.
- A long title is cut with `…` so the progress and agent part always fits.

## 2. Detail pane

### 2.1 Header

- A task's header labels its progress `SUBTASKS`. Every other kind keeps `TASKS`.
- New field `AUTHOR`, right under `STATUS`: the git author name of the commit that first added the file. A task shows its plan's author; a debt item shows its debt file's author. A file with no commit yet shows the current `git config user.name`. When neither exists, the field is left out.

### 2.2 Task list with dots

Under the header rule and above the markdown body, the detail pane lists the item's tasks, one line each, in file order:

`<dot> <short id>  <title>  <done>/<total> · <agent>`

- Dots: `●` in progress (accent color), `○` not started (dim), `✓` done (dim). A wontfix debt line uses `✓` too.
- The progress and agent part shows only on in-progress lines, like the list rows.
- What each kind lists:
  - Spec: the tasks of every plan under it, grouped with one plain line per plan (the plan's short ID and title) above its tasks.
  - Plan: its tasks.
  - Task: its steps (the checklist boxes under the task), labelled SUBTASKS in the header. A step has no short ID; its line shows the dot and the step text.
  - Bug: the tasks of every plan whose `parent` is this bug, grouped per plan like a spec.
  - Debt item: every checklist line of its debt file, with the ID `DEBT-n.m`. The line of the selected item is bright; the others are dim.
- No tasks: no list and no empty-state line.

### 2.3 Debt body

A debt item's detail pane shows the list above, then any text of the debt file outside the checklist, rendered as markdown like other bodies. Today this pane is empty.

## 3. Scroll, popup, frames

### 3.1 Per-pane scroll

- Pane [1], [2] and [3] each keep their own scroll offset. There is no scroll shared across panes.
- Only the focused pane scrolls, like lazygit (user ruling 2026-09-27). The mouse wheel scrolls the focused pane only when the pointer is over it; a wheel over any other pane does nothing (no scroll, no focus change). A click still focuses a pane. `j`/`k`, `g`/`G` and `ctrl+d`/`ctrl+u` act on the focused pane, as today.
- In a list, moving the selection keeps the selected row inside the pane.
- Choosing a different item resets the detail offset to the top.
- When a pane's content is taller than its inner height, the pane draws a scrollbar in one column inside the right border: `█` for the visible part, `░` for the rest. It also writes the position `<first visible line>/<total lines>` into its bottom border. When the content fits, neither is drawn and the column goes back to content.

### 3.2 Help popup

- Only the popup box is drawn over the screen, centered. Every cell outside the box shows the panels behind it unchanged.

### 3.3 Frames never break

- The left column (pane [1] stacked on pane [2]) is always exactly as tall as the detail pane, at every terminal size. Today `split` in `internal/tui/view.go` floors both halves at 5 rows on their own, so for a body of 12 rows or less the left column is taller than the detail pane and `View` cuts rows, which overlaps borders.
- `View` never returns more lines than the terminal height, nor a line wider than the terminal width.
- After a resize, and while scrolling any pane, the screen shows no rows or borders left over from an earlier frame. The user saw this at full terminal size, so the `split` fix alone does not cover it. The first build task reproduces it at full size; if the cause is the renderer keeping stale rows after a size change, the fix clears the screen on every size change. If the cause turns out to be something else, the build stops and reports it before fixing.

### 3.4 Corners

- Every pane and the popup use square corners (`┌ ┐ └ ┘`). No rounded corners anywhere.

## 4. Landing leaves no task hanging

PLAN-15 landed with task 6 at `doing 4/5`: its step 5 said "after the branch lands", so no one could tick it before the merge, and `acta:land` did not check. The board showed a finished task as in progress.

- `acta:land`, before the merge: run `acta show <plan id>`. When `done` is less than `total`, do not merge. Tick each box whose work is really done; report the rest to the user.
- `acta:plan`: a plan never holds a step that can only happen after landing. Such work goes in the landing report as the next action, not as a box.
- `internal/plugincheck` guards both lines in the land and plan skills.

## 5. Status word `in-progress`

A task that is under way reads `in-progress` everywhere (board, `acta show`, `acta list --json`, the TUI), never `doing`. Specs and plans already use `in-progress`; bugs keep `fixing`. Task status is derived from boxes and never stored, so no file changes.

## Testing

- List rows: one line per item at every tab; in-progress rows end with `n/m · agent`, other rows carry no progress, agent or status word; long titles cut with `…` and the tail still fits; selected row has the subtle dark background and bright text, never reverse video.
- Divider: a full-width `─` line between the groups; hidden when a group is empty.
- Header: `SUBTASKS` on a task, `TASKS` elsewhere; `AUTHOR` from the first commit, from the plan for a task, from `user.name` for an uncommitted file, left out when neither exists.
- Task list: for each kind (spec, plan, task, bug, debt item) the right lines, dots and order; nothing when empty; the selected debt line bright.
- Debt body: non-checklist text of the debt file renders under the list.
- Scroll: each pane keeps its own offset; the wheel scrolls only the focused pane, and a wheel over an unfocused pane changes nothing; scrollbar and `n/m` appear only when content overflows and match the offset.
- Popup: every cell outside the popup box equals the frame without the popup.
- Frames: for every height from 3 up and a range of widths, left column height equals detail height, `View` line count is at most the height and every line fits the width; the full-size resize/scroll repro no longer leaves stale rows.
- Corners: no `╭ ╮ ╰ ╯` in any frame.
- Land skill: guard requires the progress check before merge; plan skill: guard forbids post-land steps.
- Status word: no board item, CLI output or TUI frame ever shows `doing`; a started or half-ticked task reads `in-progress`.
