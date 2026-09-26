package tui

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"pm-board/internal/board"
	"pm-board/internal/config"
	"pm-board/internal/write"
)

func fixture(t *testing.T) (config.Config, *board.Board) {
	t.Helper()
	dir, err := filepath.Abs("../board/testdata/basic")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default(dir)
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return cfg, b
}

func newModel(t *testing.T) Model {
	t.Helper()
	cfg, b := fixture(t)
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return m
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		next, _ := m.Update(key(k))
		m = next.(Model)
	}
	return m
}

// click puts the left mouse button down on a cell of the screen.
func click(m Model, x, y int) Model {
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	return next.(Model)
}

// wheel turns the scroll wheel over a cell of the screen. up is true for a
// notch away from the user.
func wheel(m Model, x, y int, up bool) Model {
	button := tea.MouseButtonWheelDown
	if up {
		button = tea.MouseButtonWheelUp
	}
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: button})
	return next.(Model)
}

func ids(rows []row) []string {
	var out []string
	for _, r := range rows {
		out = append(out, r.id)
	}
	return out
}

func rowIDs(m Model) []string { return ids(m.openRows()) }

func doneRowIDs(m Model) []string { return ids(m.doneRows()) }

func TestTabKeyFocusesTheNextPane(t *testing.T) {
	m := newModel(t)
	if m.focus != paneOpen {
		t.Fatalf("the screen starts on pane %d", m.focus)
	}
	for _, want := range []pane{paneDone, paneDetail, paneOpen} {
		m = press(m, "tab")
		if m.focus != want {
			t.Fatalf("tab gave focus %d, want %d", m.focus, want)
		}
	}
}

func TestShiftTabFocusesThePreviousPane(t *testing.T) {
	m := newModel(t)
	for _, want := range []pane{paneDetail, paneDone, paneOpen} {
		m = press(m, "shift+tab")
		if m.focus != want {
			t.Fatalf("shift+tab gave focus %d, want %d", m.focus, want)
		}
	}
}

func TestNumberKeysFocusThatPane(t *testing.T) {
	m := newModel(t)
	for _, s := range []string{"1", "2", "3"} {
		m = press(m, s)
		if m.focus != pane(s[0]-'1') {
			t.Fatalf("%s gave focus %d", s, m.focus)
		}
	}
}

func TestOpenTabsCycleAndWrap(t *testing.T) {
	m := newModel(t)
	for _, want := range []int{tabPlans, tabTasks, tabBugs, tabSpecs} {
		m = press(m, "]")
		if m.tab != want {
			t.Fatalf("] gave tab %d, want %d", m.tab, want)
		}
	}
	for _, want := range []int{tabBugs, tabTasks, tabPlans, tabSpecs} {
		m = press(m, "[")
		if m.tab != want {
			t.Fatalf("[ gave tab %d, want %d", m.tab, want)
		}
	}
	// The tab really changed: pane [1] shows specs again.
	if got := strings.Join(rowIDs(m), " "); !strings.HasPrefix(got, "specs/2026-09-22-beta") {
		t.Fatalf("rows %q", got)
	}
}

func TestDonePaneTabsFollowTheOpenTab(t *testing.T) {
	for _, tc := range []struct {
		keys  []string
		names string
	}{
		{nil, "Done Dropped"},
		{[]string{"]"}, "Done Dropped"},
		{[]string{"]", "]"}, "Done"},
		{[]string{"]", "]", "]"}, "Fixed Wontfix"},
	} {
		m := press(newModel(t), append(tc.keys, "2")...)
		if got := strings.Join(m.doneTabNames(), " "); got != tc.names {
			t.Errorf("after %v the finished tabs are %q, want %q", tc.keys, got, tc.names)
		}
	}
}

