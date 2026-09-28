---
id: PLAN-23
hash: kdgu
---
# Pane Sort Toggle Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Every sidebar pane sorts by the date in the file name, oldest first by default, and the key `o` flips the focused pane to newest first and back, with the current order shown in the pane title.

**Architecture:** One pure function `ordered` in a new `internal/tui/order.go` sorts a slice of items by file-name date. The model keeps one flag per sidebar pane. Every row builder in `internal/tui/sidebar.go` passes its items through `ordered` before it makes rows. The title adds one faint word, and the old commit-time order of the Done pane and its git lookups go away.

**Tech Stack:** Go, Bubble Tea, lipgloss. Tests are plain `testing` in package `tui` and `board`.

**Spec:** `.acta/specs/2026-09-29-pane-sort-toggle-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- The base is branch `top-tabs` at 6112ebb, merged into `pane-sort` (293ee72), so this runs beside PLAN-19. PLAN-19 may still change the TUI; the orchestrator merges main in again before landing. If a name below (`paneTop`, `rowsOf`, `openRows`, `activeRows`, `doneRows`, `treeRows`, `helpLines`) no longer exists, stop and report; do not guess.
- The key is `o`. The two title words are exactly `oldest` and `newest`.
- Every pane starts at `oldest` on each TUI start. The choice lives in memory only.
- Search results keep the board order; the flag does not touch them.
- Comments are plain English a 10-year-old can read: short words, say why.
- Before each commit run `gofmt -l .` (must print nothing), `go vet ./...` and `go test ./...`.

## File map

| File | Task | Change |
|---|---|---|
| `internal/tui/order.go` | 1 | new: `ordered` |
| `internal/tui/order_test.go` | 1 | new: order and toggle tests |
| `internal/tui/model.go` | 1 | field `newest`, init, key `o` |
| `internal/tui/sidebar.go` | 1 | row builders call `ordered`; Done pane drops `SortTime` |
| `internal/tui/model_test.go` | 1 | delete `TestDonePaneShowsTheMostRecentlyCommittedFirst` |
| `internal/tui/frame.go` | 2 | title word |
| `internal/tui/view.go` | 2 | help line for `o` |
| `internal/tui/frame_test.go` | 2 | title and help tests |
| `internal/board/closed.go` | 3 | delete `fillCommitTimes`, `SortTime` |
| `internal/board/board.go` | 3 | delete `CommitAt` field and the `fillCommitTimes` call |
| `internal/board/closed_test.go` | 3 | delete the tests of the removed code |
| `internal/gitc/*` | 3 | delete `LastCommit` only if nothing else calls it |

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2 and Task 3, in parallel. They share no file. Task 2 needs `m.newest` from Task 1. Task 3 needs Task 1 to have removed the last reader of `SortTime`.

---

### Task 1: Order rule, per-pane flag and the `o` key

**Files:**
- Create: `internal/tui/order.go`
- Create: `internal/tui/order_test.go`
- Modify: `internal/tui/model.go` (the `Model` struct, `New`, the key switch next to `case "t", "s":`)
- Modify: `internal/tui/sidebar.go` (`activeRows`, `openRows`, `doneRows`)
- Modify: `internal/tui/model_test.go` (delete `TestDonePaneShowsTheMostRecentlyCommittedFirst`)

**verify:** Every list a sidebar pane shows (Active, each kind tab, the plans tree, Done) follows the file-name date in the direction of that pane's own flag, and no other order is left in any of them. Same date never reorders between loads. Tasks never leave their plan or their file order. An item with no date is last in both directions. Pressing `o` changes only the focused pane, never the detail box, and keeps the selected item selected. List every row builder checked and which order each one used before and after.

**Interfaces:**
- Produces: `func ordered(items []*board.Item, newest bool) []*board.Item` (returns a new slice, never sorts the input in place).
- Produces: `Model.newest []bool`, indexed by sidebar pane, length `len(sidebar)`, all false at start. Task 2 reads `m.newest[p]`.

- [x] **Step 1: Write the failing tests** in `internal/tui/order_test.go`

```go
package tui

import (
	"slices"
	"testing"

	"github.com/iyay/acta/internal/board"
)

func ids(items []*board.Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.ID
	}
	return out
}

func spec(id, date string) *board.Item {
	return &board.Item{ID: id, Kind: board.KindStory, Title: id, Date: date, Status: "open"}
}

func TestOrderedSortsByFileDateBothWays(t *testing.T) {
	in := []*board.Item{spec("specs/2026-09-20-b", "2026-09-20"), spec("specs/2026-09-01-a", "2026-09-01"), spec("specs/2026-09-10-c", "2026-09-10")}
	if got, want := ids(ordered(in, false)), []string{"specs/2026-09-01-a", "specs/2026-09-10-c", "specs/2026-09-20-b"}; !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	if got, want := ids(ordered(in, true)), []string{"specs/2026-09-20-b", "specs/2026-09-10-c", "specs/2026-09-01-a"}; !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
	if in[0].ID != "specs/2026-09-20-b" {
		t.Fatalf("ordered sorted its input in place")
	}
}

func TestOrderedSameDateFallsBackToIDBothWays(t *testing.T) {
	in := []*board.Item{spec("specs/2026-09-29-b", "2026-09-29"), spec("specs/2026-09-29-a", "2026-09-29")}
	if got, want := ids(ordered(in, false)), []string{"specs/2026-09-29-a", "specs/2026-09-29-b"}; !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	if got, want := ids(ordered(in, true)), []string{"specs/2026-09-29-b", "specs/2026-09-29-a"}; !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
}

func TestOrderedKeepsTasksUnderTheirPlanInFileOrder(t *testing.T) {
	task := func(id string) *board.Item {
		return &board.Item{ID: id, Kind: board.KindTask, Date: "2026-09-29", Status: "todo"}
	}
	in := []*board.Item{task("plans/2026-09-29-x#task-1"), task("plans/2026-09-29-x#task-2"), task("plans/2026-09-29-x#task-3")}
	want := []string{"plans/2026-09-29-x#task-1", "plans/2026-09-29-x#task-2", "plans/2026-09-29-x#task-3"}
	for _, newest := range []bool{false, true} {
		if got := ids(ordered(in, newest)); !slices.Equal(got, want) {
			t.Fatalf("newest=%v: got %v, want %v", newest, got, want)
		}
	}
}

func TestOrderedPutsItemsWithNoDateLast(t *testing.T) {
	in := []*board.Item{spec("specs/undated", ""), spec("specs/2026-09-01-a", "2026-09-01"), spec("specs/2026-09-20-b", "2026-09-20")}
	for _, newest := range []bool{false, true} {
		got := ids(ordered(in, newest))
		if got[len(got)-1] != "specs/undated" {
			t.Fatalf("newest=%v: undated item not last: %v", newest, got)
		}
	}
}

func rowIDs(rows []row) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.id
	}
	return out
}

func TestOFlipsOnlyTheFocusedPane(t *testing.T) {
	m := newModel(t)
	m.board = &board.Board{Items: []*board.Item{
		spec("specs/2026-09-20-b", "2026-09-20"),
		spec("specs/2026-09-01-a", "2026-09-01"),
		{ID: "bugs/2026-09-20-y", Kind: board.KindBug, Date: "2026-09-20", Status: "open"},
		{ID: "bugs/2026-09-01-x", Kind: board.KindBug, Date: "2026-09-01", Status: "open"},
	}}
	m = press(m, "2")
	if got, want := rowIDs(m.rowsOf(paneSpecs)), []string{"specs/2026-09-01-a", "specs/2026-09-20-b"}; !slices.Equal(got, want) {
		t.Fatalf("start: got %v, want oldest first %v", got, want)
	}
	m = press(m, "o")
	if got, want := rowIDs(m.rowsOf(paneSpecs)), []string{"specs/2026-09-20-b", "specs/2026-09-01-a"}; !slices.Equal(got, want) {
		t.Fatalf("after o: got %v, want newest first %v", got, want)
	}
	if got, want := rowIDs(m.rowsOf(paneBugs)), []string{"bugs/2026-09-01-x", "bugs/2026-09-20-y"}; !slices.Equal(got, want) {
		t.Fatalf("o in Specs moved the Bugs pane: got %v, want %v", got, want)
	}
	m = press(m, "o")
	if got, want := rowIDs(m.rowsOf(paneSpecs)), []string{"specs/2026-09-01-a", "specs/2026-09-20-b"}; !slices.Equal(got, want) {
		t.Fatalf("second o: got %v, want oldest first again %v", got, want)
	}
}

func TestOInTheDetailBoxChangesNoPane(t *testing.T) {
	m := newModel(t)
	m = press(m, "0", "o")
	for p := range len(sidebar) {
		if m.newest[p] {
			t.Fatalf("o in the detail box flipped pane %d", p)
		}
	}
}

func TestDonePaneFollowsFileDateNotCommitTime(t *testing.T) {
	m := newModel(t)
	m.board = &board.Board{Items: []*board.Item{
		{ID: "specs/2026-09-20-late", Kind: board.KindStory, Date: "2026-09-20", Status: "done"},
		{ID: "specs/2026-09-01-early", Kind: board.KindStory, Date: "2026-09-01", Status: "done"},
	}}
	m.openTab(tabSpecs)
	if got, want := rowIDs(m.rowsOf(paneDone)), []string{"specs/2026-09-01-early", "specs/2026-09-20-late"}; !slices.Equal(got, want) {
		t.Fatalf("Done pane: got %v, want oldest first %v", got, want)
	}
}
```

If `newModel`, `press` or `openTab(tabSpecs)` changed name after PLAN-19, use the helper the other tests in `model_test.go` use now, with the same meaning.

- [x] **Step 2: Delete the old test** `TestDonePaneShowsTheMostRecentlyCommittedFirst` in `internal/tui/model_test.go`. It pins the commit-time order the spec removes.

- [x] **Step 3: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'Ordered|OFlips|OInTheDetail|DonePaneFollows' -v`
Expected: FAIL to compile with `undefined: ordered` and `m.newest undefined`.

- [x] **Step 4: Write `internal/tui/order.go`**

```go
package tui

import (
	"sort"
	"strings"

	"github.com/iyay/acta/internal/board"
)

// ordered gives items sorted by the date in their file name, oldest first, or
// newest first when newest is set. Items on the same date sort by the file
// they live in, so the list never jumps between loads. The sort is stable and
// every task of a plan lives in the same file, so tasks keep their file order.
// An item with no date goes last either way, since there is nothing to place
// it by. It returns a new slice so the board keeps its own order.
func ordered(items []*board.Item, newest bool) []*board.Item {
	out := append([]*board.Item(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		a, c := out[i], out[j]
		if (a.Date == "") != (c.Date == "") {
			return c.Date == ""
		}
		if a.Date != c.Date {
			return (a.Date < c.Date) != newest
		}
		fa, fc := fileOf(a.ID), fileOf(c.ID)
		if fa != fc {
			return (fa < fc) != newest
		}
		return false
	})
	return out
}

// fileOf drops the "#task-N" end of a task ID, so a task sorts with its plan.
func fileOf(id string) string {
	file, _, _ := strings.Cut(id, "#")
	return file
}
```

- [x] **Step 5: Add the flag and the key** in `internal/tui/model.go`

In the `Model` struct, next to `tab`:

```go
	newest    []bool // per sidebar pane: true shows the newest file date first
```

In `New`, next to `tab: make([]int, len(sidebar)),`:

```go
		newest:   make([]bool, len(sidebar)),
```

In the key switch, next to `case "t", "s":`:

```go
	case "o":
		// The detail box has no list to sort.
		if m.focus != paneDetail {
			m.newest[m.focus] = !m.newest[m.focus]
		}
```

The selection is kept by item ID (`cursorOf`), so the same item stays selected after the flip. Do not move the cursor.

- [x] **Step 6: Route every row builder through `ordered`** in `internal/tui/sidebar.go`

- `activeRows`: `return toRows(ordered(going, m.newest[paneActive]), 0)`
- `openRows`: `m.treeRows(ordered(m.board.List(board.KindPlan, false), m.newest[p]))` and `toRows(ordered(m.board.List(tab.kind, false), m.newest[p]), 0)`. Leave the legacy group rows as they are.
- `doneRows`: replace the `sort.SliceStable(... SortTime ...)` block with `finished = ordered(finished, m.newest[paneDone])`. Update its comment: the file-name date decides, oldest first unless the pane was flipped. Drop the `sort` import if nothing else uses it.
- `searchRows` stays as it is.

- [x] **Step 7: Run the tests to see them pass**

Run: `go test ./internal/tui/ -run 'Ordered|OFlips|OInTheDetail|DonePaneFollows' -v`
Expected: PASS.

Run: `gofmt -l . && go vet ./... && go test ./...`
Expected: no gofmt output, vet clean, all packages PASS. If an older tui test pinned newest-first board order in a pane, it now meets the approved `oldest` default: fix its expected order to the spec rule and say which tests changed in the report.

- [x] **Step 8: Commit**

```bash
git add internal/tui/order.go internal/tui/order_test.go internal/tui/model.go internal/tui/sidebar.go internal/tui/model_test.go
git commit -m "feat(tui): sort every pane by file date, o flips the focused pane"
```

---

### Task 2: Sort word in the title and `o` in the help

**Files:**
- Modify: `internal/tui/frame.go` (`paneTop`)
- Modify: `internal/tui/view.go` (`helpLines`)
- Test: `internal/tui/frame_test.go`

**verify:** Every sidebar pane title shows exactly one sort word, and it always matches that pane's flag; the detail box never shows one. A narrow title drops the sort word before any tab name. The help popup lists `o`. List every pane checked and the word each shows before and after `o`.

**Interfaces:**
- Consumes: `Model.newest []bool` from Task 1.

- [x] **Step 1: Write the failing tests** in `internal/tui/frame_test.go`

```go
func TestTitleShowsTheSortOfEachPane(t *testing.T) {
	m := newModel(t)
	view := m.View()
	if got := strings.Count(view, "oldest"); got != len(sidebar) {
		t.Fatalf("start: %d panes say oldest, want %d", got, len(sidebar))
	}
	if strings.Contains(view, "newest") {
		t.Fatalf("start: a pane says newest before any o")
	}
	m = press(m, "2", "o")
	view = m.View()
	if got := strings.Count(view, "newest"); got != 1 {
		t.Fatalf("after o: %d panes say newest, want 1", got)
	}
	if got := strings.Count(view, "oldest"); got != len(sidebar)-1 {
		t.Fatalf("after o: %d panes say oldest, want %d", got, len(sidebar)-1)
	}
}

func TestHelpListsTheSortKey(t *testing.T) {
	if !strings.Contains(helpLines, "oldest / newest") {
		t.Fatalf("help does not list o: %q", helpLines)
	}
}
```

Add `"strings"` to the imports if the file does not have it. If a test item title or ID on the default board contains the word `oldest` or `newest`, give the test its own empty `board.Board{}`.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'TitleShowsTheSort|HelpListsTheSortKey' -v`
Expected: FAIL: `0 panes say oldest` and `help does not list o`.

- [x] **Step 3: Add the word to the title** in `paneTop`, right before `return topLine(b.w, edge, segs)`:

```go
	// The sort word goes last, so a narrow title drops it before any tab name.
	if p != paneDetail {
		word := "oldest"
		if m.newest[p] {
			word = "newest"
		}
		segs = append(segs, segment{text: " " + word + " ", style: faint})
	}
```

Check that `topLine` drops pieces from the right (its comment says so); if it no longer does, stop and report.

- [x] **Step 4: Add the help line** in `helpLines`, right after the `z` line, lined up with the other rows:

```
o                flip the sort: oldest / newest
```

- [x] **Step 5: Run the tests to see them pass**

Run: `go test ./internal/tui/ -run 'TitleShowsTheSort|HelpListsTheSortKey' -v`
Expected: PASS.

Run: `gofmt -l . && go vet ./... && go test ./...`
Expected: clean, all PASS. A golden or width test of the title may now see the extra word: update its expected text to include it and name it in the report.

- [x] **Step 6: Commit**

```bash
git add internal/tui/frame.go internal/tui/view.go internal/tui/frame_test.go
git commit -m "feat(tui): show each pane's sort in its title, list o in help"
```

---

### Task 3: Remove the commit-time lookups

**Files:**
- Modify: `internal/board/closed.go` (delete `fillCommitTimes`, `SortTime`)
- Modify: `internal/board/board.go` (delete the `CommitAt` field and the `b.fillCommitTimes()` call)
- Modify: `internal/board/closed_test.go` (delete the tests that call `SortTime` or set `CommitAt`)
- Modify: `internal/gitc/*` (delete `LastCommit` and its test only if nothing else calls it)

**verify:** No code path asks git for a last-commit time while the board loads, and nothing left reads `CommitAt` or `SortTime`. Author lookup (`fillAuthors`) still works. List every caller searched for and what `go build ./...` and `go test ./...` say.

**Interfaces:**
- Consumes: Task 1 removed the last reader of `SortTime` in `internal/tui/sidebar.go`.

- [x] **Step 1: Confirm the removed code has no reader left**

Run: `grep -rn "SortTime\|CommitAt\|fillCommitTimes\|LastCommit" --include='*.go' .`
Expected: hits only in `internal/board/closed.go`, `internal/board/board.go`, `internal/board/closed_test.go` and `internal/gitc/`. Any other hit: stop and report.

- [x] **Step 2: Delete the code and its tests**

- `internal/board/closed.go`: delete `fillCommitTimes` and `SortTime` with their comments. Keep `fillAuthors` and the `gitAuthor` / `gitUserName` variables. Drop imports that go unused.
- `internal/board/board.go`: delete the `CommitAt` line in `Item` and the `b.fillCommitTimes()` call.
- `internal/board/closed_test.go`: delete every test that calls `SortTime` or sets `CommitAt`. Keep the author tests.
- `internal/gitc`: if Step 1 showed `LastCommit` has no caller outside `internal/board/closed.go`, delete it and its test.

- [x] **Step 3: Run the checks**

Run: `grep -rn "SortTime\|CommitAt\|fillCommitTimes" --include='*.go' .`
Expected: no output.

Run: `gofmt -l . && go vet ./... && go test ./...`
Expected: clean, all PASS, including the `fillAuthors` tests.

- [x] **Step 4: Commit**

```bash
git add -A internal/board internal/gitc
git commit -m "refactor(board): drop commit-time lookups, panes sort by file date"
```
