---
created: "2026-09-30"
parent: specs/2026-09-30-tui-tick-items-design
id: PLN-0052
hash: zb024n6
started: "2026-09-30"
finished: "2026-09-30"
---
# Mark Tasks and Debt Lines Done from the TUI Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** In the TUI, `+` marks a task or a debt line done and `-` puts it back to open, `acta tick <id> --undo` does the same as `-` from the CLI, and the `?` help names one key per line.

**Architecture:** A new `write.MarkItem` sets every box of a task (or the one box of a debt line) to `x` or blank, dates the plan and spec on a done task, and commits the touched files the way `write.SetValue` does. The date code moves from `internal/cli/tick.go` into `write.TaskDates`, so the CLI and the TUI date the same way. The TUI calls `MarkItem` through an injectable `markItem` field, like `setValue`.

**Tech Stack:** Go, Bubble Tea, the repo's own `internal/gitc`.

**Spec:** `.acta/specs/2026-09-30-tui-tick-items-design.md`

**Tests:** fast `scripts/test <touched packages>`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- `+` marks the item done, `-` puts it back to open. On a task that means every box of the task; on a debt line its one box.
- `-` works from done and from wontfix. It does not remove `started` or `finished` dates already written.
- No wontfix for a task, no remove or drop command, no single-box tick from the TUI.
- The CLI tick keeps not committing, as today. Only the TUI path commits, through `write.MarkItem`.
- A TUI tick names no agent, so it writes no `.agents.json` record (`write.RecordAgent` writes nothing for an empty agent that is not starting).
- The CLI keeps exactly one action flag per call; `--undo` is one more action flag.
- Comments are plain English a 10-year-old reads back without stopping. They say why, not what.
- Run `gofmt -l .` and `go vet` on the touched packages before each commit, inside the task commit.

## File Map

| File | Task | Change |
|---|---|---|
| `internal/write/tick.go` | 1 | `TickText` and `Tick` share a state-taking core; new `UntickText`, `Untick` |
| `internal/write/mark.go` | 1 | new: `TaskDates`, `markDate`, `MarkItem` |
| `internal/write/mark_test.go` | 1 | new tests |
| `internal/write/tick_test.go` | 1 | `UntickText` tests |
| `internal/cli/tick.go` | 2 | `--undo`; `markTaskDates` and `markDate` removed, calls `write.TaskDates` |
| `internal/cli/tick_test.go` | 2 | `--undo` tests |
| `internal/tui/model.go` | 3 | `markItem` field, `+` and `-` keys, `markRow` |
| `internal/tui/model_test.go` | 3 | key tests |
| `internal/tui/view.go` | 4 | `helpLines` one key per line |
| `internal/tui/view_test.go` | 4 | help test |

## Waves

- Wave 1: Task 1, Task 4 (no shared file).
- Wave 2: Task 2, Task 3 (both need Task 1; no shared file).

---

### Task 1: `write.MarkItem`, `Untick` and `TaskDates`

**Files:**
- Modify: `internal/write/tick.go` (`TickText`, `Tick`)
- Create: `internal/write/mark.go`
- Create: `internal/write/mark_test.go`
- Modify: `internal/write/tick_test.go`

**verify:** For every item kind and state, `MarkItem` either changes exactly the boxes of that one task or debt line and nothing else in the file, or writes nothing and returns an error. A done task always leaves the same plan and spec dates `acta tick --all` leaves. A commit happens only when auto-commit is on and none of the touched files had changes of their own before; the commit holds only the touched files. List every refusal path (unknown id, legacy, worktree, wrong kind) and every commit path checked.

**Interfaces:**
- Produces: `func UntickText(src []byte, headingLine int) ([]byte, int, int, error)`, `func Untick(path string, headingLine int) (int, int, error)` (both return done and total boxes after the change), `func TaskDates(cfg config.Config, it *board.Item) ([]string, error)` (returns the paths it wrote), `func MarkItem(cfg config.Config, b *board.Board, id string, done bool) (Outcome, error)`.

- [x] **Step 1: Write the failing tests**

