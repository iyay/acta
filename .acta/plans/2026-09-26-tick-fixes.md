---
id: PLAN-10
hash: nze9
---
# pmb tick Fixes Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** Make `pmb tick --help` exit 0 and name the task id format, make the `pm:build` implementer hand-off carry the full `pmb tick` command with the task id, and make sure two processes can never hold the tick lock at once.

**Architecture:** Task 1 changes the `tick` flag handling in `cmd/pmb/tick.go`. Task 2 changes text in the `build` skill and its content check. Task 3 swaps the lock file (create-exclusive plus a one-minute stale takeover) for an OS `flock` on a file in the temp folder, named from a hash of the plan's real path. The kernel drops an `flock` when its process ends, so there is no stale lock and no takeover.

**Tech Stack:** Go (stdlib `flag`, `syscall`, `crypto/sha256`), markdown skills checked by `internal/plugincheck`.

**Spec:** Bounded design approved in chat on 2026-09-26 (no spec file):
1. `pmb tick --help` (and `-h`) exits 0 and its help text names the id format `plans/<stem>#task-N`.
2. The `pm:build` implementer hand-off names the full `pmb tick` command with the task id, never a bare `<task-id>`. Added on 2026-09-26: the same goes for `plugin/references/house-rules.md`, which the hand-off points implementers at.
3. The stale-lock takeover in `pmb tick` can never leave two processes holding the lock. Fix: `syscall.Flock` on `os.TempDir()/pmb-<hash>.lock`; Unix only (darwin, linux); no lock file next to the plan.

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- No new dependencies. Go stdlib only.
- Targets are darwin and linux only; `syscall.Flock` is fine.
- Comments in plain English a ten-year-old can read, saying why; no marker tags.
- Run `gofmt -l .` (must print nothing) and `go vet ./...` before every commit; formatting goes in the task commit.
- Stage by path, never `git add -A` or `git add .`. Never commit the plan file.
- Worktree: `/Users/iyay/Nayakatara/pm-board-tick-fixes`, branch `tick-fixes`, parent `main`.

## File Map

| File | Task | Change |
|---|---|---|
| `cmd/pmb/tick.go` | 1 | help exits 0, usage names `plans/<stem>#task-N` |
| `cmd/pmb/tick_test.go` | 1 | help test |
| `plugin/skills/build/SKILL.md` | 2 | full tick command in hand-off list and constraints |
| `plugin/skills/build/implementer-prompt.md` | 2 | full tick command with `[TASK_ID]` |
| `internal/plugincheck/skill_build_test.go` | 2 | require id format, forbid bare `<task-id>` |
| `plugin/references/house-rules.md` | 2 | full tick command in PROGRESS rule |
| `internal/plugincheck/plugin_test.go` | 2 | `TestHouseRules`: require id format, forbid bare `<task-id>` |
| `internal/write/tick.go` | 3 | `flock` lock in temp folder |
| `internal/write/tick_test.go` | 3 | crash, one-holder, no-file-in-repo tests |

## Waves

- **Wave 1:** Task 1, Task 2, Task 3 (no shared files).

---

### Task 1: `pmb tick --help` exits 0 and names the id format

**Files:**
- Modify: `cmd/pmb/tick.go` (`tickUsage`, `cmdTick`)
- Test: `cmd/pmb/tick_test.go`

**verify:** Every way to ask for help (`--help`, `-help`, `-h`, before or after a task id) exits 0 and its output names `plans/<stem>#task-N`; every bad input that exited 1 before still exits 1. List each help form and each bad-input case checked.

**Interfaces:**
- Consumes: `flags`, `parseMixed` in `cmd/pmb/main.go` (unchanged); `pmb(t, dir, stdin, args...) (stdout, stderr string, code int)` test helper.
- Produces: `tickUsage = "usage: pmb tick plans/<stem>#task-N [--step N | --all]"`.

- [x] **Step 1: Write the failing test** in `cmd/pmb/tick_test.go`:

