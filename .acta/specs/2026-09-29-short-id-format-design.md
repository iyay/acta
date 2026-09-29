---
parent: scratch/2026-09-29-fast-id-lookup
status: approved
id: SPC-0023
hash: oo47ajk
---
# Short Id Format: 3-Letter Prefix, Fixed Width

Status: design approved by the user on 2026-09-29 (Architectural). The rulings are in SCRATCH-13.

## Why

TUI list rows do not line up. `SCRATCH-3` and `SCRATCH-13` have different widths, and `SCRATCH-` is long. Every kind has the same problem. Agents also need an id they can match with one grep, with few tokens.

## Format

- Id: a 3-letter prefix, a dash, then 4 digits padded with zeros. `PLN-0030`.
- Prefixes: `PLN` (plan), `SPC` (spec), `BUG` (bug), `DBT` (debt), `SCR` (scratch).
- `IsID` accepts 4 digits and rejects `0000`. More than 4 digits is accepted only without a leading zero, so number 10000 can still be written; it then gets a problem line (below).
- Hash: 7 characters, the same length as a git short commit hash. First a lower-case letter, then a-z or 0-9. `IsHash` checks for length 7.
- The board still reads an old id (`PLAN-30`) and an old 4-character hash, so an unmigrated branch loads. It shows the new form (`PLN-0030`) and marks the file for `acta id` to rewrite.
- Tasks and debt items get a 2-digit number after a dot: `PLN-0030.03`, `DBT-0022.04`. A task number that is not a plain number, like the fix task `F1`, stays as written: `PLN-0030.F1`. This number is shown only. It is never written to a file.
- A plan with 100 or more tasks, a debt file with 100 or more items, or an id number of 10000 or more gets a problem line. The rows would stop lining up, so acta says so out loud.

## Lookup

- One function `canon(id)` turns any id into one key. It lowercases the id, maps the old prefix to the new one (`plan` to `pln`, `scratch` to `scr`, `spec` to `spc`, `debt` to `dbt`), and drops leading zeros. `PLAN-30.3`, `pln-0030.03` and `PLN-30.3` all become `pln-30.3`.
- `addAlias` stores keys in canon form, and `Board.Get` looks them up in canon form. There is one map and no table of old names.
- A hash matches by unique prefix, the way git does. `oxoq` finds `oxoqk2m` when only one hash starts with it. Two matches mean not found, with a message that says the prefix is ambiguous. A prefix shorter than 4 characters is rejected.

## Display

- TUI list and `acta list` show `PLN-0030  Title`. An id is always 8 wide and a task id is 11 wide, so the columns line up inside each tab.
- The hash shows only in the detail pane and in `acta show`.

## Migration

- There is no new command. `acta id` already runs after every merge. It now also rewrites old ids to the new format on every run. This covers the first migration and any old files a merge brings back.
- It rewrites frontmatter only:
  1. `id: PLAN-30` becomes `id: PLN-0030`.
  2. A 4-character hash gets 3 random characters added. `freeHash` checks the result so it does not clash.
  3. Ids in `closes:` become the new format, for example `[DEBT-17.1, SCRATCH-14]` becomes `[DBT-0017.01, SCR-0014]`.
- `parent:` holds a path, so it is not touched. Bodies, commit messages and memory are not touched either. `canon` still resolves the old ids written in them.
- The board reads both formats. So a branch that has not been migrated yet still loads correctly.
- The change lands as one commit through the auto-commit that `acta id` already makes: `acta: migrate ids to 3-letter prefix`.
- Run the first migration when few worktrees are open, so that fewer merges conflict on `id:` lines.

## Skills

- `land`, `brainstorm`, `plan` and `scratch` switch their examples to the new format (`SCR-0001`, `PLN-0001`).
- `acta:plan` gets a rule: 99 tasks per plan at most. Split big work into several plans.
- `acta:land` says in one sentence that `acta id` after a merge also rewrites old ids.

## Testing

Write every test first and watch it fail.

- `IsID` and `IsHash`: accept 4 digits and exactly 7 characters. Reject `PLN-30`, `PLN-0000`, `PLN-00300`, 4- and 8-character hashes, a hash that starts with a digit, and unicode.
- `canon`: `PLAN-30.3`, `pln-0030.03`, `SCRATCH-14` and `scr-14` each find the right item. `PLX-0001` and an empty string find nothing.
- Hash prefix: a unique prefix finds the item. A prefix two hashes share finds nothing and says it is ambiguous. A prefix under 4 characters is rejected.
- `acta id`: an old-format file gets its id, hash and `closes:` rewritten. A new-format file stays byte for byte the same, so a second run changes nothing. Duplicates still get renumbered.
- Problem lines: a plan with 100 tasks, and an id number of 10000.
- TUI: a list row renders as `PLN-0030  Title` at a fixed width.
- Test fixtures that use `SCRATCH-` or `PLAN-` move to the new format. Some stay in the old format on purpose, to test the aliases.

## Out of scope

- `acta path <id>`, a fast lookup that skips git. Filed as its own scratch item.
- Some list rows show the file slug instead of a title. Filed as its own scratch item.
