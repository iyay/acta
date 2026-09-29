---
id: PLN-0014
hash: qzjdymo
---
# Rename pm and pmb to acta Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** The plugin and the CLI are both called `acta`: skills `acta:*`, binary `acta` (with a `pmb` alias that warns), root folder `.acta/` (still reading `.pm/`), `.acta.yaml` / `ACTA_ROOT` / `~/.acta/voice.yaml` (still reading the old names), Go module `github.com/iyay/acta`, and a `acta migrate-root` command; this repo's own `.pm/` moves to `.acta/` last.

**Architecture:** Mechanical rename in four ordered steps so each step keeps the suite green: module path, then CLI binary and alias, then config/root/voice fallbacks plus `migrate-root`, then plugin text; the repo's own root folder moves in the final task with the new command.

**Tech Stack:** Go 1.27, git CLI.

**Spec:** `.pm/specs/2026-09-26-rename-acta-design.md`

**Worktree:** `/Users/iyay/Nayakatara/pm-board-rename-acta`, branch `rename-acta`, parent `main`. Before dispatch, main is merged in (TUI work included).

## Global Constraints

- Names, verbatim from the spec: plugin `acta`, skills `acta:<name>`, CLI `acta`, folder `cmd/acta`, root `.acta/`, config `.acta.yaml`, env `ACTA_ROOT`, voice `~/.acta/voice.yaml`, module `github.com/iyay/acta`, install `go install github.com/iyay/acta/cmd/acta@latest`.
- Root lookup order: `--root`, `ACTA_ROOT`, `PM_ROOT`, `root` in `.acta.yaml`, `root` in `.pm.yaml`, then `.acta/` if it exists, else `.pm/` if it exists, else `.acta/`.
- `pmb` alias prints exactly `pmb is now acta; this name goes away in a later version` to stderr first, then behaves like `acta`.
- Short ID prefixes (`SPEC-`, `PLAN-`, `BUG-`), `.agents.json`, frontmatter fields: unchanged.
- Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- Edit only the files each task names (task 1 and task 4 are wide by nature: they name folders). Stage by path. Never push. Never commit the plan file except in task 5 as part of the move (the orchestrator commits ticks). gofmt inside each task commit. A defect in this branch's code is fixed here, never filed in `.pm/bugs/`. Never edit test fixtures' content except the rename of names inside them that the task names.
- Comments in plain English. No marker tags.
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## Waves

All tasks run in order (each touches files the next one builds on): wave 1 task 1, wave 2 task 2, wave 3 task 3, wave 4 task 4, wave 5 task 5.

---

### Task 1: Module path github.com/iyay/acta

**Files:** `go.mod`, every `.go` file that imports `pm-board/...`.

**verify:** No file imports `pm-board/` any more; `go build ./...`, `go vet ./...` and the full suite pass; `go list -m` prints `github.com/iyay/acta`. List the count of files changed and the grep that proves none remain.

- [x] **Step 1: Failing check:** add `TestModulePath` in `internal/plugincheck` that reads `go.mod` and requires `module github.com/iyay/acta`; run it, see it fail.
- [x] **Step 2: Rename:** `go mod edit -module github.com/iyay/acta`, then rewrite imports (`gofmt -r` cannot do import paths; use `sed` over the `.go` files for the exact string `"pm-board/` → `"github.com/iyay/acta/`).
- [x] **Step 3: Gate** (green), `grep -rn '"pm-board/' --include=*.go .` prints nothing.
- [x] **Step 4: Commit** (`refactor: module path github.com/iyay/acta`), tick `#task-1 --all`.

---

### Task 2: CLI `acta` with a `pmb` alias

**Files:** move `cmd/pmb` to `cmd/acta` (`git mv`), new thin `cmd/pmb/main.go`, the shared entry point so both binaries run the same code (for example move `run(args, stdin, stdout, stderr) int` into `internal/cli` if `cmd/pmb` keeps it in `package main` today), `cmd/acta/*_test.go`, usage strings that say `pmb`.

**verify:** `acta <anything>` behaves exactly as `pmb <anything>` did (same stdout, exit code); `pmb <anything>` prints the one warning line to stderr first and otherwise matches `acta`; every usage/help string says `acta`; the test binary is built from `cmd/acta`. List the commands compared.

