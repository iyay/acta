---
id: PLN-0019
hash: dkk062l
---
# TUI Top Tabs Per Kind Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Replace the five-pane sidebar with a top tab bar (Scratches, Bugs, Debts, Specs, Plans, Activities), each kind tab holding a List pane and a Done pane, Activities holding one List of in-progress tasks, and the Plans lists drawn as a `+`/`-` tree of plans and their tasks.

**Architecture:** Task 1 adds the plan tree on today's layout: one `treeRows` builder in `internal/tui/sidebar.go`, an `openPlans` map and a `toggleRow` key handler in `model.go`, and `+`/`-`/dot marks in `rowText`. Task 2 swaps the `sidebar` table for a `topTabs` table, keeps the live cursor slots (`focus`, `last`, `sel`, `idx`, `off`, `done`) for the open tab and parks the others in `Model.tabs`, draws the tab bar as the first screen line, and rewrites the old five-pane tests. `focusPane`, scroll, scrollbars and popups keep their code.

**Tech Stack:** Go 1.27, Bubble Tea v1.3.10, lipgloss.

**Spec:** `.acta/specs/2026-09-28-tui-top-tabs-design.md`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- TUI layout change only. No change to kinds, statuses, files or the CLI. No file outside `internal/tui` changes.
- Tab names and order, exact: `Scratches`, `Bugs`, `Debts`, `Specs`, `Plans`, `Activities`. Keys `1`..`6` in that order.
- Done sub-tabs, exact: Scratches `Specced`/`Dropped`, Bugs `Fixed`/`Wontfix`, Debts `Done`/`Wontfix`, Specs `Done`/`Dropped`, Plans `Done`/`Dropped`. Activities has no Done pane.
- The TUI opens on Activities.
- Empty list detail text, exact: `No items`.
- Tree marks, exact: a plan row starts with `+` (shut) or `-` (open); a task row starts with its status dot (`dotGoing`, `dotWaiting`, `dotDone` from `internal/tui/detail.go`).
- No subtask rows in any list. No saving of tab, cursor or expand state to disk.
- In progress means `inProgress(it)` in `internal/tui/model.go`. Reuse it; do not write a second rule.
- No file in `internal/tui` grows past 800 lines.
- Comments: plain English a 10-year-old reads, say why. No marker tags, no Latin, no emoji.
- TDD: write the failing test first, run it and watch it fail for the right reason, then the minimum code, then watch it pass.
- Gates before every commit, inside the task commit: `gofmt -l .` prints nothing, `go vet ./...` clean, `go test ./...` passes.
- Existing tests that pin the five-pane sidebar are rewritten for the new layout, never deleted without a replacement. Never delete test fixtures to make a test pass.
- Never `git reset`, `rebase`, `amend` or `push`. Never file acta bugs for defects in this branch; fix them in the task.

## File map

| File | Tasks |
|---|---|
| `internal/tui/sidebar.go` (`treeRows`, plan branch of `openRows` and `doneRows`), `internal/tui/model.go` (`row.tree`, `Model.openPlans`, `toggleRow`, `" "` and `enter` keys), `internal/tui/scroll.go` (`rowText` marks, `treeMark`), `internal/tui/detail.go` (`dotOf`, `workLine` uses it), `internal/tui/plantree_test.go` (new), `internal/tui/model_test.go` (`TestEnterFocusesDetailFromBothListPanes` only) | 1 |
| `internal/tui/sidebar.go` (`topTabs` table, `tabState`, `openTab`, `panes`, `cyclePane`, rows, titles), `internal/tui/model.go` (Model fields, `New`, key switch, `focusPane`, `cycleTab`, `clickTab`), `internal/tui/frame.go` (`geometry`, `leftHeights`), `internal/tui/scroll.go` (`boxOf`), `internal/tui/view.go` (`tabBar`, `View`, `helpLines`), `internal/tui/detail.go` (`No items`), every `internal/tui/*_test.go` that names the old sidebar | 2 |

## Waves

- Wave 1: Task 1
- Wave 2: Task 2 (touches every file Task 1 touched, and needs the tree so task detail stays reachable once the Tasks inner tab goes)

---

### Task 1: Plans tree in the Plans lists

Built on today's layout, where key `3` focuses the Plans ─ Tasks pane with the Plans tab open and key `5` then shows its Done pane. Task 2 moves the two test helpers below to the new keys.

**Files:**
- Modify: `internal/tui/sidebar.go` (`openRows`, `doneRows`; add `treeRows`)
- Modify: `internal/tui/model.go` (`row` struct, `Model` struct, key switch, add `toggleRow`)
- Modify: `internal/tui/scroll.go` (`rowText`; add `treeMark`)
- Modify: `internal/tui/detail.go` (add `dotOf`, `workLine` uses it)
- Modify: `internal/tui/model_test.go` (`TestEnterFocusesDetailFromBothListPanes`)
- Create: `internal/tui/plantree_test.go`

**verify:** In every Plans list (the open list and the Done list), the rows are exactly: each plan once, and right under a plan the reader opened, each of its tasks once in file order, every status; nothing else. Every plan starts shut. Space and enter on a plan row flip only that plan; on a task row they change nothing and do not move the focus; outside a tree list enter still focuses the detail. A plan row selects plan detail and a task row selects task detail. A plan with no tasks shows `+` and adds no row when opened. List every path checked: open list, Done list, plan with tasks, plan without tasks, task row, non-tree list.

**Interfaces:**
- Produces: `row.tree bool` (set on every row a tree builds); `Model.openPlans map[string]bool` (plan ID to open; never mutated in place, always replaced); `func (m Model) treeRows(plans []*board.Item) []row`; `func (m *Model) toggleRow() bool` (true on any tree row, so the caller stops there); `func (m Model) treeMark(r row, it *board.Item) string`; `func dotOf(it *board.Item) string`.

- [x] **Step 1: Write the failing tests**

