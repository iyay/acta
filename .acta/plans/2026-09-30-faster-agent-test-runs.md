---
created: "2026-09-30"
id: PLN-0042
hash: rxv1rsf
status: approved
started: "2026-09-30"
finished: "2026-09-30"
---
# Faster Test Runs for Agents Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Only one full test suite runs on the machine at a time (`acta run-one`), agents run narrow tests during a build and the full suite only in `acta:land`, and the full suite itself uses every core.

**Architecture:** A new `acta run-one -- <cmd>` holds one `flock` per user per machine (built on `lockDir` and `safeDir` in `internal/write/tick.go`) while it runs the command. `scripts/test` is this repo's test entry: `-short` by default, `--full` under `run-one`. The skills name narrow tests for tasks and the full suite only at land. The slow sweep tests split into parallel subtests, and the git test helpers set identity with repo-local `git config` instead of `t.Setenv`, so their tests can run in parallel.

**Tech Stack:** Go (`syscall.Flock`, `os/exec`, `os/signal`, `testing` subtests), POSIX sh, skill markdown checked by `internal/plugincheck`.

**Spec:** `.acta/specs/2026-09-30-faster-agent-test-runs-design.md`

**Tests:** fast `go test -short ./...` (`scripts/test` once Task 3 lands), full `go test ./...` (`scripts/test --full` once Task 3 lands).

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- No test may check less than it does today. No sweep samples fewer sizes in a full run. A missed time target is reported with real numbers, never met by loosening a check.
- Production code changes only for `acta run-one` (Task 2). Everything else is tests, `scripts/test` and skill text.
- Plans that are already running (tui-colors, batch-author-lookup, config-file-name) are not edited.
- No new dependency in go.mod.
- Comments use plain English that a 10-year-old can read back: short words, and they say why.
- Before each commit, run `gofmt -l cmd internal scripts` (it must print nothing), `go vet ./...`, and the narrow tests of the packages the task touched. The full suite runs only in Task 7 and in `acta:land`.

## Spec deviations (need the user's yes with the plan)

1. The spec says a usage error from `acta run-one` exits 2. In acta, exit 2 means "skipped" (`exitSkipped` in `internal/cli/cli.go`), and every other usage error exits 1 (`exitBadInput`). This plan uses exit 1.
2. The spec lists "the move of the shared lock helpers". No move is needed: the new `HoldRunOne` sits in package `write` next to `lockDir` and `safeDir`, and the CLI calls it. That is a smaller diff, and the test hook `lockRoot` keeps working.
3. `scripts/test` adds `./...` when no argument starts with `./`, so `scripts/test -count=1` still tests the whole repo and not only the root folder.

## File Map

- `.acta/plans/2026-09-30-faster-agent-test-runs.md` (this file): the `## Timings` table, filled by Task 1 (before) and Task 7 (after).
- `internal/write/runone.go` (new), `internal/write/runone_test.go` (new): `HoldRunOne`, the machine-wide lock.
- `internal/cli/runone.go` (new), `internal/cli/runone_test.go` (new), `internal/cli/cli.go` (the `Run` switch and the unknown-command line): `acta run-one`.
- `cmd/acta/runone_test.go` (new): a stop signal reaches the command.
- `scripts/test` (new, mode 0755), `scripts/test_test.go` (new): the repo's test entry.
- `plugin/skills/plan/SKILL.md`, `plugin/skills/build/SKILL.md`, `plugin/skills/dispatch/SKILL.md`, `plugin/skills/land/SKILL.md`, `plugin/skills/review/SKILL.md`, and their tests `internal/plugincheck/skill_{plan,build,dispatch,land,review}_test.go`: the narrow-tests rule.
- `internal/tui/view_test.go`, `internal/tui/frame_test.go`, `internal/tui/watch_test.go` (only if `-race` trips): parallel sweeps.
- `internal/write/ops_test.go` and other `internal/write/*_test.go`, `cmd/acta/main_test.go` and other `cmd/acta/*_test.go`, `internal/cli/{cli,dispatch,doctor,migrate}_test.go` and other `internal/cli/*_test.go`, `internal/trees/trees_test.go`: git identity and `t.Parallel()`. Never `runone_test.go` in any package.

## Waves

- Wave 1: Task 1 alone. It measures the untouched base, so nothing else may run or change files while it runs.
- Wave 2: Task 2, Task 3, Task 4, Task 5. No two tasks share a file.
- Wave 3: Task 6. It edits tests in the same packages Task 2 builds in (write, cli, cmd/acta), so it waits until those packages compile with Task 2 in them.
- Wave 4: Task 7 alone. It measures, so nothing else may run.

---

### Task 1: Measure the base before any change

**Files:**
- Modify: `.acta/plans/2026-09-30-faster-agent-test-runs.md` (fill the "before" rows of `## Timings`)

**verify:** Every "before" row comes from a real run on the base commit with no change from this plan in the tree (`git diff --stat <base>..HEAD` lists only this plan file). Each row shows the command, the 1-minute load average read right before the run, the wall seconds, and whether it passed. A run taken while the load was above 4 is marked as such, never hidden or dropped.

