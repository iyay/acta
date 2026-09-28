---
id: PLAN-17
hash: qgef
---
# Scratchpad Kind and Five-Pane Sidebar Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A Scratchpad kind (`.acta/scratch/`, `acta scratch new`, `acta scratch add`, statuses raw/brainstorming/specced/dropped with specced derived from a spec's `parent`) and a TUI sidebar of five stacked panes (Active, Specs ─ Scratchpad, Plans ─ Tasks, Bugs ─ Debt, Done) with a `z` expand toggle and a `<selected> of <total>` counter.

**Architecture:** The board learns one new kind in `internal/board` (new file `scratch.go` holds the scratch status rules and the spec-to-scratch link, so `board.go` does not grow). Write ops mirror `NewBug` and `appendDebt` in a new `internal/write/scratch.go`. The TUI swaps its fixed three-value `pane` enum and its `[5]` tab arrays for one table of sidebar panes in a new `internal/tui/sidebar.go`; expand and the counter live in `frame.go` and `scroll.go`.

**Tech Stack:** Go 1.27, Bubble Tea, lipgloss.

**Spec:** `.acta/specs/2026-09-28-scratch-sidebar-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Scratch statuses, exact: `raw`, `brainstorming`, `specced`, `dropped`. Only `raw`, `brainstorming` and `dropped` are ever written to a file. `specced` is derived, never written.
- Error strings, exact: `scratch body is empty`, `specced comes from a spec's parent link`, `parent <value> not found` (same form debt uses).
- Commit messages from the CLI, exact: `acta: new scratch <stem>`, `acta: add to scratch <stem>`.
- Sidebar titles, exact: `[1] Active`, `[2] Specs ─ Scratchpad`, `[3] Plans ─ Tasks`, `[4] Bugs ─ Debt`, `[5] Done`. Detail pane key is `0`.
- Counter text, exact: `<selected> of <total>`, 1-based, bottom right of the border; `0 of 0` for an empty pane; no counter on the detail pane.
- In progress means `inProgress(it)` in `internal/tui/model.go`, extended so a scratch item with status `brainstorming` counts. Reuse it; do not write a second rule.
- No file in `internal/tui` or `internal/board` grows past 800 lines. `model.go` is 790 lines today, so new TUI code goes into `sidebar.go`.
- Comments: plain English a 10-year-old reads, say why. No marker tags, no Latin, no emoji.
- Gates before every commit: `gofmt -l .` prints nothing, `go vet ./...` clean, `go test ./...` passes.
- TDD: write the failing test first, watch it fail, then the minimum code.
- Tests run write commands only inside `t.TempDir()` repos (`repoWith` in `internal/write/ops_test.go`). Never run `acta` write commands in the worktree itself, except the `acta scratch new` lines Task 5 names and the `acta tick` lines for this plan.
- Never `git reset`, `rebase`, `amend` or `push`. Never delete test fixtures to make a test pass. Never file acta bugs for defects in this branch; fix them in the task.

## File map

| File | Tasks |
|---|---|
| `internal/board/board.go` (Kind const, status list, `Allowed`, `Closed`, `collect`, `collectFiles`, `derive` hook, call to `linkScratch`), `internal/board/ids.go` (`Prefix`), `internal/board/scratch.go` (new), `internal/board/scratch_test.go` (new), `internal/board/testdata/basic/` (new scratch fixture + spec with parent), `internal/config/config.go` (`Dirs.Scratch`), `internal/config/config_test.go`, `internal/write/ids.go` (`scanIDs` seed) , `internal/write/ids_test.go` | 1 |
| `internal/write/scratch.go` (new), `internal/write/scratch_test.go` (new), `internal/write/ops.go` (`SetValue` specced guard), `internal/cli/cli.go` (`scratch` case, usage text), `internal/cli/cli_test.go` | 2 |
| `internal/tui/sidebar.go` (new), `internal/tui/sidebar_test.go` (new), `internal/tui/model.go` (pane type, tab arrays, Model fields, `openRows`, `doneRows`, `onTab`, `doneTabNames`, `focusPane`, `cycleTab`, key switch, `inProgress`), `internal/tui/view.go` (`geom`, `tabsOf`, `helpLines`, `View`), `internal/tui/frame.go` (`geometry`), `internal/tui/scroll.go` (`boxOf`), existing tui tests | 3 |
| `internal/tui/frame.go` (heights, expand), `internal/tui/scroll.go` (counter, remove `count`), `internal/tui/view.go` (`paneView` footer), `internal/tui/sidebar.go` (`expanded` state), `internal/tui/model.go` (`z` key), `internal/tui/scroll_test.go`, `internal/tui/frame_test.go` | 4 |
| `.acta/scratch/*.md` (created by `acta scratch new`) | 5 |
| `internal/tui/detail.go` (body lines), `internal/tui/detail_test.go` | 6 |

## Waves

