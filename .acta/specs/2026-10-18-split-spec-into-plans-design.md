---
parent: scratch/2026-10-18-split-spec-into-plans
id: SPC-0147
created: "2026-10-18 20:15:10"
hash: pahb0va
---
# Split one spec into several parallel plans

## Why

With the inline executor, one session writes all the code of a plan. A big
spec then runs through one session from start to end. Splitting the spec into
a few plans lets several sessions build the same spec at once, each in its
own worktree.

Today a plan links to its spec with `parent:`, and nothing stops two plans per
spec. The board already sums the tasks of every plan under one spec
(`parentStatus` in `internal/board/board.go`), so the spec shows done only
when every plan is done. What is missing: a setting to turn splitting on,
slice rules for how to cut, a way to say one plan waits for another, and a
check that stops a build from starting too early.

## Design

### 1. Setting

- New user setting `plan_split`, values `on` or `off`, default `off`. It
  lives next to `plan_depth`: `internal/config/user.go` validates it, the
  repo `.acta.yaml` can set it too, `acta config set --plan-split` writes it.
- `acta setup` asks one new question for it.
- `acta config show` prints `plan_split: off (default)` when nothing sets it,
  the same way `plan_depth` shows its default.
- The session note writes `- Plan split: on.` only when it is on. When it is
  off, the note does not grow, so the session start byte cap stays the same.
- `/slice split` turns it on for one run, the way `/slice minimal` sets the
  depth for one run. It is never saved.

### 2. Slice

- With `plan_split` on, slice cuts the spec into 2 to 4 plans by file area.
  Two plans that do not wait on each other never touch the same file.
- A plan that needs the work of another plan lists it in its frontmatter:
  `after: [PLN-xxxx]`.
- A spec that fits one plan stays one plan. More than 4 plans means the spec
  is too big and should have been split in shape.
- Each plan sets the same `parent:` (the spec), gets its own `acta id`, and is
  committed on main.
- After writing them, slice prints one line per plan and does not start a
  build itself:

  ```
  build PLN-0160            (ready now)
  build PLN-0161            (ready now)
  build PLN-0162            (wait for PLN-0160 to land)
  ```

  The user hands each line to a session of their choice.

### 3. Build, board, land

- New command `acta ready <plan>`. Exit 0 when every plan in its `after:`
  has landed on the parent (status `done`). Exit 1 lists each plan still
  open. Unknown ids in `after:` are an error (exit 2) with the id named.
- `acta:build` runs `acta ready` before it makes the worktree. Exit 1 stops
  the build and names the plans it waits for.
- The board marks a plan whose `after:` is not all done as `waiting`.
- Land does not change: each plan lands on its own, through tidy.

## Testing

- Config: `plan_split` accepts `on` and `off`, rejects anything else, and
  shows `off (default)` when unset.
- Setup: the new question and its default.
- Session note: the line shows only when on; the byte cap test stays at its
  current value.
- `acta ready`: ready, waiting on one or more plans, unknown id.
- Board: the `waiting` mark, and no mark once every `after:` plan is done.
- Plugincheck: the new slice and build text.