**Interfaces:**
- Consumes: nothing.
- Produces: the "before" rows that Task 7 compares against.

- [x] **Step 1: Wait for a quiet machine**

Run `uptime`. If the 1-minute load average is above 4, run it again every minute, for up to 10 minutes. After 10 minutes, measure anyway and write `busy` in the load column.

- [x] **Step 2: Run each command twice, fresh**

Run each line twice, one at a time, with `uptime` right before each run:

```bash
/usr/bin/time -p go test -count=1 -short ./... 2>&1 | tail -20
/usr/bin/time -p go test -count=1 ./... 2>&1 | tail -20
```

Keep the `real` seconds, and from the full run keep the per-package time of `internal/tui`, `internal/write`, `internal/cli` and `cmd/acta`.

- [x] **Step 3: Write the rows**

Fill the before rows of the table in `## Timings` at the end of this file, one row per run, then the per-package line.

- [x] **Step 4: Commit**

```bash
git add .acta/plans/2026-09-30-faster-agent-test-runs.md
git commit -m "Record test timings before the speedup"
```

---

### Task 2: `acta run-one` runs one command at a time per machine

**Files:**
- Create: `internal/write/runone.go`, `internal/write/runone_test.go`
- Create: `internal/cli/runone.go`, `internal/cli/runone_test.go`
- Create: `cmd/acta/runone_test.go`
- Modify: `internal/cli/cli.go` (a `case "run-one":` in the `Run` switch; add `run-one` to the unknown-command list)

**verify:** Two holders of the run-one lock never overlap on any path: a second caller waits, says so once on stderr, and runs as soon as the first ends, including when the first holder is killed. Every exit path is listed with its code: the command's own exit code (0 and non-zero) passes back unchanged; a missing `--`, a bare `--`, or no args exits 1 with the usage line and runs nothing; a command that cannot start exits 3; a lock folder that is not the user's own or is open to others exits 3 and runs nothing; a stop signal (SIGINT or SIGTERM) sent to acta reaches the command, so no command outlives acta without the lock. stdin, stdout and stderr pass through byte for byte.

**Interfaces:**
- Consumes: `lockDir() (string, error)`, `safeDir(dir string) error`, the test hook `lockRoot` and `TestMain` in `internal/write/main_test.go`, and the child-process pattern of `TestHelperHoldLock` in `internal/write/tick_test.go`; `exitOK`, `exitBadInput`, `exitOther` in `internal/cli/cli.go`; `actaBin` from `cmd/acta/main_test.go`.
- Produces: `write.HoldRunOne(waiting func()) (release func(), err error)`; the command `acta run-one -- <cmd> [args...]`. Task 3 calls it by that name.

- [x] **Step 1: Write the failing lock tests**

`internal/write/runone_test.go`:

```go
package write

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestHelperHoldRunOne is not a real test. The tests below run it in a
// child process to hold the run-one lock and then get killed.
func TestHelperHoldRunOne(t *testing.T) {
	if os.Getenv("ACTA_HOLD_RUN_ONE") == "" {
		t.Skip("helper for the run-one tests")
	}
	lockRoot = os.Getenv("ACTA_LOCK_ROOT")
	if _, err := HoldRunOne(func() {}); err != nil {
		t.Fatal(err)
	}
	os.Stdout.WriteString("held\n")
	// Wait to be killed. A bare select{} trips Go's deadlock check.
	time.Sleep(time.Minute)
}

// holdInChild takes the run-one lock in a child process and returns once
// the child holds it. The caller kills the child to let go.
func holdInChild(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldRunOne$")
	cmd.Env = append(os.Environ(), "ACTA_HOLD_RUN_ONE=1", "ACTA_LOCK_ROOT="+lockRoot)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	buf := make([]byte, 5)
	if _, err := io.ReadFull(out, buf); err != nil || string(buf) != "held\n" {
		t.Fatalf("the child did not take the lock: %q %v", buf, err)
	}
	return cmd
}

func TestHoldRunOneWaitsForTheHolder(t *testing.T) {
	useLockBase(t)
	child := holdInChild(t)
	waited := 0
	got := make(chan error, 1)
	go func() {
		release, err := HoldRunOne(func() { waited++ })
		if err == nil {
			release()
		}
		got <- err
	}()
	select {
	case err := <-got:
		t.Fatalf("took the lock while the child held it: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	child.Process.Kill()
	child.Wait()
	select {
	case err := <-got:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("still waiting after the holder died")
	}
	if waited != 1 {
		t.Fatalf("waiting ran %d times, want 1", waited)
	}
}

func TestHoldRunOneFreeLockDoesNotWait(t *testing.T) {
	useLockBase(t)
	waited := false
	release, err := HoldRunOne(func() { waited = true })
	if err != nil {
		t.Fatal(err)
	}
	release()
	if waited {
		t.Fatal("said it was waiting on a free lock")
	}
	// A released lock can be taken again at once.
	release, err = HoldRunOne(func() { t.Fatal("waited on a released lock") })
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestHoldRunOneRefusesAnOpenFolder(t *testing.T) {
	dir := useLockBase(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := HoldRunOne(func() {}); err == nil {
		t.Fatal("took a lock in a folder open to other users")
	}
	if _, err := os.Stat(filepath.Join(dir, "run-one.lock")); !os.IsNotExist(err) {
		t.Fatalf("made a lock file in a folder it refused: %v", err)
	}
}
```

