---
id: PLN-0022
hash: l443mag
---
# Dispatch Reply-Back Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A dispatched recipient tells the orchestrator it is done (or blocked) with one command, `acta reply-back`, fed by a record the orchestrator writes with `acta dispatch init`; the skills make this the build skill's last step and add an idle watcher on the orchestrator side.

**Architecture:** One new file `internal/cli/dispatch.go` holds the record type, its checks, and the two commands; `cli.go` routes them. The plan's progress comes from the existing board (`loadBoard`, `board.Item.Children`, task `Status`). `herdr` runs through `exec.Command` with separate arguments. Skill text changes and their `plugincheck` guards are a separate task on separate files.

**Tech Stack:** Go standard library; existing packages `internal/cli`, `internal/board`, `internal/hook`, `internal/plugincheck`.

**Spec:** `.acta/specs/2026-09-29-reply-back-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- The record file is untrusted input. No value from it reaches a shell; `herdr` runs through `exec.Command` with separate arguments.
- Record checks: `pane` matches `^[A-Za-z0-9]+:[A-Za-z0-9]+$`; `base` is 40 lowercase hex characters; `plan` is a relative path that, joined to the repo root and cleaned, stays inside the acta root and names a plan on the board; `round` matches `^[a-z0-9][a-z0-9-]{0,63}$`.
- Exit codes follow the repo: `exitOK` 0, `exitBadInput` 1, `exitOther` 3.
- Neither command commits anything.
- Comments in plain English a 10-year-old can read; say why, not what.

---

## File map

- Create: `internal/cli/dispatch.go` (record type, checks, `cmdDispatchInit`, `cmdReplyBack`)
- Create: `internal/cli/dispatch_test.go`
- Modify: `internal/cli/cli.go` (the `switch args[0]` in `Run`, and the unknown-command message)
- Modify: `plugin/skills/build/SKILL.md`, `plugin/skills/dispatch/SKILL.md`, `plugin/skills/dispatch/herdr-delivery.md`
- Create: `internal/plugincheck/reply_back_test.go`

## Waves

- Wave 1: Task 1 and Task 3 (disjoint files)
- Wave 2: Task 2 (same files as Task 1)

### Task 1: `acta dispatch init` and the record

**Files:**
- Create: `internal/cli/dispatch.go`
- Create: `internal/cli/dispatch_test.go`
- Modify: `internal/cli/cli.go` (`Run` switch, unknown-command message)

**verify:** No input makes `acta dispatch init` write a record that fails the record checks, and every bad input exits 1 with nothing written. Enumerate every input checked (outside git, empty `--pane`, empty `--plan`, bad pane shape, plan path outside the acta root, plan not on the board, good input, second run after a new commit) and report the list. A good run writes `base` = current HEAD, `round` = the branch name when `--round` is not given, and adds `.dispatch.json` to the acta root's `.gitignore`.

**Interfaces:**
- Produces:
  - `type dispatchRecord struct { Pane, Base, Plan, Round string }` with JSON tags `pane`, `base`, `plan`, `round`.
  - `func recordPath(cfg config.Config) string` returning `filepath.Join(cfg.Root, ".dispatch.json")`.
  - `func checkRecord(cfg config.Config, b *board.Board, r dispatchRecord) (*board.Item, error)` returning the plan item when every check passes.
  - `func cmdDispatchInit(args []string, stdout, stderr io.Writer) int`.

- [x] **Step 1: Write the failing tests** in `internal/cli/dispatch_test.go`. Build a real git repo in `t.TempDir()` (`git init`, a commit holding `.acta/plans/2026-09-29-p.md` with frontmatter `id: PLAN-1`, `hash: aaaa` and two tasks `### Task 1: A` / `- [ ] a` and `### Task 2: B` / `- [ ] b`), `t.Chdir` into it, and call `Run([]string{"dispatch", "init", ...}, nil, false, &out, &errb)`. Tests:
  - `TestDispatchInitWritesHeadAndBranch`: `--pane wM:pH --plan .acta/plans/2026-09-29-p.md` exits 0; `.acta/.dispatch.json` decodes to `pane wM:pH`, `base` = `git rev-parse HEAD`, `plan .acta/plans/2026-09-29-p.md`, `round` = the branch name; `.acta/.gitignore` has the line `.dispatch.json`.
  - `TestDispatchInitSecondRunMovesBase`: init, make a new commit, init again; `base` equals the new HEAD.
  - `TestDispatchInitRejectsBadInput`: table over empty `--pane`, empty `--plan`, `--pane "wM:pH; rm"`, `--pane wMpH`, `--plan ../outside.md`, `--plan .acta/plans/missing.md`, `--round "Bad Round"`; each exits 1 and `.acta/.dispatch.json` does not exist.
  - `TestDispatchInitOutsideGitFails`: in a plain temp folder with a `.acta/plans` plan file and no `.git`, exits 1 and writes nothing.