```go
// internal/tui/plantree_test.go
package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
)

// toPlans puts the focus on the open plans list, the way a reader gets there.
func toPlans(m Model) Model { return press(m, "3") }

// toPlansDone puts the focus on the finished plans list.
func toPlansDone(m Model) Model { return press(toPlans(m), "5") }

func planModel(t *testing.T) Model {
	t.Helper()
	return sized(newModel(t), 160, 50)
}

func TestPlanRowsStartShutAndSpaceOrEnterFlipThem(t *testing.T) {
	m := toPlans(planModel(t))
	shut := []string{"plans/2026-09-23-lonely", "plans/2026-09-21-alpha"}
	if got := rowIDs(m); !slices.Equal(got, shut) {
		t.Fatalf("plans list %q, want every plan shut %q", got, shut)
	}
	open := []string{"plans/2026-09-23-lonely", "plans/2026-09-23-lonely#task-1", "plans/2026-09-21-alpha"}
	for _, k := range []string{" ", "enter"} {
		opened := press(m, k)
		if got := rowIDs(opened); !slices.Equal(got, open) {
			t.Errorf("after %q the list is %q, want %q", k, got, open)
		}
		if opened.focus == paneDetail {
			t.Errorf("%q on a plan row moved the focus to the detail", k)
		}
		if got := rowIDs(press(opened, k)); !slices.Equal(got, shut) {
			t.Errorf("a second %q left %q, want the plan shut again", k, got)
		}
	}
}

func TestTreeRowsSelectTheirOwnDetail(t *testing.T) {
	m := press(toPlans(planModel(t)), " ")
	if it := m.Selected(); it == nil || it.Kind != board.KindPlan {
		t.Fatalf("a plan row selects %v, want the plan", it)
	}
	m = press(m, "j")
	it := m.Selected()
	if it == nil || it.ID != "plans/2026-09-23-lonely#task-1" || it.Kind != board.KindTask {
		t.Fatalf("a task row selects %v, want the task", it)
	}
	before := rowIDs(m)
	for _, k := range []string{" ", "enter"} {
		after := press(m, k)
		if got := rowIDs(after); !slices.Equal(got, before) {
			t.Errorf("%q on a task row changed the list to %q", k, got)
		}
		if after.focus == paneDetail {
			t.Errorf("%q on a task row moved the focus to the detail", k)
		}
	}
}

func TestTreeRowsWearTheirMarks(t *testing.T) {
	m := toPlans(planModel(t))
	rows := m.rowsOf(m.listPane())
	if head := m.rowText(rows[1], m.board.Get(rows[1].id), 80); !strings.HasPrefix(head, "+ ") {
		t.Errorf("a shut plan row reads %q, want it to start with +", head)
	}
	m = press(m, "j", " ")
	rows = m.rowsOf(m.listPane())
	want := map[string]string{
		"plans/2026-09-21-alpha":        "- ",
		"plans/2026-09-21-alpha#task-1": "  " + dotDone + " ",
		"plans/2026-09-21-alpha#task-2": "  " + dotGoing + " ",
	}
	for _, r := range rows {
		mark, ok := want[r.id]
		if !ok {
			continue
		}
		if head := m.rowText(r, m.board.Get(r.id), 80); !strings.HasPrefix(head, mark) {
			t.Errorf("row %s reads %q, want it to start with %q", r.id, head, mark)
		}
		delete(want, r.id)
	}
	if len(want) != 0 {
		t.Errorf("rows never drawn: %v", want)
	}
}

func TestAPlanWithNoTasksAddsNoRows(t *testing.T) {
	cfg := treeCfg(t, map[string]string{".acta/plans/2026-09-20-empty.md": "# Empty plan\n"})
	m := toPlans(sized(detailModel(t, cfg), 160, 50))
	if got := rowIDs(m); !slices.Equal(got, []string{"plans/2026-09-20-empty"}) {
		t.Fatalf("plans list %q", got)
	}
	r := m.rowsOf(m.listPane())[0]
	if head := m.rowText(r, m.board.Get(r.id), 80); !strings.HasPrefix(head, "+ ") {
		t.Errorf("a plan with no tasks reads %q, want it to start with +", head)
	}
	if got := rowIDs(press(m, " ")); !slices.Equal(got, []string{"plans/2026-09-20-empty"}) {
		t.Errorf("opening a plan with no tasks gave %q, want no new row", got)
	}
}

func TestTheDonePlansTreeBehavesTheSame(t *testing.T) {
	m := toPlansDone(planModel(t))
	rows := doneRowIDs(m)
	if len(rows) == 0 {
		t.Fatal("the fixture has no finished plan, so this test proves nothing")
	}
	for _, id := range rows {
		if it := m.board.Get(id); it == nil || it.Kind != board.KindPlan {
			t.Fatalf("a shut Done tree holds %s, want plans only", id)
		}
	}
	first := m.board.Get(rows[0])
	opened := doneRowIDs(press(m, " "))
	want := append(append([]string{rows[0]}, first.Children...), rows[1:]...)
	if !slices.Equal(opened, want) {
		t.Errorf("opening %s in Done gave %q, want %q", rows[0], opened, want)
	}
}

func TestOpeningAPlanLeavesOlderModelsAlone(t *testing.T) {
	m := toPlans(planModel(t))
	_ = press(m, " ")
	if got := len(rowIDs(m)); got != 2 {
		t.Errorf("the model before space now shows %d rows, want its 2 shut plans", got)
	}
}
```

In `internal/tui/model_test.go`, `TestEnterFocusesDetailFromBothListPanes` walks every kind tab and presses enter. Enter on a plan row now opens the plan, so skip the Plans tab in that walk and move its Done box case to the Bugs pane:

```go
	for p := pane(1); p < paneDone; p++ {
		for tab := range sidebar[p].tabs {
			// Enter on a plan row opens the plan instead, which the tree tests
			// in plantree_test.go pin.
			if sidebar[p].tabs[tab].kind == board.KindPlan {
				continue
			}
```