`useLockBase(t)` already exists in `internal/write/tick_test.go` (it points `lockRoot` at a temp folder and returns it). These tests change `lockRoot`, so none of them calls `t.Parallel()`.

- [x] **Step 2: Run them and watch them fail**

Run: `go test ./internal/write/ -run 'RunOne' -count=1`
Expected: FAIL, `undefined: HoldRunOne`.

- [x] **Step 3: Write `HoldRunOne`**

`internal/write/runone.go`:

```go
package write

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// runOneName is the one lock file every acta run-one of this user shares.
// Runs fight over the machine's CPU and memory, not over one repo, so the
// name holds no repo path.
const runOneName = "run-one.lock"

// HoldRunOne takes the machine-wide run-one lock and waits as long as it
// takes. waiting runs once, the first time the lock is busy, so the caller
// can say why nothing is happening yet. The kernel drops the lock when the
// process ends, even in a crash, so a dead holder never blocks anyone.
func HoldRunOne(waiting func()) (func(), error) {
	dir, err := lockDir()
	if err != nil {
		return nil, err
	}
	// The cache folder and the acta folder under it are not there yet on a
	// first run, so make them closed to everyone else.
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		return nil, err
	}
	if err := safeDir(dir); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, runOneName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		waiting()
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	}
	if err != nil {
		f.Close()
		return nil, err
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}
```

- [x] **Step 4: Run the lock tests and watch them pass**

Run: `go test ./internal/write/ -run 'RunOne|Lock' -count=1`
Expected: PASS.

- [x] **Step 5: Write the failing CLI tests**

`internal/cli/runone_test.go`:

```go
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// runOne calls acta run-one with its lock in a temp cache folder, so a test
// never waits on a real full run of this machine. The cache folder comes
// from HOME on macOS and from XDG_CACHE_HOME on Linux, so both move.
func runOne(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	var out, errb bytes.Buffer
	code := Run(append([]string{"run-one"}, args...), strings.NewReader(stdin), false, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRunOnePassesTheExitCodeBack(t *testing.T) {
	for _, want := range []int{0, 1, 7} {
		code, _, _ := runOne(t, "", "--", "sh", "-c", "exit "+strconv.Itoa(want))
		if code != want {
			t.Errorf("exit %d came back as %d", want, code)
		}
	}
}

func TestRunOnePassesOutputAndInputThrough(t *testing.T) {
	code, out, errb := runOne(t, "from stdin\n", "--", "sh", "-c", "cat; echo to-err >&2")
	if code != 0 || out != "from stdin\n" || errb != "to-err\n" {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errb)
	}
}

func TestRunOneRefusesBadUsage(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	for _, args := range [][]string{
		nil,
		{"--"},
		{"touch", marker},
		{"-x", "--", "touch", marker},
	} {
		code, _, errb := runOne(t, "", args...)
		if code != exitBadInput || !strings.Contains(errb, "usage: acta run-one -- <command> [args...]") {
			t.Errorf("%q: code %d, stderr %q", args, code, errb)
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("a bad call still ran the command")
	}
}

func TestRunOneCommandThatCannotStart(t *testing.T) {
	code, _, errb := runOne(t, "", "--", filepath.Join(t.TempDir(), "no-such-command"))
	if code != exitOther || errb == "" {
		t.Fatalf("code %d, stderr %q", code, errb)
	}
}
```

These tests set `HOME`, so none of them calls `t.Parallel()`.

`cmd/acta/runone_test.go`:

```go
package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A stop sent to acta has to reach the command. Otherwise acta dies, the
// lock goes free, and the command keeps eating the machine with no lock.
func TestRunOneStopReachesTheCommand(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		home := t.TempDir()
		cmd := exec.Command(actaBin, "run-one", "--", "sh", "-c", "echo $$; exec sleep 30")
		cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME="+filepath.Join(home, ".cache"))
		out, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		line, err := bufio.NewReader(out).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			t.Fatal(err)
		}
		cmd.Process.Signal(sig)
		done := make(chan struct{})
		go func() { cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			t.Fatalf("%v: acta did not stop", sig)
		}
		if err := syscall.Kill(pid, 0); err == nil {
			syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("%v: the command outlived acta", sig)
		}
	}
}
```

- [x] **Step 6: Run them and watch them fail**