- [x] **Step 2: Run** `go test ./internal/cli/ -run TestDispatchInit -v`. Expected: FAIL, `unknown command "dispatch"`.

- [x] **Step 3: Implement.** In `cli.go` add:

```go
	case "dispatch":
		if len(args) < 2 || args[1] != "init" {
			fmt.Fprintln(stderr, "usage: acta dispatch init --pane <id> --plan <path> [--round <slug>]")
			return exitBadInput
		}
		return cmdDispatchInit(args[2:], stdout, stderr)
	case "reply-back":
		return cmdReplyBack(args[1:], stdout, stderr)
```

and add `dispatch init` and `reply-back` to the unknown-command message. Task 1 adds a stub `cmdReplyBack` that prints the usage and returns `exitBadInput`, so the build compiles; Task 2 replaces it.

In `dispatch.go`, write `dispatchRecord`, `recordPath`, and `checkRecord`:

```go
var (
	panePattern  = regexp.MustCompile(`^[A-Za-z0-9]+:[A-Za-z0-9]+$`)
	shaPattern   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	roundPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
)

// checkRecord treats the record as untrusted, because anyone can edit the
// file. Every value is checked before it is used for anything.
func checkRecord(cfg config.Config, b *board.Board, r dispatchRecord) (*board.Item, error) {
	if !panePattern.MatchString(r.Pane) {
		return nil, fmt.Errorf("bad pane %q", r.Pane)
	}
	if !shaPattern.MatchString(r.Base) {
		return nil, fmt.Errorf("bad base %q", r.Base)
	}
	if !roundPattern.MatchString(r.Round) {
		return nil, fmt.Errorf("bad round %q", r.Round)
	}
	if r.Plan == "" || filepath.IsAbs(r.Plan) {
		return nil, fmt.Errorf("bad plan %q", r.Plan)
	}
	abs := filepath.Clean(filepath.Join(cfg.RepoRoot, r.Plan))
	rel, err := filepath.Rel(cfg.Root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("plan %q is outside %s", r.Plan, cfg.Root)
	}
	it := b.Get(strings.TrimSuffix(filepath.ToSlash(rel), ".md"))
	if it == nil || it.Kind != board.KindPlan {
		return nil, fmt.Errorf("no plan %s", r.Plan)
	}
	return it, nil
}
```

`cmdDispatchInit`: parse `--pane`, `--plan`, `--round` with `flags("dispatch init", stderr)`; `cfg, b, code := loadBoard(*root, stderr)`; require `cfg.IsGit`; `base` from `git rev-parse HEAD` and the default `round` from `git rev-parse --abbrev-ref HEAD`, both run with `exec.Command("git", ...)` and `cmd.Dir = cfg.RepoRoot`; build the record; `checkRecord`; `hook.EnsureGitignore(cfg.Root, ".dispatch.json")` (its error is printed, not fatal); write the JSON with `os.WriteFile(recordPath(cfg), data, 0o644)`; print `dispatch record: <path>` on stdout. Any failed check prints one line on stderr and returns `exitBadInput`.

