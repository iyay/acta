---
parent: scratch/2026-10-07-tidy-commit-history
depth: minimal
closes: [SPC-0105]
id: PLN-0114
created: "2026-10-07 19:28:43"
hash: uyo6akt
started: "2026-10-07 19:32:45"
finished: "2026-10-07 20:38:24"
---
# Tidy commit history Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** With `commit_history: tidy` (the default), acta:land folds a branch into one commit per code task and moves the parent forward in a straight line, proven to hold the same files.

**Spec:** .acta/specs/2026-10-07-tidy-commit-history-design.md

**Tests:** `scripts/test ./internal/tidy/ ./internal/cli/ ./internal/config/ ./internal/setup/ ./internal/plugincheck/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand task and code first; then skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Git is driven through the real git binary, the way `internal/gitc` does it. `git merge-tree --write-tree --merge-base=` needs git 2.40 or newer; when git is older, `acta tidy` exits non-zero with a message that names the version.
- `acta tidy` never moves `<branch>`, the parent branch, or any ref outside `refs/acta/`. Commit messages, trailers and authors are copied unchanged.
- Tests build temp repos with `t.TempDir()`; never touch this repo or the real HOME.
- Run tests with `scripts/test`, never bare `go test`. Do not run `scripts/eval` inside a task.
- Never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 02
- Wave 2: Task 03, Task 04
- Wave 3: Task 05, Task 06

Each wave's tasks touch different files. Task 03 needs Task 02's package; Task 04 needs Task 01's key; Task 05 needs Task 03's command.

### Task 01: commit_history setting

**Files:** `internal/config/user.go`, `internal/config/user_test.go` (or the config test file that covers `plan_depth`), `internal/cli/config_cmd.go`, `internal/cli/config_repo_test.go`

**verify:** `commit_history` accepts only `tidy` and `full`, reads `tidy` when unset at user and repo level, and a repo value beats the user value. `acta config show` (text and `--json`) and `acta config set` handle it on every path `plan_depth` takes. List each path checked.

- [x] Red: tests for the default, both valid values, a bad value rejected with an error that names `commit_history`, the repo override, and the show/set output. Run `scripts/test ./internal/config/ ./internal/cli/ -run CommitHistory` and watch it fail.
- [x] Green: add `CommitHistory string \`yaml:"commit_history,omitempty"\`` next to `PlanDepth`, with the same validate, trim and repo-override code, and wire it into `config_cmd.go` beside `plan_depth`.
- [x] Commit: `feat(config): commit_history setting, tidy by default (SPC-0105)`

### Task 02: tidy package

**Files:** `internal/tidy/tidy.go`, `internal/tidy/remap.go`, `internal/tidy/tidy_test.go`, `internal/tidy/remap_test.go`

**verify:** For any branch, the new chain's final tree equals the branch tree plus only the hash-remap edits in planning-root files, or `Run` returns an error and writes no ref. No commit is lost: every branch commit's change sits in exactly one new commit. List each fold path (before the first kept commit, after it, after the review marker, on a replay clash, `--onto` parent chores) and the test that covers it.

- [x] Red: tests on temp repos for: the keep rule (touches a file outside the planning root and type not `chore`/`docs`/`polish`/`wiki`); a planning-only commit before the first kept commit rides forward; one after folds back; every commit after the first review-notes or `polish` commit folds into the last kept commit; a replay clash folds into the commit before; dates never go backward; messages and authors unchanged; short and full old hashes in planning files remapped to the new commit at the same length, unknown hashes left; a forced tree mismatch returns an error and no `refs/acta/tidy/<branch>`; `Onto` replays `chore(...)` parent commits first and folds them forward; a non-chore parent commit makes `Onto` fold nothing. Run `scripts/test ./internal/tidy/` and watch it fail.
- [x] Green: `func Run(repo string, opt Options) (Result, error)` with `Options{Base, Branch, Onto, PlanningRoot string}` and `Result{OldCount, NewCount int; Tip string}`. Replay with `git merge-tree --write-tree --merge-base=<c>^ <current> <c>`; build commits with `git commit-tree` and the kept commit's author/committer env; write `refs/acta/tidy/<branch>` only after the tree check passes. `remap.go` holds the old-to-new map and the file rewrite.
- [x] Commit: `feat(tidy): fold a branch into one commit per code task (SPC-0105)`

