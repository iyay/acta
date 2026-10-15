package tui

import (
	"fmt"
	"maps"
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
		{"Scratchpad", board.KindScratch, "Specced Dropped", false},
		{"Bugs", board.KindBug, "Fixed Wontfix", false},
		{"Debt", board.KindDebtItem, "Done Wontfix", false},
		{"Specs", board.KindStory, "Done Dropped", false},
		{"Plans", board.KindPlan, "Done Dropped", true},
		{"Activity", "", "", false},
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
	if topTabs[tabScratchpad].name != "Scratchpad" || topTabs[tabPlans].name != "Plans" || topTabs[tabActivity].name != "Activity" {
		t.Error("the tab constants do not match the table")
	}
}

func TestTUIOpensOnActivity(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	if m.top != tabActivity || m.focus != paneList {
		t.Fatalf("opens on tab %d pane %d, want Activity and its List", m.top, m.focus)
	}
	// Activity opens on the head of the group, and the task under way sits
	// one row below it.
	if it := m.Selected(); it == nil || it.ID != "plans/2026-09-21-alpha" || it.Kind != board.KindPlan {
		t.Errorf("opens on %v, want the plan the task under way belongs to", it)
	}
	if it := press(m, "j").Selected(); it == nil || it.ID != "plans/2026-09-21-alpha#task-2" || it.Kind != board.KindTask {
		t.Errorf("one row down is %v, want the one task in progress", it)
	}
	bar := plain(strings.Split(m.View(), "\n")[1])
	if !strings.HasPrefix(bar, "│ 1 Scratchpad  2 Bugs  3 Debt  4 Specs  5 Plans  6 Activity") {
		t.Errorf("the names line is %q, want the tab box with Activity marked", bar)
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
	if got := press(m, "7").top; got != tabActivity {
		t.Errorf("7 is no tab but moved to tab %d", got)
	}
}

func TestLeftRightWalkTheTabs(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 160, 50)
	if got := press(m, "right").top; got != tabScratchpad {
		t.Errorf("right from Activity opens tab %d, want Scratchpad", got)
	}
	if got := press(m, "left").top; got != tabPlans {
		t.Errorf("left from Activity opens tab %d, want Plans", got)
	}
	walked := m
	for i := range topTabs {
		walked = press(walked, "right")
		if want := (tabActivity + 1 + i) % len(topTabs); walked.top != want {
			t.Fatalf("right %d times: tab %d, want %d", i+1, walked.top, want)
		}
	}
	for i := range topTabs {
		walked = press(walked, "left")
		if want := (tabActivity - 1 - i + 2*len(topTabs)) % len(topTabs); walked.top != want {
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
			t.Fatalf("tab %d times on Activity focused a Done pane", i+1)
		}
	}
	a.focusPane(paneDone)
	if a.focus == paneDone {
		t.Error("Activity took the focus into a Done pane it does not have")
	}
	if got := len(a.geometry().side); got != 1 {
		t.Errorf("Activity draws %d list panes, want 1", got)
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
	if got := press(m, "tab", "]").done; got != 0 {
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
	for _, i := range []int{tabActivity, tabScratchpad} {
		m = press(m, tabKey(i))
		if got := plainLines(m.detailLines(80)); len(got) != 1 || got[0] != "No items" {
			t.Errorf("tab %s with no rows shows %q, want No items", topTabs[i].name, got)
		}
	}
}

// activityTaskIDs gives the ids of the Activity rows that are tasks, so a
// test reads the work under way and leaves the group heads out.
func activityTaskIDs(m Model) []string {
	var out []string
	for _, r := range m.rowsOf(paneList) {
		if m.board.Get(r.id).Kind == board.KindTask {
			out = append(out, r.id)
		}
	}
	return out
}

func TestActivityListsOnlyInProgressTasks(t *testing.T) {
	t.Parallel()

	cfg := treeCfg(t, map[string]string{
		".acta/plans/2026-09-20-one.md": "# One\n\n### Task 1: A\n\n- [x] a\n- [ ] b\n",
		".acta/plans/2026-09-21-two.md": "# Two\n\n### Task 1: B\n\n- [x] a\n- [ ] b\n\n### Task 2: C\n\n- [ ] c\n",
		".acta/bugs/2026-09-22-lag.md":  "---\nstatus: fixing\n---\n# Lag\n\n## Symptom\nx\n",
	})
	m := sized(detailModel(t, cfg), 160, 50)
	got := activityTaskIDs(m)
	slices.Sort(got)
	want := []string{"plans/2026-09-20-one#task-1", "plans/2026-09-21-two#task-1"}
	if !slices.Equal(got, want) {
		t.Errorf("Activity holds %q, want the in-progress task of each plan %q", got, want)
	}
	// Every top level row of the list is a head the tasks hang under, so a
	// task never stands on its own here.
	for _, r := range m.rowsOf(paneList) {
		if r.depth == 0 && m.board.Get(r.id).Kind == board.KindTask {
			t.Errorf("the task %s stands at the top level, want a head", r.id)
		}
	}
	// Every task in progress of the real fixture is on show and nothing that
	// is only open, raw, todo or finished sneaks in.
	full := sized(newModel(t), 160, 50)
	var wantAll []string
	for _, it := range full.board.Items {
		if it.Kind == board.KindTask && inProgress(it) {
			wantAll = append(wantAll, it.ID)
		}
	}
	if len(wantAll) == 0 {
		t.Fatal("the fixture holds no task in progress, so this test proves nothing")
	}
	slices.Sort(wantAll)
	listed := activityTaskIDs(full)
	slices.Sort(listed)
	if !slices.Equal(listed, wantAll) {
		t.Errorf("Activity holds %q, want %q", listed, wantAll)
	}
}

// TestActivityDropsStartedDoneTask checks a task that was started and then
// finished is not under way. The started record stays on disk for good, so
// the status has to win over it.
func TestActivityDropsStartedDoneTask(t *testing.T) {
	t.Parallel()

	for _, status := range []string{"done", "fixed"} {
		cfg := treeCfg(t, map[string]string{
			".acta/plans/2026-09-20-one.md": "# One\n\n### Task 1: A\n\n- [ ] a\n",
		})
		m := sized(detailModel(t, cfg), 160, 50)
		id := "plans/2026-09-20-one#task-1"
		task := m.board.Get(id)
		if task == nil {
			t.Fatalf("no task %s on the board", id)
		}
		task.Started = true
		task.Status = status
		if inProgress(task) {
			t.Errorf("a started task at %s counts as under way", status)
		}
		if got := activityTaskIDs(m); slices.Contains(got, id) {
			t.Errorf("Activity still lists the %s task %s", status, id)
		}
		for _, r := range m.rowsOf(paneList) {
			if r.depth == 0 && r.id == task.PlanID {
				t.Errorf("Activity still shows the plan head of the %s task", status)
			}
		}
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

// A search narrows the box the reader is in and nothing else: every row it
// shows is a row that box shows with no query, and each one either matches or
// is a head with a match under it. This holds on every tab, in both boxes,
// with every plan opened so the trees have children to search.
func TestSearchStaysInActivePane(t *testing.T) {
	t.Parallel()

	base := newModel(t)
	base.openPlans = everyPlanOpen(base.board)
	base.groupOpen = true
	for top := range topTabs {
		for _, p := range []pane{paneList, paneDone} {
			for _, q := range []string{"crash", "a", "e", "zzzq"} {
				m := base
				m.top, m.focus = top, p
				plain := m.rowsOf(p)
				m.query = q
				got := m.rowsOf(p)
				where := fmt.Sprintf("tab %s pane %d query %q", topTabs[top].name, p, q)
				checkSearchRows(t, m, where, plain, got, q)
			}
		}
	}
}

func checkSearchRows(t *testing.T, m Model, where string, plain, got []row, q string) {
	t.Helper()
	// got must be plain with some rows left out, in the same order.
	j := 0
	for _, r := range got {
		for j < len(plain) && plain[j] != r {
			j++
		}
		if j == len(plain) {
			t.Errorf("%s: row %s is not a row of this box", where, r.id)
			return
		}
		j++
	}
	matches := func(r row) bool {
		it := m.board.Get(r.id)
		return it != nil && strings.Contains(strings.ToLower(it.Title+"\x00"+it.Ref+"\x00"+it.Slug), q)
	}
	for i, r := range got {
		if matches(r) {
			continue
		}
		if i+1 < len(got) && got[i+1].depth > r.depth {
			continue
		}
		t.Errorf("%s: row %s neither matches nor heads a match", where, r.id)
	}
}

// Plans start shut, so a search has to look inside them: a task that matches
// shows up under its plan even when the reader never opened that plan.
func TestSearchFindsTaskUnderShutPlan(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabPlans), "/", "second step")
	got := m.rowsOf(paneList)
	for i, r := range got {
		it := m.board.Get(r.id)
		if it == nil || it.Kind != board.KindTask || !strings.Contains(strings.ToLower(it.Title), "second step") {
			continue
		}
		if i == 0 || got[i-1].id != it.PlanID {
			t.Fatalf("task %s is not under its plan head: %v", r.id, ids(got))
		}
		return
	}
	t.Fatalf("the matching task is missing: %v", ids(got))
}

// The same word lives on two tabs. Each tab finds only its own item.
func TestSearchKeepsToItsTab(t *testing.T) {
	t.Parallel()

	bugs := ids(press(newModel(t), tabKey(tabSpecs), "/", "alpha").rowsOf(paneList))
	if !slices.Contains(bugs, "specs/2026-09-20-alpha") || slices.Contains(bugs, "plans/2026-09-21-alpha") {
		t.Errorf("Specs tab search: %v", bugs)
	}
	plans := ids(press(newModel(t), tabKey(tabPlans), "/", "alpha").rowsOf(paneList))
	if !slices.Contains(plans, "plans/2026-09-21-alpha") || slices.Contains(plans, "specs/2026-09-20-alpha") {
		t.Errorf("Plans tab search: %v", plans)
	}
}

// A search shows every head open, so the fold keys must not change a fold
// the reader cannot see. After esc the tree is the way it was.
func TestFoldKeysDuringSearchKeepTheFolds(t *testing.T) {
	t.Parallel()

	for _, keys := range [][]string{{" "}, {"h"}, {"l"}, {"j", "h"}} {
		m := press(newModel(t), tabKey(tabPlans), "l", "/", "step", "enter")
		before := maps.Clone(m.openPlans)
		m = press(press(m, keys...), "esc")
		if !maps.Equal(m.openPlans, before) {
			t.Errorf("keys %q during a search changed the plan folds: %v, want %v", keys, m.openPlans, before)
		}
	}
	// l only opens, so it needs a plan that starts shut to show a change.
	shut := press(newModel(t), tabKey(tabPlans), "/", "step", "enter", "l", "esc")
	if len(shut.openPlans) != 0 {
		t.Errorf("l on a shut plan during a search opened it: %v", shut.openPlans)
	}
	a := press(newModel(t), tabKey(tabActivity), "/", "step", "enter", " ", "esc")
	if len(a.shutActs) != 0 {
		t.Errorf("space on an Activity head during a search shut it: %v", a.shutActs)
	}
}

// The legacy group lists its files during a search, so its arrow says open
// and enter on it leaves the fold alone.
func TestSearchOpensTheLegacyGroup(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs), "/", "old story", "enter")
	got := m.rowsOf(paneList)
	if !slices.Contains(ids(got), "docs/superpowers/specs/2026-01-01-old") {
		t.Fatalf("the legacy file is missing: %v", ids(got))
	}
	if txt := m.rowText(row{id: groupRowID, group: true}, nil, 40); !strings.HasPrefix(txt, "▾") {
		t.Errorf("the group arrow reads %q during a search", txt)
	}
	for i, r := range got {
		if r.group {
			m.moveTo(i)
		}
	}
	if m = press(m, "enter"); m.groupOpen {
		t.Error("enter on the group row during a search opened the fold")
	}
}

