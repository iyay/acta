---
parent: scratch/2026-10-18-split-spec-into-plans
id: SPC-0147
created: "2026-10-18 20:15:10"
hash: pahb0va
started: "2026-10-18 20:23:18"
finished: "2026-10-18 20:35:16"
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
- `acta setup` asks it as its own step, right after "Plan detail":
  title "Plan split", options `off: one plan per spec` (first, the
  default) and `on: big specs become 2 to 4 plans, one session each`. The
  answered line, the text-mode form and the setup summary show it like
  every other step. Setup writes `plan_split` with the rest of the user
  config.
- The no-wizard fallback in `plugin/skills/setup/SKILL.md` asks it too:
  `acta config show` marks it `(default)` when it was never set, the same
  way it marks an unset `plan_depth`.
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

- New command `acta ready <plan> --parent <ref>`. Build picks its parent
  branch first (the branch the user names, else the default branch), then
  calls `acta ready` with it.
- Landed means present in the committed tree of that ref, never the local
  board: `acta ready` reads each `after:` plan with `git show
  <ref>:<plan path>` and counts it landed only when the file exists there
  and every task box in it is ticked. A plan done in another worktree, or
  landed on a different branch, is not landed. The worktree branches from
  the same ref, so it holds every plan it waits for.
- Exit 0: every `after:` plan has landed on the ref. Exit 1: lists each
  plan not landed yet.
- Exit 2, with the cause named: an id that is not known, an id that is not
  a plan (a spec, bug, debt or scratch id), a plan listed in its own
  `after:`, or a cycle. A cycle prints its chain, for example
  `PLN-0160 -> PLN-0162 -> PLN-0160`. Slice runs the same check over the
  plans it writes, so a bad `after:` never reaches main.
- `acta:build` runs `acta ready` before it makes the worktree. Exit 1 or 2
  stops the build and shows the output.
- A reused worktree may predate the land it waits for. So `acta ready`
  also takes `--head <rev>` (build passes the worktree HEAD whenever it
  reuses a worktree), and then checks a second thing for every `after:`
  plan: the parent commit where that plan landed is an ancestor of the
  head (`git merge-base --is-ancestor`). The landing commit is the first
  commit on the ref whose copy of the plan has every box ticked. Not an
  ancestor: exit 1, naming the plan and saying the worktree must take in
  the parent first. Build stops; it never rebases or merges on its own.
- The board marks a plan whose `after:` is not all done as `waiting`.
- Each plan lands on its own, through tidy.

### 4. Spec dates with several plans

Today every `acta tick` writes `started:` (and `finished:` once the spec
is done) into the parent spec, from inside the worktree. With two plans in
two worktrees, both write `started:` with different times, so the second
tidy land clashes. And a worktree that only sees its own plan done may
never write `finished:`.

Rule: when a spec has more than one plan, a tick in a worktree never
writes the spec's dates. Only the plan file gets its own dates there.
Land writes the spec dates on the parent branch, after the plan lands:

- `started:` is the earliest `started:` of the spec's plans on the parent.
  Each land may move it earlier, never later, so a plan that started first
  but lands last still sets it.
- `finished:` is written when the landing plan makes every plan of the
  spec done on the parent, and only then.

Order of a land that writes spec dates, inside the tidy flow: move the
parent to the tidy tip, write the spec dates, commit them with
`acta commit <spec path>`, and only then update `refs/acta/last-land` and
clean up the worktree. Right after, `git status --porcelain` in the main
checkout must show no tracked change; a leftover stops the land report as
not done, so the next land never meets a dirty parent.
A spec with one plan keeps today's behaviour.

## Testing

- Config: `plan_split` accepts `on` and `off`, rejects anything else, and
  shows `off (default)` when unset.
- Setup: the Plan split step sits after Plan detail, starts on off, keeps a saved value, and the answered line shows the pick.
- Session note: the line shows only when on; the byte cap test stays at its
  current value.
- `acta ready`: ready; waiting on one or more plans; a plan done only in
  another worktree or landed only on another branch counts as not landed;
  unknown id, non-plan id, self-reference and a two- and three-plan cycle
  each exit 2 with the cause (and chain for a cycle).
- Spec dates: two plans of one spec built in two real worktrees, started
  at different times, both land through tidy with no clash; the spec ends
  with the earliest `started:` and one `finished:` written at the second
  land, in both land orders (first-started plan lands first, and lands
  last); after each land the main checkout has no tracked change. A
  one-plan spec still gets its dates from tick.
- `acta ready --head`: a worktree made before its `after:` plan landed
  exits 1 until it takes in the parent; once it does, exit 0.
- Board: the `waiting` mark, and no mark once every `after:` plan is done.
- Plugincheck: the new slice and build text.