```go
func TestTickHelp(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"tick", "--help"},
		{"tick", "-help"},
		{"tick", "-h"},
		{"tick", "plans/x#task-1", "-h"},
	} {
		out, errOut, code := pmb(t, dir, "", args...)
		if code != 0 {
			t.Errorf("%v: exit %d, want 0", args, code)
		}
		if !strings.Contains(out+errOut, "plans/<stem>#task-N") {
			t.Errorf("%v: help does not name the id format:\n%s%s", args, out, errOut)
		}
	}
}
```

- [x] **Step 2: Run it, watch it fail**

Run: `go test -count=1 -run 'TestTickHelp|TestTickCommand' ./cmd/pmb/`
Expected: `TestTickHelp` FAILS with `exit 1, want 0` and "does not name the id format"; `TestTickCommand` passes.

- [x] **Step 3: Write the minimal code** in `cmd/pmb/tick.go`:

```go
const tickUsage = "usage: pmb tick plans/<stem>#task-N [--step N | --all]"
```

In `cmdTick`, right after `fs, root := flags("tick", stderr)` and the two flag lines, set the usage so `-h` shows the id format:

```go
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), tickUsage)
		fs.PrintDefaults()
	}
	pos, err := parseMixed(fs, args)
	// Asking for help is not a mistake, so it exits 0.
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}
	if err != nil || len(pos) != 1 || (*step > 0) == *all {
```

Add `"flag"` to the imports.

- [x] **Step 4: Run tests, watch them pass**

Run: `go test -count=1 ./cmd/pmb/`
Expected: PASS (including `TestTickCommand` bad-input cases still exiting 1).

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add cmd/pmb/tick.go cmd/pmb/tick_test.go
git commit -m "fix(pmb): tick --help exits 0 and names the plans/<stem>#task-N id"
```

---

### Task 2: `pm:build` hand-off carries the full `pmb tick` command

**Files:**
- Modify: `plugin/skills/build/SKILL.md` (Task Loop hand-off list line starting `- right after each step`, and the `**Constraints:**` line under Common Mistakes)
- Modify: `plugin/skills/build/implementer-prompt.md` (step 6)
- Modify: `plugin/references/house-rules.md` (the `PROGRESS:` line)
- Test: `internal/plugincheck/skill_build_test.go`
- Test: `internal/plugincheck/plugin_test.go` (`TestHouseRules`)

**verify:** No text in the `build` skill folder or in `plugin/references/house-rules.md` tells an implementer to run `pmb tick` without the id format `plans/<stem>#task-N`; the bare form `pmb tick <task-id>` cannot come back in any of those files. List every `pmb tick` mention checked.

**Interfaces:**
- Consumes: `CheckSkill(t, SkillRule{...})` in `internal/plugincheck` (reads every `.md` in the skill folder).
- Produces: nothing other tasks use.

- [x] **Step 1: Write the failing test.** In `internal/plugincheck/skill_build_test.go`, add to `Must`:

```go
			"pmb tick plans/<stem>#task-N --step <n>", "pmb tick [TASK_ID] --step <n>",
```

and add to `MustNot`:

```go
			"pmb tick <task-id>",
```

In `internal/plugincheck/plugin_test.go`, `TestHouseRules`: add `"pmb tick plans/<stem>#task-N --step <n>"` to the `want` list and `"pmb tick <task-id>"` to the `bad` list.

- [x] **Step 2: Run it, watch it fail**

Run: `go test -count=1 -run 'TestSkillBuild|TestHouseRules' ./internal/plugincheck/`
Expected: `TestHouseRules` FAILS with `house-rules.md missing "pmb tick plans/<stem>#task-N --step <n>"` and `house-rules.md still has "pmb tick <task-id>"`; `TestSkillBuild` FAILS with `missing required text: pmb tick plans/<stem>#task-N --step <n>`, `missing required text: pmb tick [TASK_ID] --step <n>` and `has forbidden text: pmb tick <task-id>`.