Run: `go test ./internal/cli/ -run RunOne -count=1` and `go test ./cmd/acta/ -run RunOne -count=1`
Expected: FAIL, `unknown command "run-one"` (exit 1 instead of the command's code).

- [x] **Step 7: Write the command**

`internal/cli/runone.go`:

```go
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/iyay/acta/internal/write"
)

// cmdRunOne runs one command while it holds the machine-wide run-one lock,
// so two full test runs from two worktrees never fight over the CPU. The
// command's own output and exit code come back unchanged.
func cmdRunOne(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 || args[0] != "--" {
		fmt.Fprintln(stderr, "usage: acta run-one -- <command> [args...]")
		return exitBadInput
	}
	release, err := write.HoldRunOne(func() {
		fmt.Fprintln(stderr, "waiting for another run-one to finish")
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	defer release()
	cmd := exec.Command(args[1], args[2:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	// Catch a stop only now, after the wait, so a stop while waiting still
	// ends acta at once. From here on the stop goes to the command, or acta
	// would die and leave the command running with no lock.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	go func() {
		for s := range sig {
			cmd.Process.Signal(s)
		}
	}()
	err = cmd.Wait()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() >= 0 {
		return exit.ExitCode()
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitOther
	}
	return exitOK
}
```

In `internal/cli/cli.go`, inside the `Run` switch, before `default:`:

```go
	case "run-one":
		return cmdRunOne(args[1:], stdin, stdout, stderr)
```

and in the `default:` line, change `dispatch init or reply-back` to `dispatch init, reply-back or run-one`.

- [x] **Step 8: Run them and watch them pass**

Run: `go test ./internal/write/ ./internal/cli/ ./cmd/acta/ -count=1 -race`
Expected: PASS.

- [x] **Step 9: Commit**

```bash
gofmt -l cmd internal scripts && go vet ./...
git add internal/write/runone.go internal/write/runone_test.go internal/cli/runone.go internal/cli/runone_test.go internal/cli/cli.go cmd/acta/runone_test.go
git commit -m "Add acta run-one to run one full suite at a time per machine"
```

---

### Task 3: `scripts/test`, the repo's test entry

**Files:**
- Create: `scripts/test` (mode 0755)
- Test: `scripts/test_test.go`

**verify:** Every way to call the script maps to exactly one `go` call, listed with what it runs: no args; `--full` alone; package args; flag args with no package; `--full` with flags; `--full` with packages. Only `--full` goes through `acta run-one`, a plain call never does, and every call runs from the repo root whatever the caller's folder. The script never drops or reorders the caller's args.

**Interfaces:**
- Consumes: the command name `acta run-one -- <cmd>` from Task 2 (by name only; the test fakes `go`).
- Produces: `scripts/test` and `scripts/test --full`, which Task 4 names in the skills and Task 7 times.

- [x] **Step 1: Write the failing test**

`scripts/test_test.go`:

```go
package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// callScript runs scripts/test with a fake go that prints its folder and its
// args, so the test sees the exact go call without running any tests.
func callScript(t *testing.T, args ...string) string {
	t.Helper()
	bin := t.TempDir()
	fake := "#!/bin/sh\necho \"$(pwd) $*\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs("test")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(script, args...)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "PATH="+bin+":/usr/bin:/bin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestScriptMapsEachCallToOneGoCall(t *testing.T) {
	t.Parallel()

	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, "test -short ./..."},
		{[]string{"./internal/tui/"}, "test -short ./internal/tui/"},
		{[]string{"-count=1"}, "test -short -count=1 ./..."},
		{[]string{"-run", "TestX", "./internal/tui/"}, "test -short -run TestX ./internal/tui/"},
		{[]string{"--full"}, "run ./cmd/acta run-one -- go test ./..."},
		{[]string{"--full", "-count=1"}, "run ./cmd/acta run-one -- go test -count=1 ./..."},
		{[]string{"--full", "./internal/cli/"}, "run ./cmd/acta run-one -- go test ./internal/cli/"},
	} {
		if got := callScript(t, tc.args...); got != root+" "+tc.want {
			t.Errorf("%q: got %q, want %q", tc.args, got, root+" "+tc.want)
		}
	}
}
```

- [x] **Step 2: Run it and watch it fail**

Run: `go test ./scripts/ -count=1`
Expected: FAIL, the script `test` is not there (`no such file or directory`).

- [x] **Step 3: Write the script**

`scripts/test`, then `chmod 0755 scripts/test`:

```sh
#!/bin/sh
# Runs this repo's Go tests. A plain run skips the slow screen-size sweeps.
# --full runs every test, but first waits for any other full run on this
# machine to end, so agents in other worktrees do not fight over the CPU.
# With no package given, the whole repo is tested.
set -eu
cd "$(dirname "$0")/.."
full=
if [ "${1:-}" = "--full" ]; then
	full=1
	shift
fi
pkgs=./...
for a in "$@"; do
	case $a in ./*) pkgs= ;; esac
done
if [ -n "$full" ]; then
	exec go run ./cmd/acta run-one -- go test "$@" $pkgs
fi
exec go test -short "$@" $pkgs
```

- [x] **Step 4: Run it and watch it pass**

Run: `go test ./scripts/ -count=1`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l cmd internal scripts && go vet ./...
git add scripts/test scripts/test_test.go
git commit -m "Add scripts/test: short by default, --full waits its turn"
```

---

### Task 4: Skills run narrow tests, and the full suite only at land

**Files:**
- Modify: `plugin/skills/plan/SKILL.md` (new `## Test commands` section after `## Verify lines are properties`; a `**Tests:**` line in the header template after `**Spec:**`)
- Modify: `plugin/skills/build/SKILL.md` (Step 3 baseline sentence; the first sentence under `## Close`)
- Modify: `plugin/skills/dispatch/SKILL.md` (the `GATES` line of the brief; the test line of the Phase 2 check block)
- Modify: `plugin/skills/land/SKILL.md` (step 1 precondition; step 6)
- Modify: `plugin/skills/review/SKILL.md` (the reviewers-may-run-tests sentence)
- Test: `internal/plugincheck/skill_plan_test.go`, `skill_build_test.go`, `skill_dispatch_test.go`, `skill_land_test.go`, `skill_review_test.go`
- Added during build (orchestrator ruling, spec section 3.1 wins): `plugin/skills/build/implementer-prompt.md` (it told implementers to run the full suite before each commit), `plugin/skills/tdd/SKILL.md` (checklist line "All tests pass"), `internal/plugincheck/skill_tdd_test.go`

**verify:** No skill still sends an agent to the full suite outside `acta:land`: list every place in `plugin/skills/` that names a test run (`grep -rn "test suite\|go test\|npm test\|pytest\|cargo test\|run the tests" plugin/skills/`) and say for each whether it is narrow, fast, or the land gate. The land gate always runs the full suite once under `acta run-one`, and skips the rerun after the merge only when the merge tree equals the branch tree. The text stays stack-neutral: Go appears only as an example, next to other stacks. Every skill stays under its `MaxLines`, and the skills total stays under the 4240 cap in `TestTotalSkillSize`.

**Interfaces:**
- Consumes: the names `acta run-one --`, `scripts/test`, `scripts/test --full` from Tasks 2 and 3.
- Produces: skill text only.

- [x] **Step 1: Write the failing checks**

Add to the `Must` list of each rule:

- `skill_plan_test.go`: `"## Test commands"`, `"Never `./...` in a task"`, `"**Tests:**"`
- `skill_build_test.go`: `"The full suite waits for `acta:land`"`, `"scripts/test"`, and to `MustNot`: `"or `go test ./...`, whichever the project uses"`, `"run the full test suite and the type checks, show the output, then use"`
- `skill_dispatch_test.go`: `"GATES (from the worktree): <the plan's fast test command>"`, and to `MustNot`: `"<one-shot test runner>"`
- `skill_land_test.go`: `"acta run-one -- <full command>"`, `"HEAD^{tree}"`, `"tree same as branch, gates reused"`
- `skill_review_test.go`: `"never the full suite"`

- [x] **Step 2: Run them and watch them fail**

Run: `go test ./internal/plugincheck/ -run 'TestSkill(Plan|Build|Dispatch|Land|Review)$' -count=1`
Expected: FAIL, one "missing" line per new string.

- [x] **Step 3: Change the skill text**

`plugin/skills/plan/SKILL.md`, a new section right before `## Waves`:

```markdown
## Test commands

Every run step in a task names the narrowest command that proves it: the package or file the task touched, for example `go test ./internal/tui/ -run TestX`, `pytest tests/test_x.py::test_y` or `npm test -- x.test.ts`. Never `./...` in a task, and never the whole suite. The full suite runs once, in `acta:land`. Why: agents in other worktrees share the same machine, and each full run slows every other one down.

The header names both commands on a `**Tests:**` line: the fast one (a short mode, or the touched packages) and the full one. When the repo has `scripts/test`, they are `scripts/test` and `scripts/test --full`.
```

In the header template, right after the `**Spec:** ...` paragraph:

```markdown
**Tests:** [fast command, full command — see Test commands]
```

`plugin/skills/build/SKILL.md`, Step 3 sentence becomes:

```markdown
Run the project's fast tests to make sure the workspace starts clean: `scripts/test` when the repo has it, else the plan's `**Tests:**` fast command (a short mode such as `go test -short ./...`, or `npm test`, `cargo test`, `pytest` on the touched parts). The full suite waits for `acta:land`.
```

and the first sentence under `## Close` becomes:

```markdown
When every task is committed: run the fast tests and the type checks, show the output, then use `acta:review` over `<parent>..HEAD`.
```

`plugin/skills/dispatch/SKILL.md`, the brief line becomes:

```
GATES (from the worktree): <the plan's fast test command>; typecheck; git diff --stat vs <base-sha> shows only plan files.
```

and in the Phase 2 check block, `<one-shot test runner> 2>&1 | tail -20       # real counts, shown` becomes `<the plan's fast test command> 2>&1 | tail -20   # real counts, shown`.

`plugin/skills/land/SKILL.md`, in step 1, after "the full test suite and type checks green with the output shown", add the sentence: `Run the full suite as `acta run-one -- <full command>` (`scripts/test --full` already does this), so only one full run uses the machine at a time.` Step 6 becomes:

```markdown
6. Run the gates again on the merge result, unless `git rev-parse HEAD^{tree}` prints the same tree as `git rev-parse <branch>^{tree}`: then the merge holds exactly the files the gates just passed, so write "tree same as branch, gates reused" in the report instead. Red: say so plainly and leave the merge for the user.
```

`plugin/skills/review/SKILL.md`, in the sentence "they may run tests to prove a finding", change that clause to "they may run the narrow tests that prove a finding, never the full suite".

- [x] **Step 4: Run them and watch them pass**

Run: `go test ./internal/plugincheck/ -count=1`
Expected: PASS, including `TestTotalSkillSize`.

- [x] **Step 5: Commit**

```bash
gofmt -l cmd internal scripts && go vet ./...
git add plugin/skills internal/plugincheck
git commit -m "Skills run narrow tests; the full suite runs once, at land, under run-one"
```

---

### Task 5: Sweep tests spread over every core

**Files:**
- Modify: `internal/tui/view_test.go` (`TestTheTabBoxHoldsItsWidthOnEveryScreen`, `TestViewNeverOverflowsAnyWindow`, `TestNoRoundedCorners`, `TestThumbSitsOnTheBorderNotInside`)
- Modify: `internal/tui/frame_test.go` (`TestViewFitsEveryTerminalSize`)
- Modify: `internal/tui/watch_test.go` (only if Step 4 shows a race there)

**verify:** Each of the five tests checks exactly the sizes, tabs, panes, popups and scroll spots it checks today, in both `-short` and full runs (count the checked cases before and after with a temporary counter, and do not commit the counter). No two parallel subtests share a value one of them writes: list for each test what is shared (read-only) and what each subtest builds for itself. A failure still names its size. `go test -race` passes for the package.

**Interfaces:**
- Consumes: `sweep`, `edgeWidths`, `edgeHeights`, `sized`, `press`, `newModel`, `withColors`, `checkThumbOnTheBorder`, `topTabs`, `tabKey` in the tui tests.
- Produces: nothing new.

- [x] **Step 1: Time the package before**

Run: `go test -count=1 ./internal/tui/ -run 'TestViewNeverOverflowsAnyWindow|TestNoRoundedCorners|TestThumbSitsOnTheBorderNotInside|TestViewFitsEveryTerminalSize|TestTheTabBoxHoldsItsWidthOnEveryScreen' -v 2>&1 | grep -E '^(--- |ok)'`
Keep the times; they go in the commit message. This change edits tests only, so there is no red test; the proof is the same case count, a green race run and the times.

- [x] **Step 2: Split each sweep's outer loop into parallel subtests**

`TestViewNeverOverflowsAnyWindow` becomes:

```go
func TestViewNeverOverflowsAnyWindow(t *testing.T) {
	t.Parallel()

	// One subtest per width, so the sizes spread over every core instead of
	// one long loop holding the whole package up.
	for _, w := range sweep(30, 200, edgeWidths) {
		t.Run(fmt.Sprintf("w%d", w), func(t *testing.T) {
			t.Parallel()
			for _, h := range sweep(10, 60, edgeHeights) {
				base := sized(newModel(t), w, h)
				for name, m := range map[string]Model{
					"open":   base,
					"detail": press(base, tabKey(tabPlans)),
					"help":   press(base, "?"),
				} {
					for _, line := range strings.Split(sized(m, w, h).View(), "\n") {
						if got := lipgloss.Width(line); got > w {
							t.Fatalf("%dx%d %s: line is %d cells wide, the window is %d: %q", w, h, name, got, w, line)
						}
					}
				}
			}
		})
	}
}
```

Do the same in `TestTheTabBoxHoldsItsWidthOnEveryScreen` (outer loop `w`) and `TestNoRoundedCorners` (outer loop `w`; `focuses` stays built once before the loop, it is only read). Keep the loop body as it is, only wrapped.

`TestViewFitsEveryTerminalSize`: wrap the outer `h` loop the same way, `t.Run(fmt.Sprintf("h%d", h), ...)`, and move `m := newModel(t)` inside the subtest. Today one model serves every size; its `dcache` is a pointer, so parallel subtests sharing it would race.

`TestThumbSitsOnTheBorderNotInside` stays serial at the top, because `withColors` sets a color setting the whole package shares. Inside `withColors`, wrap the loops in one group subtest that waits for its parallel children, so they all run while the colors are on:

```go
	withColors(func() {
		sizes := [][2]int{{80, 30}, {160, 50}}
		if testing.Short() {
			// One size keeps the check on every tab and pane. The full run adds the wide one.
			sizes = sizes[:1]
		}
		// The group returns only when every subtest in it is done, so all of
		// them run while the colors are on.
		t.Run("group", func(t *testing.T) {
			for _, size := range sizes {
				for i := range topTabs {
					for _, p := range append(press(newModel(t), tabKey(i)).panes(), paneDetail) {
						t.Run(fmt.Sprintf("%dx%d/%s/%d", size[0], size[1], topTabs[i].name, p), func(t *testing.T) {
							t.Parallel()
							for _, at := range []string{"top", "middle", "end"} {
								checkThumbOnTheBorder(t, i, p, at, size[0], size[1])
							}
						})
					}
				}
			}
		})
	})
```

Add `"fmt"` to the imports where it is missing.

- [x] **Step 3: Check the case count is the same**

With a temporary counter (an `atomic.Int64` bumped once per checked case), run each of the five tests with and without `-short` before and after the change; the counts must match. Remove the counter before the commit.

- [x] **Step 4: Run the package, with the race detector**

Run: `go test -count=1 ./internal/tui/` and `go test -count=1 -race -short ./internal/tui/`
Expected: PASS. If the race detector reports `TestWatchGathersEventsIntoOneReload` in `internal/tui/watch_test.go` (the read of `loads` after the `select`), read it under the lock:

```go
	mu.Lock()
	got := loads
	mu.Unlock()
	if got != 1 {
		t.Fatalf("loads = %d, want 1 for a burst of writes", got)
	}
```

- [x] **Step 5: Commit**

```bash
gofmt -l cmd internal scripts && go vet ./...
git add internal/tui
git commit -m "Split tui size sweeps into parallel subtests (before: <times>, after: <times>)"
```

---

### Task 6: Git test helpers without `t.Setenv`, and more parallel tests

**Files:**
- Modify: `internal/write/ops_test.go` (`repoWith`) and any other `internal/write/*_test.go` that sets `GIT_AUTHOR_*` or `GIT_COMMITTER_*`
- Modify: `cmd/acta/main_test.go` (`TestMain`, `fixtureRepo`) and other `cmd/acta/*_test.go` test functions
- Modify: `internal/cli/cli_test.go` (`debtRepo`), `internal/cli/dispatch_test.go` (its repo helper), `internal/cli/doctor_test.go` (`doctorRepo`), `internal/cli/migrate_test.go` (`migrateRepo`), and other `internal/cli/*_test.go` test functions
- Modify: `internal/trees/trees_test.go` (`setup`, a new `TestMain`)
- Never: any `runone_test.go` (Task 2 owns them; they must stay serial)

**verify:** No test in write, cmd/acta, cli or trees sets `GIT_AUTHOR_*` or `GIT_COMMITTER_*` any more (`grep -rn 'GIT_AUTHOR\|GIT_COMMITTER' internal/write internal/cli internal/trees cmd/acta` prints nothing; build ruling: the one allowed hit is `internal/write/ids_test.go`, which sets `GIT_AUTHOR_DATE` and `GIT_COMMITTER_DATE` on one git child's `cmd.Env`, not on the test process), and no test leans on the user's own git identity: the four packages pass with `GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1`. A test gets `t.Parallel()` only when neither it nor any helper it calls uses `t.Setenv`, `os.Setenv`, `t.Chdir` or `os.Chdir`, or writes a package-level variable (such as `lockRoot` or `runTUI`); list for each package how many tests are parallel before and after, and name the reason each remaining serial test has to stay serial. `go test -race` passes for the four packages.

**Interfaces:**
- Consumes: the helpers named above; `gitRun(t, dir, args...)` in `internal/write`.
- Produces: nothing new.

- [x] **Step 1: Count the parallel tests before**

Run, per package: `grep -c 't.Parallel()' <pkg>/*_test.go | awk -F: '{s+=$2} END {print s}'` and `grep -c '^func Test' <pkg>/*_test.go | awk -F: '{s+=$2} END {print s}'` for `internal/write`, `internal/cli`, `internal/trees`, `cmd/acta`. Keep the numbers for the commit message. This change edits tests only, so there is no red test; the proof is the green race run, the no-global-identity run and the counts.

- [x] **Step 2: Set identity in the repo, not in the env**

In each helper, delete the `t.Setenv("GIT_AUTHOR_NAME", ...)` style lines (and loops) and, right after the helper's `git init`, set the identity inside that repo. In `repoWith`:

```go
	gitRun(t, dir, "init", "-q", "-b", "main")
	// Name the author inside the repo, not in the env, so tests that make
	// commits can still run side by side.
	gitRun(t, dir, "config", "user.name", "test")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "add", ".")
```

In helpers that use a local `run := func(args ...string)`, do the same with `run("config", "user.name", "test")` and `run("config", "user.email", "test@example.com")` right after `run("init", ...)`. A helper that runs `git clone` or makes a second repo sets the same two lines in that repo too.

In `cmd/acta/main_test.go`, delete `t.Setenv("PM_ROOT", "")` from `fixtureRepo` and add, at the top of `TestMain`:

```go
	// An inherited PM_ROOT would point every test at the wrong board. Clear
	// it once here, so no helper has to set env and block parallel tests.
	os.Unsetenv("PM_ROOT")
```

In `internal/trees/trees_test.go`, delete `t.Setenv("PM_ROOT", "")` from `setup` and add:

```go
// TestMain clears an inherited PM_ROOT once, so no helper has to set env
// and block parallel tests.
func TestMain(m *testing.M) {
	os.Unsetenv("PM_ROOT")
	os.Exit(m.Run())
}
```

- [x] **Step 3: Prove no test leans on the user's git identity**

Run: `GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 go test -count=1 ./internal/write/ ./internal/cli/ ./internal/trees/ ./cmd/acta/`
Expected: PASS. A failure that is not about identity (for example a branch name) is reported in the task report, not patched here.

- [x] **Step 4: Add `t.Parallel()` where it is safe**

For each top-level test in the four packages, add `t.Parallel()` as the first line, followed by one blank line (the style the tui tests use), only when the rule in the verify line holds. Leave every other test as it is.

- [x] **Step 5: Run the four packages, with the race detector**

Run: `go test -count=1 -race ./internal/write/ ./internal/cli/ ./internal/trees/ ./cmd/acta/`
Expected: PASS. Then count the parallel tests again as in Step 1.

- [x] **Step 6: Commit**

```bash
gofmt -l cmd internal scripts && go vet ./...
git add internal/write internal/cli internal/trees cmd/acta
git commit -m "Git test helpers set identity in the repo; run safe tests in parallel (parallel before/after: <counts>)"
```

---

### Task 7: Measure after, and the full gates

**Files:**
- Modify: `.acta/plans/2026-09-30-faster-agent-test-runs.md` (fill the "after" rows of `## Timings`)

**verify:** Every "after" row comes from a real run on the finished branch, taken the same way as the "before" rows (same quiet-machine rule, two runs each, load shown), so the two sets compare fairly. Both the full suite and the race run pass. The report states the before and after numbers side by side, and says plainly whether the 40s full-suite target was met; a miss is reported with the real numbers.

**Interfaces:**
- Consumes: the "before" rows from Task 1; `scripts/test` from Task 3.
- Produces: the "after" rows the landing report quotes.

- [x] **Step 1: Wait for a quiet machine** (same rule as Task 1 Step 1)

- [x] **Step 2: Run each command twice**

```bash
/usr/bin/time -p scripts/test -count=1 2>&1 | tail -20
/usr/bin/time -p scripts/test --full -count=1 2>&1 | tail -20
```

`scripts/test` runs the same `go test -short ./...` as the before row, and `scripts/test --full` runs the same `go test ./...` under `acta run-one`. Keep the `real`, `user` and `sys` seconds and, from the full run, the per-package times of `internal/tui`, `internal/write`, `internal/cli` and `cmd/acta`.

- [x] **Step 3: Run the race gate**

Run: `go test -count=1 -race -short ./...`
Expected: PASS.

- [x] **Step 4: Write the rows and commit**

Fill the after rows in `## Timings`, then:

```bash
git add .acta/plans/2026-09-30-faster-agent-test-runs.md
git commit -m "Record test timings after the speedup"
```

---

## Timings

Filled by Task 1 (before) and Task 7 (after). Wall time in seconds, from `/usr/bin/time -p`.

| when | command | load (1 min) | wall (s) | result |
|---|---|---|---|---|
| before | `go test -count=1 -short ./...` | busy (9.46) | 39.36 | pass (1100 tests, 15 packages) |
| before | `go test -count=1 -short ./...` | busy (21.79) | 63.55 | pass (1100 tests, 15 packages) |
| before | `go test -count=1 ./...` | busy (64.40) | 90.72 | pass (all 14 test packages ok) |
| before | `go test -count=1 ./...` | busy (47.34) | 128.76 | pass (all 14 test packages ok) |
| after | `scripts/test -count=1` | busy (35.80) | 35.75 | pass (all 15 test packages ok) |
| after | `scripts/test -count=1` | busy (40.60) | 34.37 | pass (all 15 test packages ok) |
| after | `scripts/test --full -count=1` | busy (46.66) | 45.84 | pass (all 15 test packages ok, no run-one wait) |
| after | `scripts/test --full -count=1` | busy (57.95) | 55.35 | pass (all 15 test packages ok, no run-one wait) |

Per package, full run (tui / write / cli / cmd/acta):

- before: run 1: 86.502 / 35.849 / 34.618 / 33.149; run 2: 123.113 / 48.418 / 49.391 / 49.017 (the load stayed above 4 for the whole 10-minute wait, so every before run was busy)
- after: run 1: 41.906 / 33.643 / 36.945 / 16.742; run 2: 51.188 / 39.946 / 43.218 / 17.638 (the load stayed above 4 for the whole 10-minute wait and jumped past 150 near its end, so every after run was busy too)

CPU seconds (user / sys, from `/usr/bin/time -p`). The load was high for every before run, so CPU time is the fairer number to compare; wall time is still the one the user feels.

- before, short: 57.17 / 72.81 and 57.26 / 72.60
- before, full: 105.90 / 109.01 and 106.61 / 108.22
- after, short: 50.85 / 68.39 and 50.07 / 70.49
- after, full: 87.44 / 108.56 and 87.65 / 106.13
