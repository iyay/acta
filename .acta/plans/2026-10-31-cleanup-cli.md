---
parent: specs/2026-10-31-cleanup-command-design
depth: minimal
id: PLN-0201
created: "2026-10-31 07:34:53"
hash: b24s9l3
---
# Cleanup CLI and data implementation plan

**Goal:** Add the `reason` field, the scratch `done` status, closing debt lines through `acta set`, and the `acta cleanup` command with `--json`, `--fix` and `apply`.
**Spec:** `.acta/specs/2026-10-31-cleanup-command-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Cleanup never deletes a file; every change is a status change with a one-line reason, and git keeps the history.
`--fix` and `apply` check every change before the first write; when any check fails they write nothing, print each failing line, and exit non-zero.
Write tests run in a temp clone, because write commands auto-commit.
Help text in `internal/cli/help.go` lists every new command and flag; `acta help cleanup` works.

## Waves

Wave 1: Task 1. Wave 2: Task 2 and Task 3 (no shared files). Wave 3: Task 4. Wave 4: Task 5. Wave 5: Task 6.

### Task 1: reason field and scratch done status

**Files:** Modify `internal/board/board.go`, `internal/board/scratch.go`, `internal/write/ops.go`, `internal/cli/cli.go`, `internal/cli/help.go`; tests in `internal/board/`, `internal/write/`, `internal/cli/`.
**Seam:** `board.Load` items (`Item.Reason`, scratch `Status`), `write.SetValue(cfg, b, id, "reason"|"status", value)`, `cli.Run` for `set` and `show`.
**verify:** A scratch item, bug, spec or plan round-trips `reason` through `acta set` and `acta show`; a written scratch `done` shows `done` with or without a spec link, exactly like `dropped`; `reason` with an empty value, a newline, or on a task or debt item is refused; `specced` stays unwritable.
- [ ] Red: Board, write and CLI tests for reason round-trip, scratch `done` beating a link, and the refused inputs fail because no `reason` field or `done` scratch status exists.
- [ ] Change: Parse `reason` into `Item.Reason`, add `done` to the scratch statuses and to `scratchStatus` next to `dropped`, accept `reason` in `SetValue` and print it in `acta show`.
- [ ] Commit: `feat(board): reason field and scratch done status`.

### Task 2: close a debt line with acta set

**Files:** Create `internal/write/debtline.go`, `internal/write/debtline_test.go`; modify `internal/write/ops.go`, `internal/cli/cli.go`, `internal/cli/help.go`, `internal/cli/cli_test.go`.
**Seam:** `write.CloseDebtLine(cfg, b, id, status, reason string) (Outcome, error)` and `cli.Run` for `acta set <DBT-n.mm> status done|wontfix --reason "<text>"`.
**verify:** `done` writes `[x]` and `wontfix` writes `[-]` with ` (<reason>)` at the end of the line; a second run replaces the suffix this command wrote and never stacks two; a missing reason, another status, an unknown line id, or a reason with a newline or parentheses-only text is refused with no file change; old lines such as `- [x] ... (stale)` that were not written by this command keep their text; every other field on a debt line is still refused as today.
- [ ] Red: Write and CLI tests for both statuses, the replace-not-stack rerun and each refused input fail because `SetValue` refuses every debt line.
- [ ] Change: Add `CloseDebtLine` that rewrites one box and its reason suffix and commits like `acta set`, route debt-line `status` to it, and parse `--reason` in `cmdSet`.
- [ ] Commit: `feat(write): close a debt line with a reason`.

### Task 3: TUI Done tab and reason in the detail pane

**Files:** Modify `internal/tui/sidebar.go`, `internal/tui/detail.go`, `internal/tui/sidebar_test.go`, `internal/tui/detail_test.go`.
**Seam:** `topTabs` done tabs and the detail pane rows built in `detail.go`.
**verify:** The Scratchpad shows Specced, Done and Dropped tabs, and each scratch item lands in exactly one of them by status; the detail pane shows a REASON row only when the item has a reason, for every kind that can carry one.
- [ ] Red: Sidebar and detail tests fail because the Scratchpad has no Done tab and the detail pane has no REASON row.
- [ ] Change: Add `{"Done", "done"}` to the Scratchpad done tabs and a REASON row next to FIXED in the detail pane.
- [ ] Commit: `feat(tui): Done tab and reason row`.

### Task 4: acta cleanup lists candidates

**Files:** Create `internal/cleanup/candidates.go`, `internal/cleanup/candidates_test.go`, `internal/cli/cleanup.go`, `internal/cli/cleanup_test.go`; modify `internal/cli/cli.go`, `internal/cli/help.go`.
**Seam:** `cleanup.Candidates(b *board.Board, repo string, now time.Time) ([]Candidate, error)` with `Candidate{ID, Kind, Title, Path, Status, AgeDays, Why}`, and `cli.Run` for `acta cleanup [--json]`.
**verify:** Every rule in the spec's candidate list is a case: open debt line; raw scratch with no spec link; brainstorming scratch with no link that is quiet for 14 days and not one that is quiet for 13; open bug; in-progress plan with an open task and no commit for 14 days, where a chore tick commit or a Task trailer resets the clock and a missing `started` falls back to `created`; draft or approved spec with no plan created over 14 days ago. Items only in another worktree and plans with a linked worktree never appear. Output is grouped by kind, oldest first; `--json` prints one object per line with the fields `id`, `kind`, `title`, `path`, `status`, `age_days`, `why`. The command writes nothing.
- [ ] Red: Candidate tests over a temp repo with dated commits fail because the `cleanup` package does not exist.
- [ ] Change: Build `Candidates` from the board, `git log` dates for the scratch and plan files, `commits.Find` for Task trailers and the worktree list, then add the `cleanup` command with text and `--json` output.
- [ ] Commit: `feat(cleanup): list stale planning candidates`.

### Task 5: acta cleanup --fix

**Files:** Create `internal/cleanup/fix.go`, `internal/cleanup/fix_test.go`; modify `internal/cli/cleanup.go`, `internal/cli/cleanup_test.go`, `internal/cli/help.go`.
**Seam:** `cleanup.Fix(cfg config.Config, b *board.Board) (Outcome, error)` and `cli.Run` for `acta cleanup --fix`.
**verify:** One commit removes every written status the board flags as `written status X ignored, derived Y` and marks `done` every plan with at least one task, all ticked, that shows neither `done` nor `review`; a plan with no tasks, a plan with a linked worktree, and a plan already done are left alone; with nothing to fix it makes no commit; when one change fails its check nothing is written.
- [ ] Red: Fix tests fail because `--fix` is not a flag of `acta cleanup`.
- [ ] Change: Collect every change, check them all, write the files and commit them with `gitc.CommitPaths` in one commit.
- [ ] Commit: `feat(cleanup): fix statuses that need no judgment`.

### Task 6: acta cleanup apply

**Files:** Create `internal/cleanup/apply.go`, `internal/cleanup/apply_test.go`; modify `internal/cli/cleanup.go`, `internal/cli/cleanup_test.go`, `internal/cli/help.go`.
**Seam:** `cleanup.Apply(cfg config.Config, b *board.Board, decisions []Decision) (Outcome, error)` with `Decision{ID, Status, Reason, FixedIn}` read from JSON lines, and `cli.Run` for `acta cleanup apply` with stdin.
**verify:** A batch of decisions across debt lines, scratch, bugs, plans and specs lands as one commit with each status and reason written; `fixed` needs `fixed_in` and any other status refuses it; dropping a plan also drops its spec in the same commit with reason `plan PLN-<n> dropped: <plan reason>` unless another plan of that spec is not `done` or `dropped`. Each of these refuses the whole batch with no file change and names the failing line: malformed JSON, unknown id, a status not allowed for the kind, a missing or multi-line reason, a bad sha, a plan with a linked worktree.
- [ ] Red: Apply tests fail because `apply` is not a subcommand of `acta cleanup`.
- [ ] Change: Parse JSON lines into decisions, check every decision, write each item through the existing frontmatter and debt-line helpers without committing, add the spec cascade, then commit every touched path with `gitc.CommitPaths`.
- [ ] Commit: `feat(cleanup): apply approved decisions in one commit`.

## State
