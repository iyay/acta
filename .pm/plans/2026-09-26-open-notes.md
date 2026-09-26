# Close open review NOTEs Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** Close the open review NOTEs from the tick-fixes plan (f09c483) and the dispatch-notes plan (d19a71e): one lock folder for every process of a user, stronger lock tests, a full `pmb tick -h`, and plugin text that names the exact tick and show commands.

**Architecture:** The tick lock moves from `os.TempDir()` to a fixed per-user folder `/tmp/pmb-<uid>` (mode 0700, checked before use). `pmb tick -h` prints the flag list after the usage line. Plugin skill text is fixed in place and each file gets its own test guard.

**Tech Stack:** Go 1.27, `syscall.Flock`, Go tests over markdown skills in `internal/plugincheck`.

**Spec:** No spec file (Bounded, design approved in chat on 2026-09-26). Source: the NOTE lists in project memory `tick-fixes-review-notes` and `notefix-review-notes`.

**Worktree:** created by pm:build at `../pm-board-open-notes`, branch `open-notes`, parent `main`.

## Global Constraints

- Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- Edit only the files each task names. Stage by path. Never push. Never commit the plan file.
- Comments in plain English a ten-year-old can read, saying why. No marker tags.
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## File map

- `internal/write/tick.go`, `internal/write/tick_test.go`: lock folder and lock tests (Task 1).
- `cmd/pmb/tick.go`, `cmd/pmb/tick_test.go`: tick help text (Task 2).
- `plugin/skills/build/SKILL.md`, `plugin/references/house-rules.md`, `plugin/skills/dispatch/SKILL.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/skill_dispatch_test.go`, `internal/plugincheck/plugin_test.go`: plugin text (Task 3).

## Waves

- Wave 1: Tasks 1, 2 and 3 (no shared files, no dependency).

---

### Task 1: One lock folder per user

**Files:**
- Modify: `internal/write/tick.go` (func `lock`, new `lockBase` var and `lockPath` func)
- Test: `internal/write/tick_test.go`

**verify:** Every `pmb tick` process run by one user on one plan locks the same file, whatever `$TMPDIR` says, and no path lets a lock file land in a folder the user does not own or that others can write to. List every way the lock path is built and every check on the folder.

**Interfaces:**
- Consumes: nothing.
- Produces: `var lockBase = "/tmp"` (tests point it at `t.TempDir()`), `func lockPath(plan string) (string, error)` returning `<lockBase>/pmb-<uid>/pmb-<16 hex>.lock`. `lock(plan string) (func(), error)` keeps its signature.

- [x] **Step 1: Write the failing tests**

Add `"fmt"` to the imports of `tick_test.go`. Add:

```go
func TestLockPathIgnoresTMPDIR(t *testing.T) {
	p := filepath.Join(t.TempDir(), "plan.md")
	t.Setenv("TMPDIR", t.TempDir())
	a, err := lockPath(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", t.TempDir())
	b, err := lockPath(p)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("lock path follows TMPDIR: %s vs %s", a, b)
	}
	want := filepath.Join(lockBase, fmt.Sprintf("pmb-%d", os.Getuid()))
	if filepath.Dir(a) != want {
		t.Fatalf("lock folder = %s, want %s", filepath.Dir(a), want)
	}
}

// useLockBase points the lock folder at a fresh temp folder for one test.
func useLockBase(t *testing.T) string {
	old := lockBase
	lockBase = t.TempDir()
	t.Cleanup(func() { lockBase = old })
	return filepath.Join(lockBase, fmt.Sprintf("pmb-%d", os.Getuid()))
}

func TestLockMakesPrivateFolder(t *testing.T) {
	dir := useLockBase(t)
	unlock, err := lock(filepath.Join(t.TempDir(), "plan.md"))
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	fi, err := os.Lstat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.IsDir() || fi.Mode().Perm() != 0o700 {
		t.Fatalf("lock folder mode = %v, want a 0700 folder", fi.Mode())
	}
}

func TestLockRejectsUnsafeFolder(t *testing.T) {
	for name, setup := range map[string]func(dir string){
		"symlink": func(dir string) { os.Symlink(t.TempDir(), dir) },
		"others can write": func(dir string) {
			os.Mkdir(dir, 0o700)
			os.Chmod(dir, 0o777)
		},
		"plain file": func(dir string) { os.WriteFile(dir, nil, 0o600) },
	} {
		t.Run(name, func(t *testing.T) {
			dir := useLockBase(t)
			setup(dir)
			if unlock, err := lock(filepath.Join(t.TempDir(), "plan.md")); err == nil {
				unlock()
				t.Fatal("lock used an unsafe folder")
			}
		})
	}
}
```

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/write -run 'TestLockPathIgnoresTMPDIR|TestLockMakesPrivateFolder|TestLockRejectsUnsafeFolder' -count=1`
Expected: FAIL to build with `undefined: lockPath` and `undefined: lockBase`.

- [x] **Step 3: Write the implementation**

In `internal/write/tick.go`, add `"io/fs"` to the imports (for `fs.ErrExist`), and replace the top of `lock` up to the `os.OpenFile` call:

```go
// lockBase is where the per-user lock folder lives. It is fixed, not
// os.TempDir, so two processes with a different TMPDIR still share one lock.
var lockBase = "/tmp"