- [x] **Step 4: Run** `go test ./...`. Expected: PASS.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/cli/dispatch.go internal/cli/dispatch_test.go internal/cli/cli.go
git commit -m "feat(cli): add acta dispatch init"
```

### Task 2: `acta reply-back`

**Files:**
- Modify: `internal/cli/dispatch.go` (replace the `cmdReplyBack` stub)
- Modify: `internal/cli/dispatch_test.go`

**verify:** `acta reply-back` sends a review request only when the record passes every check and every task of the plan is done, and sends nothing on any other path. Enumerate every path checked (no record, broken JSON, bad pane, bad sha, plan outside the root, open tasks, all done, `--blocked` with a reason, `--blocked ""`, `herdr` missing, `herdr` failing) with what each sends and its exit code, and report the list.

**Interfaces:**
- Consumes: `dispatchRecord`, `recordPath`, `checkRecord` from Task 1.
- Produces: `func cmdReplyBack(args []string, stdout, stderr io.Writer) int`.

- [x] **Step 1: Write the failing tests.** Reuse Task 1's repo helper and run `acta dispatch init` first. A fake `herdr` is a shell script written to a temp folder that appends each argument on its own line to a file (`printf '%s\n' "$@" >> "$HERDR_LOG"`) and exits with `$HERDR_EXIT`; prepend its folder with `t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))`, set `HERDR_LOG`, `HERDR_EXIT=0` and `HERDR_PANE_ID=wM:p9`. Tests:
  - `TestReplyBackRefusesOpenTasks`: no box ticked; exit 1; stderr names `plans/2026-09-29-p#task-1` and `#task-2`; the herdr log does not exist.
  - `TestReplyBackSendsReviewWhenDone`: tick both boxes in the plan file and commit; exit 0; the log is exactly `agent`, `prompt`, `wM:pH`, `/acta:review <base>..<head> - plan .acta/plans/2026-09-29-p.md, round <branch>, pane wM:p9`.
  - `TestReplyBackBlockedSendsProse`: open tasks, `--blocked "tests need a db"`; exit 0; last log line is `<branch> blocked: tests need a db - pane wM:p9`.
  - `TestReplyBackRejectsBadRecords`: table that overwrites `.acta/.dispatch.json` with: nothing (file removed), `{`, a bad pane, a 7-character base, plan `../x.md`; each exits 1 and the log does not exist.
  - `TestReplyBackEmptyBlockedFails`: `--blocked ""`; exit 1; no log.
  - `TestReplyBackHerdrFailureExits3`: all done, `HERDR_EXIT=2`; exit 3. And with `PATH` set to an empty temp folder: exit 3.

- [x] **Step 2: Run** `go test ./internal/cli/ -run TestReplyBack -v`. Expected: FAIL (the stub returns 1 everywhere, so the done and blocked tests fail).

- [x] **Step 3: Implement** `cmdReplyBack`: parse `--blocked` (use `fs.Visit` to tell "not given" from "given empty"); `cfg, b, code := loadBoard(*root, stderr)`; read `recordPath(cfg)` and `json.Unmarshal` it (errors exit 1); `plan, err := checkRecord(cfg, b, r)` (exit 1); `own := os.Getenv("HERDR_PANE_ID")`, and when it is empty or fails `panePattern`, use `unknown`.
  - `--blocked` given: empty reason exits 1; else text = `fmt.Sprintf("%s blocked: %s - pane %s", r.Round, reason, own)`.
  - Otherwise: list every child id of `plan.Children` whose `b.Get(id).Status != "done"`; any open task prints `open tasks:` plus one id per line on stderr and exits 1. Else `head` from `git rev-parse HEAD` in `cfg.RepoRoot`, text = `fmt.Sprintf("/acta:review %s..%s - plan %s, round %s, pane %s", r.Base, head, r.Plan, r.Round, own)`.
  - Send with `cmd := exec.Command("herdr", "agent", "prompt", r.Pane, text)`; capture its stderr; `exec.ErrNotFound` or a non-zero exit prints herdr's stderr and exits `exitOther`. Success prints `sent to <pane>: <text>` on stdout.