### Task 03: acta tidy command

**Files:** `internal/cli/tidy.go`, `internal/cli/tidy_test.go`, `internal/cli/cli.go`

**verify:** `acta tidy <base> <branch> [--onto <ref>]` prints exactly `tidy: <old> commits -> <new>, tree ok` and exits 0 on success; on any error (bad args, unknown ref, old git, tree mismatch) it exits non-zero, prints why on stderr, and leaves every ref as it was. List each exit path checked.

- [x] Red: CLI tests on a temp repo for success output, missing args, unknown ref, and a mismatch. Run `scripts/test ./internal/cli/ -run Tidy` and watch it fail.
- [x] Green: parse args, resolve the planning root the way other commands do (`internal/config`), call `tidy.Run`, add `case "tidy":` to `cli.go` and a line in its usage text.
- [x] Commit: `feat(cli): acta tidy command (SPC-0105)`

### Task 04: setup asks commit_history

**Files:** `internal/setup/form.go`, `internal/setup/plan.go`, `internal/setup/form_test.go`, `internal/setup/plan_test.go`

**verify:** The wizard asks commit history once, defaults to the current value or `tidy`, and the saved user config holds the answer. No other question changes order or wording. List each check.

- [x] Red: tests for the new question title and options, its default, and the saved value. Run `scripts/test ./internal/setup/` and watch it fail.
- [x] Green: add the question after the plan depth question, with options `tidy` (one commit per task, straight history) and `full` (every commit plus a merge commit).
- [x] Commit: `feat(setup): ask for commit history (SPC-0105)`

### Task 05: land skill tidy path and wiki

**Files:** `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_land_test.go`, `.acta/wiki/commit-subject-form.md`, `.acta/wiki/land-reuse-gates-planning-only.md`, `.acta/wiki/land-tidy-history.md`, `.acta/wiki/index.md` and `.acta/wiki/log.md` if the wiki keeps them

**verify:** With `commit_history: tidy` the land skill runs gates, then `acta tidy`, picks the fold point (`@{upstream}`, else `refs/acta/last-land`, else none), moves the parent with `merge --ff-only` or the guarded `reset --keep`, sets `refs/acta/last-land`, puts the last task commit in `fixed_in`, and deletes the branch with `-D` only after tidy passed; with `full` it reads as today. A failed tidy never falls back to a plain merge. The skill stays under the plugincheck caps. List each step checked.

- [x] Red: `skill_land_test.go` requires the land skill to name `acta tidy`, `commit_history`, `refs/acta/last-land`, `--ff-only`, `reset --keep`, and the no-fallback rule. Run `scripts/test ./internal/plugincheck/ -run Land` and watch it fail.
- [x] Green: add the tidy path to the land skill (steps 5, 7 and 8 split by mode), rewrite the two wiki pages and add `land-tidy-history` (keep rule, fold rules, why `-D` is safe), following the wiki format in `plugin/skills/build/wiki.md`. Run `acta wiki check`.
- [x] Commit: `feat(land): tidy history before landing (SPC-0105)`

### Task 06: Version 0.1.36

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.36, and `internal/plugincheck` passes. List each file and the version it holds.

- [x] Green: change 0.1.35 to 0.1.36 in all three files. Run `scripts/test ./internal/plugincheck/`.
- [x] Commit: `chore: version 0.1.36`

## State

### Next

Took over from omp (tab closed). Task 01 done (e950533) by omp.
Task 02 and Task 04 running as Claude sonnet subagents in parallel.
Then Task 03 (needs 02), Task 05 (needs 03), Task 06.

### Findings

Review round 1: Standards axis BLOCKED. tidy.Run builds on the current base tip but forces the last tree to the branch tree, so parent commits made after the fork are undone (tidy.go:100,129). Confirmed by reading the code. Waiting on Spec axis before the fix round.

## Fix round 1

### Task 07: Tidy keeps parent work the branch does not have

