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
	t.Parallel()

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
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	if m.top != tabActivities || m.focus != paneList {
		t.Fatalf("opens on tab %d pane %d, want Activities and its List", m.top, m.focus)
	}
	if it := m.Selected(); it == nil || it.ID != "plans/2026-09-21-alpha#task-2" {
		t.Errorf("opens on %v, want the one task in progress", it)
	}
	bar := plain(strings.Split(m.View(), "\n")[1])
	if !strings.HasPrefix(bar, "│ 1 Scratches  2 Bugs  3 Debts  4 Specs  5 Plans  6 Activities") {
		t.Errorf("the names line is %q, want the tab box with Activities marked", bar)
	}
}

func TestNumberKeysOpenTheirTab(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	for i := range topTabs {
		walked = press(walked, "left")
		if want := (tabActivities - 1 - i + 2*len(topTabs)) % len(topTabs); walked.top != want {
			t.Fatalf("left %d times: tab %d, want %d", i+1, walked.top, want)
		}
	}
}

func TestTabCyclesOnlyTheOpenTabsPanes(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

// TestListPaneTitleSaysOpenOrTasks walks every tab, so no tab can keep the
// old word and every tab with a kind says the same word.
func TestListPaneTitleSaysOpenOrTasks(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		m := press(sized(newModel(t), 160, 40), tabKey(i))
		want := "Open"
		if topTabs[i].kind == "" {
			want = "Tasks"
		}
		if got := m.tabsOf(paneList); len(got) != 1 || got[0] != want {
			t.Errorf("tab %s: title %q, want %q", topTabs[i].name, got, want)
		}
		top := plain(m.paneTop(paneList, m.geometry().side[paneList], m.edge(paneList)))
		if !strings.Contains(top, want) || strings.Contains(top, "List") {
			t.Errorf("tab %s: top border %q, want %q and no List", topTabs[i].name, top, want)
		}
	}
}

func TestFirstVisitSelectsTheTopRowAndShowsIt(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	m := press(sized(newModel(t), 160, 50), tabKey(tabBugs), "j", "tab")
	// Oldest first puts alpha before lonely, so step down to lonely before
	// opening it: lonely has one task, alpha has two.
	m = press(m, tabKey(tabPlans), "j", " ", tabKey(tabBugs))
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

// TestAShrunkListClampsTheCursor shrinks the list under the cursor from
// three rows to two. The cursor has to land on the last row that is left,
// which is neither the row it was on nor the first row of the new list.
func TestAShrunkListClampsTheCursor(t *testing.T) {
	t.Parallel()

	bugs := press(sized(newModel(t), 160, 50), tabKey(tabBugs), "G")
	if got := cursorOf(bugs.rowsOf(paneList), bugs.sel[paneList], bugs.idx[paneList]); got != 2 {
		t.Fatalf("before the reload the cursor is on row %d, want the last of the 3 bugs", got)
	}
	m := press(bugs, tabKey(tabSpecs))
	// Oldest first puts lag last, so the reload has to drop lag itself (not
	// the oldest bug) to make the cursor actually clamp to a new row.
	cfg := treeCfg(t, map[string]string{
		".acta/specs/2026-09-15-really-bug.md": "---\ntype: bug\n---\n# Really a bug\n",
		".acta/bugs/2026-09-26-open.md":        "# Open\n\n## Symptom\nx\n",
	})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	next, _ := m.Update(reloadMsg{b: b})
	m = press(next.(Model), tabKey(tabBugs))
	if got := len(m.rowsOf(paneList)); got != 2 {
		t.Fatalf("the reloaded list has %d rows, want the 2 bugs left", got)
	}
	if got := cursorOf(m.rowsOf(paneList), m.sel[paneList], m.idx[paneList]); got != 1 {
		t.Errorf("after the list shrank the cursor is on row %d, want the last row left (1)", got)
	}
	if it := m.Selected(); it == nil || it.ID != "bugs/2026-09-26-open" {
		t.Errorf("after the list shrank the cursor is on %v, want the last row left", it)
	}
}

func TestAnEmptyListShowsNoItems(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	// Every task in progress of the real fixture is on show, oldest file date
	// first, and nothing that is only open, raw, todo or finished sneaks in.
	full := sized(newModel(t), 160, 50)
	var going []*board.Item
	for _, it := range full.board.Items {
		if it.Kind == board.KindTask && inProgress(it) {
			going = append(going, it)
		}
	}
	if len(going) == 0 {
		t.Fatal("the fixture holds no task in progress, so this test proves nothing")
	}
	var ids_ []string
	for _, it := range ordered(going, false) {
		ids_ = append(ids_, it.ID)
	}
	listed := ids(full.rowsOf(paneList))
	if !slices.Equal(listed, ids_) {
		t.Errorf("Activities holds %q, want %q", listed, ids_)
	}
}

// TestNoDividerRow walks every pane of every tab and checks no row is a rule:
// each row is a real item or the folded legacy group, and no pane draws a
// full-width line of dashes, even where in-progress and not-started items sit
// next to each other.
func TestNoDividerRow(t *testing.T) {
	t.Parallel()

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

// TestScratchDetailNamesTheLinkedSpec checks the detail of a scratch item
// names the spec that came out of it, the way a plan names the spec it
// carries. The first board gives the spec a number ID, so the line reads
// SPC-0009; the shared fixture's linked spec has none, so the same line names
// it by path there.
func TestScratchDetailNamesTheLinkedSpec(t *testing.T) {
	t.Parallel()

	cfg := treeCfg(t, map[string]string{
		".acta/scratch/2026-09-28-idea.md":        "# Themes\n\nCatet aja dulu.\n",
		".acta/specs/2026-09-28-themes-design.md": "---\nid: SPEC-9\nparent: scratch/2026-09-28-idea\n---\n# Themes design\n\nThe spec the idea became.\n",
	})
	lines := plainLines(detailLines(t, cfg, "scratch/2026-09-28-idea"))
	if !strings.Contains(strings.Join(lines, "\n"), "SPC-0009") {
		t.Errorf("the detail of a scratch item does not name the spec it produced:\n%s", strings.Join(lines, "\n"))
	}
	fixtureCfg, _ := fixture(t)
	used := plainLines(detailLines(t, fixtureCfg, "scratch/2026-09-28-idea-used"))
	if !strings.Contains(strings.Join(used, "\n"), "specs/2026-09-28-from-scratch-design") {
		t.Errorf("the detail of the used idea does not name its spec:\n%s", strings.Join(used, "\n"))
	}
}