func TestDonePaneTabsCycleAndWrap(t *testing.T) {
	m := press(newModel(t), "2")
	if m.doneTab != 0 {
		t.Fatalf("done tab %d", m.doneTab)
	}
	m = press(m, "]")
	if m.doneTab != 1 {
		t.Fatalf("] gave done tab %d", m.doneTab)
	}
	m = press(m, "]")
	if m.doneTab != 0 {
		t.Fatalf("the finished tabs should wrap, got %d", m.doneTab)
	}
	m = press(m, "[")
	if m.doneTab != 1 {
		t.Fatalf("[ gave done tab %d", m.doneTab)
	}
	// A task is only ever done, so there is no second tab to move to.
	m = press(m, "1", "]", "]", "2", "]", "]", "]")
	if m.doneTab != 0 {
		t.Fatalf("Tasks should stay on Done, got %d", m.doneTab)
	}
	// Coming from a tab that has two finished tabs, the one without a second
	// drops back to its only tab.
	m = press(newModel(t), "]", "]", "]", "2", "]", "1", "[", "2")
	if m.tab != 2 || m.doneTab != 0 {
		t.Fatalf("Tasks should be back on Done, tab %d done %d", m.tab, m.doneTab)
	}
}

func TestPaneDetailHasNoTabs(t *testing.T) {
	m := press(newModel(t), "]", "]", "2", "]", "3", "]", "[", "]")
	if m.tab != 2 || m.doneTab != 0 {
		t.Fatalf("pane [3] moved the tabs: tab %d done %d", m.tab, m.doneTab)
	}
}

func TestEachTabHoldsItsOwnItems(t *testing.T) {
	for _, tc := range []struct {
		keys []string
		rows string
	}{
		{nil, "specs/2026-09-22-beta specs/2026-09-20-alpha specs/2026-09-18-weird specs/2026-09-17-broken " + groupRowID},
		{[]string{"]"}, "plans/2026-09-23-lonely plans/2026-09-21-alpha"},
		{[]string{"]", "]"}, "plans/2026-09-23-lonely#task-1 plans/2026-09-21-alpha#task-2"},
		{[]string{"]", "]", "]"}, "bugs/2026-09-26-open specs/2026-09-15-really-bug"},
	} {
		m := press(newModel(t), tc.keys...)
		if got := strings.Join(rowIDs(m), " "); got != tc.rows {
			t.Errorf("after %v pane [1] holds %q, want %q", tc.keys, got, tc.rows)
		}
	}
}

func TestDonePaneHoldsTheFinishedItemsOfTheOpenTab(t *testing.T) {
	m := press(newModel(t), "2")
	if got := strings.Join(doneRowIDs(m), " "); got != "specs/2026-09-16-finished" {
		t.Fatalf("Specs done %q", got)
	}
	m = press(m, "]")
	if got := strings.Join(doneRowIDs(m), " "); got != "specs/2026-09-19-dropped" {
		t.Fatalf("Specs dropped %q", got)
	}
	// Pane [2] keeps the finished tab it was on, so step back to Done.
	m = press(m, "1", "]", "2", "[")
	if got := strings.Join(doneRowIDs(m), " "); got != "plans/2026-09-28-dotted-tasks plans/2026-09-27-orphan plans/2026-09-26-dash-tasks plans/2026-09-25-crash-fix plans/2026-09-16-finished" {
		t.Fatalf("Plans done %q", got)
	}
	m = press(m, "1", "]", "2")
	if got := strings.Join(doneRowIDs(m), " "); got != "plans/2026-09-28-dotted-tasks#task-2.1 plans/2026-09-28-dotted-tasks#task-2.2 plans/2026-09-27-orphan#task-1 plans/2026-09-26-dash-tasks#task-F-1 plans/2026-09-26-dash-tasks#task-F-2 plans/2026-09-25-crash-fix#task-F1 plans/2026-09-21-alpha#task-1 plans/2026-09-16-finished#task-1" {
		t.Fatalf("Tasks done %q", got)
	}
	m = press(m, "1", "]", "2")
	if got := strings.Join(doneRowIDs(m), " "); got != "bugs/2026-09-24-crash" {
		t.Fatalf("Bugs fixed %q", got)
	}
	m = press(m, "]")
	if got := doneRowIDs(m); len(got) != 0 {
		t.Fatalf("no bug is wontfix, got %v", got)
	}
}

