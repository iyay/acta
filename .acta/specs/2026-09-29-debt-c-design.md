---
id: SPC-0016
hash: a5kicgy
---
# Debt Group C: Land Guard, Bounded Graph, Fix Brief, Round Slug

Status: approved by the user on 2026-09-29 (Bounded). Closes DEBT-17.1, DEBT-17.2, DEBT-16.1 and DEBT-16.2, found as USER items in the debt triage of 2026-09-29.

## 1. Land checks the main checkout first (DEBT-17.2)

File: `plugin/skills/land/SKILL.md`, the Parent step.

Before the merge, the main checkout must be clean (`git status --porcelain` shows no tracked change; untracked files do not count) and `git rev-parse --abbrev-ref HEAD` must equal the parent branch. Either one fails: stop and report. Never check out, stash or reset to make it pass.

Why: without it, a main checkout that sits on another branch gets the merge.

## 2. Bounded review loops back to the short spec (DEBT-17.1)

File: `plugin/skills/brainstorm/SKILL.md`, the flow graph.

Bounded gets its own spec-review node, so its "changes requested" edge goes back to "Write short spec" and its "approved" edge goes to "Invoke acta:plan skill". The Architectural nodes and edges stay as they are.

## 3. Fix brief has no hand-filled range (DEBT-16.1)

File: `plugin/skills/dispatch/SKILL.md`, the fix-round step that says "REPLY-BACK over `<fixed-from>..<new-head>`".

It becomes: run `acta dispatch init` again, and the REPLY-BACK line is the same one sentence as in the brief. No range placeholder is left anywhere in `plugin/skills/dispatch/`.

## 4. Round defaults to a slug of the branch (DEBT-16.2)

File: `internal/cli/dispatch.go`, `cmdDispatchInit`.

When `--round` is empty, the branch name becomes a slug: lower case, every run of characters outside `a-z0-9` becomes one `-`, leading and trailing `-` are dropped, then it is cut to 64 characters (and a trailing `-` left by the cut is dropped). If the slug is empty or still fails `roundPattern`, exit 1 with `cannot make a round from branch "<branch>": pass --round <slug>` and write nothing. An explicit `--round` is never changed; it must match `roundPattern` as today. A detached `HEAD` becomes `head`, which is fine since the round is only a label.

## Testing

Each guard red first.

- Go, in `internal/cli/dispatch_test.go`: `feat/X` gives `feat-x`; `--a--b--` gives `a-b`; `HEAD` gives `head`; a 70-character name is cut to at most 64 and ends in no `-`; `///` exits 1 with the `pass --round` message and writes no record; `--round "Bad Round"` still exits 1.
- plugincheck: land names the clean-checkout and HEAD-equals-parent check; `<fixed-from>` does not appear in `plugin/skills/dispatch/`; the brainstorm graph has an edge from the Bounded review node back to "Write short spec".
- `gofmt -l .` empty, `go vet ./...` ok, `go test ./...` green.

## Out of scope

The other 29 USER items from the triage (groups A, B, D, E).