In `internal/write/tick_test.go`:

```go
func TestUntickTextClearsEveryBoxOfTheTaskOnly(t *testing.T) {
	src := []byte("# P\n\n### Task 1: A\n- [x] a\n- [x] b\n\n### Task 2: B\n- [x] c\n")
	out, done, total, err := UntickText(src, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := "# P\n\n### Task 1: A\n- [ ] a\n- [ ] b\n\n### Task 2: B\n- [x] c\n"
	if string(out) != want || done != 0 || total != 2 {
		t.Fatalf("got %q %d/%d", out, done, total)
	}
}
```

In `internal/write/mark_test.go` (package `write`, uses `repoWith`, `gitRun`, `mustLoad` from `ops_test.go`):

```go
package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var markFiles = map[string]string{
	".acta/specs/2026-09-30-s.md": "---\nid: SPC-0001\n---\n# S\n",
	".acta/plans/2026-09-30-p.md": "---\nid: PLN-0002\nparent: specs/2026-09-30-s\n---\n# P\n\n### Task 1: One\n- [ ] a\n- [ ] b\n\n### Task 2: Two\n- [ ] c\n",
	".acta/debt/2026-09-30-d.md":  "---\nid: DBT-0003\n---\n# D\n\n- [ ] x\n- [-] y\n",
}

func TestMarkItemDoneTicksTheWholeTaskDatesAndCommits(t *testing.T) {
	cfg := repoWith(t, markFiles)
	b := mustLoad(t, cfg)
	o, err := MarkItem(cfg, b, "plans/2026-09-30-p#task-1", true)
	if err != nil || !o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	plan := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md"))
	if !strings.Contains(plan, "- [x] a\n- [x] b\n") || !strings.Contains(plan, "- [ ] c") {
		t.Fatalf("plan boxes: %q", plan)
	}
	if !strings.Contains(plan, "started:") {
		t.Fatalf("plan not dated: %q", plan)
	}
	spec := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/specs/2026-09-30-s.md"))
	if !strings.Contains(spec, "started:") {
		t.Fatalf("spec not dated: %q", spec)
	}
	if st := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); st != "" {
		t.Fatalf("left uncommitted: %q", st)
	}
}

func TestMarkItemOpenClearsTaskAndDebtLine(t *testing.T) {
	cfg := repoWith(t, markFiles)
	if _, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", true); err != nil {
		t.Fatal(err)
	}
	if _, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-1", false); err != nil {
		t.Fatal(err)
	}
	plan := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md"))
	if !strings.Contains(plan, "- [ ] a\n- [ ] b\n") {
		t.Fatalf("task not open: %q", plan)
	}
	b := mustLoad(t, cfg)
	var wontfix string
	for _, it := range b.Items {
		if it.Path == filepath.Join(cfg.RepoRoot, ".acta/debt/2026-09-30-d.md") && it.Status == "wontfix" {
			wontfix = it.ID
		}
	}
	if wontfix == "" {
		t.Fatal("no wontfix debt line in the board")
	}
	if _, err := MarkItem(cfg, b, wontfix, false); err != nil {
		t.Fatal(err)
	}
	debt := readFile(t, filepath.Join(cfg.RepoRoot, ".acta/debt/2026-09-30-d.md"))
	if !strings.Contains(debt, "- [ ] x\n- [ ] y") {
		t.Fatalf("debt line not open: %q", debt)
	}
}

func TestMarkItemRefusesAndWritesNothing(t *testing.T) {
	cfg := repoWith(t, markFiles)
	b := mustLoad(t, cfg)
	for _, id := range []string{"nope", "specs/2026-09-30-s", "plans/2026-09-30-p"} {
		if _, err := MarkItem(cfg, b, id, true); err == nil {
			t.Errorf("%s: no error", id)
		}
	}
	if st := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); st != "" {
		t.Fatalf("refusal wrote files: %q", st)
	}
}

func TestMarkItemLeavesADirtyFileUncommitted(t *testing.T) {
	cfg := repoWith(t, markFiles)
	path := filepath.Join(cfg.RepoRoot, ".acta/plans/2026-09-30-p.md")
	if err := os.WriteFile(path, []byte(readFile(t, path)+"\nnote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	o, err := MarkItem(cfg, mustLoad(t, cfg), "plans/2026-09-30-p#task-2", true)
	if err != nil || o.Committed || !o.Skipped {
		t.Fatalf("outcome %+v err %v", o, err)
	}
}

```