**Files:** `internal/tidy/tidy.go`, `internal/tidy/tidy_test.go`, `internal/cli/tidy.go`, `internal/cli/tidy_test.go`, `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_land_test.go`, `plugin/README.md`, `.acta/wiki/land-tidy-history.md`

**verify:** For any base and branch, the tidy tip's tree equals the three-way merge of the base tip and the branch (from their merge base) plus only remap edits, so no parent commit's content is lost on either the `--ff-only` or the `--onto`/`reset --keep` path; when that merge clashes, tidy exits non-zero and writes no ref. No git failure is ever read as a clash or ignored. List every path checked: parent moved with a code commit, parent moved with chore commits only (with and without `--onto`), merge clash, unknown `--onto` ref, a git error during replay, a bad branch name.

- [x] Red: tests in `internal/tidy` for a parent that moved after the fork (code commit on base: the tidy tip keeps it and adds the branch change), the `--onto` path with chore commits on the parent (their content survives in the tip), a clashing parent change (error, no ref), an unknown `--onto` ref (error), and a bad branch name (error before any commit is built); CLI test for the new output line. Watch them fail.
- [x] Green: the target tree is `git merge-tree --write-tree --merge-base=<merge-base of base and branch> <base> <branch>`; a clash returns an error. The last node takes the target tree (plus remap) and `proof` compares against the target, not the branch tree. Walk `<merge-base>..<branch>`. `replay` returns real git errors (exit code other than 1). `ontoStart` returns an error for an unknown ref or a failed rev-list. Check the branch with `git check-ref-format --branch` first. Put `--end-of-options` before user revs in `rev-parse`. Fix the `removeIndex` comment to say plainly why the error is ignored. The CLI line becomes `tidy: <old> commits -> <new>, tree ok, parent <short sha of base read>, folded <n>` where folded counts parent chore commits folded.
- [x] Green: land skill tidy path: pick the fold point before running tidy; record the parent tip from the `tidy:` line and check it before moving the parent; use `reset --keep` only when `folded` is above 0, else `--ff-only`; reuse gates only when `git diff --name-only <branch> refs/acta/tidy/<branch>` lists only planning-root paths, else run the gates on the tidy tip before moving the parent. Update the land skill frontmatter description and `plugin/README.md` so they no longer say every land is `merge --no-ff`. Update `land-tidy-history.md` (why `-D` is safe now rests on the merge proof). Extend `skill_land_test.go` for the new rules. Stay under the plugincheck caps.
- [x] Commit: `fix(tidy): keep parent work the branch does not have (PLN-0114 fix round 1)`

## Polish

### Task 08: Review polish

**verify:** Each NOTE below is applied, and nothing else changes.

- [x] Land skill tidy step 5 compares like with like: `git rev-parse --short=7 <parent>` against the `parent` sha of the `tidy:` line.
- [x] Land skill tidy step 3: a red gate on the tidy tip stops the land.
- [x] Land skill Red Flags gets back "Relying on partial verification" and the Rationalization row "Partial check is enough → Partial proves nothing", within the current byte cap.
- [x] `TestBadBranchNameIsRefused` uses a name that exists as a ref but that `check-ref-format --branch` refuses, or asserts the "not a valid branch name" message, so it fails when the check is removed.
- [x] Doc comments in `internal/tidy/tidy.go` match the code: `replay` (returns clash and err), `proof` (compares against the merge of base and branch), and `Result` (what `Parent` and `Folded` mean).
- [x] Commit: `polish: review notes for PLN-0114`

## Review notes

- Round 1 BLOCKED (both axes): tidy forced the last tree to the branch tree, so parent work made after the fork was undone. Fixed in fix round 1; round 2 CLEAN on both axes; polish CLEAN on both axes.
- Hashes whose change lands in the last tidy commit stay old in planning files, since a commit cannot name its own hash; land takes fixed_in from refs/acta/tidy/<branch>.
- Task 05 raised the land skill byte cap in internal/plugincheck/budget_test.go from 6962 to 9100, a file the plan did not list.
- A branch named HEAD is now refused by acta tidy; land always passes a real branch name.
- plan() in internal/tidy/tidy.go shadows err from replay with an inner var; it works.
- The polish reworded land steps 1 to 4 and 7 to fit the cap, with no rule removed and no meaning changed.
