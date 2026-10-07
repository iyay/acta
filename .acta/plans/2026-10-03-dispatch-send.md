---
parent: specs/2026-10-03-review-and-skill-diet-design
depth: minimal
id: PLN-0079
created: "2026-10-03 21:01:58"
hash: pv9qwfx
started: "2026-10-03 21:05:18"
finished: "2026-10-03 21:17:56"
---
# Dispatch Send and Close Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** One `acta dispatch send` call writes the brief and the record, finds or makes the omp tab, sets the goal and checks the todo list, `acta dispatch close` closes the tab, and one short `build/dispatch.md` replaces the two dispatch files.

**Spec:** `.acta/specs/2026-10-03-review-and-skill-diet-design.md` (section 1, plan 1 of 3)

**Tests:** fast `scripts/test ./internal/cli` or `scripts/test ./internal/plugincheck` (the package a task names), full `scripts/test --full`; land also runs `scripts/eval` with the branch binary on PATH.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- `acta dispatch init` and `acta reply-back` keep working as they are. `send` reuses the record code and `checkRecord` in `internal/cli/dispatch.go`; it never writes a record by another path.
- Every herdr call goes through `exec.Command` with separate arguments, never a shell string. Plan text, note text and pane text never become shell text.
- The send never sends `/new`. A fresh `omp` process is the new session.
- The waits (ready at most 60 s, goal mark, about 20 s before the checkpoint read, the poll step) are package variables, so tests set them to milliseconds. No test sleeps for real seconds.
- Exit codes: 0 ok or checkpoint unconfirmed; 1 (`exitBadInput`) bad input, nothing sent; 2 delivery failed, with herdr's reason; 4 checkpoint drift. The spec's "1 for drift" moves to 4 because 1 already means bad input in this CLI; the spec line is fixed in the same commit as this plan.
- Herdr names match `[a-z][a-z0-9_-]{0,31}`. The slug is `roundFromBranch(branch)` cut to 32 characters; a slug that still does not match is refused.
- Every run step uses `scripts/test <package> -run <Name>`, never bare `go test` and never `./...`.
- Comments are plain English a 10-year-old can read. They say why, not what.

## Waves

- Wave 1: Task 1, Task 2 (no shared file: the brief in `dispatch_brief.go`, herdr in `dispatch_herdr.go`, each with its own test file and its own fake).
- Wave 2: Task 3, Task 4 (Task 3 is `internal/cli`, Task 4 is `plugin/` and `internal/plugincheck`; Task 3 needs Tasks 1 and 2).
- Wave 3: Task 5 (the manifests; plugincheck reads them, so it runs alone after Task 4).

### Task 1: The brief comes from the plan

**Files:**
- Create: `internal/cli/dispatch_brief.go`, `internal/cli/dispatch_brief_test.go`

**verify:** For every plan the board can parse, the brief either names every task the round covers, each with its own `verify:` line, plus the plan's `## Waves` text as written, or the call returns an error and writes no file. No round ever picks a task outside its range: no `--round` takes the tasks before the first `## Fix round`, any other round takes only the tasks under the last `## Fix round <n>`, a missing fix round section is an error, and `polish` takes no tasks and needs a non-empty note. Every brief holds the plan and spec paths, worktree, branch, parent, base, the absolute `house-rules.md` path, each `MEMORY.md` path that exists (`~/.claude/memory/MEMORY.md` and `~/.claude/projects/<main checkout path with every / and . turned into ->/memory/MEMORY.md`), the plan's `**Tests:**` line, the note, `SKILL: load build`, the line "never run bare `go test`" with the pre-tool hook reason, and the one REPLY-BACK line. List every round kind and every refusal checked.

- [x] Failing test: table tests over small plan files in a temp dir (first dispatch, fix round, two fix rounds, missing fix round, polish with and without a note, a task with no `verify:`, a missing rules file) that call `buildBrief` and check the text and the errors; they fail because `buildBrief` does not exist.
- [x] Code: `buildBrief(planPath string, src []byte, in briefInput) (string, error)` reads tasks with `board.Parse`, picks the round's tasks by heading line, pulls each `**verify:**` line and the `## Waves` and `**Tests:**` text, and fills the brief; `briefInput` holds round, note, rules path, worktree, branch, parent, base and home.
- [x] Commit: `dispatch: build the brief from the plan`

### Task 2: herdr steps in Go

**Files:**
- Create: `internal/cli/dispatch_herdr.go`, `internal/cli/dispatch_herdr_test.go` (with its own scripted fake herdr, so `fakeHerdr` in `dispatch_test.go` stays as it is)

**verify:** No herdr path can send a goal to the wrong pane or wait forever. For every outcome of `herdr agent get <slug>` (found idle, found working, not found, herdr error), the steps either reuse the found pane, refuse a working agent with nothing sent, or make one tab with `--workspace` from `$HERDR_PANE_ID`, `--label <slug>`, `--no-focus` and `--cwd <worktree>`, run `omp` in it and rename it to the slug. `/new` is never sent. Every wait ends at its limit with a clear error: omp not ready, goal mark not shown after one retry. A held goal (`⏸ Goal` or "Resume the current goal first") is cleared with `/goal drop` and two enters before the goal is sent. The checkpoint read happens exactly once and gives `ok`, `drift` with the missing task ids, or `unconfirmed` when no todo list shows. List every herdr call sequence checked, with the calls in order.