- Wave 1: Task 1, Task 6 (no shared file)
- Wave 2: Task 2, Task 3 (write/cli and tui share no file; both only need Task 1's kind)
- Wave 3: Task 4 (touches tui files Task 3 changed)
- Wave 4: Task 5 (needs the `acta scratch new` from Task 2)

---

### Task 1: Scratch kind on the board

**Files:**
- Modify: `internal/board/board.go` (Kind consts near line 17, status lists near line 77, `Allowed` near 85, `Closed` near 99, `collect` parts list near 224, `collectFiles` map near 265, `derive` near 505, the link step where `linkDebt` is called near 193)
- Modify: `internal/board/ids.go` (`Prefix`)
- Create: `internal/board/scratch.go`, `internal/board/scratch_test.go`
- Create fixtures: `internal/board/testdata/basic/.acta/scratch/2026-09-28-idea-raw.md`, `.../2026-09-28-idea-used.md`, `.../2026-09-28-idea-dropped.md`, and one spec `.../.acta/specs/2026-09-28-from-scratch-design.md` with `parent: scratch/2026-09-28-idea-used`. Check the real fixture root path under `internal/board/testdata/basic` first and match it. Existing tests that count items in this fixture must be updated to the new counts, never by deleting fixtures.
- Modify: `internal/config/config.go` (`Dirs.Scratch`, default `scratch`, override block in `Load` next to `Debt`), `internal/config/config_test.go`
- Modify: `internal/write/ids.go` (`scanIDs` seed map gets `"SCRATCH": 1`), `internal/write/ids_test.go`

**verify:** A scratch item's shown status is always one of the four scratch words and follows one rule: a written `dropped` wins; else any spec whose `parent` names the item makes it `specced` (source `derived`); else the written status; else `raw`. List every combination checked (written raw/brainstorming/dropped/none × linked/not linked) with the status each gave. No other kind's status or ID prefix changes: list the kinds checked. A spec whose `parent` names nothing gets exactly one problem and no crash.

**Interfaces:**
- Produces: `board.KindScratch Kind = "scratch"`; `board.Allowed(board.KindScratch)` returns `[]string{"raw", "brainstorming", "dropped"}` (the settable ones; `specced` is left out on purpose so no UI or CLI offers it); `board.Closed("specced") == true`; `board.Prefix(board.KindScratch, false) == "SCRATCH"`; `config.Dirs.Scratch string` (yaml `scratch`); a scratch `*Item` has `Children` holding the linked spec IDs.

- [x] **Step 1: Write the failing tests**

```go
// internal/board/scratch_test.go
package board

import (
	"slices"
	"testing"
)

func TestScratchStatusFollowsOneRule(t *testing.T) {
	b := loadBasic(t) // use the loader the other board tests use on testdata/basic
	cases := []struct {
		id, status, source string
	}{
		{"scratch/2026-09-28-idea-raw", "raw", ""},
		{"scratch/2026-09-28-idea-used", "specced", "derived"},
		{"scratch/2026-09-28-idea-dropped", "dropped", ""},
	}
	for _, c := range cases {
		it := b.Get(c.id) // use the board's real lookup method
		if it == nil {
			t.Fatalf("%s not loaded", c.id)
		}
		if it.Kind != KindScratch || it.Status != c.status {
			t.Errorf("%s: kind %s status %s, want scratch %s", c.id, it.Kind, it.Status, c.status)
		}
		if c.source != "" && it.StatusSource != c.source {
			t.Errorf("%s: source %s, want %s", c.id, it.StatusSource, c.source)
		}
	}
	used := b.Get("scratch/2026-09-28-idea-used")
	if !slices.Contains(used.Children, "specs/2026-09-28-from-scratch-design") {
		t.Errorf("used item children = %v, want the spec", used.Children)
	}
}

func TestWrittenDroppedBeatsSpecLink(t *testing.T) {
	// Build a board where a spec links to an item written as dropped.
	// The item must stay dropped.
}

func TestSpecParentNotFoundIsOneProblem(t *testing.T) {
	// A spec with parent: scratch/missing gets exactly
	// []string{"parent scratch/missing not found"} and the board still loads.
}

func TestScratchStatusesAndPrefix(t *testing.T) {
	if got := Allowed(KindScratch); !slices.Equal(got, []string{"raw", "brainstorming", "dropped"}) {
		t.Errorf("Allowed = %v", got)
	}
	for _, s := range []string{"specced", "dropped"} {
		if !Closed(s) {
			t.Errorf("Closed(%q) = false", s)
		}
	}
	for _, s := range []string{"raw", "brainstorming"} {
		if Closed(s) {
			t.Errorf("Closed(%q) = true", s)
		}
	}
	if Prefix(KindScratch, false) != "SCRATCH" {
		t.Error("prefix")
	}
	if Prefix(KindStory, false) != "SPEC" || Prefix(KindBug, false) != "BUG" || Prefix(KindDebt, false) != "DEBT" {
		t.Error("other prefixes changed")
	}
}
```

Write the two stub tests in full: each builds its own temp root with files (copy the pattern the board tests use for one-off roots) and asserts the stated result. Add in `config_test.go` a test that `.acta.yaml` with `dirs: {scratch: ideas}` gives `cfg.Dirs.Scratch == "ideas"` and the default is `"scratch"`. Add in `write/ids_test.go` a test that `acta id` over a root with two new scratch files gives them `SCRATCH-1` and `SCRATCH-2`, and leaves existing SPEC/BUG numbers alone.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/board/ ./internal/config/ ./internal/write/ -run 'Scratch|Dropped|ParentNotFound' -v`
Expected: FAIL, `undefined: KindScratch` and missing `Dirs.Scratch`.

- [x] **Step 3: Write the minimal code**

```go
// board.go
KindScratch Kind = "scratch"

scratchStatuses = []string{"raw", "brainstorming", "dropped"}

// in Allowed
case KindScratch:
	return append([]string(nil), scratchStatuses...)

// in Closed
case "done", "fixed", "dropped", "wontfix", "specced":

// collect parts list and collectFiles map gain
{s.dirs.Scratch, KindScratch}
cfg.Dirs.Scratch: KindScratch
```

```go
// ids.go Prefix
case k == KindScratch:
	return "SCRATCH"
```

```go
// scratch.go
package board

import "strings"

// linkScratch ties each spec that names a scratch item as its parent to that
// item. A spec is where a raw idea ends up, so the item learns it was used.
func (b *Board) linkScratch() {
	for _, it := range b.Items {
		if it.Kind != KindStory {
			continue
		}
		want := field(it.front, "parent") // use the real field that holds a file's frontmatter
		if !strings.HasPrefix(want, "scratch/") && !strings.HasPrefix(want, b.scratchDir+"/") {
			continue
		}
		target := b.byID[want]
		if target == nil || target.Kind != KindScratch {
			it.Problems = append(it.Problems, "parent "+want+" not found")
			continue
		}
		target.Children = append(target.Children, it.ID)
	}
}

// scratchStatus picks the one status a scratch item shows. Dropped is a
// choice the user wrote down, so it wins over a spec link.
func scratchStatus(written string, linked bool) (status, source string) {
	switch {
	case written == "dropped":
		return "dropped", ""
	case linked:
		return "specced", "derived"
	case written != "":
		return written, ""
	}
	return "raw", ""
}
```

In `derive`, handle `KindScratch` before the generic `parentStatus` path (the way the debt block is special-cased), calling `scratchStatus(it.fmStatus, len(it.Children) > 0)`. Call `b.linkScratch()` right after the debt link step and before `derive`. Match the real names of the frontmatter holder and the configured scratch dir; the snippet shows intent. A written status outside the scratch list gets the same problem other kinds get for a bad status.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/board/ ./internal/config/ ./internal/write/ -v`
Expected: PASS, including every existing test.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/board internal/config internal/write/ids.go internal/write/ids_test.go
git commit -m "feat(board): scratch kind with specced derived from a spec's parent"
```

---

### Task 2: `acta scratch new` and `acta scratch add`

**Files:**
- Create: `internal/write/scratch.go`, `internal/write/scratch_test.go`
- Modify: `internal/write/ops.go` (`SetValue` status case)
- Modify: `internal/cli/cli.go` (command switch near line 88, unknown-command text near line 111), `internal/cli/cli_test.go`

**verify:** No failing call ever leaves a file written, changed or committed. List every failure path checked (empty stdin for new, whitespace-only stdin, bad slug, existing file, unknown ID for add, a non-scratch ID for add, empty stdin for add, `acta set ... status specced`) with the error each printed and the proof the tree was unchanged (`git status --porcelain` empty, log unchanged). Every success path commits exactly one commit with the exact message. Unicode bodies come back byte for byte.

**Interfaces:**
- Consumes: `board.KindScratch`, `config.Dirs.Scratch`, `scanIDs` seed `SCRATCH` (Task 1); `finish`, `dirtyBefore`, `Outcome`, `bad` in `ops.go`.
- Produces: `write.NewScratch(cfg config.Config, slug, title string, body []byte) (Outcome, error)`; `write.AppendScratch(cfg config.Config, b *board.Board, id string, text []byte) (Outcome, error)`; CLI `acta scratch new <slug> [--title T] < body.md` and `acta scratch add <SCRATCH-n> < text.md`, both printing the item's short ID and path on success, exit code `exitBadInput` on bad input.

- [x] **Step 1: Write the failing tests**

```go
// internal/write/scratch_test.go
package write

import (
	"strings"
	"testing"
)

func TestNewScratchWritesRawItemAndCommits(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	body := []byte("catet aja dulu: sort newest first ✓\n")
	out, err := NewScratch(cfg, "newest-first", "", body)
	if err != nil {
		t.Fatal(err)
	}
	got := readFile(t, out.Path) // use the helper the other write tests use
	for _, want := range []string{"id: SCRATCH-1", "status: raw", "title: newest-first", string(body)} {
		if !strings.Contains(got, want) {
			t.Errorf("file lacks %q:\n%s", want, got)
		}
	}
	if msg := gitRun(t, cfg.Root, "log", "-1", "--format=%s"); strings.TrimSpace(msg) != "acta: new scratch "+stem(out.Path) {
		t.Errorf("commit message %q", msg)
	}
}

func TestNewScratchRefusesAndWritesNothing(t *testing.T) {
	cases := map[string]struct {
		slug string
		body string
	}{
		"empty body":      {"idea", ""},
		"blank body":      {"idea", " \n\t\n"},
		"bad slug":        {"Bad Slug!", "x"},
		"empty slug":      {"", "x"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := repoWith(t, baseFiles)
			before := gitRun(t, cfg.Root, "log", "--format=%H")
			if _, err := NewScratch(cfg, c.slug, "", []byte(c.body)); err == nil {
				t.Fatal("want an error")
			}
			if s := gitRun(t, cfg.Root, "status", "--porcelain"); s != "" {
				t.Errorf("tree changed: %s", s)
			}
			if after := gitRun(t, cfg.Root, "log", "--format=%H"); after != before {
				t.Error("a commit was made")
			}
		})
	}
	// Also: the same slug twice on one day fails the second time and changes nothing.
	// Also: empty body error text is exactly "scratch body is empty".
}

