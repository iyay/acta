---
name: cleanup
description: "acta: Use when the user asks to clean up, tidy or check stale planning items: scratch, bugs, debt lines, plans or specs that are done, dropped, closed, built or no longer valid. Proposes each change with evidence and writes only what the user approves."
---

# Cleanup

Cleanup closes planning items whose work is over. It never deletes a file: each close is a status and a one-line reason, and git keeps the history. Run it only when the user asks.

**Subagent models.** Name no model for any subagent; the harness default picks it.

## Steps

1. Run `acta cleanup`. When it lists fixes, show them and wait for a yes. Then run `acta cleanup --fix`. A file it names as not committed blocks the fix: tell the user and go on.
2. Run `acta cleanup --json`. Start one read-only subagent per kind that has candidates, all in parallel. Give each its list and the rules below.
3. Each subagent returns one proposal per item: keep, done, dropped, wontfix or fixed. Each proposal carries evidence (a `file:line`, a commit sha or a spec id) and a one-line reason. Subagents never write: no file edit, no `acta` write command, no commit.
4. Check every piece of evidence yourself. The sha must exist in this repo, and the `file:line` must hold what the proposal claims. Weak evidence becomes keep.
5. Send one message per kind in the current channel, at most 10 numbered items per message. Each item shows the proposal, the evidence and the reason. A plan drop also drops its spec when no other open plan of that spec is left: show that spec drop as a line under the plan. The user answers ok, or names the numbers to skip.
6. Pipe the approved items of that message to `acta cleanup apply` as one batch, one JSON object per line, so each batch is one commit. A refused batch writes nothing: read the error, fix or drop that line, and show the user the change.

Keep writes nothing. A skipped item stays as it is.

## What each close means

| Kind | Built | Not valid or no longer needed |
|---|---|---|
| Debt line | `done`, reason `done in <sha>` | `wontfix`, reason `stale` or a short cause |
| Scratch item | `done`, reason names the commit | `dropped` |
| Bug | `fixed` with `fixed_in` | `wontfix`, when it does not reproduce and has no fix commit |
| Plan | through `--fix` only | `dropped` |
| Spec | not a cleanup case | `dropped` |

A scratch item a spec already covers shows specced on its own: propose keep. `dropped` means the work was not done or is not valid; never drop finished work.

## Apply lines

```
{"id":"DBT-0012.03","status":"done","reason":"done in 4dcc44c"}
{"id":"BUG-0007","status":"fixed","fixed_in":"cb1f58b","reason":"runner stops at the final reply"}
{"id":"PLN-0031","status":"dropped","reason":"no work for 40 days, replaced by PLN-0044"}
```

`fixed_in` goes only with `fixed`. Every line needs a reason. An item a linked worktree is building cannot be closed from here.