- [x] Failing test: a scripted fake herdr that logs each call and answers `agent get`, `agent read` and `tab create` from numbered reply files, plus cases for each outcome above, all with millisecond waits; they fail because the functions do not exist.
- [x] Code: `findOrMakeTab(slug, worktree string) (pane string, err error)`, `deliverGoal(slug, goal string) error` (ready wait, held goal clear, prompt, goal mark check with one retry) and `checkpoint(slug string, ids []string) (verdict string, missing []string, pane string)`, each through one `herdr(args ...string) (string, error)` helper.
- [x] Commit: `dispatch: drive the herdr steps in Go`

### Task 3: `acta dispatch send` and `acta dispatch close`

**Files:**
- Create: `internal/cli/dispatch_send.go`, `internal/cli/dispatch_send_test.go`
- Modify: `internal/cli/cli.go` (the `dispatch` case, usage text, the unknown-command list), `internal/cli/dispatch.go` (move the record writing out of `cmdDispatchInit` into `writeRecord`, so init and send share it)

**verify:** `send` never sends anything to herdr until the record and the brief are both written and checked, and every refusal (no `HERDR_ENV=1`, no `$HERDR_PANE_ID`, not a git repo, bad plan, bad round, no rules file, bad slug, a brief error) exits 1 with nothing sent. On success it prints exactly these lines: slug, pane, base, brief path, checkpoint verdict, and the watcher command `herdr agent wait <slug> --until idle --until done`; drift prints the pane text and exits 4; a herdr failure exits 2 with herdr's reason. `close` resolves the pane from the slug and runs `herdr pane close <pane>`, and refuses when the slug does not resolve. `dispatch init` output and behaviour are unchanged. List every flag, every refusal and every exit code checked.

- [x] Failing test: end-to-end tests in a temp git repo with a plan and the Task 2 scripted fake herdr: first send, fix round send (no `tab create`, no `pane run`), polish with `--note-file -` from stdin, each refusal, drift, herdr failure, `close` found and not found; plus the existing `dispatch init` tests still pass; they fail because `send` and `close` do not exist.
- [x] Code: `cmdDispatchSend` with `--plan`, `--round`, `--note-file` (a path or `-`), `--rules` (absolute path of `house-rules.md`, must exist), then `writeRecord`, `buildBrief` to `.claude/dispatch/<slug>-brief.md`, `findOrMakeTab`, `deliverGoal` with the goal text (one-line plan title, `ultrathink orchestrate` as plain words, the brief path, the REPLY-BACK line), `checkpoint`, and the printed lines; `cmdDispatchClose` with no flags.
- [x] Commit: `dispatch: add send and close`

### Task 4: One short dispatch.md

**Files:**
- Modify: `plugin/skills/build/dispatch.md` (rewrite)
- Delete: `plugin/skills/build/herdr-delivery.md`
- Test: `internal/plugincheck/build_dispatch_test.go`, `internal/plugincheck/reply_back_test.go` (`TestDispatchBriefLoadsBuild` and the `herdr-delivery.md` check go: Task 1 tests the brief text now), `internal/plugincheck/budget_test.go` (drop the `herdr-delivery.md` cap, set the `dispatch.md` cap to its new size, at most 8000 bytes)

**verify:** No rule the old two files enforced is lost without a home: each one is either in the new `dispatch.md`, tested in Go by Tasks 1 to 3, or owned by the skill the spec names (`acta:slice` for property verify lines, `acta:review` for the stopping rule, `references/house-rules.md` for spreading work), and the only rules dropped are `/new`, the "New session started" check and the wave check, as the spec says. The new file says: it applies only with `HERDR_ENV=1`, else `dispatch` runs as `subagent`; run `acta dispatch send` with `--rules` set to this skill's base dir plus `../../references/house-rules.md`; read its lines and exit code; never wait, end the turn; start the printed watcher in the background and handle its four outcomes; verify from git; fix rounds and polish with `send --round`; `acta dispatch close` after `acta:land`; `Bugs found by recipient`; harvest omp memory; the advisor in one line. The file is at most 8000 bytes. List every pinned phrase of the old tests and where its rule lives now.

- [x] Failing test: rewrite the pins in `build_dispatch_test.go` to the new file (keep the bans list and add `herdr-delivery.md`, `New session started` and `/new` to it), drop the moved pins, and lower the caps; they fail because the old files still exist and are too big.
- [x] Code: write the new `dispatch.md` from scratch in plain short sentences, and `git rm` `herdr-delivery.md`.
- [x] Commit: `build: one short dispatch.md on top of acta dispatch send`

### Task 5: Version 0.1.3

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three manifests say `0.1.3` and plugincheck's version test passes; no other file pins a version.

