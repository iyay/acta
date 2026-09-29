---
id: SPEC-6
hash: cwpp
---
# Review NOTEs become tracked tech debt

Status: approved by the user on 2026-09-26. Builds on `.pm/specs/2026-09-26-file-contract-design.md`, `.pm/specs/2026-09-26-short-ids-design.md` and `.pm/specs/2026-09-26-tui-panes-design.md` (built after the TUI panes work lands).

## Why

A review sorts findings into BLOCKER and NOTE. BLOCKERs become fix tasks; NOTEs today go only into agent memory ("A NOTE stays a NOTE: one line in memory"). They are real tech debt, but they are invisible on the board, have no ID, cannot be ticked, and are scattered over memory files. The user wants them tracked as a backlog (2026-09-26).

## Goals

- Every NOTE a review keeps lands in a file on the board, with its own ID, and can be ticked done or marked won't-fix.
- One file per reviewed plan, not one per NOTE.
- A debt item can be picked up like any other work ("work on DEBT-3.2").

## Non-goals

- No priority, owner or due date fields.
- No automatic fix of debt items.
- BLOCKERs do not change: they stay fix tasks in the plan.

## 1. File and folder

- New folder `<root>/debt/` (config key `dirs.debt`, default `debt`), next to specs, plans and bugs.
- One file per reviewed plan: `debt/<YYYY-MM-DD>-<plan stem>.md`, dated the day of the final review.
- Content:

```markdown
---
id: DEBT-3
hash: t9qe
parent: plans/2026-09-26-short-ids
---
# Review NOTEs: Short IDs

- [ ] `gitc.FirstSeen` has no `--follow`: a renamed file counts from the rename commit.
- [ ] `FixDuplicates` and `candidates` are just over the 50-line limit.
- [x] `plan/SKILL.md` line 20 repeats "right after".
- [-] `NewBug` counts numbers from the main tree only (duplicates are fixed at land).
```

- `parent` is the plan the NOTEs came from.
- Each checklist line is one debt item. Its status comes from the box: `[ ]` open, `[x]` done, `[-]` wontfix.
- Lines outside the checklist (for example a short intro) are allowed and ignored by the board.

## 2. Board

- New kind `debt`. The file is a `debt` item with its own number and hash (`DEBT-3`, `DEBT-t9qe`), like plans in the short IDs spec.
- Each checklist line is a child item of kind `debt-item`, numbered by its position among the checklist lines: `DEBT-3.1`, `DEBT-3.2`, … and `DEBT-t9qe.1`. Path ID: `debt/<stem>#item-N`.
- Status of the file: `done` when every line is `[x]` or `[-]`, else `open`. Progress `done/total` counts `[x]` and `[-]` as closed.
- `pmb show`, `pmb tick` and `pmb set` accept debt IDs. `pmb tick DEBT-3.2 --all` sets the line to `[x]`. New `pmb tick DEBT-3.2 --wontfix` sets it to `[-]`.
- The board reports a problem when `parent` names a plan that does not exist.
- Numbering rule for lines: position only. Reordering or deleting a line renumbers the ones after it, which is fine because lines are short-lived; the hash form of the file stays stable.

## 3. CLI: `pmb debt new`

- `pmb debt new <plan id> [--title T] < notes.md` reads NOTEs from stdin, one per line (a leading `- ` is optional), and writes `debt/<today>-<plan stem>.md` with the checklist, `parent`, `id` and `hash`, then auto-commits like `pmb bug new` (`pm: new debt <stem>`).
- Title defaults to `Review NOTEs: <plan title>`.
- If the file for that plan already exists (a second review of the same plan), the new NOTEs are appended to its checklist, never duplicated (a line with the same text is skipped).
- Empty stdin: exit 1, no file.
- The plan ID accepts every ID form (`PLAN-3`, hash, path).

## 4. Skills

- `pm:review`, "Where findings go": a NOTE is no longer kept in memory. When the plan's final review round is CLEAN, the orchestrator collects every NOTE from all rounds of that plan and runs `pmb debt new <plan id>` on the branch before `pm:land`. The debt file merges with the branch.
- `pm:land`: the landing report names the debt file and its number of items.
- `pm:brainstorm` / `pm:plan`: a plan that works a debt item sets `parent: debt/<stem>` and names the item IDs it closes; `pm:land` ticks those items with `pmb tick <DEBT-n.m> --all` after the merge.
- The house rules and dispatch brief template say the same, so a recipient that reviews never writes NOTEs to memory.

## 5. TUI

- Pane `[1]` tabs become `Specs ─ Plans ─ Tasks ─ Bugs ─ Debt`. The Debt tab lists open debt items (lines), grouped under their file's short ID in the meta line.
- Pane `[2]` for Debt: `Done ─ Wontfix`.
- Detail pane for a debt item: `ID`, `DEBT` (the line text), `STATUS`, `FROM` (the parent plan's short ID and title), `FILE`.

## 6. Moving the current NOTEs

- One-time move, done as the last task: every open NOTE currently in project memory files (`open-notes-review-notes`, `short-ids-review-notes` and the NOTE lines in the others) becomes a debt file per source plan through `pmb debt new`. NOTEs already fixed since are written as `[x]`.
- After the move, those memory files are deleted and the memory index updated, so each NOTE lives in one place.

## Testing

- Parse: a debt file gives one `debt` item and one child per checklist line; `[ ]`, `[x]`, `[-]` map to open, done, wontfix; non-checklist lines are ignored; file status and progress follow the lines.
- IDs: `DEBT-n`, `DEBT-hash`, `DEBT-n.m`, `DEBT-hash.m` and the path forms resolve; lower case works.
- `pmb debt new`: writes the file with parent, ids and checklist; commit once; appends to an existing file without duplicate lines; empty stdin exits 1; bad plan ID exits 1.
- `pmb tick DEBT-n.m --all` and `--wontfix` set `[x]` and `[-]`.
- Board problem for a missing parent plan.
- Plugin text: `pm:review` names `pmb debt new` and no longer says NOTEs stay in memory; `pm:land` reports the debt file and ticks closed debt items; guarded per skill in `internal/plugincheck`.
- TUI: Debt tab and its Done/Wontfix pane list the right lines; detail header fields.