// A search looks inside an Activity head the reader shut.
func TestSearchLooksInsideShutActivityHead(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabActivity), " ", "/", "second step")
	if !slices.Contains(ids(m.rowsOf(paneList)), "plans/2026-09-21-alpha#task-2") {
		t.Fatalf("the task under the shut head is missing: %v", ids(m.rowsOf(paneList)))
	}
}

// A query of spaces is no search: no fold opens and no row goes.
func TestBlankQueryIsNoSearch(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabPlans))
	want := ids(m.rowsOf(paneList))
	if got := ids(press(m, "/", " ").rowsOf(paneList)); !slices.Equal(got, want) {
		t.Errorf("a blank query changed the rows: %v, want %v", got, want)
	}
}

// A search narrows the box that has the focus. The other box draws as if no
// search ran: the same rows, with the same plans open or shut.
func TestSearchLeavesTheOtherBoxAlone(t *testing.T) {
	t.Parallel()

	base := newModel(t)
	for top := range topTabs {
		if len(topTabs[top].done) == 0 {
			continue
		}
		for _, focus := range []pane{paneList, paneDone} {
			other := paneDone
			if focus == paneDone {
				other = paneList
			}
			m := base
			m.top, m.focus = top, focus
			want := ids(m.rowsOf(other))
			m.query = "a"
			if got := ids(m.rowsOf(other)); !slices.Equal(got, want) {
				t.Errorf("tab %s focus %d: the other box shows %v, want %v", topTabs[top].name, focus, got, want)
			}
		}
	}
}