func TestAppendScratchAddsTextAfterOneBlankLine(t *testing.T) {
	// new, then AppendScratch(cfg, board, "SCRATCH-1", "answer 1\n"):
	// body ends with "\n\nanswer 1\n", one commit "acta: add to scratch <stem>".
	// Works the same when the item is dropped.
}

func TestAppendScratchRefusesAndChangesNothing(t *testing.T) {
	// unknown id SCRATCH-99, a BUG id, empty text: each errors, file bytes and log unchanged.
}

func TestSetValueRefusesSpecced(t *testing.T) {
	// SetValue(cfg, b, "SCRATCH-1", "status", "specced") fails with exactly
	// "specced comes from a spec's parent link"; "brainstorming" and "dropped" succeed.
}
```

Write every stub test in full. In `internal/cli/cli_test.go`, add one test per command that runs the CLI entry point with stdin, in a temp repo, and checks the exit code, the printed ID and one bad-input case (`acta scratch` with no subcommand prints a usage line naming `scratch new` and `scratch add`).

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/write/ ./internal/cli/ -run 'Scratch|Specced' -v`
Expected: FAIL, `undefined: NewScratch`.

- [x] **Step 3: Write the minimal code**

`NewScratch` follows `NewBug`: build `cfg.Root/<acta root>/<Dirs.Scratch>/<date>-<slug>.md` with the same slug check and exists check as `bugPath` (share the check, do not copy it), refuse `strings.TrimSpace(body) == ""` with `bad("scratch body is empty")` before touching disk, write frontmatter `id`, `hash`, `title` (slug when empty), `status: raw`, `created: <date>`, then the body exactly, then `finish(cfg, path, "acta: new scratch "+stem, dirty)`.