and change `m := press(newModel(t), "3", "5")` to `m := press(newModel(t), "4", "5")`.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -run 'Plan|Tree|Enter' -v`
Expected: FAIL to build with `m.openPlans undefined` and `m.treeMark undefined`, or once stubbed, FAIL with the plans list holding no task rows.

- [x] **Step 3: Write the code**

`internal/tui/detail.go`, add `dotOf` above `workLine` and use it there:

```go
// dotOf is the status dot of an item: finished, under way, or waiting. The
// tree rows and the detail lists both wear it, so the two always agree.
func dotOf(it *board.Item) string {
	switch {
	case board.Closed(it.Status):
		return dotDone
	case inProgress(it):
		return dotGoing
	}
	return dotWaiting
}
```

```go
func workLine(it *board.Item, on bool, w int) string {
	mark, brush := dotOf(it), faint
	if mark == dotGoing {
		brush = work
	}
	if on {
		brush = lipgloss.NewStyle()
	}
```

(the rest of `workLine` stays as it is).

`internal/tui/model.go`, the `row` struct gets one field and the `Model` one field:

```go
type row struct {
	id    string
	group bool
	depth int
	tree  bool // a row of a plans tree, which space and enter act on
}
```

```go
	openPlans map[string]bool // the plans the reader opened in a tree list
```

Add `toggleRow` next to `enter`, and add `"maps"` to the imports:

```go
// toggleRow opens or shuts the plan under the cursor of a tree list. It says
// true on any tree row, so space and enter leave a task row alone and never
// jump to the detail from a tree.
func (m *Model) toggleRow() bool {
	if m.focus == paneDetail {
		return false
	}
	rows := m.listOf()
	i := m.cursor()
	if i < 0 || !rows[i].tree {
		return false
	}
	if rows[i].depth > 0 {
		return true
	}
	// A new map each time, so an older copy of the model keeps the tree it
	// drew.
	open := make(map[string]bool, len(m.openPlans)+1)
	maps.Copy(open, m.openPlans)
	open[rows[i].id] = !open[rows[i].id]
	m.openPlans = open
	m.keepVisible(m.listPane())
	return true
}
```

In the key switch:

```go
	case " ":
		m.toggleRow()
```

```go
	case "enter":
		if m.toggleRow() {
			return m, nil
		}
		return m.enter()
```

`internal/tui/sidebar.go`, add `treeRows` and use it in the two plan lists:

```go
// treeRows lays plans out as a tree: one row per plan, and under a plan the
// reader opened, one row per task in file order, whatever its status. Each
// row's id is the item it selects, so the detail box needs nothing else.
func (m Model) treeRows(plans []*board.Item) []row {
	out := make([]row, 0, len(plans))
	for _, p := range plans {
		out = append(out, row{id: p.ID, tree: true})
		if !m.openPlans[p.ID] {
			continue
		}
		for _, id := range p.Children {
			out = append(out, row{id: id, depth: 1, tree: true})
		}
	}
	return out
}
```

In `openRows`, right after `tab, ok := m.tabOf(p)` and its `!ok` check:

```go
	if tab.kind == board.KindPlan {
		return m.treeRows(m.board.List(board.KindPlan, false))
	}
```

In `doneRows`, after the sort:

```go
	if tab, ok := m.tabOf(m.follows); ok && tab.kind == board.KindPlan {
		return m.treeRows(finished)
	}
	return toRows(finished, 0)
```

`internal/tui/scroll.go`, in `rowText` replace the `head :=` line and add `treeMark` below `rowText`:

```go
	lead := strings.Repeat("  ", r.depth)
	if r.tree {
		lead += m.treeMark(r, it) + " "
	}
	head := lead + name + "  " + it.Title
```

```go
// treeMark is what a tree row starts with: + on a shut plan, - on an open
// one, and the status dot on a task, so the reader sees what opens and what
// is done without reading the detail.
func (m Model) treeMark(r row, it *board.Item) string {
	if r.depth > 0 {
		return dotOf(it)
	}
	if m.openPlans[it.ID] {
		return "-"
	}
	return "+"
}
```

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -v`
Expected: PASS, the whole package. A test outside this task that pressed enter on a plan row to reach the detail is changed to press `0`, since enter on a plan row now opens the plan.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui
git commit -m "feat(tui): plans lists fold open into their tasks"
```

---

### Task 2: Top tab bar, one tab per kind

**Files:**
- Modify: `internal/tui/sidebar.go` (replace the `sidebar` table and its helpers)
- Modify: `internal/tui/model.go` (`Model` fields, `New`, `focusPane`, `cycleTab`, key switch, `clickTab`)
- Modify: `internal/tui/frame.go` (`geometry`, `leftHeights`)
- Modify: `internal/tui/scroll.go` (`boxOf`)
- Modify: `internal/tui/view.go` (`helpLines`, `View`; add `tabBar`)
- Modify: `internal/tui/detail.go` (`detailLines` empty case)
- Modify: `internal/tui/plantree_test.go` (`toPlans`, `toPlansDone` bodies)
- Rewrite: `internal/tui/sidebar_test.go`
- Modify: `internal/tui/model_test.go`, `internal/tui/view_test.go`, `internal/tui/scroll_test.go`, `internal/tui/detail_test.go`, `internal/tui/frame_test.go`, `internal/tui/trees_test.go` where they name the old sidebar

**verify:** Every key path lands in the open tab only: `1`..`6` and `←`/`→` open exactly the tab named, `tab`/`shift+tab` never focus a pane of another tab and never a Done pane on Activities, `[`/`]` change the Done sub-tab only while the Done pane has the focus. Every tab shows its own rows on every visit: the first visit selects the top List row and the detail shows it; a return visit gets back its pane, cursor row, Done sub-tab and open plans; a list that shrank puts the cursor on its last row; an empty list shows `No items`. Activities lists every in-progress task of every plan and nothing else. List every tab checked for each of these, and every old five-pane test with the test that replaced it.

**Interfaces:**
- Consumes: `treeRows`, `toggleRow`, `row.tree`, `Model.openPlans`, `dotOf` from Task 1.
- Produces: `type topTab struct{ name string; kind board.Kind; done []doneTab; tree bool }`; `var topTabs = [...]topTab{...}`; consts `tabScratches`, `tabBugs`, `tabDebts`, `tabSpecs`, `tabPlans`, `tabActivities` (0..5); consts `paneList = pane(0)`, `paneDone = pane(1)`, `paneDetail = pane(2)`, `sidePanes = 2`, `boxes = 3`; `type tabState struct{ focus, last pane; sel []string; idx, off []int; done int }`; `func (m *Model) openTab(i int)`; `func (m Model) panes() []pane`; `func (m *Model) cyclePane(step int)`; `func (m Model) activityRows() []row`; `func (m Model) tabBar() string`; Model fields `top int`, `tabs [len(topTabs)]tabState`, `done int`. Removed: `sidebar`, `sidebarTab`, `sidebarPane`, `paneActive`, `paneSpecs`, `panePlans`, `paneBugs`, `tabOf`, `activeRows`, `Model.follows`, `Model.tab`.

- [x] **Step 1: Write the failing tests**

Replace `internal/tui/sidebar_test.go` whole. It keeps `TestNoDividerRow` (walk rewritten) and `TestScratchDetailNamesTheLinkedSpec` (unchanged body). The old tests go with a named replacement each: `TestNumberKeysFocusTheirBox` becomes `TestNumberKeysOpenTheirTab`, `TestTabCyclesSixBoxes` becomes `TestTabCyclesOnlyTheOpenTabsPanes`, `TestSidebarTitles` becomes `TestTopTabsTable`, `TestActiveHoldsOnlyWorkInProgress` becomes `TestActivitiesListsOnlyInProgressTasks`, `TestDoneFollowsLastSidebarPane` becomes `TestDoneSubTabKeysOnlyActOnTheDonePane` plus the Done columns of `TestTopTabsTable`.

```go
// internal/tui/sidebar_test.go
package tui

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
)