// The box without the focus draws the same lines with or without a query,
// fold marks and the group arrow included.
func TestSearchDrawsTheOtherBoxPlain(t *testing.T) {
	t.Parallel()

	base := sized(newModel(t), 160, 50)
	for top := range topTabs {
		if len(topTabs[top].done) == 0 {
			continue
		}
		for _, focus := range []pane{paneList, paneDone} {
			other := paneDone
			if focus == paneDone {
				other = paneList
			}
			m := base
			m.top, m.focus, m.last = top, focus, focus
			b := m.geometry().at(other)
			want := m.listView(other, b.w, b)
			m.query = "a"
			if got := m.listView(other, b.w, b); !slices.Equal(got, want) {
				t.Errorf("tab %s focus %d: the other box draws\n%q\nwant\n%q", topTabs[top].name, focus, got, want)
			}
		}
	}
}

// Moving the focus to the other list box ends the search, by key or by
// click, so a click lands on the row the reader saw.
func TestFocusChangeEndsTheSearch(t *testing.T) {
	t.Parallel()

	m := press(sized(newModel(t), 160, 50), tabKey(tabPlans), "/", "dotted", "enter")
	if m.query != "dotted" {
		t.Fatalf("query %q, want dotted", m.query)
	}
	if got := press(m, "tab"); got.query != "" || got.searching {
		t.Errorf("tab to Done kept the search: query %q searching %v", got.query, got.searching)
	}
	if got := press(m, "enter"); got.query != "dotted" {
		t.Errorf("moving to the detail box dropped the query: %q", got.query)
	}
	drawn := m.rowsOf(paneDone)
	if len(drawn) == 0 {
		t.Fatal("the Done box has no rows to click")
	}
	b := m.geometry().at(paneDone)
	got := click(m, b.x+2, b.y+1)
	if got.focus != paneDone || got.query != "" {
		t.Fatalf("click on Done: focus %d query %q", got.focus, got.query)
	}
	if id := got.rowsOf(paneDone)[got.cursor()].id; id != drawn[0].id {
		t.Errorf("click on the first Done row selected %s, want %s", id, drawn[0].id)
	}
}

// A tab that opens on the other list box ends the search, the same as moving
// the focus there. A tab that opens on the same box keeps it.
func TestTabSwitchKeepsSearchOnlyInTheSameBox(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs), "tab", tabKey(tabPlans), "/", "a", "enter")
	if got := press(m, tabKey(tabSpecs)); got.query != "" {
		t.Errorf("Specs opened on Done and kept the query %q", got.query)
	}
	if got := press(m, tabKey(tabBugs)); got.query != "a" {
		t.Errorf("Bugs opened on List and lost the query: %q", got.query)
	}
}