- [x] Failing test: none needed; plugincheck's existing version test is the check, run before and after.
- [x] Code: change `0.1.2` to `0.1.3` in the three files.
- [x] Commit: `version 0.1.3`

## Fix round 1

Review round 1 (`3476539..d6f6698`): Spec axis BLOCKED, Standards axis CLEAN. Two BLOCKERs and four `[fix]` NOTEs, all in this one task.

### Task 6: Checkpoint sees a made-up todo list, close runs before land, and four small fixes

**Files:**
- Modify: `internal/cli/dispatch_herdr.go`, `internal/cli/dispatch_herdr_test.go`, `internal/cli/dispatch_send.go`, `internal/cli/dispatch_send_test.go`, `internal/cli/dispatch_brief.go`, `plugin/skills/build/SKILL.md` (the `## Close` paragraph), `plugin/skills/review/SKILL.md` (the polish dispatch line), `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/build_dispatch_test.go` or `skill_build_test.go` (one pin for the close order), `internal/plugincheck/budget_test.go` (only if a cap must move; caps may only go down or stay)

**verify:** (1) The checkpoint never reports `ok` or `unconfirmed` for a todo list that is on screen but leaves out task ids: `unconfirmed` only when the read shows no omp todo card (the card's title line is `Todo`, see `pi-tui/src/tools/todo.ts` in omp 18), and any missing id on a screen with that card is `drift` with the missing ids, including when every id is missing. List every pane shape checked (no card, card with all ids, card with some, card with none, ids in text but no card). (2) No skill text tells the orchestrator to run `acta dispatch close` after `acta:land`: build `SKILL.md` `## Close` and `dispatch.md` agree that close runs from the worktree before land, and a plugincheck pin guards it. (3) The brief's parent line names the branch the main checkout is on, not a fixed `main`; a main checkout on `master` gives `master`. (4) `review/SKILL.md` sends a dispatched polish with `acta dispatch send --round polish`, and its pin matches. (5) `isNotFound` treats only herdr's `agent_not_found` answer as a missing agent, as its comment says; any other "not found" text is a herdr error. (6) No comment in the diff names a plan task number. List every file changed.

- [x] Failing test: add checkpoint cases for each pane shape in (1), a send test with the main checkout on `master`, an `isNotFound` case for "socket not found", and the two plugincheck pins for (2) and (4); they fail on today's code.
- [x] Code: detect the `Todo` card line in `checkpoint`; read the parent with `gitIn(mainCheckout, "rev-parse", "--abbrev-ref", "HEAD")`; narrow `isNotFound`; reword the `## Close` sentence in build `SKILL.md` and the polish line in review `SKILL.md`; fix the `briefTasks` comment.
- [x] Commit: `dispatch: fix round 1 (made-up todo list is drift, close before land, real parent)`

## Review notes

- deliverGoal retry: when the first /goal was accepted but its mark took longer than herdrGoalWait, a second /goal goes on top of the running goal.
- exitDelivery = 2 shares its number with exitSkipped in cli.go.
- acta dispatch close does not check HERDR_ENV; outside herdr it exits 2, which is harmless and matches reply-back.
- waitIdle counts agent status done as ready; that matches herdr's meaning and TestHerdrDeliverGoalDoneAgentIsReady pins it.
- The brief no longer carries the recipient-side "never write NOTEs to memory" line or the sandbox hint to copy house-rules.md next to the brief.
- While omp still streams the todo call, the pane shows only the `Todo init` header, so a checkpoint read in that short window reports drift on every id.
- An omp todo error card (`✗ Todo`) counts as a card, so it reports drift with every id missing.
- A main checkout on a detached head refuses with exit 3, like the other git read failures.
- newSendEnvOn in dispatch_send_test.go repeats git config the worktree already shares through the common dir.
- Task 4's verify line says close after acta:land; Task 6 replaced that with close from the worktree right before land, and the code and skills follow Task 6.
- dispatch.md "Close the tab first (below)" is vague on its own; the next lines make it clear.
- dispatch.md drift line says real drift "skips wave 1 tasks"; a skipped later-wave task that is on screen is no longer named.
- The ascii header regex matches lowercase `[x]` only; omp draws it lowercase today.
- The drift line says "omp shows 8 rows"; it is 8 rows per phase.
- The `t-` id branch also matches inside hyphenated words like commit-t-1; the cost is only a missed drift, never a false one.
- An empty todo list drawn as unframed `[x] Todo 0 tasks` now reports drift, as the unicode presets already did.
- The exit 3 line says "a git read failed", but exit 3 also comes from a missing home folder or a failed record or brief write; the action is the same.
- dispatch.md exit 4 says "yield" without saying the printed watcher still starts.
- The removed line "A fact that belongs in the plan goes in the plan." was the only steer of plan facts out of the note file.
- In the two timing tests a wait that never stops fails only through the go test timeout, since the elapsed checks run after waitFor returns.
- Full suite at c904909+polish first failed on TestHerdrFindOmpNeverStartsEndsAtTheLimit (30 ms ready limit, one shell per scripted herdr call under load); 0fc6b4d widened two timing tests, then `scripts/test --full -count=1` was green.
