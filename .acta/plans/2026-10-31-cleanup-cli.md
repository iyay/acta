---
parent: specs/2026-10-31-cleanup-command-design
depth: minimal
id: PLN-0201
created: "2026-10-31 07:34:53"
hash: b24s9l3
started: "2026-10-31 07:36:26"
finished: "2026-10-31 07:54:39"
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
- [x] Red: Board, write and CLI tests for reason round-trip, scratch `done` beating a link, and the refused inputs fail because no `reason` field or `done` scratch status exists.
- [x] Change: Parse `reason` into `Item.Reason`, add `done` to the scratch statuses and to `scratchStatus` next to `dropped`, accept `reason` in `SetValue` and print it in `acta show`.
- [x] Commit: `feat(board): reason field and scratch done status`.

### Task 2: close a debt line with acta set

**Files:** Create `internal/write/debtline.go`, `internal/write/debtline_test.go`; modify `internal/write/ops.go`, `internal/cli/cli.go`, `internal/cli/help.go`, `internal/cli/cli_test.go`.
**Seam:** `write.CloseDebtLine(cfg, b, id, status, reason string) (Outcome, error)` and `cli.Run` for `acta set <DBT-n.mm> status done|wontfix --reason "<text>"`.
**verify:** `done` writes `[x]` and `wontfix` writes `[-]` with ` (<reason>)` at the end of the line; a second run replaces the suffix this command wrote and never stacks two; a missing reason, another status, an unknown line id, or a reason with a newline or parentheses-only text is refused with no file change; old lines such as `- [x] ... (stale)` that were not written by this command keep their text; every other field on a debt line is still refused as today.
- [x] Red: Write and CLI tests for both statuses, the replace-not-stack rerun and each refused input fail because `SetValue` refuses every debt line.
- [x] Change: Add `CloseDebtLine` that rewrites one box and its reason suffix and commits like `acta set`, route debt-line `status` to it, and parse `--reason` in `cmdSet`.
- [x] Commit: `feat(write): close a debt line with a reason`.

### Task 3: TUI Done tab and reason in the detail pane

**Files:** Modify `internal/tui/sidebar.go`, `internal/tui/detail.go`, `internal/tui/sidebar_test.go`, `internal/tui/detail_test.go`.
**Seam:** `topTabs` done tabs and the detail pane rows built in `detail.go`.
**verify:** The Scratchpad shows Specced, Done and Dropped tabs, and each scratch item lands in exactly one of them by status; the detail pane shows a REASON row only when the item has a reason, for every kind that can carry one.
- [x] Red: Sidebar and detail tests fail because the Scratchpad has no Done tab and the detail pane has no REASON row.
- [x] Change: Add `{"Done", "done"}` to the Scratchpad done tabs and a REASON row next to FIXED in the detail pane.
- [x] Commit: `feat(tui): Done tab and reason row`.

### Task 4: acta cleanup lists candidates

**Files:** Create `internal/cleanup/candidates.go`, `internal/cleanup/candidates_test.go`, `internal/cli/cleanup.go`, `internal/cli/cleanup_test.go`; modify `internal/cli/cli.go`, `internal/cli/help.go`.
**Seam:** `cleanup.Candidates(b *board.Board, repo string, now time.Time) ([]Candidate, error)` with `Candidate{ID, Kind, Title, Path, Status, AgeDays, Why}`, and `cli.Run` for `acta cleanup [--json]`.
**verify:** Every rule in the spec's candidate list is a case: open debt line; raw scratch with no spec link; brainstorming scratch with no link that is quiet for 14 days and not one that is quiet for 13; open bug; in-progress plan with an open task and no commit for 14 days, where a chore tick commit or a Task trailer resets the clock and a missing `started` falls back to `created`; draft or approved spec with no plan created over 14 days ago. Items only in another worktree and plans with a linked worktree never appear. Output is grouped by kind, oldest first; `--json` prints one object per line with the fields `id`, `kind`, `title`, `path`, `status`, `age_days`, `why`. The command writes nothing.
- [x] Red: Candidate tests over a temp repo with dated commits fail because the `cleanup` package does not exist.
- [x] Change: Build `Candidates` from the board, `git log` dates for the scratch and plan files, `commits.Find` for Task trailers and the worktree list, then add the `cleanup` command with text and `--json` output.
- [x] Commit: `feat(cleanup): list stale planning candidates`.

### Task 5: acta cleanup --fix