// tabKey is the key that opens tab i of the bar.
func tabKey(i int) string { return strconv.Itoa(i + 1) }

// everyPlanOpen opens every plan of a board, so a walk over the lists also
// meets every task.
func everyPlanOpen(b *board.Board) map[string]bool {
	open := map[string]bool{}
	for _, p := range b.List(board.KindPlan, true) {
		open[p.ID] = true
	}
	return open
}

func TestTopTabsTable(t *testing.T) {
	want := []struct {
		name string
		kind board.Kind
		done string
		tree bool
	}{
		{"Scratches", board.KindScratch, "Specced Dropped", false},
		{"Bugs", board.KindBug, "Fixed Wontfix", false},
		{"Debts", board.KindDebtItem, "Done Wontfix", false},
		{"Specs", board.KindStory, "Done Dropped", false},
		{"Plans", board.KindPlan, "Done Dropped", true},
		{"Activities", "", "", false},
	}
	if len(topTabs) != len(want) {
		t.Fatalf("%d tabs, want %d", len(topTabs), len(want))
	}
	for i, w := range want {
		tab := topTabs[i]
		var done []string
		for _, d := range tab.done {
			done = append(done, d.name)
			// The status a sub-tab lists is its own name in lower case.
			if d.status != strings.ToLower(d.name) {
				t.Errorf("%s sub-tab %s lists status %q", tab.name, d.name, d.status)
			}
		}
		if tab.name != w.name || tab.kind != w.kind || strings.Join(done, " ") != w.done || tab.tree != w.tree {
			t.Errorf("tab %d is %+v, want %+v", i, tab, w)
		}
	}
	if topTabs[tabScratches].name != "Scratches" || topTabs[tabPlans].name != "Plans" || topTabs[tabActivities].name != "Activities" {
		t.Error("the tab constants do not match the table")
	}
}

func TestTUIOpensOnActivities(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	if m.top != tabActivities || m.focus != paneList {
		t.Fatalf("opens on tab %d pane %d, want Activities and its List", m.top, m.focus)
	}
	if it := m.Selected(); it == nil || it.ID != "plans/2026-09-21-alpha#task-2" {
		t.Errorf("opens on %v, want the one task in progress", it)
	}
	bar := plain(strings.Split(m.View(), "\n")[0])
	if !strings.HasPrefix(bar, " Scratches  Bugs  Debts  Specs  Plans  [Activities]") {
		t.Errorf("the top line is %q, want the tab bar with Activities marked", bar)
	}
}

func TestNumberKeysOpenTheirTab(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	for from := range topTabs {
		for to := range topTabs {
			got := press(m, tabKey(from), tabKey(to))
			if got.top != to || got.focus != paneList {
				t.Errorf("from %d press %s: tab %d pane %d, want tab %d on its List", from, tabKey(to), got.top, got.focus, to)
			}
		}
	}
	if got := press(m, "7").top; got != tabActivities {
		t.Errorf("7 is no tab but moved to tab %d", got)
	}
}

func TestLeftRightWalkTheTabs(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	if got := press(m, "right").top; got != tabScratches {
		t.Errorf("right from Activities opens tab %d, want Scratches", got)
	}
	if got := press(m, "left").top; got != tabPlans {
		t.Errorf("left from Activities opens tab %d, want Plans", got)
	}
	walked := m
	for i := range topTabs {
		walked = press(walked, "right")
		if want := (tabActivities + 1 + i) % len(topTabs); walked.top != want {
			t.Fatalf("right %d times: tab %d, want %d", i+1, walked.top, want)
		}
	}
}

func TestTabCyclesOnlyTheOpenTabsPanes(t *testing.T) {
	m := press(sized(newModel(t), 160, 50), tabKey(tabPlans))
	for i, want := range []pane{paneDone, paneDetail, paneList} {
		m = press(m, "tab")
		if m.focus != want || m.top != tabPlans {
			t.Fatalf("tab %d times: tab %d pane %d, want Plans pane %d", i+1, m.top, m.focus, want)
		}
	}
	for i, want := range []pane{paneDetail, paneDone, paneList} {
		m = press(m, "shift+tab")
		if m.focus != want || m.top != tabPlans {
			t.Fatalf("shift+tab %d times: tab %d pane %d, want Plans pane %d", i+1, m.top, m.focus, want)
		}
	}
	a := sized(newModel(t), 160, 50)
	for i := range 4 {
		a = press(a, "tab")
		if a.focus == paneDone {
			t.Fatalf("tab %d times on Activities focused a Done pane", i+1)
		}
	}
	a.focusPane(paneDone)
	if a.focus == paneDone {
		t.Error("Activities took the focus into a Done pane it does not have")
	}
	if got := len(a.geometry().side); got != 1 {
		t.Errorf("Activities draws %d list panes, want 1", got)
	}
}

func TestDoneSubTabKeysOnlyActOnTheDonePane(t *testing.T) {
	m := press(sized(newModel(t), 160, 50), tabKey(tabBugs))
	if got := press(m, "]").done; got != 0 {
		t.Errorf("] on the List pane moved the Done sub-tab to %d", got)
	}
	m = press(m, "tab")
	if got := strings.Join(m.tabsOf(paneDone), " "); got != "Fixed Wontfix" {
		t.Errorf("the Bugs Done pane shows %q", got)
	}
	if got := strings.Join(doneRowIDs(m), " "); got != "bugs/2026-09-24-crash" {
		t.Errorf("Fixed holds %q", got)
	}
	m = press(m, "]")
	if m.done != 1 || len(doneRowIDs(m)) != 0 {
		t.Errorf("] on Done: sub-tab %d rows %q, want Wontfix and no rows", m.done, doneRowIDs(m))
	}
	if m = press(m, "["); m.done != 0 {
		t.Errorf("[ on Done left sub-tab %d, want Fixed", m.done)
	}
	if got := press(m, "0", "]").done; got != 0 {
		t.Errorf("] on the detail moved the Done sub-tab to %d", got)
	}
}

func TestFirstVisitSelectsTheTopRowAndShowsIt(t *testing.T) {
	for i := range topTabs {
		m := press(sized(newModel(t), 160, 50), tabKey(i))
		rows := m.rowsOf(paneList)
		if len(rows) == 0 {
			t.Fatalf("tab %s has no rows in the fixture, so this test proves nothing", topTabs[i].name)
		}
		it := m.Selected()
		if it == nil || it.ID != rows[0].id {
			t.Errorf("tab %s first visit selects %v, want %s", topTabs[i].name, it, rows[0].id)
			continue
		}
		if detail := strings.Join(plainLines(m.detailLines(100)), "\n"); !strings.Contains(detail, idText(it)) {
			t.Errorf("tab %s detail does not show %s:\n%s", topTabs[i].name, idText(it), detail)
		}
	}
}