// lockPath gives the lock file for a plan, inside a folder only this user
// can use. /tmp is shared, so the folder is checked before we trust it.
func lockPath(plan string) (string, error) {
	real, err := filepath.Abs(plan)
	if err != nil {
		return "", err
	}
	if r, err := filepath.EvalSymlinks(real); err == nil {
		real = r
	}
	dir := filepath.Join(lockBase, fmt.Sprintf("pmb-%d", os.Getuid()))
	sum := sha256.Sum256([]byte(real))
	return filepath.Join(dir, "pmb-"+hex.EncodeToString(sum[:8])+".lock"), nil
}

// safeDir makes the lock folder, or checks the one already there. Someone
// else could make it first to steal or block our locks, so it must be a
// real folder, ours, and closed to everyone else.
func safeDir(dir string) error {
	if err := os.Mkdir(dir, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !fi.IsDir() || !ok || int(st.Uid) != os.Getuid() || fi.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("lock folder %s is not a private folder owned by you", dir)
	}
	return nil
}
```

and in `lock`:

```go
func lock(plan string) (func(), error) {
	path, err := lockPath(plan)
	if err != nil {
		return nil, err
	}
	if err := safeDir(filepath.Dir(path)); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	// rest of the function stays as it is
```

Update the doc comment on `lock` so it says "a private folder under /tmp" instead of "the temp folder".

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/write -run 'TestLock|TestTick' -count=1`
Expected: PASS.

- [x] **Step 5: Fix the helper that crashes itself**

In `TestHelperHoldLock`, replace `select {}` with:

```go
	// Wait to be killed. A bare select{} trips Go's deadlock check and the
	// child would crash before the parent kills it.
	time.Sleep(time.Minute)
```

- [x] **Step 6: Make TestLockOneHolder fail when the lock does nothing**

In `TestLockOneHolder`, between `n := holders.Add(1)` handling and `holders.Add(-1)`, add:

```go
			// Hold the lock a moment, so a lock that does nothing lets
			// several holders overlap and the test goes red.
			time.Sleep(5 * time.Millisecond)
```

Check it: temporarily make `lock` return `func() {}, nil` on its first line, run `go test ./internal/write -run TestLockOneHolder -count=1`, see FAIL with `holders at once, want 1`, then undo that line.

- [x] **Step 7: Run the gate**

Run: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`
Expected: all PASS.

- [x] **Step 8: Commit**

```bash
git add internal/write/tick.go internal/write/tick_test.go
git commit -m "fix(tick): lock in a private per-user folder under /tmp"
```

Then run `pmb tick plans/2026-09-26-open-notes#task-1 --all`.

---

### Task 2: `pmb tick -h` lists the flags

**Files:**
- Modify: `cmd/pmb/tick.go` (the `fs.Usage` line)
- Test: `cmd/pmb/tick_test.go` (func `TestTickHelp`)

**verify:** Every help form (`-h`, `-help`, `--help`, and `-h` after an id) prints the id shape and both `-step` and `-all` with their text. List every form checked.

**Interfaces:**
- Consumes: nothing.
- Produces: nothing new.

- [x] **Step 1: Write the failing test**

In `TestTickHelp`, after the `plans/<stem>#task-N` check inside the loop, add:

```go
		for _, flag := range []string{"-step", "-all", "tick every checkbox of the task"} {
			if !strings.Contains(out+errOut, flag) {
				t.Errorf("%v: output %q does not list %q", args, out+errOut, flag)
			}
		}
```

- [x] **Step 2: Run the test to see it fail**

Run: `go test ./cmd/pmb -run TestTickHelp -count=1`
Expected: FAIL with `does not list "-step"`.

- [x] **Step 3: Write the implementation**

In `cmd/pmb/tick.go`, replace the `fs.Usage` line:

```go
	// Show the id shape and the flags on -h/--help so agents copy it right.
	fs.Usage = func() {
		fmt.Fprintln(stderr, tickUsage)
		fs.PrintDefaults()
	}
```

- [x] **Step 4: Run the test to see it pass**

Run: `go test ./cmd/pmb -run 'TestTick' -count=1`
Expected: PASS.

- [x] **Step 5: Run the gate**

Run: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`
Expected: all PASS.

- [x] **Step 6: Commit**

```bash
git add cmd/pmb/tick.go cmd/pmb/tick_test.go
git commit -m "fix(tick): list --step and --all on -h"
```

Then run `pmb tick plans/2026-09-26-open-notes#task-2 --all`.

---

### Task 3: Plugin text names exact commands

**Files:**
- Modify: `plugin/skills/build/SKILL.md` (implementer constraints bullet with "run it again with --all"; the `## Close` paragraph)
- Modify: `plugin/references/house-rules.md` (the `PROGRESS:` line)
- Modify: `plugin/skills/dispatch/SKILL.md` (loop item 5 `**Comprehension checkpoint**`; section `## Comprehension checkpoint`)
- Test: `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/skill_dispatch_test.go`, `internal/plugincheck/plugin_test.go` (func `TestHouseRules`)

**verify:** No text an agent reads can be taken to mean `--step N --all`, a short task id, a bare `pmb show`, or a second checkpoint read, and each fixed file has its own test that goes red if that file alone is reverted. List every file and line checked, and which test guards it.

**Interfaces:**
- Consumes: nothing.
- Produces: nothing new.

- [x] **Step 1: Write the failing tests**

In `skill_build_test.go`, in `TestSkillBuild`:
- in `Must`, replace `"--all right after"` with `"pmb tick plans/<stem>#task-N --all"` and add `"pmb show <plan id> --json"`, `"progress.done"`;
- in `MustNot`, add `"run it again with --all"`.

Add a per-file guard for `build/SKILL.md` below `TestBuildTickRuleEverywhere`:

```go
// TestBuildSkillTickCommands reads SKILL.md on its own. CheckSkill looks at
// the whole build folder, so a revert of SKILL.md alone could stay green.
func TestBuildSkillTickCommands(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{"run `pmb tick plans/<stem>#task-N --all`", "pmb show <plan id> --json", "progress.done"} {
		if !strings.Contains(txt, want) {
			t.Errorf("build/SKILL.md missing %q", want)
		}
	}
	if strings.Contains(txt, "run it again with --all") {
		t.Error("build/SKILL.md still says \"run it again with --all\"")
	}
}
```

Also in `TestBuildTickRuleEverywhere`, change the check from `"--all right after"` to `"pmb tick [TASK_ID] --all"` (that command is already in implementer-prompt.md; this pins the command, not a phrase).

In `plugin_test.go` `TestHouseRules`:
- in the `want` list, replace `"--all right after"` with `"pmb tick plans/<stem>#task-N --all"` and add `"progress.done"`;
- in the `bad` list, add `"<its task id>"`.

In `skill_dispatch_test.go`, add a per-file guard:

```go
// TestDispatchCheckpointOneRead reads SKILL.md and herdr-delivery.md one at
// a time. CheckSkill joins the folder, so reverting one file stayed green.
func TestDispatchCheckpointOneRead(t *testing.T) {
	for file, wants := range map[string][]string{
		"SKILL.md":          {"exactly one read", "checkpoint unconfirmed", "gets its own one read", "outside this rule"},
		"herdr-delivery.md": {"exactly one read", "checkpoint unconfirmed"},
	} {
		b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", file))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range wants {
			if !strings.Contains(string(b), want) {
				t.Errorf("dispatch/%s missing %q", file, want)
			}
		}
	}
	b, _ := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", "SKILL.md"))
	if n := strings.Count(string(b), "checkpoint unconfirmed"); n < 2 {
		t.Errorf("dispatch/SKILL.md says \"checkpoint unconfirmed\" %d times; the loop item and the section both need it", n)
	}
}
```

Change the import line of `skill_dispatch_test.go` to import `os`, `path/filepath`, `strings`, `testing`.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/plugincheck -count=1`
Expected: FAIL naming the missing strings in `build/SKILL.md`, `house-rules.md` and `dispatch/SKILL.md`.