**Files:** Create `internal/cleanup/fix.go`, `internal/cleanup/fix_test.go`; modify `internal/cli/cleanup.go`, `internal/cli/cleanup_test.go`, `internal/cli/help.go`.
**Seam:** `cleanup.Fix(cfg config.Config, b *board.Board) (Outcome, error)` and `cli.Run` for `acta cleanup --fix`.
**verify:** One commit removes every written status the board flags as `written status X ignored, derived Y` and marks `done` every plan with at least one task, all ticked, that shows neither `done` nor `review`; a plan with no tasks, a plan with a linked worktree, and a plan already done are left alone; with nothing to fix it makes no commit; when one change fails its check nothing is written.
- [x] Red: Fix tests fail because `--fix` is not a flag of `acta cleanup`.
- [x] Change: Collect every change, check them all, write the files and commit them with `gitc.CommitPaths` in one commit.
- [x] Commit: `feat(cleanup): fix statuses that need no judgment`.

### Task 6: acta cleanup apply

**Files:** Create `internal/cleanup/apply.go`, `internal/cleanup/apply_test.go`; modify `internal/cli/cleanup.go`, `internal/cli/cleanup_test.go`, `internal/cli/help.go`.
**Seam:** `cleanup.Apply(cfg config.Config, b *board.Board, decisions []Decision) (Outcome, error)` with `Decision{ID, Status, Reason, FixedIn}` read from JSON lines, and `cli.Run` for `acta cleanup apply` with stdin.
**verify:** A batch of decisions across debt lines, scratch, bugs, plans and specs lands as one commit with each status and reason written; `fixed` needs `fixed_in` and any other status refuses it; dropping a plan also drops its spec in the same commit with reason `plan PLN-<n> dropped: <plan reason>` unless another plan of that spec is not `done` or `dropped`. Each of these refuses the whole batch with no file change and names the failing line: malformed JSON, unknown id, a status not allowed for the kind, a missing or multi-line reason, a bad sha, a plan with a linked worktree.
- [x] Red: Apply tests fail because `apply` is not a subcommand of `acta cleanup`.
- [x] Change: Parse JSON lines into decisions, check every decision, write each item through the existing frontmatter and debt-line helpers without committing, add the spec cascade, then commit every touched path with `gitc.CommitPaths`.
- [x] Commit: `feat(cleanup): apply approved decisions in one commit`.

## Fix round 1

### Task 7: Round 1 review fixes

**verify:** No cleanup path reports success while it skipped or refused part of the user's input: every apply line holds exactly one JSON object, plain acta cleanup only reads and exits 0, a flag pair that cannot both be honoured is refused, and the cleanup tests run under testguard

- [ ] B1: internal/cleanup has no test file calling testguard.Watch(), so plugincheck TestEveryTestPackageWatches fails; add internal/cleanup/main_test.go like internal/board/main_test.go.
- [ ] B2: plain acta cleanup exits 1 when a --fix target file is dirty, because the read-only preview calls cleanup.Plan and maps its error to exitBadInput; the listing must stay read-only and exit 0, naming the dirty file instead of the fix.
- [ ] B3: ParseDecisions decodes only the first JSON value of a line, so a line with two objects applies the first and silently drops the second with exit 0; trailing content on a line must be an error that names the line, and the batch writes nothing.
- [ ] [fix] Spec and plan text say the debt suffix is " (<reason>)"; the code writes " (<status>: <reason>)". Also say SCR-0064 is already handled by the board, which derives done from the boxes for a written approved or in-progress plan.
- [ ] [fix] acta cleanup --json --fix ignores --json and prints text; refuse the pair as bad input.
- [ ] [fix] The spec cascade in withSpecs skips the Legacy and Worktree guard; leave a legacy or worktree-only spec alone instead of writing it.
- [ ] [fix] A brainstorming scratch with no commit shows "quiet 0 days"; fall back to its created date like a plan does.
- [ ] [fix] The apply commit subject counts the decisions before the cascade; count the items written.
- [ ] [fix] No board-package test reads the reason frontmatter field into Item.Reason.
- [ ] Commit with the trailer `Task: PLN-b24s9l3#7`

## State

### Next

Task 4: new package internal/cleanup; Candidates(b, repo, now). Item.Line, Item.Created/StartedOn strings; gitc has CommitPaths(In). commits.Find(repo, refs) keys hash#tasknum.

### Findings

One TUI run failed once with a clipboard screen dump and passed twice after with no change; the known TUI flake, not this work.
write.RemoveField leaves '{}' in the frontmatter when it removes the only field; real specs always carry id and hash, so --fix never hits it.

### Open rulings

User said yes (2026-10-31 07:52): a written dropped on a spec beats its plans; done in task 6.