`AppendScratch` resolves the ID on the board, refuses when missing or `Kind != board.KindScratch` or the text is blank, appends `"\n" + text` so there is one blank line before it (add a trailing newline to the old body first if it lacks one), and commits `acta: add to scratch <stem>`.

In `SetValue`, before the generic status check:

```go
if field == "status" && it.Kind == board.KindScratch && value == "specced" {
	return Outcome{}, bad("specced comes from a spec's parent link")
}
```

In `cli.go`, add:

```go
case "scratch":
	if len(args) < 2 || (args[1] != "new" && args[1] != "add") {
		fmt.Fprintln(stderr, "usage: acta scratch new <slug> [--title T] < body.md | acta scratch add <SCRATCH-n> < text.md")
		return exitBadInput
	}
	if args[1] == "new" {
		return cmdScratchNew(args[2:], stdin, stdout, stderr)
	}
	return cmdScratchAdd(args[2:], stdin, stdout, stderr)
```

and add `scratch new` and `scratch add` to the unknown-command message. Both commands read stdin only; no editor path.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/write/ ./internal/cli/ -v`
Expected: PASS.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/write internal/cli
git commit -m "feat(cli): acta scratch new and acta scratch add"
```

---

### Task 3: Five-pane sidebar

**Files:**
- Create: `internal/tui/sidebar.go`, `internal/tui/sidebar_test.go`
- Modify: `internal/tui/model.go` (`pane` type and consts near 19, `tabKinds`/`tabNames` near 37, `doneTabs` near 45, Model fields `sel`/`idx`/`doneSel`/`doneIdx`/`off`/`tab`/`doneTab`/`last` near 107, `openRows` near 231, `doneRows` near 262, `onTab`, `doneTabNames`, `focusPane` near 412, `cycleTab` near 425, key switch near 484, `inProgress` near 213)
- Modify: `internal/tui/view.go` (`geom` near 33, `tabsOf` near 114, `helpLines` near 65, `View`)
- Modify: `internal/tui/frame.go` (`geometry`), `internal/tui/scroll.go` (`boxOf`)
- Update existing tests in `internal/tui/*_test.go` that press `3` for detail or count three panes; change them to the new keys, never delete their checks.

**verify:** Every item on the board shows in exactly the panes the spec gives it: open items in their kind pane's tab, closed items only in Done, in-progress items also in Active, and nothing that is only open/raw/todo in Active. List each kind checked with the pane and tab it appeared in. Every key `0`-`5`, `tab`, `shift+tab` lands on the box the spec names from every starting box: list the 6×8 moves checked. No hard-coded pane count (`%3`, `[3]`, `[5]`) is left in `internal/tui`: list the grep run and its output.

**Interfaces:**
- Consumes: `board.KindScratch`, `board.Closed("specced")` (Task 1).
- Produces: in `sidebar.go`:

```go
// sidebarPane is one stacked box on the left. Its tabs are the kinds it shows.
type sidebarPane struct {
	title string       // "Active", "Specs", ...
	tabs  []sidebarTab // empty for Active and Done
}

type sidebarTab struct {
	name string
	kind board.Kind
	done []doneTab // the Done pane's tabs when this tab had focus last
}

// sidebar is the whole left column, top to bottom. Index i is key i+1.
var sidebar = []sidebarPane{
	{title: "Active"},
	{title: "Specs ─ Scratchpad", tabs: []sidebarTab{
		{"Specs", board.KindStory, []doneTab{{"Done", "done"}, {"Dropped", "dropped"}}},
		{"Scratchpad", board.KindScratch, []doneTab{{"Specced", "specced"}, {"Dropped", "dropped"}}},
	}},
	{title: "Plans ─ Tasks", tabs: []sidebarTab{
		{"Plans", board.KindPlan, []doneTab{{"Done", "done"}, {"Dropped", "dropped"}}},
		{"Tasks", board.KindTask, []doneTab{{"Done", "done"}}},
	}},
	{title: "Bugs ─ Debt", tabs: []sidebarTab{
		{"Bugs", board.KindBug, []doneTab{{"Fixed", "fixed"}, {"Wontfix", "wontfix"}}},
		{"Debt", board.KindDebtItem, []doneTab{{"Done", "done"}, {"Wontfix", "wontfix"}}},
	}},
	{title: "Done"},
}

const (
	paneActive = 0
	paneDone   = len(sidebar) - 1 // computed from the table, so it follows the table
)
```

  `pane` stays an `int` type; `paneDetail` becomes `pane(len(sidebar))`. Per-pane state (selected ID, index, tab, scroll offset) becomes slices sized from `sidebar`, with the detail offset last. Keys: `"1"`..`"5"` focus `pane(k-'1')`, `"0"` focuses `paneDetail`, `tab`/`shift+tab` cycle `len(sidebar)+1` boxes. The Done pane follows `m.last` (the last sidebar pane with focus, never Done or detail) and that pane's current tab; when `m.last == paneActive`, Done lists every closed item with no tabs, in today's closed order. Active rows are every non-legacy item where `inProgress(it)` holds, in today's in-progress order, no tabs. `inProgress` also returns true for `it.Kind == board.KindScratch && it.Status == "brainstorming"`. The in-progress divider row (`dividerRowID`) and its code are removed. Pane titles are `[n] <title>` with the tab names drawn the way `tabsOf` draws them today. `helpLines` lists `1-5 panes · 0 detail`.

- [x] **Step 1: Write the failing tests**

```go
// internal/tui/sidebar_test.go
package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(m Model, k string) Model {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	return next.(Model)
}

func TestNumberKeysFocusTheirBox(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	for k, want := range map[string]pane{"1": 0, "2": 1, "3": 2, "4": 3, "5": 4, "0": paneDetail} {
		for _, start := range []string{"1", "2", "3", "4", "5", "0"} {
			if got := key(key(m, start), k).focus; got != want {
				t.Errorf("from %s press %s: focus %d, want %d", start, k, got, want)
			}
		}
	}
}

func TestTabCyclesSixBoxes(t *testing.T) {
	// From pane 0, tab six times visits 1,2,3,4,5(detail),0 in order;
	// shift+tab walks the same ring backwards.
}

func TestSidebarTitles(t *testing.T) {
	v := sized(newModel(t), 160, 50).View()
	for _, want := range []string{"[1]─Active", "[2]─Specs", "Scratchpad", "[3]─Plans", "Tasks", "[4]─Bugs", "Debt", "[5]─Done"} {
		if !strings.Contains(v, want) {
			t.Errorf("view lacks %q", want)
		}
	}
	// Match the real title join the frame uses (today "[n]─Title"); adjust the
	// strings to it, but every title and tab name above must appear.
}

func TestActiveHoldsOnlyWorkInProgress(t *testing.T) {
	// Using the basic fixture (Task 1 added scratch files; add one
	// brainstorming scratch item and make sure a fixing bug and an
	// in-progress plan exist): Active rows contain each in-progress item of
	// every kind and none whose status is open, raw, todo or closed.
}

func TestDoneFollowsLastSidebarPane(t *testing.T) {
	// Focus 2, switch to the Scratchpad tab with "]", focus 5:
	// Done tabs read "Specced" and "Dropped" and rows are the specced and
	// dropped scratch items. Focus 4 (Bugs), then 5: tabs "Fixed", "Wontfix".
	// Focus 1 (Active), then 5: no tabs, rows are every closed item.
}

func TestNoDividerRow(t *testing.T) {
	// No row in any sidebar pane is a divider, even when a pane holds both
	// in-progress and not-started items.
}
```