func TestReturnVisitKeepsPaneRowAndTree(t *testing.T) {
	m := press(sized(newModel(t), 160, 50), tabKey(tabBugs), "j", "tab")
	m = press(m, tabKey(tabPlans), " ", tabKey(tabBugs))
	if m.focus != paneDone {
		t.Errorf("back on Bugs the focus is on pane %d, want Done", m.focus)
	}
	if got := cursorOf(m.rowsOf(paneList), m.sel[paneList], m.idx[paneList]); got != 1 {
		t.Errorf("back on Bugs the List cursor is on row %d, want 1", got)
	}
	m = press(m, tabKey(tabPlans))
	if got := len(m.rowsOf(paneList)); got != 3 {
		t.Errorf("back on Plans the list has %d rows, want the opened plan still open (3)", got)
	}
}

func TestAShrunkListClampsTheCursor(t *testing.T) {
	m := press(sized(newModel(t), 160, 50), tabKey(tabBugs), "G", tabKey(tabSpecs))
	cfg := treeCfg(t, map[string]string{".acta/bugs/2026-09-20-only.md": "# Only bug\n\n## Symptom\nx\n"})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	next, _ := m.Update(reloadMsg{b: b})
	m = press(next.(Model), tabKey(tabBugs))
	if it := m.Selected(); it == nil || it.ID != "bugs/2026-09-20-only" {
		t.Errorf("after the list shrank the cursor is on %v, want the last row left", it)
	}
}

func TestAnEmptyListShowsNoItems(t *testing.T) {
	cfg := treeCfg(t, map[string]string{".acta/bugs/2026-09-20-only.md": "# Only bug\n\n## Symptom\nx\n"})
	m := sized(detailModel(t, cfg), 160, 50)
	for _, i := range []int{tabActivities, tabScratches} {
		m = press(m, tabKey(i))
		if got := plainLines(m.detailLines(80)); len(got) != 1 || got[0] != "No items" {
			t.Errorf("tab %s with no rows shows %q, want No items", topTabs[i].name, got)
		}
	}
}

func TestActivitiesListsOnlyInProgressTasks(t *testing.T) {
	cfg := treeCfg(t, map[string]string{
		".acta/plans/2026-09-20-one.md": "# One\n\n### Task 1: A\n\n- [x] a\n- [ ] b\n",
		".acta/plans/2026-09-21-two.md": "# Two\n\n### Task 1: B\n\n- [x] a\n- [ ] b\n\n### Task 2: C\n\n- [ ] c\n",
		".acta/bugs/2026-09-22-lag.md":  "---\nstatus: fixing\n---\n# Lag\n\n## Symptom\nx\n",
	})
	m := sized(detailModel(t, cfg), 160, 50)
	got := ids(m.rowsOf(paneList))
	slices.Sort(got)
	want := []string{"plans/2026-09-20-one#task-1", "plans/2026-09-21-two#task-1"}
	if !slices.Equal(got, want) {
		t.Errorf("Activities holds %q, want the in-progress task of each plan %q", got, want)
	}
}