- [x] **Step 1: Failing tests:** in the CLI tests, build both binaries in `TestMain`; `TestPmbAliasWarns` (stderr starts with the warning line, stdout equals `acta`'s for `list --all`); `TestUsageSaysActa` (`acta` with no args and `acta tick -h` mention `acta`, not `pmb`).
- [x] **Step 2: Run to see them fail.**
- [x] **Step 3: Implement** the move and the alias.
- [x] **Step 4: Gate.**
- [x] **Step 5: Commit** (`feat(cli): the CLI is acta; pmb warns and forwards`), tick `#task-2 --all`.

---

### Task 3: `.acta/` root, old names as fallback, `acta migrate-root`

**Files:** `internal/config/config.go`, `internal/config/config_test.go`, `internal/voice/voice.go`, `internal/voice/voice_test.go`, new `cmd/acta/migrate.go` (or in the shared CLI package) and its test, the dispatch switch.

**verify:** Every step of the root lookup order resolves as the spec lists, including `.pm/` only, `.acta/` only, both (`.acta/` wins), neither (`.acta/`), `PM_ROOT` and `.pm.yaml` honored when the new ones are missing; voice reads `~/.acta/voice.yaml`, falls back to `~/.pm/voice.yaml`, always writes the new file; `acta migrate-root` moves `.pm/` to `.acta/` (and `.pm.yaml` to `.acta.yaml`) with `git mv` in one commit, and refuses with exit 1 when `.acta/` exists, when `.pm/` is missing, or when `.pm/` has uncommitted changes. List every lookup case and refusal checked.

- [x] **Step 1: Failing tests** for each case above (temp repos, `t.Setenv` for env vars, a temp HOME for voice).
- [x] **Step 2: Run to see them fail.**
- [x] **Step 3: Implement.**
- [x] **Step 4: Gate.**
- [x] **Step 5: Commit** (`feat: .acta root with .pm fallback, and acta migrate-root`), tick `#task-3 --all`.

---

### Task 4: Plugin text says acta

**Files:** everything under `plugin/` (manifests, `package.json`, skills, references, hooks, omp extension, README), `internal/plugincheck/*_test.go`.

**verify:** No plugin file names a `pm:` skill, a bare `pmb ` command, or `.pm/` as the default root, except the one sentence that explains the `pmb` alias; manifest and package names are `acta`; the session-start hook text names `acta:*` skills; every plugincheck test passes with the new required and forbidden strings, plus one new test that fails if any `pm:` skill name or bare `pmb ` command returns. List the grep commands that prove it.

- [x] **Step 1: Failing test:** `TestNoOldNames` in plugincheck walks `plugin/` and fails on `pm:` skill names, a bare `pmb ` command, or `.pm/` (allowing the alias sentence); update the `Must`/`MustNot` lists to the new names.
- [x] **Step 2: Run to see it fail.**
- [x] **Step 3: Rewrite** the text (`pm:` → `acta:`, `pmb` → `acta`, `.pm/` → `.acta/`, `PM_ROOT` → `ACTA_ROOT`, plugin name `pm` → `acta`), keeping each file's meaning.
- [x] **Step 4: Gate.**
- [x] **Step 5: Commit** (`feat(plugin): the plugin is acta`), tick `#task-4 --all`.

---

### Task 5: This repo moves to `.acta/`

**Files:** `.pm/` → `.acta/` (whole folder, via the new command), `.pm.yaml` if present.

**verify:** After `go run ./cmd/acta migrate-root`, this repo has `.acta/` and no `.pm/`, the move is one commit, `go run ./cmd/acta list --all` shows the same items (same short IDs, same statuses) as before the move, and the full gate passes. List the item count before and after.

- [x] **Step 1:** Record `go run ./cmd/acta list --all --json | jq length` and the list of short IDs.
- [x] **Step 2:** Run `go run ./cmd/acta migrate-root` (it commits). The orchestrator must have committed the plan file's ticks before this step, so `.pm/` has no uncommitted changes.
- [x] **Step 3:** Compare the list after the move with step 1; run the gate.
- [x] **Step 4:** Tick `plans/2026-09-27-rename-acta#task-5 --all` with `go run ./cmd/acta tick ...` (the plan now lives under `.acta/plans/`).

## Fix round 1

### Task F1: migrate-root commits only the move and never half-moves

**Files:** `internal/cli/migrate.go`, its test file.

**verify:** On every repo state, `acta migrate-root` either moves `.pm/` (and `.pm.yaml`) to `.acta/` (and `.acta.yaml`) in exactly one commit that holds only those renames and a board that still lists every item afterwards, or refuses with exit 1 and changes nothing (working tree, index and HEAD byte-identical to before). List every repo state checked: clean; unrelated file staged; unrelated file modified but not staged; untracked `.pm.yaml`; modified `.pm.yaml`; `.pm.yaml` with `root: .pm`; `.pm.yaml` with other settings only; `.acta/` exists; `.pm/` missing; changes under `.pm/`.

Review round 1 BLOCKERs (both axes), each needs a red test first:

1. `internal/cli/migrate.go:66-67`: `git commit -q -m ...` has no pathspec, so anything the user had staged lands in the "move root folder" commit (`echo s > other.txt; git add other.txt; acta migrate-root` gives a commit with `A other.txt`). Expected: commit with a pathspec limited to `.pm .acta .pm.yaml .acta.yaml`, and the unrelated staged file stays staged and uncommitted afterwards.
2. `internal/cli/migrate.go:59-66`: `.pm.yaml` handling. (a) An untracked or modified `.pm.yaml` passes the dirty check (which only looks at `.pm`), then the rename step fails after `.pm/` was already moved and staged: half-move, exit 3, and `.pm.yaml` `root: .pm` then points the board at a folder that is gone. (b) A tracked `.pm.yaml` with `root: .pm` is renamed as is, so `.acta.yaml` says `root: .pm` and `acta list --all` prints nothing. Spec §3: rename `.pm.yaml` only when it only holds settings. Expected: before any `git mv`, refuse (exit 1, nothing changed) when `.pm.yaml` is untracked or modified; when it holds `root: .pm` (or `root: ./.pm`), write it as `.acta.yaml` without that `root` line (the default lookup finds `.acta/`), in the same single commit; if any step after the first `git mv` fails, undo the moves so the repo is back as it was.

- [x] **Step 1: Write the failing tests** (one per repo state in the verify line that is not already covered).
- [x] **Step 2: Run them to see them fail.**
- [x] **Step 3: Fix** `migrate.go`.
- [x] **Step 4: Gate:** `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- [x] **Step 5: Commit** (`fix(cli): migrate-root commits only the move and never half-moves`), tick `plans/2026-09-27-rename-acta#task-F1 --all` with `go run ./cmd/acta tick`.

## Fix round 2

### Task F2: migrate-root refuses when .acta.yaml already exists

**Files:** `internal/cli/migrate.go`, its test file.

**verify:** When `.acta.yaml` exists in any form (tracked or untracked file, ignored file, directory, symlink) before the move, `acta migrate-root` exits 1 before any `git mv` and leaves working tree, index and HEAD byte-identical; every other state from F1 behaves as before. List every form checked.

Review round 2 BLOCKER (Spec axis): `internal/cli/migrate.go:42-45` / `checkMovable` checks only `.acta/`, not `.acta.yaml`. (A) Untracked `.acta.yaml` with `root: .pm`, no `.pm.yaml`: rc 0, the move commits, `.acta.yaml` still points at `.pm`, `acta list --all` prints nothing. (B) Tracked `.pm.yaml` plus `.acta.yaml` as a directory: rc 0, the commit holds `.pm.yaml -> .acta.yaml/.pm.yaml`, the board fails with "is a directory". Expected: refuse (exit 1, nothing changed) whenever `.acta.yaml` exists, checked with `os.Lstat` before the first move.

- [x] **Step 1: Failing tests** for each form above.
- [x] **Step 2: Run them to see them fail.**
- [x] **Step 3: Fix** in `checkMovable`.
- [x] **Step 4: Gate.**
- [x] **Step 5: Commit** (`fix(cli): migrate-root refuses an existing .acta.yaml`), tick `plans/2026-09-27-rename-acta#task-F2 --all` with `go run ./cmd/acta tick`.