- [x] **Step 3: Fix `plugin/skills/build/SKILL.md`**

In the implementer constraints bullet, replace:

`right after the task's commit, run it again with --all right after the commit so no box stays open; never commit the plan file.`

with:

`right after the task's commit, run `pmb tick plans/<stem>#task-N --all` so no box stays open; never commit the plan file.`

(keep the inner backticks around the command, as the bullet already uses them for the `--step` command).

In `## Close`, replace:

`Before the review, every task of the plan shows all its boxes done (`pmb show`).`

with:

`Before the review, run `pmb show <plan id> --json` and check that `progress.done` equals `progress.total`. If a box is still open, tick it with `pmb tick plans/<stem>#task-N --all` when that task is committed, or finish the task first.`

- [x] **Step 4: Fix `plugin/references/house-rules.md`**

In the `PROGRESS:` line, replace:

`Run pmb tick <its task id> --all right after the commit so no box is left open. As the recipient, check with pmb show <plan id> --json that every task shows all boxes done before the reply-back.`

with:

`Right after your own task commit, run pmb tick plans/<stem>#task-N --all with the full id (a short id like task-1 fails with "unknown id"), so no box is left open. As the recipient, run pmb show <plan id> --json before the reply-back and check that progress.done equals progress.total.`

- [x] **Step 5: Fix `plugin/skills/dispatch/SKILL.md`**

