---
parent: debt/2026-10-03-dispatch-send
depth: minimal
closes: [SPC-0094, DBT-0071.01, DBT-0071.02, DBT-0071.03, DBT-0078.01]
id: PLN-0103
created: "2026-10-06 09:25:50"
hash: dd3qpd3
started: "2026-10-06 09:30:59"
finished: "2026-10-06 09:50:38"
---
# dispatch checkpoint verdict Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `checkpoint` in `internal/cli/dispatch_herdr.go` says drift only when omp's newest todo card really lists fewer or other tasks than the plan.

**Spec:** .acta/specs/2026-10-06-dispatch-checkpoint-verdict.md

**Tests:** `scripts/test ./internal/cli/ -run 'Checkpoint|Herdr'` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- `checkpoint(slug string, ids []string) (string, []string, string)` keeps its signature and its three verdict constants; the plan's task count is `len(ids)`.
- Tests use the fake herdr already in `internal/cli/dispatch_herdr_test.go`; no test calls the real herdr or omp.
- Run tests with `scripts/test`, never bare `go test`.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02
- Wave 3: Task 03
- Wave 4: Task 04
- Wave 5: Task 05

Tasks 01 to 04 all edit `internal/cli/dispatch_herdr.go` and its test file, so they run one after another.

### Task 01: Read only the newest round

**Files:** `internal/cli/dispatch_herdr.go`, `internal/cli/dispatch_herdr_test.go`

**verify:** No card drawn before the newest `goalMark` line can change the verdict, and pane text with no goal mark is always `unconfirmed`. List every pane shape tested (old card above new goal, no goal mark, new card only).

- [x] Red: a pane with an old round's card naming task 1 and 2 above the newest `🎯 Goal` line, and a new card naming only task 3 and 4, for ids `3,4`: today it reads the old card too; the test wants `ok` from the new part only. Also a pane with no goal mark wants `unconfirmed`.
- [x] Green: after the read, cut `text` to what follows the last line containing `goalMark`; with no such line return `unconfirmed`.
- [x] Commit: `fix(dispatch): checkpoint reads only the newest round (DBT-0071.03)`

### Task 02: Framed header and its task count

**Files:** `internal/cli/dispatch_herdr.go`, `internal/cli/dispatch_herdr_test.go`

**verify:** Every header shape omp draws (glyph header, unframed `[x] Todo`, framed `+--- [x] Todo 3 tasks ---+`) is found, and words like Todos, mytodo or a sentence about a todo are never a header. List every header form tested, matching and not matching.

- [x] Red: replace `TestHerdrCheckpointAsciiHeader`'s unframed-only case with a table that adds the framed ascii header; the framed one fails today.
- [x] Green: set `todoCardHeader` to `(?m)^[^\p{L}\p{N}\n]*(?:\[x\][^\p{L}\p{N}\n]*)?Todo(?:\s|$)` and add a helper `cardTotal(header string) (int, bool)` that reads `<N> tasks` from the header line.
- [x] Commit: `fix(dispatch): checkpoint finds the framed ascii todo header (DBT-0071.02)`

### Task 03: Folded list is not drift

**Files:** `internal/cli/dispatch_herdr.go`, `internal/cli/dispatch_herdr_test.go`

**verify:** With a task count in the header, `drift` comes only from a count below `len(ids)`; ids missing under a count at or above it give `unconfirmed`; with no count, today's rule holds. List every count and seen-ids combination tested.

- [x] Red: `Todo 12 tasks` with 8 rows shown for 12 ids wants `unconfirmed`; `Todo 3 tasks` for 5 ids wants `drift`; `Todo 2 tasks` naming both of 2 ids wants `ok`; a header with no count, some ids missing, wants `drift`.
- [x] Green: in `checkpoint`, use `cardTotal`; apply the three rules from the spec before the old some-missing rule.
- [x] Commit: `fix(dispatch): a folded todo list is unconfirmed, not drift (DBT-0071.01)`

### Task 04: Generic task text is not an id

**Files:** `internal/cli/dispatch_herdr.go`, `internal/cli/dispatch_herdr_test.go`

**verify:** No text outside the card's rows, and no `task <n> of <m>` phrase anywhere, can count as id n. List every decoy tested.

- [x] Red: a card whose only id-like text is `Task 1 of 3`, for ids `1,2`, wants `unconfirmed` (today `drift`); a `task 2` line above the header wants it not counted.
- [x] Green: look for ids only in lines after the header line, and drop matches followed by `\s+of\s+\d`.
- [x] Commit: `fix(dispatch): task N of M text is not a task id (DBT-0078.01)`

### Task 05: Version 0.1.26

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.26, and `internal/plugincheck` passes. List each file and the version it holds.

- [x] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [x] Green: bump the patch from 0.1.25 to 0.1.26 in all three files.
- [x] Commit: `chore: version 0.1.26`

## Polish

### Task 06: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] `internal/cli/dispatch_herdr.go`: move the header-and-rows split out of `checkpoint` into a small helper so `checkpoint` is under 50 lines, with no change in behaviour.
- [x] `internal/cli/dispatch_herdr.go`: the task-count regex reads a one-task header too (`Todo 1 task`); add that case to the folded-list test table.
- [x] Commit: `polish: review notes for PLN-0103`

## Review notes

- On a counted drift, the missing list can name ids omp listed but folded away; only the message changes, not the verdict.
- The header regex still matches a symbol-led sentence like "- Todo list here", as before this plan.
- A card header on the pane's last line gives no rows, so every id counts as missing; acceptable with nothing visible.
- missingIDs compiles one regexp per id per call; one call per send, so it does not matter.
- cardCount has no word break after tasks?, so "Todo 3 tasksomething" reads 3; omp never prints that.
- The polish split out two helpers (cardHeaderRows, missingIDs), one more than the NOTE asked; no behaviour change.