- [x] **Step 3: Change the text.**

In `plugin/skills/build/SKILL.md`, replace the hand-off list line

```
- right after each step, run `pmb tick <task-id> --step <n>` from the worktree so the board shows live progress; never commit the plan file.
```

with

```
- right after each step, run the full tick command from the worktree, with this task's id written out: `pmb tick plans/<stem>#task-N --step <n>` (for example `pmb tick plans/2026-09-26-tick-fixes#task-3 --step 2`), so the board shows live progress; never commit the plan file.
```

In the same file, replace

```
**Constraints:** stage by path, `pmb tick` after each step, never commit the plan file.
```

with

```
**Constraints:** stage by path, `pmb tick plans/<stem>#task-N --step <n>` after each step, never commit the plan file.
```

In `plugin/skills/build/implementer-prompt.md`, replace step 6

```
    6. Right after each step, run `pmb tick <task-id> --step <n>` from the
       worktree so the board shows live progress; never commit the plan file.
```

with

```
    6. Right after each step, run `pmb tick [TASK_ID] --step <n>` from the
       worktree so the board shows live progress; never commit the plan file.
       [TASK_ID] is this task's id, written out in full by the orchestrator:
       plans/<stem>#task-N, for example plans/2026-09-26-tick-fixes#task-3.
```

In `plugin/references/house-rules.md`, replace

```
PROGRESS: right after each step of a ticket, run pmb tick <task-id> --step <n> from the worktree, so the board shows live progress.
```

with

```
PROGRESS: right after each step of a ticket, run pmb tick plans/<stem>#task-N --step <n> from the worktree, with your task's id written out (for example pmb tick plans/2026-09-26-tick-fixes#task-3 --step 2), so the board shows live progress.
```

Keep the rest of that line as it is.

- [x] **Step 4: Run tests, watch them pass**

Run: `go test -count=1 ./internal/plugincheck/`
Expected: PASS (line cap 680 still holds).

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add plugin/skills/build/SKILL.md plugin/skills/build/implementer-prompt.md plugin/references/house-rules.md internal/plugincheck/skill_build_test.go internal/plugincheck/plugin_test.go
git commit -m "fix(plugin): build hand-off names the full pmb tick command and task id"
```

---

### Task 3: Tick lock can never have two holders

**Files:**
- Modify: `internal/write/tick.go` (`Tick`, `lock`)
- Test: `internal/write/tick_test.go`

**verify:** On every path, at most one `Tick` holds the lock for a plan at any moment, and a holder that dies without unlocking never blocks the next `Tick`. List every path checked: normal unlock, crashed holder, many concurrent ticks, a leftover old-style `plan.md.lock` file from the previous version, and the timeout error.

**Interfaces:**
- Consumes: nothing new.
- Produces: `func lock(plan string) (func(), error)` takes the plan path (not a lock path). `Tick(path string, headingLine, step int) (int, int, error)` keeps its signature.

- [x] **Step 1: Write the failing tests** in `internal/write/tick_test.go`. Add `"os/exec"`, `"sync/atomic"` and `"time"` to the imports. Replace the `p + ".lock"` check at the end of `TestTickFileConcurrent` with a check that the plan folder holds only the plan:

```go
	if ents, _ := os.ReadDir(filepath.Dir(p)); len(ents) != 1 {
		t.Fatalf("tick left files next to the plan: %v", ents)
	}
```

Then add:

```go
// TestHelperHoldLock is not a real test. TestTickAfterHolderDies runs it
// in a child process to take the lock and then get killed.
func TestHelperHoldLock(t *testing.T) {
	p := os.Getenv("PMB_HOLD_LOCK")
	if p == "" {
		t.Skip("helper for TestTickAfterHolderDies")
	}
	if _, err := lock(p); err != nil {
		t.Fatal(err)
	}
	os.Stdout.WriteString("held\n")
	select {}
}

func TestTickAfterHolderDies(t *testing.T) {
	p := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(p, []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldLock$")
	cmd.Env = append(os.Environ(), "PMB_HOLD_LOCK="+p)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 5)
	if _, err := io.ReadFull(out, buf); err != nil || string(buf) != "held\n" {
		t.Fatalf("helper did not take the lock: %q %v", buf, err)
	}
	cmd.Process.Kill()
	cmd.Wait()

	start := time.Now()
	if _, _, err := Tick(p, 3, 2); err != nil {
		t.Fatalf("a dead holder blocked the tick: %v", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("tick waited %v after the holder died", d)
	}
}

func TestLockOneHolder(t *testing.T) {
	p := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(p, []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	// A lock file left by the old version, old enough to look stale.
	// It must not let anyone skip the lock.
	old := time.Now().Add(-2 * time.Minute)
	os.WriteFile(p+".lock", nil, 0o644)
	os.Chtimes(p+".lock", old, old)

	var holders, most atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock, err := lock(p)
			if err != nil {
				t.Error(err)
				return
			}
			n := holders.Add(1)
			for {
				m := most.Load()
				if n <= m || most.CompareAndSwap(m, n) {
					break
				}
			}
			holders.Add(-1)
			unlock()
		}()
	}
	wg.Wait()
	if most.Load() != 1 {
		t.Fatalf("%d holders at once, want 1", most.Load())
	}
}

func TestLockTimeout(t *testing.T) {
	p := filepath.Join(t.TempDir(), "plan.md")
	unlock, err := lock(p)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if _, err := lock(p); err == nil || !strings.Contains(err.Error(), "locked") {
		t.Fatalf("second lock: err = %v, want a locked error", err)
	}
}
```

Add `"io"` to the imports too.

- [x] **Step 2: Run them, watch them fail**

Run: `go test -count=1 -run 'TestTickFileConcurrent|TestTickAfterHolderDies|TestLockOneHolder|TestLockTimeout' ./internal/write/`
Expected: FAIL. `TestTickAfterHolderDies` fails with `a dead holder blocked the tick: plan is locked by ...` (the old lock file stays behind with a fresh time). `TestLockOneHolder` fails because the old `lock` treats its argument as the lock file itself: `plan.md` already exists and is fresh, so the goroutines time out with `plan is locked by ...`.

- [x] **Step 3: Write the minimal code** in `internal/write/tick.go`. In `Tick`, change `lock(path + ".lock")` to `lock(path)`. Replace `lock` with:

```go
// lock takes an OS lock (flock) on a file in the temp folder, named from
// the plan's real path. The kernel drops the lock when the process ends,
// even in a crash, so no lock is ever left behind to take over, and two
// processes can never both hold it. Keeping the file out of the repo means
// no stray file shows up in git status.
func lock(plan string) (func(), error) {
	real, err := filepath.Abs(plan)
	if err != nil {
		return nil, err
	}
	if r, err := filepath.EvalSymlinks(real); err == nil {
		real = r
	}
	sum := sha256.Sum256([]byte(real))
	path := filepath.Join(os.TempDir(), "pmb-"+hex.EncodeToString(sum[:8])+".lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				f.Close()
			}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			f.Close()
			return nil, err
		}
		if time.Now().After(deadline) {
			f.Close()
			return nil, fmt.Errorf("plan %s is locked by another pmb tick", plan)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
```

Add `"crypto/sha256"`, `"encoding/hex"`, `"path/filepath"` and `"syscall"` to the imports. Update the `Tick` doc comment: "An OS lock keeps two implementers in one worktree from overwriting each other's ticks."

- [x] **Step 4: Run tests, watch them pass**

Run: `go test -count=1 -race ./internal/write/ && go test -count=1 ./...`
Expected: PASS, no race reports.

- [x] **Step 5: Commit**

```bash
gofmt -l . && go vet ./...
git add internal/write/tick.go internal/write/tick_test.go
git commit -m "fix(pmb): tick lock uses flock so two processes never both hold it"
```