Loop item 5, replace:

`5. **Comprehension checkpoint**: one read ~20s later — todo list must name the ticket ids, else drift; correct and re-dispatch.`

with:

`5. **Comprehension checkpoint**: exactly one read ~20s later — todo list must name the ticket ids, else drift; correct and re-dispatch. Todo list not up yet: report "checkpoint unconfirmed" and yield.`

In `## Comprehension checkpoint`, after the sentence ending `the reply-back is the real signal.`, insert:

` A re-dispatch after drift gets its own one read, under the same rule. Reads to find out why something failed, after a reply-back or a stuck report, are outside this rule.`

- [x] **Step 6: Run the tests to see them pass**

Run: `go test ./internal/plugincheck -count=1`
Expected: PASS. Then revert only `plugin/skills/dispatch/SKILL.md` with `git stash push plugin/skills/dispatch/SKILL.md`, run the same command, see FAIL in `TestDispatchCheckpointOneRead`, then `git stash pop`. Do the same for `plugin/skills/build/SKILL.md` (expect FAIL in `TestBuildSkillTickCommands`).

- [x] **Step 7: Run the gate**

Run: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`
Expected: all PASS.

- [x] **Step 8: Commit**

```bash
git add plugin/skills/build/SKILL.md plugin/references/house-rules.md plugin/skills/dispatch/SKILL.md internal/plugincheck/skill_build_test.go internal/plugincheck/skill_dispatch_test.go internal/plugincheck/plugin_test.go
git commit -m "fix(plugin): name exact tick and show commands, one checkpoint read per dispatch"
```

Then run `pmb tick plans/2026-09-26-open-notes#task-3 --all`.