Write every stub test in full against the real fixture IDs.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'Number|TabCycles|Sidebar|Active|DoneFollows|NoDivider' -v`
Expected: FAIL (focus for key `4` does not move; `[1]─Active` missing).

- [x] **Step 3: Write the code**

Move the tab tables into `sidebar.go` as shown in Interfaces. Replace every `[5]`/`[3]` array and every `%3` with values sized from `sidebar`. Make `geometry` stack `len(sidebar)` boxes of equal height in the left column (the last box takes the leftover lines so the column always fills the height), keep detail on the right, and keep the narrow-terminal path showing only the focused box. `boxOf` indexes the new box slice. Remove the divider row code. Update `helpLines`. A scratch item's detail shows its linked specs through the same children list the detail pane already draws; add a test that the used fixture item's detail names `SPEC-` of the linked spec. Then run `grep -nE '%3|\[3\]|\[5\]' internal/tui/*.go` and clear every hit that counts panes or tabs.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -v`
Expected: PASS, including every existing tui test after its key updates.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
wc -l internal/tui/*.go
git add internal/tui
git commit -m "feat(tui): five-pane sidebar with Active and Scratchpad"
```

---

### Task 4: Expand with `z` and the item counter

**Files:**
- Modify: `internal/tui/frame.go` (left column heights), `internal/tui/sidebar.go` (`expanded` state helpers), `internal/tui/model.go` (`z` key; clear expand in `focusPane`), `internal/tui/scroll.go` (remove `count`, add `itemCount`), `internal/tui/view.go` (`paneView` footer: counter only on sidebar panes, right aligned), `internal/tui/frame.go` (`paneBottom` right-aligned foot)
- Test: `internal/tui/frame_test.go`, `internal/tui/scroll_test.go`, `internal/tui/view_test.go`

**verify:** For every terminal height from 3 to 60 and widths 40, 80, 160, with each sidebar pane expanded and none expanded, the left column fills the height exactly, every other sidebar pane gets at least 3 lines when `height >= 3*(len(sidebar)-1) + 3`, and below that the other panes show only their title bar; no line is wider than the terminal. List the ranges checked and what each gave. The counter always equals the selected row's 1-based place and the pane's item count, never a line number: list the panes and selections checked, the empty pane case, and the detail pane (no counter).

**Interfaces:**
- Consumes: `sidebar`, `paneDetail`, per-pane state from Task 3.
- Produces: `func itemCount(selected, total int) string` returning `"<selected> of <total>"` with `selected` 1-based, and `"0 of 0"` when `total == 0`; Model field `expanded int` (`-1` when none); `func (m Model) leftHeights(h int) []int` returning one height per sidebar pane.

- [x] **Step 1: Write the failing tests**

```go
func TestItemCount(t *testing.T) {
	cases := []struct {
		sel, total int
		want       string
	}{{0, 0, "0 of 0"}, {1, 1, "1 of 1"}, {3, 20, "3 of 20"}, {236, 236, "236 of 236"}}
	for _, c := range cases {
		if got := itemCount(c.sel, c.total); got != c.want {
			t.Errorf("itemCount(%d, %d) = %q, want %q", c.sel, c.total, got, c.want)
		}
	}
}

func TestLeftHeightsFillAndExpand(t *testing.T) {
	m := newModel(t)
	others := len(sidebar) - 1
	for h := 3; h <= 60; h++ {
		for e := -1; e < len(sidebar); e++ {
			m.expanded = e
			hs := m.leftHeights(h)
			sum := 0
			for _, x := range hs {
				sum += x
			}
			if sum != h {
				t.Fatalf("h=%d expanded=%d: heights %v sum %d", h, e, hs, sum)
			}
			if e < 0 {
				continue
			}
			for i, x := range hs {
				if i == e {
					continue
				}
				if h >= 3*others+3 && x != 3 {
					t.Errorf("h=%d expanded=%d: pane %d got %d lines, want 3", h, e, i, x)
				}
				if h < 3*others+3 && x != 1 {
					t.Errorf("h=%d expanded=%d: pane %d got %d lines, want title only", h, e, i, x)
				}
			}
		}
	}
}

func TestZTogglesAndFocusRestores(t *testing.T) {
	// focus 3, press z: expanded == 2. press z: expanded == -1.
	// focus 3, z, then press 4: expanded == -1.
	// focus 0 (detail), press z: expanded stays -1.
}

func TestCounterShowsSelectedItemNotLine(t *testing.T) {
	// focus 3 (Plans), press j twice: the pane's bottom border ends with
	// "3 of <plans count>" at the right. An empty pane shows "0 of 0".
	// The detail pane's bottom border holds no " of ".
}
```

The title-only fallback below `3*others+3` lines is how "shrink to their title bar" reads in code; if the frame needs 2 lines for a title-only box on this layout, use that value and state it in the test name.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'ItemCount|LeftHeights|ZToggles|Counter' -v`
Expected: FAIL, `undefined: itemCount`.

- [x] **Step 3: Write the code**

Add `expanded` to Model, starting at `-1`. `z` on a sidebar pane toggles `m.expanded` between that pane and `-1`; `z` on detail does nothing; `focusPane` sets `m.expanded = -1` whenever the focus moves to another box. `leftHeights` gives equal shares when `expanded < 0` (the last pane takes the leftover), else 3 lines for each other pane and the rest for the expanded one, or 1 line each below the threshold. `geometry` uses `leftHeights`. Delete `count` and its call; sidebar panes pass `itemCount(selectedIndex+1, len(rows))` (or `0, 0`) as the foot, and `paneBottom` draws it at the right end of the bottom border. Detail passes no foot. Add `z expand` to `helpLines`.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -v`
Expected: PASS.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
wc -l internal/tui/*.go
git add internal/tui
git commit -m "feat(tui): z expands a sidebar pane, counter shows selected of total"
```

---

### Task 5: Move the backlog ideas into Scratchpad

**Files:**
- Create through the CLI only: six files in `.acta/scratch/`

**verify:** Every open idea the orchestrator hands over becomes exactly one scratch item with status `raw`, its body equal byte for byte to the text given, and one commit each with the exact message. List each slug, its SCRATCH id and its commit hash. No other file in the repo changes.

**Interfaces:**
- Consumes: `acta scratch new` (Task 2), built from this worktree.

The orchestrator puts six body files in the brief (they come from agent memory, outside the repo). Slugs, in this order: `newest-first-sort`, `themes`, `drag-select-copy`, `jev-routing-spike`, `skill-rules-spec`, `harness-spec`.

- [x] **Step 1: Build the CLI from the worktree**

Run: `go build -o /tmp/acta-scratch-sidebar ./cmd/acta`
Expected: exit 0.

- [x] **Step 2: File each idea**

```bash
for s in newest-first-sort themes drag-select-copy jev-routing-spike skill-rules-spec harness-spec; do
  /tmp/acta-scratch-sidebar scratch new "$s" < "<brief dir>/$s.md" || break
done
```

Expected: six lines, each with a SCRATCH id; `git log --oneline -6` shows six `acta: new scratch ...` commits.

- [x] **Step 3: Check**

Run: `/tmp/acta-scratch-sidebar list --json` and confirm six scratch items with status `raw`; `git status --porcelain` prints nothing; `go test ./...` passes.

- [x] **Step 4: Report the IDs**

Reply with the slug to SCRATCH id list, so the orchestrator can point the memory note at them.

---

### Task 6: Tabs in a body never break the frame

Fixes `.acta/bugs/2026-09-28-detail-tabs-break-frame.md`. When this plan lands, the orchestrator runs `acta set bugs/2026-09-28-detail-tabs-break-frame fixed_in <merge sha>`.

**Files:**
- Modify: `internal/tui/detail.go` (where the detail body text is turned into lines, before any width is measured)
- Test: `internal/tui/detail_test.go`

**verify:** No text the detail pane can show ever holds a tab character once it reaches the screen, and no detail line is ever wider than the pane. List every path that puts text into the detail pane (body, header fields, task list, debt body, anything else found) and whether each goes through the tab expansion. Leading tabs, tabs mid-line, several tabs in a row, and a tab right after a wide unicode character all expand to the next multiple of 8 columns.

**Interfaces:**
- Produces: `func expandTabs(s string) string` in `detail.go`: each `\t` becomes spaces up to the next column that is a multiple of 8, counting columns with `lipgloss.Width` of the text before it on the same line.

- [x] **Step 1: Write the failing tests**

```go
func TestExpandTabsToNextStop(t *testing.T) {
	cases := []struct{ in, want string }{
		{"\tx", "        x"},
		{"ab\tx", "ab      x"},
		{"\t\tx", "                x"},
		{"12345678\tx", "12345678        x"},
		{"a\nb\tc", "a\nb       c"},
		{"界\tx", "界      x"},
		{"no tabs", "no tabs"},
		{"", ""},
	}
	for _, c := range cases {
		if got := expandTabs(c.in); got != c.want {
			t.Errorf("expandTabs(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetailWithTabsFitsThePane(t *testing.T) {
	// Load a board whose selected plan body holds a Go code block indented
	// with tabs. Render the view at 80x30 and 160x50, scroll the detail pane
	// down line by line to the end. At every step: no line of View() holds
	// "\t", and no line is wider than the terminal.
}
```

Write the second test in full with a temp fixture holding a plan with tab-indented code.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'ExpandTabs|DetailWithTabs' -v`
Expected: FAIL, `undefined: expandTabs`.

- [x] **Step 3: Write the minimal code**

Write `expandTabs` and call it once on the raw body text (and any other detail text path the verify list finds) before it is rendered and wrapped, so every later width count sees spaces.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -v`
Expected: PASS.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui/detail.go internal/tui/detail_test.go
git commit -m "fix(tui): expand tabs in the detail pane so lines fit"
```

---

## Added tasks (approved by the user in chat on 2026-09-28, Bounded)

Waves: wave 5 = Task 7, wave 6 = Task 8 (both touch `internal/tui/view.go`).

### Task 7: Scrollbar thumb on the right border

**Files:**
- Modify: `internal/tui/scroll.go` (`scrollbar`), `internal/tui/view.go` (`paneView`, where `scrollbar(total, b.inner, first, b.inner)` is called), `internal/tui/frame.go` (right border drawing)
- Test: `internal/tui/scroll_test.go`, `internal/tui/view_test.go`

**verify:** No pane ever spends an inner column on a scrollbar, and the thumb is only ever drawn on the pane's own right border line. For every pane (all five sidebar panes and detail), at sizes 80x30 and 160x50, scrolled to top, middle and end, and with content that fits (no thumb at all): the last cell of every body row is the border cell; rows where the thumb sits show `┃` in the pane's border color (active color when focused, dim border color when not); every other border row shows the normal `│`; no `░` or `█` appears anywhere; the inner width equals the full pane width minus the two border columns. List every pane × size × position checked and what it gave.

**Interfaces:**
- Consumes: `scrollbar(total, visible, first, h int) []string` (today it returns `░`/`█` cells for an inner column).
- Produces: `scrollbar` returns, per row, whether the thumb covers that row (`[]bool`, length `h`; all false when `total <= visible`); the thumb stays proportional to `visible/total` with a minimum of 1 row, placed the way the current code places it. The right border is drawn from that list: `┃` where true, `│` where false, both in the border style the pane already uses.

- [x] **Step 1: Write the failing tests**

```go
func TestScrollbarRowsAreThumbOnly(t *testing.T) {
	cases := []struct {
		total, visible, first, h int
		want                     string // one char per row: T thumb, . border
	}{
		{10, 10, 0, 10, ".........."}, // fits: no thumb
		{0, 10, 0, 10, ".........."},
		{100, 10, 0, 10, "T........."},
		{100, 10, 90, 10, ".........T"},
		{20, 10, 5, 10, ".....TTTTT"},
		{1000, 5, 500, 5, "..T.."},
	}
	for _, c := range cases {
		rows := scrollbar(c.total, c.visible, c.first, c.h)
		got := ""
		for _, on := range rows {
			if on {
				got += "T"
			} else {
				got += "."
			}
		}
		if got != c.want {
			t.Errorf("scrollbar(%d,%d,%d,%d) = %s, want %s", c.total, c.visible, c.first, c.h, got, c.want)
		}
	}
}

func TestThumbSitsOnTheBorderNotInside(t *testing.T) {
	// For every pane, at 80x30 and 160x50, scrolled top/middle/end: strip ANSI
	// from View(), take the pane's rows, and check the right border column holds
	// only "│" or "┃", the column left of it is pane content (never "░", "█",
	// "┃"), and at least one "┃" shows when the pane's content is taller than
	// the pane. Check the thumb's style matches the pane's border style
	// (focused vs not) by rendering the thumb cell with the same style helper.
}
```

Adjust the exact rows in the table to the current proportional placement if one case differs by one row, and say so in the reply; the rule under test is: no thumb when content fits, thumb touches the top at the start and the bottom at the end, at least 1 row.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'ScrollbarRows|ThumbSits' -v`
Expected: FAIL (scrollbar returns strings; thumb drawn inside the pane).

- [x] **Step 3: Write the code**

Change `scrollbar` to return `[]bool`. In `paneView`, stop reserving the inner column for the bar, so the content gets that column back. Draw the right border row by row: `┃` where the thumb is, `│` elsewhere, in the pane's border style. Remove the `░`/`█` glyphs.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -v`
Expected: PASS.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui
git commit -m "feat(tui): scrollbar thumb on the right border, no track"
```

### Task 8: Dim everything behind a popup

**Files:**
- Modify: `internal/tui/view.go` (`cover`, `View`)
- Test: `internal/tui/view_test.go`

**verify:** While any popup is open (the `?` help, the `t` and `s` pickers, the new-bug slug prompt), no cell outside the popup box keeps its own color, and no cell inside the box is dimmed. List every popup kind checked, at 80x30 and 160x50. Closing the popup gives a `View()` byte-equal to the one before it opened.

**Interfaces:**
- Consumes: `cover(body string) string`, `popupBox() string`.
- Produces: `cover` draws the body behind the box with ANSI styling stripped and one faint grey style applied (`lipgloss.NewStyle().Faint(true).Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "240"})`), and draws the box as today. No new Model state.

- [x] **Step 1: Write the failing tests**

```go
func TestPopupDimsTheBackground(t *testing.T) {
	for _, open := range []string{"?", "t", "s", "n"} {
		for _, size := range [][2]int{{80, 30}, {160, 50}} {
			m := sized(newModel(t), size[0], size[1])
			before := m.View()
			withPopup := key(m, open)
			v := withPopup.View()
			// Every line outside the popup box's rows and columns renders in the
			// dim style only: re-render its plain text with the dim style and
			// compare. Every popup box line matches popupBox() output.
			_ = v
			closed := key(withPopup, "esc")
			if closed.View() != before {
				t.Errorf("popup %q at %v: view not restored after esc", open, size)
			}
		}
	}
}
```

Write the dim check in full: use `ansi.Strip` from the ANSI package already in go.mod (or lipgloss's own helper) to get plain text, then compare each background segment to `dim.Render(plain)`. Pick the key that opens each popup from the real key switch; skip a popup kind only if it cannot open on the fixture, and say which.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run PopupDims -v`
Expected: FAIL (background keeps its colors).

- [x] **Step 3: Write the code**

In `cover`, before placing the box, strip the styling from each body segment that sits outside the box and render it with the dim style. Keep the box exactly as today.

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -v`
Expected: PASS.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui
git commit -m "feat(tui): dim the screen behind a popup"
```