// TestNoDividerRow walks every pane of every tab and checks no row is a rule:
// each row is a real item or the folded legacy group, and no pane draws a
// full-width line of dashes, even where in-progress and not-started items sit
// next to each other.
func TestNoDividerRow(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	m.openPlans = everyPlanOpen(m.board)
	_, b := fixture(t)
	checked := 0
	for i := range topTabs {
		mm := press(m, tabKey(i))
		for _, p := range mm.panes() {
			mm.focusPane(p)
			rows := mm.rowsOf(p)
			if len(rows) == 0 {
				continue
			}
			checked++
			var going, waiting bool
			for _, r := range rows {
				if r.id == groupRowID {
					continue
				}
				it := b.Get(r.id)
				if it == nil {
					t.Errorf("tab %s pane %d holds a row %q that is no item", topTabs[i].name, p, r.id)
					continue
				}
				if inProgress(it) {
					going = true
				} else {
					waiting = true
				}
			}
			if !going || !waiting {
				continue
			}
			for _, line := range innerLines(mm, paneBox(mm, p)) {
				if strings.Trim(strings.TrimSpace(line), "─") == "" && strings.TrimSpace(line) != "" {
					t.Errorf("tab %s pane %d draws the divider row %q", topTabs[i].name, p, line)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no pane held a row, so this test proves nothing")
	}
}
```

Keep `TestScratchDetailNamesTheLinkedSpec` at the end of the file with its body unchanged.

In `internal/tui/plantree_test.go` move the two helpers to the new keys:

```go
// toPlans puts the focus on the open plans list, the way a reader gets there.
func toPlans(m Model) Model { return press(m, tabKey(tabPlans)) }

// toPlansDone puts the focus on the finished plans list.
func toPlansDone(m Model) Model { return press(toPlans(m), "tab") }
```

In `internal/tui/model_test.go`, teach `key` the arrow keys, next to the other cases:

```go
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
```

In `internal/tui/detail_test.go`, replace the pane walk inside `detailLines` (from `for p := pane(0); p < paneDone; p++ {` to the `t.Fatalf`) with:

```go
	// A finished item sits in a Done pane and a task sits under its plan, so
	// every tab, both panes, every Done sub-tab and every open plan is walked.
	m.openPlans = everyPlanOpen(b)
	for i := range topTabs {
		m.openTab(i)
		for _, p := range m.panes() {
			for d := range max(1, len(topTabs[i].done)) {
				m.done = d
				m.focus = p
				rows, sel, idx := m.slotOf(p)
				for _, r := range rows {
					if r.id == it.ID {
						*sel, *idx = it.ID, 0
						return m.detailLines(100)
					}
				}
			}
		}
	}
	t.Fatalf("%s is on no list of the board", id)
	return nil
```

and replace the body of `onItem` after the `it == nil` check with:

```go
	for i := range topTabs {
		m = press(m, tabKey(i))
		m.openPlans = everyPlanOpen(m.board)
		for _, p := range m.panes() {
			m.focusPane(p)
			for d := range max(1, len(topTabs[i].done)) {
				m.done = d
				for n, r := range m.rowsOf(p) {
					if r.id == it.ID {
						m.moveTo(n)
						m.focus = paneDetail
						return press(m, "g")
					}
				}
			}
		}
	}
	t.Fatalf("no pane of the board lists %s", id)
	return m
```

Every other test in `model_test.go`, `view_test.go`, `scroll_test.go`, `frame_test.go` and `trees_test.go` that names the old sidebar is moved by this table, keeping what it asserts:

| Old | New |
|---|---|
| key `1`, `keyOf(paneActive)`, `m.activeRows()` | `tabKey(tabActivities)`, `m.activityRows()`; an assert about a non-task kind in Active moves to that kind's tab |
| `onTab(t, m, paneSpecs, 0)`, key `2` | `press(m, tabKey(tabSpecs))` |
| `onTab(t, m, paneSpecs, 1)` (Scratchpad) | `press(m, tabKey(tabScratches))` |
| `onTab(t, m, panePlans, 0)`, key `3`, `keyOf(panePlans)` | `press(m, tabKey(tabPlans))` |
| `onTab(t, m, panePlans, 1)` (Tasks) | `press(m, tabKey(tabPlans), " ", "j")` to reach the first task, or `tabKey(tabActivities)` for a task in progress |
| `onTab(t, m, paneBugs, 0)`, key `4` | `press(m, tabKey(tabBugs))` |
| `onTab(t, m, paneBugs, 1)` (Debt) | `press(m, tabKey(tabDebts))` |
| key `5`, `keyOf(paneDone)` after a kind | `"tab"` from that tab's List pane |
| `m.tab[paneDone]` | `m.done` |
| `m.openRows(paneSpecs)`, `m.openRows(panePlans)` | `press(m, tabKey(tabSpecs)).openRows()`, `press(m, tabKey(tabPlans)).openRows()`; a plans assert expecting task rows opens the plans first with `m.openPlans = everyPlanOpen(m.board)` |
| `m.follows = p`, `m.tab[p] = tab` | `m.openTab(i)` |
| `for p := ...; p < paneDone` with `sidebar[p].tabs` | `for i := range topTabs` with `m.panes()` |
| `paneKeys`, `keyOf`, `paneOfKey`, `onTab` helpers | removed; use `tabKey` and `"0"` |
| `[1]─Active` and the other five titles | `─List`, `─Done ─ Dropped` (or that tab's sub-tabs), `[0]─Detail`, and the tab bar line |
| hard-coded pane `y` values | one line lower: the tab bar owns screen line 0 |

`TestGeometryPlacesThePanes` at 120 by 40 on the Plans tab expects two side boxes `{0, 1, 36, 19}` and `{0, 20, 36, 19}`, the detail box `{36, 1, 84, 38}`, and on Activities one side box `{0, 1, 36, 38}`. `TestViewShowsTheSidebarAndTheDetail` expects the tab bar line ` Scratches  Bugs  Debts  Specs  [Plans]  Activities` after `tabKey(tabPlans)`, and `─List`, `─Done ─ Dropped`, `[0]─Detail`. `TestViewHelpPopupCoversThePanes` also checks the help lists `1-6`, `← →`, `tab shift+tab`, `[ ]` and `space enter`.

- [x] **Step 2: Run the tests to see them fail**

Run: `go test ./internal/tui/ -v`
Expected: FAIL to build with `undefined: topTabs`, `undefined: tabActivities`, `m.top undefined`, `m.openTab undefined`.

- [x] **Step 3: Write the code**

`internal/tui/sidebar.go` becomes this file whole (the `treeRows` from Task 1 stays at its end, unchanged):

```go
package tui

import (
	"slices"
	"sort"

	"github.com/iyay/acta/internal/board"
)

// doneTab is one tab of the Done pane: the name in its title and the status it
// lists.
type doneTab struct {
	name   string
	status string
}

// topTab is one tab of the bar on the top line: its name, the kind it lists,
// the finished sub-tabs of its Done pane, and tree for the one tab whose lists
// fold plans open to show their tasks. Activities has no kind and no Done pane.
type topTab struct {
	name string
	kind board.Kind
	done []doneTab
	tree bool
}

// topTabs is the bar, left to right. Index i is key i+1. It is an array, so
// len(topTabs) is a constant the Model can size its saved tabs with.
var topTabs = [...]topTab{
	{name: "Scratches", kind: board.KindScratch, done: []doneTab{{"Specced", "specced"}, {"Dropped", "dropped"}}},
	{name: "Bugs", kind: board.KindBug, done: []doneTab{{"Fixed", "fixed"}, {"Wontfix", "wontfix"}}},
	{name: "Debts", kind: board.KindDebtItem, done: []doneTab{{"Done", "done"}, {"Wontfix", "wontfix"}}},
	{name: "Specs", kind: board.KindStory, done: []doneTab{{"Done", "done"}, {"Dropped", "dropped"}}},
	{name: "Plans", kind: board.KindPlan, done: []doneTab{{"Done", "done"}, {"Dropped", "dropped"}}, tree: true},
	{name: "Activities"},
}

// The tab numbers, in the order of topTabs, so code and tests name a tab
// instead of counting.
const (
	tabScratches = iota
	tabBugs
	tabDebts
	tabSpecs
	tabPlans
	tabActivities
)

const (
	// paneList is the top box of a tab: its open items.
	paneList = pane(0)
	// paneDone is the box under it: the finished items of the Done sub-tab.
	paneDone = pane(1)
	// paneDetail is the box on the right.
	paneDetail = pane(2)
)

// sidePanes is how many list boxes a tab can have, and boxes adds the detail.
const (
	sidePanes = 2
	boxes     = sidePanes + 1
)

// paneKey is what a box wears at the start of its title. The detail box keeps
// its key, since 0 still focuses it; the list boxes have no key of their own.
func paneKey(p pane) string {
	if p == paneDetail {
		return "─[0]─"
	}
	return "─"
}

// tabState is what a tab keeps while another one is open: the focused pane,
// the list the detail showed, the cursor and offset of each pane, and the
// Done sub-tab. sel is nil until the first visit.
type tabState struct {
	focus, last pane
	sel         []string
	idx, off    []int
	done        int
}

// freshTab is how a tab looks on its first visit: the List pane has the focus
// and its top row is selected, because an empty cursor falls on row 0.
func freshTab() tabState {
	return tabState{
		focus: paneList, last: paneList,
		sel: make([]string, sidePanes), idx: make([]int, sidePanes), off: make([]int, boxes),
	}
}

// openTab puts tab i on screen. The tab that was open keeps its place, so a
// later visit finds the same pane, row and Done sub-tab. A row that went away
// since is caught by cursorOf, which falls back to the last row that is left.
func (m *Model) openTab(i int) {
	if i == m.top || i < 0 || i >= len(topTabs) {
		return
	}
	m.tabs[m.top] = tabState{m.focus, m.last, m.sel, m.idx, m.off, m.done}
	s := m.tabs[i]
	if s.sel == nil {
		s = freshTab()
	}
	m.top = i
	m.focus, m.last, m.sel, m.idx, m.off, m.done = s.focus, s.last, s.sel, s.idx, s.off, s.done
	// The room and the detail place belonged to the tab that is gone now.
	m.expanded = -1
	m.off[paneDetail] = 0
	m.keepVisible(m.listPane())
}

// panes gives the list boxes of the open tab, top to bottom. Activities has
// no Done pane.
func (m Model) panes() []pane {
	if len(topTabs[m.top].done) == 0 {
		return []pane{paneList}
	}
	return []pane{paneList, paneDone}
}

// cyclePane walks tab and shift+tab over the boxes of the open tab and the
// detail box, and wraps around, so the focus never lands in another tab.
func (m *Model) cyclePane(step int) {
	ring := append(m.panes(), paneDetail)
	at := slices.Index(ring, m.focus)
	m.focusPane(ring[(at+step+len(ring))%len(ring)])
}

// doneTabOf gives the Done sub-tab the open tab shows, and false on a tab with
// no Done pane.
func (m Model) doneTabOf() (doneTab, bool) {
	done := topTabs[m.top].done
	if len(done) == 0 {
		return doneTab{}, false
	}
	return done[clamp(m.done, 0, len(done)-1)], true
}

// doneTabNames gives the sub-tab names the Done pane of the open tab shows.
func (m Model) doneTabNames() []string {
	var out []string
	for _, d := range topTabs[m.top].done {
		out = append(out, d.name)
	}
	return out
}

// tabsOf gives the names a box draws in its title.
func (m Model) tabsOf(p pane) []string {
	switch p {
	case paneList:
		return []string{"List"}
	case paneDone:
		return m.doneTabNames()
	}
	return nil
}

// onTab gives the name a box has open, so its title never drops that name
// even when the box is too narrow to draw every name.
func (m Model) onTab(p pane) int {
	if p == paneDone {
		return clamp(m.done, 0, max(0, len(m.doneTabNames())-1))
	}
	return 0
}

// rowsOf gives the rows a box shows: the finished items in Done, the tasks
// under way on Activities, and the open items of the tab anywhere else.
func (m Model) rowsOf(p pane) []row {
	switch {
	case p == paneDone:
		return m.doneRows()
	case topTabs[m.top].kind == "":
		return m.activityRows()
	}
	return m.openRows()
}

// searchRows gives a searching reader every item whose text matches, whatever
// box asked, so one search never depends on which box has the focus.
func (m Model) searchRows() []row {
	return toRows(m.board.Search(m.query), 0)
}

// activityRows gives Activities every task whose work has begun, of every
// plan, in the order the board lists them. inProgress is the one rule that
// decides.
func (m Model) activityRows() []row {
	if m.query != "" {
		return m.searchRows()
	}
	var going []*board.Item
	for _, it := range m.board.List(board.KindTask, true) {
		if inProgress(it) {
			going = append(going, it)
		}
	}
	return toRows(going, 0)
}

// openRows gives the List box the items of the open tab that are not
// finished, in the order the board lists them. Plans come as a tree. The
// files outside the root folder have no tab of their own, so the Specs tab
// ends with a single row that opens them.
func (m Model) openRows() []row {
	if m.query != "" {
		return m.searchRows()
	}
	tab := topTabs[m.top]
	items := m.board.List(tab.kind, false)
	if tab.tree {
		return m.treeRows(items)
	}
	rows := toRows(items, 0)
	if tab.kind != board.KindStory {
		return rows
	}
	legacy := m.board.Untyped(false)
	if len(legacy) == 0 {
		return rows
	}
	rows = append(rows, row{id: groupRowID, group: true})
	if m.groupOpen {
		rows = append(rows, toRows(legacy, 1)...)
	}
	return rows
}

// toggleExpand gives the focused list box the room of the whole column, or
// takes the room back when it already has it. The detail box sits outside the
// column, so z there does nothing. The box that grows shows more rows, so its
// offset goes back inside what it can really show.
func (m *Model) toggleExpand() {
	if m.focus == paneDetail {
		return
	}
	if m.expanded == int(m.focus) {
		m.expanded = -1
	} else {
		m.expanded = int(m.focus)
	}
	m.clampOff(m.focus)
}

// doneRows gives the Done box the finished items of the open tab's sub-tab,
// the ones touched last at the top: the date of the last commit on the file
// decides, and the date in the file name is the fallback. Plans come as a
// tree.
func (m Model) doneRows() []row {
	d, ok := m.doneTabOf()
	if !ok {
		return nil
	}
	tab := topTabs[m.top]
	var finished []*board.Item
	for _, it := range m.board.List(tab.kind, true) {
		if it.Status == d.status {
			finished = append(finished, it)
		}
	}
	sort.SliceStable(finished, func(i, j int) bool {
		return finished[i].SortTime() > finished[j].SortTime()
	})
	if tab.tree {
		return m.treeRows(finished)
	}
	return toRows(finished, 0)
}
```

`internal/tui/model.go`:

Replace the `follows` and `tab` fields with:

```go
	top       int                     // the open tab, an index into topTabs
	tabs      [len(topTabs)]tabState  // the saved place of each tab that is not open
	done      int                     // the Done sub-tab the open tab shows
```

In `New`, replace the `tab:`, `sel:`, `idx:` and `off:` lines so the model opens on Activities with a fresh place:

```go
func New(cfg config.Config, b *board.Board, dark bool) Model {
	s := freshTab()
	return Model{
		cfg: cfg, board: b, width: 120, height: 40, now: time.Now(), version: "dev",
		open:     defaultOpen,
		top:      tabActivities,
		focus:    s.focus,
		last:     s.last,
		sel:      s.sel,
		idx:      s.idx,
		off:      s.off,
		expanded: -1,
```

(the `load`, `setValue` and `render` lines stay as they are).

`focusPane` drops the `follows` lines and refuses a Done pane the tab does not have:

```go
// focusPane moves the focus. The detail box shows the list box that had it
// last. A tab with no Done pane never gives it the focus.
func (m *Model) focusPane(p pane) {
	if m.focus == p || !slices.Contains(append(m.panes(), paneDetail), p) {
		return
	}
	// The room belongs to the box that has the focus, so a box that is no
	// longer the one on top of the screen cannot keep it.
	m.expanded = -1
	m.focus = p
	if p != paneDetail {
		m.last = p
	}
	m.keepVisible(p)
}
```

(add `"slices"` to the imports).

`cycleTab` only works on the Done pane:

```go
// cycleTab walks the Done pane along its sub-tabs and wraps around. The other
// boxes have no sub-tabs, so [ and ] do nothing there.
func (m *Model) cycleTab(step int) {
	n := len(m.doneTabNames())
	if m.focus != paneDone || n == 0 {
		return
	}
	m.done = (m.done + step + n) % n
	m.keepVisible(paneDone)
}
```

The key switch cases for moving around become:

```go
	case "0":
		m.focusPane(paneDetail)
	case "1", "2", "3", "4", "5", "6":
		m.openTab(int(k.String()[0] - '1'))
	case "left":
		m.openTab((m.top + len(topTabs) - 1) % len(topTabs))
	case "right":
		m.openTab((m.top + 1) % len(topTabs))
	case "tab":
		m.cyclePane(1)
	case "shift+tab":
		m.cyclePane(-1)
```

`clickTab`:

```go
// clickTab switches the Done box to the sub-tab the click landed on. The
// other boxes hold one name, so a click there only takes the focus.
func (m *Model) clickTab(p pane, idx int) {
	if p == paneDone {
		m.done = clamp(idx, 0, max(0, len(m.doneTabNames())-1))
	}
	m.keepVisible(p)
}
```

`internal/tui/frame.go`, the tab bar owns the top line:

```go
func (m Model) geometry() geom {
	// The top line is the tab bar and the last line the status line, so the
	// boxes share the rest. A terminal with fewer rows gets the rows it has.
	bodyH := max(0, m.height-2)
	panes := m.panes()
	g := geom{wide: m.width >= 60, leftW: clamp(m.width*3/10, 28, 48), side: make([]box, len(panes))}
	y := 1
	for p, h := range m.leftHeights(bodyH) {
		g.side[p] = m.box(pane(p), 0, y, g.leftW, h)
		y += h
	}
	g.detail = m.box(paneDetail, g.leftW, 1, max(0, m.width-g.leftW), bodyH)
	if g.wide {
		return g
	}
	g.full = m.box(m.focus, 0, 1, m.width, bodyH)
```

(the rest of `geometry` stays). In `leftHeights`, count the open tab's boxes instead of the old table:

```go
func (m Model) leftHeights(h int) []int {
	n := len(m.panes())
	others := n - 1
	if m.expanded < 0 || m.expanded >= n {
		return split(h, n)
	}
	out := make([]int, n)
```

(the rest stays).

`internal/tui/scroll.go`, `boxOf` measures an off-screen box with the same body height:

```go
		return m.box(p, 0, 1, m.width, max(0, m.height-2))
```

`internal/tui/view.go`, add `tabBar` below `View`:

```go
// tabBar draws the top line: every tab name in order, the open one in
// brackets and the accent, so the reader always sees where they are.
func (m Model) tabBar() string {
	names := make([]string, len(topTabs))
	for i, t := range topTabs {
		names[i] = faint.Render(t.name)
		if i == m.top {
			names[i] = accent.Bold(true).Render("[" + t.name + "]")
		}
	}
	return " " + strings.Join(names, "  ")
}
```

In `View`, put the bar above the boxes before the lines are cut and fitted:

```go
	lines := append([]string{m.tabBar()}, strings.Split(body, "\n")...)
```

and replace `helpLines`:

```go
const helpLines = `1-6 ← →          open a tab, previous / next tab
tab shift+tab    move between the panes of the tab
0                focus the detail
[ ]              switch the Done tab, on the Done pane
space enter      open or shut a plan row
z                expand the focused pane
j k g G          move a list, scroll the detail
ctrl+d ctrl+u    page down and up
enter            focus the detail on the row
e                open the row in the editor
t s n            set a value, new bug
esc              back to the list, close this help
/ r q            search, reload, quit
?                close this help`
```

`internal/tui/detail.go`, in `detailLines`, an empty list says so:

```go
	if it == nil {
		if len(m.board.Items) == 0 {
			return cut("this repo has no .acta/ yet.\n\npress n to write the first bug, or let the agent plugin create specs and plans.", w)
		}
		if len(m.listOf()) == 0 {
			return []string{faint.Render("No items")}
		}
		return []string{faint.Render("enter opens the group")}
	}
```

- [x] **Step 4: Run the tests to see them pass**

Run: `go test ./internal/tui/ -v`
Expected: PASS, the whole package, including every test moved by the table above.

- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui
git commit -m "feat(tui): top tab bar with one tab per kind"
```

## Fix round 1

### Task 3: The tab bar always shows the open tab

Review round 1 (range da82e5b..b1b2c57) found one BLOCKER. `tabBar()` (`internal/tui/view.go`) builds one line with every tab name, and `pad` / `fit` cut it from the right. On a terminal narrower than the bar (about 51 columns), the open tab can be cut away: at 40 columns on the start tab Activities the top line reads ` Scratches  Bugs  Debts  Specs  Plans  [`, and pane titles only say `List`, so the reader cannot tell where they are. At da82e5b the focused pane title always kept its open name below 60 columns.

**Files:** `internal/tui/view.go`, `internal/tui/view_test.go`

**verify:** At every width from 1 to 200 columns and for every open tab, the top line holds the open tab's name in brackets whole, or as much of it as the width allows when even the name alone does not fit, and it never shows a cut name of another tab. When the bar does not fit, names are dropped from the side farthest from the open tab first, the same way `dropOrder` in `internal/tui/title.go` drops pane-title parts. List every width band checked (fits, drops some names, only the open name, narrower than the open name) for each tab.

- [x] **Step 1: Write the failing test** in `internal/tui/view_test.go`: for each of the 6 tabs and each width 1..200, open the tab, set the width, render, take the first line with ANSI stripped, and assert it contains `[<name>]` when the width is at least len(name)+3, and that every other name on the line appears whole.
- [x] **Step 2: Run** `go test ./internal/tui/ -run TestTabBar -v`. Expected: FAIL at 40 columns on Activities.
- [x] **Step 3: Implement** in `tabBar(width int)`: start from all names; while the joined width is over the screen width, drop the name farthest from the open tab (ties: drop the right one); when only the open name is left and it still does not fit, let `fit` cut it. Pass `m.width` from the caller.
- [x] **Step 4: Run** `go test ./...`. Expected: PASS.
- [x] **Step 5: Gates and commit**

```bash
gofmt -l . && go vet ./... && go test ./...
git add internal/tui/view.go internal/tui/view_test.go .acta/plans/2026-09-28-tui-top-tabs.md
git commit -m "fix(tui): keep the open tab in a narrow tab bar"
```