- [x] **Step 4: Run** `go test ./...`. Expected: PASS.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/cli/dispatch.go internal/cli/dispatch_test.go
git commit -m "feat(cli): add acta reply-back"
```

### Task 3: Skill text and guards

**Files:**
- Modify: `plugin/skills/build/SKILL.md`
- Modify: `plugin/skills/dispatch/SKILL.md`
- Modify: `plugin/skills/dispatch/herdr-delivery.md`
- Create: `internal/plugincheck/reply_back_test.go`

**verify:** No skill file still tells a recipient to hand-fill a reply-back command or lists allowed acta commands, and every place that used to carry the old reply-back text now points at `acta reply-back` or `acta dispatch init`. Enumerate every old reply-back, `<new-head-sha>` and `--start` restate you found in the three files and what replaced each, and report the list.

**Interfaces:**
- Consumes: the command names `acta dispatch init` and `acta reply-back` from the spec (Tasks 1 and 2 build them).
- Produces: nothing code uses.

- [x] **Step 1: Write the failing guards** in `internal/plugincheck/reply_back_test.go`, reading files the way the package's other tests do (find a sibling test that reads `plugin/skills/...` and reuse its path helper):
  - `TestBuildSkillEndsWithReplyBack`: `build/SKILL.md` contains `acta reply-back` and `.dispatch.json`.
  - `TestDispatchBriefLoadsBuild`: `dispatch/SKILL.md` contains `SKILL: load build`.
  - `TestNoSkillWhitelistsActaCommands`: no `.md` under `plugin/skills` contains `only allowed` (case-insensitive).
  - `TestNoHandFilledReplyBack`: no `.md` under `plugin/skills/dispatch` contains `<new-head-sha>`.
  - `TestDispatchInitBeforeGoal`: `dispatch/herdr-delivery.md` contains `acta dispatch init`.

- [x] **Step 2: Run** `go test ./internal/plugincheck/ -run 'TestBuildSkillEndsWithReplyBack|TestDispatchBriefLoadsBuild|TestNoSkillWhitelistsActaCommands|TestNoHandFilledReplyBack|TestDispatchInitBeforeGoal' -v`. Expected: FAIL on the first, second, fourth and fifth.

- [x] **Step 3: Edit the skills** as spec section 5 and 6 say:
  - `build/SKILL.md`: add the last step "When `<acta root>/.dispatch.json` exists, run `acta reply-back` after the last task commit. Exit 1 lists open tasks: finish them and run it again. A real blocker: `acta reply-back --blocked \"<reason>\"`. A dispatch recipient does not stop before `acta reply-back` exits 0."
  - `dispatch/SKILL.md`: the brief format gets the line `SKILL: load build (omp: build) and tdd before the todo list, and follow them for every task.`; add "Briefs never list allowed acta commands; they may only forbid acta write commands the build skill does not call."; replace the "Reply-back" section's command block and every `/goal` reply-back tail with one sentence: after the last commit the build skill runs `acta reply-back`; add `acta dispatch init --pane $HERDR_PANE_ID --plan <plan> [--round <slug>]` in the worktree before each `/goal` (first dispatch and each fix round); add the watcher (spec section 6) after the comprehension checkpoint, and add it as the one exception in "Never wait".
  - `herdr-delivery.md`: drop the `--start` and tick restates (the build skill owns them); replace the reply-back command in the `/goal` templates with the one sentence above; add the `acta dispatch init` step before `/goal` and the watcher step after the checkpoint.

- [x] **Step 4: Run** `go test ./...`. Expected: PASS.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add plugin/skills/build/SKILL.md plugin/skills/dispatch/SKILL.md plugin/skills/dispatch/herdr-delivery.md internal/plugincheck/reply_back_test.go
git commit -m "feat(plugin): reply-back through acta, idle watcher in dispatch"
```