func TestMoveKeysInListPanes(t *testing.T) {
	m := newModel(t)
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("first selection %s", m.Selected().ID)
	}
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("after j: %s", m.Selected().ID)
	}
	m = press(m, "k")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("after k: %s", m.Selected().ID)
	}
	m = press(m, "ctrl+d")
	if m.openRows()[m.cursor()].id != groupRowID {
		t.Fatalf("ctrl+d should jump a page: %s", m.openRows()[m.cursor()].id)
	}
	m = press(m, "ctrl+u")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("ctrl+u should step back a page: %s", m.Selected().ID)
	}
	// The same keys move pane [2].
	m = press(m, "2", "j", "j")
	if m.Selected() == nil || m.Selected().Status != "done" {
		t.Fatalf("pane [2] did not move: %v", m.Selected())
	}
}

func TestTopAndBottomKeysInListPanes(t *testing.T) {
	m := newModel(t)
	m = press(m, "G")
	if m.openRows()[m.cursor()].id != groupRowID {
		t.Fatal("G should land on the last row")
	}
	m = press(m, "j")
	if m.openRows()[m.cursor()].id != groupRowID {
		t.Fatal("j past the end should stay on the last row")
	}
	m = press(m, "g", "k")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatal("k at the top should stay on the first row")
	}
}

func TestScrollKeysInPaneDetail(t *testing.T) {
	m := press(newModel(t), "3")
	if m.scroll != 0 {
		t.Fatalf("scroll %d", m.scroll)
	}
	m = press(m, "ctrl+d")
	if m.scroll != pageLines {
		t.Fatalf("ctrl+d scrolled to %d", m.scroll)
	}
	m = press(m, "ctrl+u", "ctrl+u")
	if m.scroll != 0 {
		t.Fatalf("scroll went below 0: %d", m.scroll)
	}
	m = press(m, "j", "j")
	if m.scroll != 2 {
		t.Fatalf("j should scroll one line: %d", m.scroll)
	}
	m = press(m, "g")
	if m.scroll != 0 {
		t.Fatalf("g should go back to the top: %d", m.scroll)
	}
	m = press(m, "G")
	if m.scroll == 0 {
		t.Fatal("G should go to the bottom of the body")
	}
}

func TestScrollKeysDoNothingInListPanes(t *testing.T) {
	m := press(newModel(t), "3", "ctrl+d", "1")
	if m.scroll != 0 {
		t.Fatalf("leaving pane [3] should drop the scroll: %d", m.scroll)
	}
	m = press(m, "ctrl+d")
	if m.scroll != 0 {
		t.Fatalf("a list pane should not scroll the body: %d", m.scroll)
	}
	m = press(m, "ctrl+u", "j", "k")
	if m.scroll != 0 {
		t.Fatalf("a list pane should not scroll the body: %d", m.scroll)
	}
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("j and k should still move: %v", m.Selected())
	}
}

func TestClickOnARowSelectsItAndFocusesItsPane(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	// A row takes three lines: the title, the dim meta line and a blank one.
	m = click(m, 2, g.open.y+1+rowLines)
	if m.focus != paneOpen || m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// The blank line under a row belongs to no row, so it only takes focus.
	m = click(m, 2, g.open.y+1+2)
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("the blank line should not select: %v", m.Selected())
	}
	// The same works in pane [2].
	m = press(newModel(t), "]", "2")
	g = m.geometry()
	m = click(m, 2, g.done.y+1+rowLines)
	if m.focus != paneDone || m.Selected().ID != "plans/2026-09-27-orphan" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
}

