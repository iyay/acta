---
parent: scratch/2026-10-31-cleanup-command
closes: [SCR-0064]
id: SPC-0184
created: "2026-10-31 07:24:03"
hash: b7hlrpb
started: "2026-10-31 07:36:26"
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

SCR-0064 (plans that stay open after every task is ticked) is already
handled by the board: a plan with a written `approved` or `in-progress`
takes its status from its boxes, so every box ticked reads `done`. The
`--fix` rule below covers the other written words.

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
  task, has no linked worktree, and has had no commit for 14 days. A commit
  counts when it touches the plan file or carries a `Task: PLN-<hash>#<n>`
  trailer for the plan; chore commits such as ticks count too. A plan with no
  such commit counts from its `started` date, or from `created` when `started`
  is missing.
- A spec with status `draft` or `approved` that has no plan and was created
  more than 14 days ago.

Past these thresholds, age only sorts the list. No item is closed because of
its age alone.

An item that lives only in another worktree is skipped, because that work is
still running. A plan that has a linked worktree is skipped for the same
reason, even when its file is also on the main checkout.

`--json` prints one JSON object per candidate with the fields `id`, `kind`,
`title`, `path`, `status`, `age_days` and `why` (the rule that made it a
candidate).

`acta cleanup --fix` makes the changes that need no judgment, in one commit:

- It removes a written status that the board already ignores. The board flags
  these as `written status X ignored, derived Y`; five specs had one on
  2026-10-31.
- It marks a plan `done` when it has at least one task, every task is ticked,
  and the plan does not show `done`, `review` or `dropped`.

`acta cleanup apply` reads the approved decisions on stdin, one JSON object
per line, and writes them in one commit. Fields:

- `id`: the item id, for example `SCR-0081`, `BUG-0037` or `DBT-0145.01`.
- `status`: the new status, allowed for that kind (see the table below).
- `reason`: one line of text, required.
- `fixed_in`: a commit sha, required when `status` is `fixed` and refused
  otherwise.

Example lines:

```
{"id":"DBT-0145.01","status":"done","reason":"done in 4dcc44c"}
{"id":"SCR-0081","status":"dropped","reason":"replaced by SPC-0183"}
{"id":"BUG-0037","status":"fixed","fixed_in":"cb1f58b","reason":"runner stops at the final reply"}
{"id":"PLN-0176","status":"dropped","reason":"no work for 30 days, idea moved to SCR-0080"}
```

Both `--fix` and `apply` check every change before the first write: the id
exists, the status is allowed for the kind, the reason is present, and no plan
has a linked worktree. When any check fails, they write nothing, print each
failing line, and exit non-zero, so a batch never half lands.

## Data and write commands

- New one-line frontmatter field `reason`, on scratch items, bugs, specs and
  plans. `acta set <id> reason "<text>"` writes it. `acta show` and the TUI
  detail pane show it.
- Scratch items get a written status `done`. It means the idea was built
  without a spec, and its reason must name the commit. Like `dropped`, a
  written `done` beats a spec link. The TUI Scratchpad gets a Done tab next to
  Specced and Dropped.
- `acta set <debt line> status done|wontfix --reason "<text>"` closes one
  debt line, for example `acta set DBT-0145.01 status wontfix --reason stale`.
  The reason is required. `done` writes `[x]` and `wontfix` writes `[-]`. The
  reason goes at the end of the line as ` (<status>: <reason>)`, for example
  ` (wontfix: stale)`. When the line already
  ends in a reason this command wrote, the new reason replaces it, so a second
  run never stacks two. Today `acta set` refuses debt lines, and only
  `acta debt drop-low` or a hand edit can close one. Old lines such as
  `- [x] ... (stale)` stay as they are.

Each `acta set` call commits on its own, like today. `apply` writes a whole
batch in one commit.

Closing statuses by kind:

| Kind | Built | Not valid or no longer needed |
| --- | --- | --- |
| Debt line | `[x]`, reason `done in <sha>` | `[-]`, reason `stale` or a short cause |
| Scratch item | `done`, reason names the commit | `dropped` with reason |
| Bug | `fixed` with `fixed_in` | `wontfix` with reason, for a bug that does not reproduce and has no fix commit |
| Plan | `done` through `--fix` only | `dropped` with reason |
| Spec | not a cleanup case | `dropped` with reason |

When `apply` drops a plan, it also drops the plan's spec in the same commit,
unless the spec has another plan whose status is not `done` or `dropped`. The
spec gets the reason `plan PLN-<n> dropped: <plan reason>`. The approval
message in the skill shows this spec change as a line under the plan, so the
user sees both writes before saying yes.

## The acta:cleanup skill

1. Run `acta cleanup`. Show the `--fix` changes and wait for a yes. Then run
   `acta cleanup --fix`.
2. Run `acta cleanup --json`. Each candidate costs an evidence check, so with
   more than 10 candidates in total the skill first tells the user the count per kind,
   the subagents it would start and the approval messages it needs, and asks:
   all, some kinds, or the N oldest. Nothing starts before the answer. Then it
   starts one read-only subagent per kind that has candidates in scope, all in
   parallel.
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

- Go tests for the candidate rules: each kind, the 14-day rule with chore
  commits counted, the `created` fallback, and the skip for items in another
  worktree and for plans with a linked worktree.
- Go tests for `--fix` and for `apply`: one commit per batch, a whole batch
  refused when one line fails a check, a plan with no tasks left alone, and the
  spec dropped with its plan only when no other open plan remains.
- Go tests for `acta set reason`, for scratch `done` beating a spec link, for
  a debt line moving to `[x]` or `[-]` with its suffix, and for a second run
  replacing the suffix.
- TUI tests for the Done tab and for the reason in the detail pane.
- The new skill lives in `plugin/skills/cleanup/`. It is added to the skill
  list in `internal/hook/hook.go` and `plugin/hooks/default-rules.md`, and it
  passes `internal/plugincheck`.
- Two eval cases. With a few candidates the agent sends proposals with
  evidence and writes nothing before the user approves. With more than 10 it
  says the count per kind and asks how much to check before any subagent
  starts.

## Plan split

Two plans. The first covers the CLI and the data: `reason`, scratch `done`,
debt lines through `acta set`, `acta cleanup`, `--fix`, `apply` and the TUI.
The second covers the skill, the hook list, the eval and the wiki page. It
runs after the first plan lands.