If a task id in the board is not `plans/<stem>#task-N`, print `b.Items` ids once and use the real form; the form above is what `acta tick` takes today. `readFile` already lives in `ids_test.go`; reuse it.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/write -run 'TestUntickText|TestMarkItem' -v`
Expected: FAIL to build, `undefined: UntickText` and `undefined: MarkItem`.

- [x] **Step 3: Write the code**

In `internal/write/tick.go`, move the body of `TickText` into `boxText(src []byte, headingLine, step int, state byte)`; the one line that writes a box becomes `lines[i] = tickBoxRe.ReplaceAllString(lines[i], "${1}"+string(state)+"${3}")`. Then:

```go
func TickText(src []byte, headingLine, step int) ([]byte, int, int, error) {
	return boxText(src, headingLine, step, 'x')
}

// UntickText clears every box of the task whose heading is on headingLine,
// so a task marked done by mistake can go back to open.
func UntickText(src []byte, headingLine int) ([]byte, int, int, error) {
	return boxText(src, headingLine, 0, ' ')
}
```

Move the body of `Tick` into `rewriteTask(path string, headingLine int, edit func([]byte) ([]byte, int, int, error)) (int, int, error)`, keeping the lock and the tmp-file rename. `Tick` calls it with `func(src []byte) ([]byte, int, int, error) { return TickText(src, headingLine, step) }`, and new `Untick(path string, headingLine int)` calls it with `UntickText`.

Create `internal/write/mark.go`:

```go
package write

import (
	"bytes"
	"fmt"
	"os"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
)

// TaskDates writes the dates a task tick gives its plan and the spec or bug
// above it. Work began the moment any task is ticked, and work ends when
// every task of the plan is done, and the spec a little later: when every
// plan under it is done too. The board is read again because the tick that
// just happened is what the file says. It gives back the files it changed.
func TaskDates(cfg config.Config, it *board.Item) ([]string, error) {
	fresh, err := board.Load(cfg)
	if err != nil {
		return nil, err
	}
	var wrote []string
	mark := func(path string, set func([]byte) ([]byte, error)) error {
		changed, err := markDate(path, set)
		if changed {
			wrote = append(wrote, path)
		}
		return err
	}
	plan := fresh.Get(it.PlanID)
	if err := mark(plan.Path, MarkStarted); err != nil {
		return wrote, err
	}
	if plan.Done == plan.Total {
		if err := mark(plan.Path, MarkFinishedOnce); err != nil {
			return wrote, err
		}
	}
	spec := fresh.Get(plan.SpecID)
	if spec == nil {
		return wrote, nil
	}
	if err := mark(spec.Path, MarkStarted); err != nil {
		return wrote, err
	}
	if spec.Status == "done" {
		return wrote, mark(spec.Path, MarkFinishedOnce)
	}
	return wrote, nil
}

// markDate runs one date writer over a file and writes the result back only
// when the bytes really changed, so a tick never touches a file it has
// nothing new to say about.
func markDate(path string, set func([]byte) ([]byte, error)) (bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	out, err := set(src)
	if err != nil {
		return false, err
	}
	if bytes.Equal(src, out) {
		return false, nil
	}
	return true, os.WriteFile(path, out, 0o644)
}