func TestClickOnATabNameSwitchesTab(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	m = click(m, g.open.x+g.open.tabs[tabTasks].x+1, g.open.y)
	if m.tab != tabTasks {
		t.Fatalf("clicking Tasks gave tab %d", m.tab)
	}
	// The click also gave the pane the focus, and no row changed.
	if m.focus != paneOpen || len(rowIDs(m)) != 2 {
		t.Fatalf("focus %d rows %v", m.focus, rowIDs(m))
	}
	// Tasks have one finished tab, so move to Bugs before clicking the second.
	m = click(m, g.open.x+g.open.tabs[tabBugs].x+1, g.open.y)
	m = press(m, "2")
	g = m.geometry()
	m = click(m, g.done.x+g.done.tabs[1].x+1, g.done.y)
	if m.focus != paneDone || m.doneTab != 1 {
		t.Fatalf("focus %d done tab %d", m.focus, m.doneTab)
	}
	// Clicking the dash before a name belongs to no tab.
	m = click(m, g.done.x+g.done.tabs[1].x-1, g.done.y)
	if m.doneTab != 1 {
		t.Fatalf("a click on the dash moved the tab to %d", m.doneTab)
	}
}

func TestClickInsideAPaneOnlyFocusesIt(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	// Below the last row of pane [1] is its own border, so use the wide gutter
	// of the left column instead.
	m = click(m, g.open.x+g.open.w-1, g.open.y+g.open.h-2)
	if m.focus != paneOpen {
		t.Fatalf("focus %d", m.focus)
	}
	m = click(m, 119, 5)
	if m.focus != paneDetail || m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// A click on the status line does nothing at all.
	m = click(m, 10, 39)
	if m.focus != paneDetail {
		t.Fatalf("focus %d", m.focus)
	}
}

func TestWheelActsOnThePaneUnderThePointer(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	m = wheel(m, 2, g.open.y+2, true)
	if m.focus != paneOpen || m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("the wheel should not move off the first row: %v", m.Selected())
	}
	m = wheel(m, 2, g.open.y+2, false)
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("the wheel should move down: %v", m.Selected())
	}
	m = wheel(m, 2, g.open.y+2, false)
	m = wheel(m, 2, g.open.y+2, true)
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("the wheel should move back up: %v", m.Selected())
	}
	// The pane under the pointer wins over the one that has the focus.
	m = press(newModel(t), "3")
	m = wheel(m, 2, g.open.y+2, false)
	if m.focus != paneOpen || m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// Over pane [2] the wheel moves that pane.
	m = press(newModel(t), "]", "2")
	m = wheel(m, 2, g.done.y+2, false)
	if m.focus != paneDone || m.Selected().ID != "plans/2026-09-27-orphan" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// Over pane [3] it scrolls the body.
	m = press(newModel(t), "3")
	m = wheel(m, 119, 5, false)
	if m.scroll == 0 {
		t.Fatal("the wheel should scroll pane [3]")
	}
	m = press(newModel(t), "3")
	m = wheel(m, 119, 5, true)
	if m.scroll != 0 {
		t.Fatalf("the wheel should scroll back to the top: %d", m.scroll)
	}
}

func TestKeysContinueFromAClickedRow(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	m = click(m, 2, g.open.y+1+2*rowLines)
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-17-broken" {
		t.Fatalf("j moved to %v", m.Selected())
	}
	m = click(m, 2, g.open.y+1)
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("j did not start from the clicked row: %v", m.Selected())
	}
}

func TestKeyADoesNothing(t *testing.T) {
	m := newModel(t)
	before := m
	after := press(m, "a", "a")
	if strings.Join(rowIDs(after), " ") != strings.Join(rowIDs(before), " ") ||
		strings.Join(doneRowIDs(after), " ") != strings.Join(doneRowIDs(before), " ") ||
		after.focus != before.focus || after.tab != before.tab || after.doneTab != before.doneTab ||
		after.Selected().ID != before.Selected().ID || after.status != before.status {
		t.Fatalf("a changed the screen: %+v", after)
	}
}

