# Short IDs for specs, plans, tasks and bugs

Status: approved by the user on 2026-09-26. Builds on `.pm/specs/2026-09-26-file-contract-design.md` (section 1, IDs).

## Why

Today an item's ID is its path: `plans/2026-09-26-spec-line-md#task-1`. It is too long to say in chat. The user wants to write "work on PLAN-12.3" and have any agent, `pmb` and the TUI understand it.

## Goals

- Every spec, plan and bug has a short number ID that is easy to say: `SPEC-4`, `PLAN-12`, `BUG-7`.
- Every task has one too, built from its plan: `PLAN-12.3`.
- Every item also has a permanent hash ID that never changes: `PLAN-k3f2`.
- Both forms work anywhere a path ID works today. Path IDs keep working.

## Non-goals

- No change to the path ID rules in the file contract, and no ID in file names (user 2026-09-26): a number can change at land, and a rename would break every path that points at the file.
- No IDs for items in legacy folders (pmb never writes there).
- No rename of the `type: story` field (the TUI tab rename is in the TUI spec).

## 1. Format

| Kind | Number ID | Hash ID |
|---|---|---|
| spec | `SPEC-4` | `SPEC-m2x9` |
| plan | `PLAN-12` | `PLAN-k3f2` |
| task | `PLAN-12.3` | `PLAN-k3f2.3` |
| bug | `BUG-7` | `BUG-q8d1` |

- The prefix is the whole word, upper case, then a dash.
- A number is digits only, starting at 1, counted per prefix.
- A hash is 4 characters: a lower-case letter, then 3 lower-case letters or digits. It always starts with a letter, so `PLAN-1234` is always a number and `PLAN-k3f2` is always a hash.
- A task adds `.` and the task number exactly as its heading writes it: `### Task 3` gives `.3`, `### Task F1` gives `.F1`.
- Lookup ignores case: `plan-12.3` finds `PLAN-12.3`.

## 2. Where IDs live

- Specs, plans and bugs carry two new frontmatter fields: `id: PLAN-12` and `hash: k3f2`. The hash field holds only the 4 characters; the prefix comes from the kind.
- Tasks carry nothing. Their IDs come from their plan's IDs plus the task number.
- A file with no `id` or `hash` still works. It shows only its path ID until IDs are given to it.

## 3. Giving IDs

- New command `pmb id [<path-id>...]`. With no argument it gives IDs to every item in the root folder that has none. It writes the two fields and auto-commits like `pmb set`, one commit for the whole run.
- `pmb bug new` gives the new bug its IDs at once.
- `pm:brainstorm` runs `pmb id` right after it writes a spec, and `pm:plan` right after it writes a plan, so the number exists while the work is in progress.
- The next number is one more than the highest number of that prefix seen on the board, which includes other worktrees and unmerged branches (the board already reads them).
- A new hash is random and is checked against every hash on the board; on a clash pmb picks again.
- `pmb id` never changes an ID that is already there.

## 4. Duplicates after merge

Two worktrees can give the same next number (both see `PLAN-12` as the highest and both take `PLAN-13`).

- `pm:land` runs `pmb id --fix-duplicates` on the parent branch right after the merge.
- For each number held by more than one item, the item whose file reached the parent branch last gets the next free number; the first one keeps its number. The hash never changes.
- pmb prints each change (`PLAN-13 -> PLAN-14 (plans/2026-09-27-x)`), and `pm:land` puts that line in its landing report so the user knows the old number moved.
- The hash is the stable name. Anything written down for the long term (commit messages, memory notes) should use the hash.

## 5. Using IDs

- `pmb show`, `pmb tick` and `pmb set` accept a number ID, a hash ID or a path ID.
- `pmb list` and `pmb show` print the number ID first, then the path.
- `pmb list --json` adds `id` and `hash` to each item.
- The TUI shows the number ID at the start of each row (layout lives in the TUI spec).
- An unknown short ID gives `unknown id PLAN-99`, exit 1, like an unknown path ID today.
- The board reports a problem for two items with the same number or the same hash, like it does today for a duplicate task number.

## 6. Existing files

- One run of `pmb id` on `main` after this lands gives IDs to every current spec, plan and bug, oldest file first by the date in its name, so older work gets smaller numbers.

## Testing

- Parse: `id` and `hash` read from frontmatter; a task's IDs built from its plan and heading (`3`, `F1`).
- Lookup: number, hash and path IDs all find the same item; lower case works; `PLAN-1234` is read as a number, `PLAN-k3f2` as a hash; unknown short ID gives exit 1.
- `pmb id`: gives IDs only to items without them; never changes existing ones; next number counts other worktrees; hash clash picks again; one commit per run; legacy items untouched.
- `pmb bug new`: new bug has both IDs.
- `--fix-duplicates`: the later-added item gets the next free number, the hash stays, the change line prints; no duplicates means no change and no commit.
- Board problem line for a duplicate number or hash.
- Plugin text: `pm:brainstorm`, `pm:plan` and `pm:land` name the `pmb id` steps, guarded in `internal/plugincheck`.