// MarkItem sets every box of a task, or the one box of a debt line, to done
// or back to open. It is what the TUI + and - keys call, so a person gets one
// commit per press, the same way the status popup works.
func MarkItem(cfg config.Config, b *board.Board, id string, done bool) (Outcome, error) {
	it := b.Get(id)
	switch {
	case it == nil:
		return Outcome{}, bad("unknown id %s", id)
	case it.Legacy:
		return Outcome{}, bad("%s is a legacy file; move it into the root folder first", id)
	case it.Worktree != "":
		return Outcome{}, bad("%s is shown from worktree %s; edit it there", id, it.Worktree)
	case it.Kind != board.KindTask && it.Kind != board.KindDebtItem:
		return Outcome{}, bad("%s is not a task or a debt line", id)
	}
	// Check every file this press can touch before writing any of them.
	// Changes a person already had in one of them are not ours to commit.
	paths := []string{it.Path}
	if it.Kind == board.KindTask && done {
		if plan := b.Get(it.PlanID); plan != nil {
			if spec := b.Get(plan.SpecID); spec != nil {
				paths = append(paths, spec.Path)
			}
		}
	}
	dirty := false
	for _, p := range paths {
		d, err := dirtyBefore(cfg, p)
		if err != nil {
			return Outcome{}, err
		}
		dirty = dirty || d
	}
	state, word := byte(' '), "open"
	if done {
		state, word = 'x', "done"
	}
	wrote := []string{it.Path}
	switch {
	case it.Kind == board.KindDebtItem:
		if err := TickLine(it.Path, it.Line, state); err != nil {
			return Outcome{}, err
		}
	case done:
		if _, _, err := Tick(it.Path, it.Line, 0); err != nil {
			return Outcome{}, err
		}
		dated, err := TaskDates(cfg, it)
		wrote = append(wrote, dated...)
		if err != nil {
			return Outcome{Path: it.Path}, err
		}
	default:
		if _, _, err := Untick(it.Path, it.Line); err != nil {
			return Outcome{}, err
		}
	}
	msg := fmt.Sprintf("acta: %s %s", id, word)
	o := Outcome{Path: it.Path}
	switch {
	case !cfg.AutoCommit:
		o.Reason = "auto_commit is off"
	case dirty:
		o.Skipped, o.Reason = true, "file had other uncommitted changes"
	default:
		r := gitc.CommitPaths(cfg.RepoRoot, unique(wrote), msg)
		o.Committed, o.Reason, o.Skipped = r.Committed, r.Reason, !r.Committed
	}
	return o, nil
}

// unique drops repeats, since the plan file is both ticked and dated.
func unique(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}
```

If `bad`, `dirtyBefore` or `unique` clash with an existing name in the package, rename the new one and say so in the task report.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/write -v -run 'TestUntickText|TestMarkItem|TestTick'`
Expected: PASS, and the old `TestTick*` tests still pass.

- [x] **Step 5: Commit**

```bash
gofmt -l internal/write && go vet ./internal/write
git add internal/write/tick.go internal/write/tick_test.go internal/write/mark.go internal/write/mark_test.go
git commit -m "write: MarkItem sets a task or debt line done or open and commits it"
```

---

### Task 2: `acta tick --undo`

**Files:**
- Modify: `internal/cli/tick.go` (`tickUsage`, flags, action switch, `markTaskDates`, `markDate`)
- Modify: `internal/cli/tick_test.go`

**verify:** `--undo` never mixes with another action flag, puts every box of a task or the one box of a debt line back to open from any state, writes no dates and no agent record, and leaves every other line of the file as it was. After the move, every tick path that dated a plan or spec before still dates it the same way. List each flag mix and each item kind checked.

**Interfaces:**
- Consumes: `write.Untick(path string, headingLine int) (int, int, error)`, `write.TickLine(path string, line int, state byte) error`, `write.TaskDates(cfg config.Config, it *board.Item) ([]string, error)` from Task 1.

- [x] **Step 1: Write the failing tests**

In `internal/cli/tick_test.go` (uses `tickRepo`, `runTick`, `debtFile`, `read`):