func TestHelpSwallowsKeysUntilItCloses(t *testing.T) {
	m := press(newModel(t), "?")
	if !m.help {
		t.Fatal("? should open the help")
	}
	before := strings.Join(rowIDs(m), " ")
	m = press(m, "j", "k", "]", "3", "/", "a", "t", "n")
	if !m.help || m.focus != paneOpen || m.tab != 0 || m.searching || m.slug != nil || m.popup != nil {
		t.Fatalf("a key got through the help: %+v", m)
	}
	if strings.Join(rowIDs(m), " ") != before {
		t.Fatalf("the list moved under the help: %v", rowIDs(m))
	}
	if _, cmd := m.Update(key("q")); cmd != nil {
		t.Fatal("q should not quit while the help is open")
	}
	m = press(m, "esc")
	if m.help {
		t.Fatal("esc should close the help")
	}
	m = press(m, "?", "?")
	if m.help {
		t.Fatal("? should close the help")
	}
	m = press(m, "?", "esc", "j")
	if m.help || m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("the keys should work again once the help is closed: %+v", m)
	}
}

func TestAMinuteTickMovesTheClock(t *testing.T) {
	m := newModel(t)
	if m.now.IsZero() {
		t.Fatal("the clock should start at the real time")
	}
	at := time.Date(2026, 9, 26, 20, 46, 0, 0, time.UTC)
	next, cmd := m.Update(clockMsg(at))
	m = next.(Model)
	if !m.now.Equal(at) {
		t.Fatalf("clock %s", m.now)
	}
	if cmd == nil {
		t.Fatal("a tick should ask for the next one")
	}
}

func TestItemKeysUseTheFocusedListPane(t *testing.T) {
	// Pane [2] has the focus, so s works on the finished plan it selected.
	m := press(newModel(t), "]", "2", "s")
	if m.popup == nil || m.popup.field != "status" {
		t.Fatalf("pane [2] popup %+v", m.popup)
	}
	if m.popup.idx != 3 {
		t.Fatalf("the popup should start on the row's own status done, got %d", m.popup.idx)
	}
	// The same key with pane [1] focused works on that pane's row instead.
	other := press(newModel(t), "s")
	if other.popup == nil || other.popup.field != "status" || other.popup.idx != 0 {
		t.Fatalf("pane [1] popup %+v", other.popup)
	}
	m = press(m, "esc", "n")
	if m.slug == nil {
		t.Fatal("n should ask for a slug")
	}
	m = press(m, "esc", "/")
	if !m.searching {
		t.Fatal("/ should start a search")
	}
	if _, cmd := press(m, "esc").Update(key("r")); cmd == nil {
		t.Fatal("r should reload the board")
	}
	if _, cmd := newModel(t).Update(key("q")); cmd == nil {
		t.Fatal("q should quit")
	}
}

func TestGeometryPlacesThePanes(t *testing.T) {
	m := sized(newModel(t), 120, 40)
	g := m.geometry()
	if !g.wide || g.leftW != 36 {
		t.Fatalf("wide %v leftW %d", g.wide, g.leftW)
	}
	// Pane [1] takes about two thirds of the height, pane [2] the rest, and
	// the last line of the screen is the status line.
	if g.open.x != 0 || g.open.y != 0 || g.open.w != 36 || g.open.h != 26 {
		t.Fatalf("pane [1] %+v", g.open)
	}
	if g.done.x != 0 || g.done.y != 26 || g.done.w != 36 || g.done.h != 13 {
		t.Fatalf("pane [2] %+v", g.done)
	}
	if g.detail.x != 36 || g.detail.y != 0 || g.detail.w != 84 || g.detail.h != 39 {
		t.Fatalf("pane [3] %+v", g.detail)
	}
	// The border takes two lines and every row three, so a pane shows whole
	// rows only, and the window follows the cursor.
	if g.open.inner != 24 || g.open.rows != 8 || g.open.first != 0 {
		t.Fatalf("pane [1] holds %+v", g.open)
	}
	if g.done.inner != 11 || g.done.rows != 3 || g.done.first != 0 {
		t.Fatalf("pane [2] holds %+v", g.done)
	}
	// Pane [2] shows three rows at a time, so its window slides down to keep
	// the last of the five finished plans in sight.
	g = press(m, "]", "2", "G").geometry()
	if g.done.first != 2 || g.done.rows != 3 {
		t.Fatalf("pane [2] should scroll to the cursor: %+v", g.done)
	}
	// The left column is 30% of the width, held between 28 and 48.
	for _, w := range []int{60, 80, 100, 120, 200} {
		if got := sized(newModel(t), w, 40).geometry().leftW; got != clamp(w*3/10, 28, 48) {
			t.Errorf("at %d columns the left column is %d", w, got)
		}
	}
	// Below 60 columns only the focused pane is on screen, full width.
	n := sized(press(newModel(t), "2"), 40, 20).geometry()
	if n.wide || n.full.w != 40 || n.full.h != 19 {
		t.Fatalf("narrow screen %+v", n)
	}
	if n.open.w != 0 || n.done.w != 0 || n.detail.w != 0 {
		t.Fatal("a narrow screen should only draw the focused pane")
	}
	if n.full.x != 0 || n.full.y != 0 || n.full.w != 40 || n.full.h != 19 {
		t.Fatalf("the focused pane should take the whole screen: %+v", n.full)
	}
	// The tab names sit where the title draws them.
	if len(g.open.tabs) != 4 || g.open.tabs[0] != (tabBox{x: 4, w: 5}) || g.open.tabs[1].x != 12 {
		t.Fatalf("pane [1] tabs %+v", g.open.tabs)
	}
	if len(g.done.tabs) != 2 || len(g.detail.tabs) != 0 {
		t.Fatalf("pane [2] tabs %+v pane [3] tabs %+v", g.done.tabs, g.detail.tabs)
	}
}

