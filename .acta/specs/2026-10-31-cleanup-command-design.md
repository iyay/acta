---
parent: scratch/2026-10-31-cleanup-command
closes: [SCR-0064]
id: SPC-0184
created: "2026-10-31 07:24:03"
hash: b7hlrpb
---
# Cleanup of stale planning items

## Problem

Planning items pile up after their work is over. On 2026-10-31 this repo had
73 open debt files holding 111 open lines, 13 raw scratch items, one open bug
and one plan in progress with two open tasks. Some of these are already built,
some no longer apply and some were never valid. Nobody checks them, so the
board shows work that is not really open.

Deciding that an item is built or no longer valid needs someone to read the
code. A plain CLI rule cannot do that. Finding the candidates and writing the
result is mechanical work that the CLI can do.

## Scope

This is Architectural work. It adds a CLI command `acta cleanup`, a skill
`acta:cleanup`, a `reason` field, a written `done` status for scratch items,
and a way to close one debt line from the CLI.

It covers debt lines, scratch items, bugs, plans and specs. It never deletes a
file. Cleaning means a status change with a short reason, and git keeps the
history. It runs only when the user asks. Nothing starts it at land.

SCR-0064 (plans that stay open after every task is ticked) is closed by the
`--fix` rule below.

## The acta cleanup command

`acta cleanup` reads the board and prints candidates grouped by kind, oldest
first. It writes nothing. `--json` prints the same list for the skill.

A candidate is any of these:

- An open debt line.
- A scratch item with status `raw` and no spec link.
- A scratch item with status `brainstorming`, no spec link, and no commit
  touching its file for 14 days. A younger one may be a brainstorm that is
  still going.
- An open bug.
- A stuck plan. That is a plan that shows `in-progress`, has at least one open
  task, and has no commit naming it for 14 days. `acta commits` already finds
  those commits. A plan with no such commit counts from its `started` date.
- A spec with status `draft` or `approved` that has no plan and was created
  more than 14 days ago.

Past these thresholds, age only sorts the list. No item is closed because of
its age alone.

An item that lives only in another worktree is skipped, because that work is
still running.

`acta cleanup --fix` makes the changes that need no judgment, in one commit:

- It removes a written status that the board already ignores. The board flags
  these as `written status X ignored, derived Y`; five specs had one on
  2026-10-31.
- It marks a plan `done` when every task is ticked but the plan does not show
  `done` or `review`.

`acta cleanup apply` reads the approved decisions on stdin, one per line, and
writes them in one commit. Each decision names the item id, the new status and
the reason. Bugs may also carry the `fixed_in` sha. It refuses the whole batch
when any line is invalid, so a batch never half lands.

## Data and write commands

- New one-line frontmatter field `reason`, on scratch items, bugs, specs and
  plans. `acta set <id> reason "<text>"` writes it. `acta show` and the TUI
  detail pane show it.
- Scratch items get a written status `done`. It means the idea was built
  without a spec, and its reason must name the commit. Like `dropped`, a
  written `done` beats a spec link. The TUI Scratchpad gets a Done tab next to
  Specced and Dropped.
- `acta set` on a debt line (for example `DBT-0145.01`) accepts `status done`
  and `status wontfix`. It rewrites the box to `[x]` or `[-]` and adds the
  reason as a suffix in parentheses, the same shape as the existing `(stale)`
  lines. Today `acta set` refuses debt lines, and only `acta debt drop-low` or
  a hand edit can close one.

Each write commits like `acta set` does today.

Closing statuses by kind:

| Kind | Built | Not valid or no longer needed |
| --- | --- | --- |
| Debt line | `[x]`, reason `done in <sha>` | `[-]`, reason `stale` or a short cause |
| Scratch item | `done`, reason names the commit | `dropped` with reason |
| Bug | `fixed` with `fixed_in` | `wontfix` with reason, for a bug that does not reproduce and has no fix commit |
| Plan | `done` through `--fix` only | `dropped` with reason |
| Spec | not a cleanup case | `dropped` with reason |

When a stuck plan is dropped, its spec is dropped too unless the spec has
another plan that is still alive.

## The acta:cleanup skill

1. Run `acta cleanup`. Show the `--fix` changes and wait for a yes. Then run
   `acta cleanup --fix`.
2. Run `acta cleanup --json`. Start one read-only subagent per kind that has
   candidates, all in parallel.
3. Each subagent returns, for each item, one proposal: keep, done, dropped,
   wontfix or fixed. Each proposal carries evidence (a `file:line`, a commit
   sha or a spec id) and a one-line reason. Subagents never write.
4. The main agent checks each piece of evidence. The sha must exist, and the
   `file:line` must hold what the proposal claims. A proposal with weak
   evidence becomes keep.
5. Send one message per kind in the current channel, at most 10 numbered items
   per message. Each item shows the proposal, the evidence and the reason. The
   user answers ok, or names the numbers to skip.
6. Pipe the approved decisions of that message to `acta cleanup apply`, so
   each approved batch is one commit.

Keep writes nothing. A skipped item stays as it is.

## Wiki

At build, rewrite `.acta/wiki/scratch-dropped-vs-specced.md`. It must say that
`done` means built without a spec and that its reason names the commit, while
`dropped` still means not done or not valid.

## Testing

- Go tests for the candidate rules: each kind, the 14-day rule, and the skip
  for items that live only in another worktree.
- Go tests for `--fix` and for `apply`: one commit per batch, and a whole
  batch refused when one line is invalid.
- Go tests for `acta set reason`, for scratch `done` beating a spec link, and
  for a debt line moving to `[x]` or `[-]` with its suffix.
- TUI tests for the Done tab and for the reason in the detail pane.
- The new skill lives in `plugin/skills/cleanup/`. It is added to the skill
  list in `internal/hook/hook.go` and `plugin/hooks/default-rules.md`, and it
  passes `internal/plugincheck`.
- One eval case. The agent sends proposals with evidence and writes nothing
  before the user approves.

## Plan split

Two plans. The first covers the CLI and the data: `reason`, scratch `done`,
debt lines through `acta set`, `acta cleanup`, `--fix`, `apply` and the TUI.
The second covers the skill, the hook list, the eval and the wiki page. It
runs after the first plan lands.