```go
func TestTickUndoPutsATaskBackToOpen(t *testing.T) {
	dir := tickRepo(t)
	if code, _, errOut := runTick(t, dir, "--all", "plans/2026-09-27-n#task-1"); code != 0 {
		t.Fatalf("tick --all: %d %s", code, errOut)
	}
	code, out, errOut := runTick(t, dir, "--undo", "plans/2026-09-27-n#task-1")
	if code != 0 || !strings.Contains(out, "0/1") {
		t.Fatalf("undo: %d %q %q", code, out, errOut)
	}
	if got := read(t, filepath.Join(dir, ".acta", "plans", "2026-09-27-n.md")); !strings.Contains(got, "- [ ] a") {
		t.Fatalf("task not open: %q", got)
	}
}

func TestTickUndoPutsADebtLineBackToOpen(t *testing.T) {
	dir := tickRepo(t)
	if code, _, errOut := runTick(t, dir, "--wontfix", "DEBT-1.1"); code != 0 {
		t.Fatalf("wontfix: %d %s", code, errOut)
	}
	if code, _, errOut := runTick(t, dir, "--undo", "DEBT-1.1"); code != 0 {
		t.Fatalf("undo: %d %s", code, errOut)
	}
	if got := read(t, debtFile(dir)); !strings.Contains(got, "- [ ] a\n- [ ] b") {
		t.Fatalf("debt line not open: %q", got)
	}
}

func TestTickUndoMixedWithAnotherActionIsRefused(t *testing.T) {
	dir := tickRepo(t)
	for _, other := range []string{"--all", "--start", "--wontfix", "--step=1"} {
		if code, _, _ := runTick(t, dir, "--undo", other, "plans/2026-09-27-n#task-1"); code == 0 {
			t.Errorf("--undo %s: exit 0", other)
		}
	}
}
```

Check the existing tests in this file for the real id forms of the task and the first debt line and use those if they differ from `plans/2026-09-27-n#task-1` and `DEBT-1.1`.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/cli -run 'TestTickUndo' -v`
Expected: FAIL, `flag provided but not defined: -undo`.

- [x] **Step 3: Write the code**

In `internal/cli/tick.go`:

1. `const tickUsage = "usage: acta tick <id> [--step N | --all | --start | --wontfix | --undo]"`.
2. Add `undo := fs.Bool("undo", false, "put the task or debt line back to open")` and add `*undo` to the one-action list: `[]bool{*step > 0, *all, *start, *wontfix, *undo}`.
3. In the action switch, before the `case it.Kind == board.KindDebtItem:` branch that ticks, add:

```go
	case *undo:
		// Undo is for a slip of the finger. It clears the boxes and nothing
		// else: dates already written stay, and nobody is recorded.
		if it.Kind == board.KindDebtItem {
			err = write.TickLine(it.Path, it.Line, ' ')
			if err == nil {
				fmt.Fprintf(stdout, "%s open\n", it.ID)
			}
		} else {
			var done, total int
			done, total, err = write.Untick(it.Path, it.Line)
			if err == nil {
				fmt.Fprintf(stdout, "%s %d/%d\n", it.ID, done, total)
			}
		}
		if err != nil {
			fmt.Fprintln(stderr, err)
			if errors.Is(err, write.ErrBadInput) {
				return exitBadInput
			}
			return exitOther
		}
		return exitOK