func TestSelectionIsPerTab(t *testing.T) {
	m := press(newModel(t), "j", "]", "]", "]", "[", "[", "[")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("the open tab lost its selection: %v", m.Selected())
	}
	m = press(newModel(t), "]", "2", "j", "]", "[")
	if m.Selected().ID != "plans/2026-09-27-orphan" {
		t.Fatalf("the finished tab lost its selection: %v", m.Selected())
	}
}

func TestSelectedFollowsTheLastFocusedListPane(t *testing.T) {
	m := press(newModel(t), "]", "2", "j", "3")
	if m.Selected().ID != "plans/2026-09-27-orphan" {
		t.Fatalf("pane [3] should keep showing pane [2]: %v", m.Selected())
	}
	// Moving the cursor inside pane [2] keeps it the one on show.
	m = press(m, "shift+tab", "j", "3")
	if m.Selected().ID != "plans/2026-09-26-dash-tasks" {
		t.Fatalf("pane [3] should keep showing pane [2]: %v", m.Selected())
	}
	// Pane [1] takes over as soon as it has the focus.
	m = press(m, "1")
	if m.Selected().ID != "plans/2026-09-23-lonely" {
		t.Fatalf("pane [1] should show its own row: %v", m.Selected())
	}
}

func TestUntypedGroupRow(t *testing.T) {
	m := newModel(t)
	if got := rowIDs(m); got[len(got)-1] != groupRowID {
		t.Fatalf("rows %v", got)
	}
	m = press(m, "G", "enter")
	if !m.groupOpen || rowIDs(m)[len(rowIDs(m))-1] != "docs/superpowers/specs/2026-01-01-old" {
		t.Fatalf("group did not open: %v", rowIDs(m))
	}
	m = press(m, "enter")
	if m.groupOpen {
		t.Fatal("enter on the group row should close it")
	}
	// The other tabs have no untyped files to show.
	if got := rowIDs(press(newModel(t), "]")); len(got) != 2 {
		t.Fatalf("plans rows %v", got)
	}
}

func TestSearch(t *testing.T) {
	m := press(newModel(t), "/", "c", "r", "a", "s", "h")
	if !m.searching || m.query != "crash" {
		t.Fatalf("searching %v query %q", m.searching, m.query)
	}
	for _, id := range rowIDs(m) {
		it := m.board.Get(id)
		if !strings.Contains(strings.ToLower(it.Title+it.Slug+it.Ref), "crash") {
			t.Errorf("row %s does not match", id)
		}
	}
	m = press(m, "backspace")
	if m.query != "cras" {
		t.Fatalf("backspace: %q", m.query)
	}
	m = press(m, "enter")
	if m.searching || m.query != "cras" {
		t.Fatal("enter should stop typing and keep the query")
	}
	m = press(m, "esc")
	if m.query != "" {
		t.Fatal("esc should clear the query")
	}
	m = press(m, "/", "x", "esc")
	if m.searching || m.query != "" {
		t.Fatal("esc while typing should stop and clear")
	}
}