```

   Declare `var err error` above the switch if none is in scope there, or reuse the one from `parseMixed`.
4. Delete `markTaskDates` and `markDate` from this file, and change the call to `if _, err := write.TaskDates(cfg, it); err != nil {`. Drop any import that is now unused (`bytes`, `config`).

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/cli -run 'TestTick' -v`
Expected: PASS, old tick tests included.

- [x] **Step 5: Commit**

```bash
gofmt -l internal/cli && go vet ./internal/cli
git add internal/cli/tick.go internal/cli/tick_test.go
git commit -m "cli: acta tick --undo puts a task or debt line back to open"
```

---

### Task 3: `+` and `-` keys in the TUI

**Files:**
- Modify: `internal/tui/model.go` (`Model` fields, `New`, `key` switch, new `markRow`)
- Modify: `internal/tui/model_test.go`

**verify:** `+` and `-` call `markItem` only on a task row or a debt line row that is not from a worktree and not legacy, with `true` for `+` and `false` for `-`; every other row, and no row, writes nothing and the status line says why. After a call the status line shows the outcome and the board reloads. The keys do nothing while the help, a popup, the slug prompt or search is open. List every row kind and every open overlay checked.

**Interfaces:**
- Consumes: `write.MarkItem(cfg config.Config, b *board.Board, id string, done bool) (write.Outcome, error)` from Task 1.
- Produces: `Model.markItem func(id string, done bool) (write.Outcome, error)`.

- [x] **Step 1: Write the failing tests**

In `internal/tui/model_test.go`:

```go
func TestPlusAndMinusMarkATaskRow(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	var got []string
	m.markItem = func(id string, done bool) (write.Outcome, error) {
		got = append(got, fmt.Sprintf("%s %v", id, done))
		return write.Outcome{Committed: true}, nil
	}
	// Open a plan to reach a task row, the same way TestPopupRefusals does.
	m = press(m, tabKey(tabPlans), " ", "j")
	it := m.Selected()
	if it == nil || it.Kind != board.KindTask {
		t.Fatalf("not on a task row: %+v", it)
	}
	m = press(m, "+", "-")
	if want := it.ID + " true," + it.ID + " false"; strings.Join(got, ",") != want {
		t.Fatalf("markItem got %v, want %s", got, want)
	}
	if !strings.Contains(m.status, "committed") {
		t.Fatalf("status %q", m.status)
	}
}

func TestPlusMarksADebtLineRow(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	var got string
	m.markItem = func(id string, done bool) (write.Outcome, error) {
		got = id
		return write.Outcome{Committed: true}, nil
	}
	m = press(m, tabKey(tabDebts))
	it := m.Selected()
	if it == nil || it.Kind != board.KindDebtItem {
		t.Fatalf("not on a debt line row: %+v", it)
	}
	m = press(m, "+")
	if got != it.ID {
		t.Fatalf("markItem got %q, want %q", got, it.ID)
	}
}

func TestPlusRefusesRowsThatAreNotTasksOrDebtLines(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	called := false
	m.markItem = func(string, bool) (write.Outcome, error) {
		called = true
		return write.Outcome{}, nil
	}
	m = press(m, tabKey(tabBugs), "+")
	if called || !strings.Contains(m.status, "task or a debt line") {
		t.Fatalf("bug row: called %v status %q", called, m.status)
	}
	m = press(m, "?", "+")
	if called {
		t.Fatal("+ under the help called markItem")
	}
}
```

Add `fmt` to the imports if missing. If the Debts tab opens with no row selected, press `j` first, and say so in the report.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui -run 'TestPlus' -v`
Expected: FAIL to build, `m.markItem undefined`.

- [x] **Step 3: Write the code**

In `internal/tui/model.go`:

1. Next to `setValue` in `Model`: `markItem func(id string, done bool) (write.Outcome, error)`.
2. In `New`, next to `setValue`:

```go
		markItem: func(id string, done bool) (write.Outcome, error) {
			fresh, err := board.Load(cfg)
			if err != nil {
				return write.Outcome{}, err
			}
			return write.MarkItem(cfg, fresh, id, done)
		},
```

3. In the key switch, next to `case "t", "s":`:

```go
	case "+":
		return m.markRow(true)
	case "-":
		return m.markRow(false)
```

4. After `openPopup`:

```go
// markRow marks the task or debt line under the cursor done or back to open.
// The same rows the status popup refuses are refused here, so a key press
// never writes a file the popup would not.
func (m Model) markRow(done bool) (tea.Model, tea.Cmd) {
	it := m.Selected()
	switch {
	case it == nil:
		m.status = "nothing selected"
		return m, nil
	case it.Kind != board.KindTask && it.Kind != board.KindDebtItem:
		m.status = "+ and - work on a task or a debt line; use s for the status"
		return m, nil
	case it.Worktree != "":
		m.status = "shown from worktree " + it.Worktree + "; edit it there"
		return m, nil
	case it.Legacy:
		m.status = "legacy file, move it into the root folder first"
		return m, nil
	}
	word := "open"
	if done {
		word = "done"
	}
	o, err := m.markItem(it.ID, done)
	m.status = outcomeText(fmt.Sprintf("acta: %s %s", it.ID, word), o, err)
	return m, m.reloadCmd()
}
```

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui -run 'TestPlus|TestPopup|TestStatusPopup' -v`
Expected: PASS.

- [x] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/model.go internal/tui/model_test.go
git commit -m "tui: + marks a task or debt line done, - puts it back to open"
```

---

### Task 4: help names one key per line

**Files:**
- Modify: `internal/tui/view.go` (`helpLines`)
- Modify: `internal/tui/view_test.go`

**verify:** Every help line names one key, or one pair that are opposites, and says in plain words what it does; every key the key switch in `model.go` handles is in the help; no old packed line (`t s n`, `/ r q`, `space enter`) can come back in any form. The help box fits a 26-line screen whole. List every key checked against the switch and every screen height checked.

**Interfaces:**
- Consumes: nothing.
- Produces: `helpLines` text, which the existing `helpText` lays out.

- [x] **Step 1: Write the failing test**

In `internal/tui/view_test.go`:

```go
// TestHelpNamesOneKeyPerLine keeps the help plain: each line is one key, or
// two keys that are opposites, so nobody has to guess which key does what.
func TestHelpNamesOneKeyPerLine(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"s": "set the status; dropped and wontfix close an item",
		"t": "set the type: spec or bug",
		"n": "new bug",
		"+": "mark the task or debt line done",
		"-": "put the task or debt line back to open",
		"/": "search the rows",
		"r": "reload the board from disk",
		"q": "quit",
	}
	got := map[string]string{}
	for _, ln := range strings.Split(helpLines, "\n") {
		k, w, ok := strings.Cut(ln, "  ")
		if !ok {
			t.Fatalf("line has no key column: %q", ln)
		}
		got[strings.TrimSpace(k)] = strings.TrimSpace(w)
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("key %q: got %q, want %q", k, got[k], w)
		}
	}
	for _, packed := range []string{"t s n", "/ r q", "space enter"} {
		if _, ok := got[packed]; ok {
			t.Errorf("packed line %q is back", packed)
		}
	}
	if n := len(strings.Split(helpLines, "\n")); n > 24 {
		t.Errorf("help has %d lines; more than 24 does not fit a 26-line screen", n)
	}
}
```

- [x] **Step 2: Run the test to see it fail**

Run: `go test ./internal/tui -run 'TestHelpNamesOneKeyPerLine' -v`
Expected: FAIL, `key "s": got "", want ...`.

- [x] **Step 3: Write the code**

Replace `helpLines` in `internal/tui/view.go` with exactly this (24 lines; the gap between key and text is at least two spaces, which `helpText` splits on):

```go
const helpLines = `1-6              open a tab
← →              previous / next tab
tab shift+tab    next / previous pane of the tab
[ ]              previous / next Done tab, on the Done pane
j k              move down / up a list, scroll the detail
g G              go to the top / bottom
ctrl+d ctrl+u    page down / up
space            open or shut a plan row
h l              shut / open a plan row, h on a task too
enter            open or shut a plan row, or focus the detail
z                expand the focused pane
o                flip the sort: oldest / newest
/                search the rows
r                reload the board from disk
e                open the row in the editor
y                copy the id of the row
s                set the status; dropped and wontfix close an item
t                set the type: spec or bug
n                new bug
+                mark the task or debt line done
-                put the task or debt line back to open
esc              back to the list, close this help
q                quit
?                close this help`
```

Before writing, read the key switch in `internal/tui/model.go` (`case "q", "ctrl+c":` and below). If it handles a key this list does not name, add a line for it and drop nothing else; say so in the report.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui -run 'TestHelp' -v`
Expected: PASS, including `TestHelpListsTheSortKey`, `TestHelpKeysSitInARightAlignedColumn` and the `h l` and no-`0` checks in `model_test.go`.

- [x] **Step 5: Commit**

```bash
gofmt -l internal/tui && go vet ./internal/tui
git add internal/tui/view.go internal/tui/view_test.go
git commit -m "tui: help names one key per line, with + and -"
```