func TestPopupRefusals(t *testing.T) {
	m := press(newModel(t), "]", "]", "s")
	if m.popup != nil || !strings.Contains(m.status, "checkboxes") {
		t.Fatalf("task popup: %v %q", m.popup, m.status)
	}
	m = press(newModel(t), "G", "enter", "j", "t")
	if m.popup != nil || !strings.Contains(m.status, "legacy") {
		t.Fatalf("legacy popup: %v %q", m.popup, m.status)
	}
}

func TestStatusPopupSetsValue(t *testing.T) {
	m := newModel(t)
	var got []string
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		got = []string{id, field, value}
		return write.Outcome{Committed: true}, nil
	}
	m = press(m, "]", "]", "]", "s")
	if m.popup == nil || m.popup.field != "status" || strings.Join(m.popup.options, ",") != "open,fixing,fixed,wontfix" || m.popup.idx != 0 {
		t.Fatalf("popup %+v", m.popup)
	}
	m = press(m, "j", "j", "enter")
	if strings.Join(got, " ") != "bugs/2026-09-26-open status fixed" {
		t.Fatalf("setValue got %v", got)
	}
	if m.popup != nil || !strings.Contains(m.status, "committed") {
		t.Fatalf("popup %v status %q", m.popup, m.status)
	}
}

func TestPopupEscAndOutcomes(t *testing.T) {
	m := press(newModel(t), "]", "]", "]", "t", "esc")
	if m.popup != nil {
		t.Fatal("esc should close the popup")
	}
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		return write.Outcome{Skipped: true, Reason: "HEAD is detached"}, nil
	}
	m = press(m, "t", "enter")
	if !strings.Contains(m.status, "not committed: HEAD is detached") {
		t.Fatalf("status %q", m.status)
	}
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		return write.Outcome{}, errors.New("boom")
	}
	m = press(m, "t", "enter")
	if !strings.Contains(m.status, "error: boom") {
		t.Fatalf("status %q", m.status)
	}
}

func TestReloadKeepsSelection(t *testing.T) {
	cfg, b := fixture(t)
	m := press(newModel(t), "j") // alpha
	next, _ := m.Update(reloadMsg{b: b})
	m = next.(Model)
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("selection moved on reload: %s", m.Selected().ID)
	}

	// Drop alpha from the board: the cursor stays at the same row number.
	var kept []*board.Item
	for _, it := range b.Items {
		if it.ID != "specs/2026-09-20-alpha" {
			kept = append(kept, it)
		}
	}
	smaller, _ := board.Load(cfg)
	smaller.Items = kept
	next, _ = m.Update(reloadMsg{b: smaller})
	m = next.(Model)
	if m.Selected() == nil || m.Selected().ID != "specs/2026-09-18-weird" {
		t.Fatalf("after alpha vanished: %v", m.Selected())
	}

	next, _ = m.Update(reloadMsg{err: errors.New("disk")})
	m = next.(Model)
	if !strings.Contains(m.status, "reload failed") {
		t.Fatalf("status %q", m.status)
	}
}

func TestWatchFailedGoesManual(t *testing.T) {
	next, _ := newModel(t).Update(WatchFailed(errors.New("too many files")))
	m := next.(Model)
	if !m.manual || !strings.Contains(m.status, "press r") {
		t.Fatalf("manual %v status %q", m.manual, m.status)
	}
}

func TestNewBugSlugInput(t *testing.T) {
	m := press(newModel(t), "n", "a", "B", "-", "1", " ", "backspace")
	if m.slug == nil || *m.slug != "a-" {
		t.Fatalf("slug %v", m.slug)
	}
	m = press(m, "esc")
	if m.slug != nil {
		t.Fatal("esc should cancel the slug prompt")
	}
}
