package tui

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/write"
)

func fixture(t *testing.T) (config.Config, *board.Board) {
	t.Helper()
	cfg := config.Default(fixtureRoot(t))
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return cfg, b
}

var (
	fixtureOnce sync.Once
	fixtureDir  string
	fixtureErr  error
)

// fixtureRoot gives the fixture folder, copied out of this repo once for the
// whole run. The tests read it outside the repo because pane [2] orders
// finished items by the last commit on the file: inside the repo this repo's
// own history would decide what the tests expect, and every later commit
// touching a fixture would break them.
func fixtureRoot(t *testing.T) string {
	t.Helper()
	fixtureOnce.Do(func() {
		src, err := filepath.Abs("../board/testdata/basic")
		if err != nil {
			fixtureErr = err
			return
		}
		tmp, err := os.MkdirTemp("", "pm-board-tui")
		if err != nil {
			fixtureErr = err
			return
		}
		dir := filepath.Join(tmp, "basic")
		if err := copyTree(src, dir); err != nil {
			fixtureErr = err
			return
		}
		fixtureDir = dir
	})
	if fixtureErr != nil {
		t.Fatal(fixtureErr)
	}
	return fixtureDir
}

func TestMain(m *testing.M) {
	code := m.Run()
	if fixtureDir != "" {
		os.RemoveAll(filepath.Dir(fixtureDir))
	}
	os.Exit(code)
}

// copyTree writes every file of src under dst.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		to := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(to, 0o755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(to, b, 0o644)
	})
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
	for _, want := range []int{tabPlans, tabTasks, tabBugs, tabDebt, tabSpecs} {
		m = press(m, "]")
		if m.tab != want {
			t.Fatalf("] gave tab %d, want %d", m.tab, want)
		}
	}
	for _, want := range []int{tabDebt, tabBugs, tabTasks, tabPlans, tabSpecs} {
		m = press(m, "[")
		if m.tab != want {
			t.Fatalf("[ gave tab %d, want %d", m.tab, want)
		}
	}
	// The tab really changed: pane [1] shows specs again.
	if got := strings.Join(rowIDs(m), " "); !strings.HasPrefix(got, "specs/2026-09-20-alpha") {
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
		{[]string{"]", "]", "]", "]"}, "Done Wontfix"},
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
		{nil, "specs/2026-09-20-alpha \x00divider specs/2026-09-22-beta specs/2026-09-18-weird specs/2026-09-17-broken " + groupRowID},
		{[]string{"]"}, "plans/2026-09-21-alpha \x00divider plans/2026-09-23-lonely"},
		{[]string{"]", "]"}, "plans/2026-09-21-alpha#task-2 \x00divider plans/2026-09-23-lonely#task-1"},
		{[]string{"]", "]", "]"}, "bugs/2026-09-26-open specs/2026-09-15-really-bug"},
		{[]string{"]", "]", "]", "]"}, "debt/2026-09-27-orphan-debt#item-1"},
	} {
		m := press(newModel(t), tc.keys...)
		if got := strings.Join(rowIDs(m), " "); got != tc.rows {
			t.Errorf("after %v pane [1] holds %q, want %q", tc.keys, got, tc.rows)
		}
	}
}

// debtModel builds a small board with one plan and one debt file that names
// it as parent, so the Debt tab tests do not depend on the shared basic
// fixture (which already carries its own orphan debt file with a broken
// parent, used by the board package's tests).
func debtModel(t *testing.T) Model {
	t.Helper()
	cfg := treeCfg(t, map[string]string{
		".acta/plans/2026-09-26-short-ids.md": "---\nid: PLAN-3\nhash: k3f2\n---\n# Short IDs\n",
		".acta/debt/2026-09-27-short-ids.md":  "---\nid: DEBT-1\nhash: t9qe\nparent: plans/2026-09-26-short-ids\n---\n# Review NOTEs: Short IDs\n\n- [ ] a\n- [x] b\n- [-] c\n",
	})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return m
}

func TestDebtTabListsOnlyTheOpenLine(t *testing.T) {
	m := openTab(t, debtModel(t), tabDebt)
	if got := strings.Join(rowIDs(m), " "); got != "debt/2026-09-27-short-ids#item-1" {
		t.Fatalf("Debt open rows %q, want only the open line", got)
	}
	if it := m.Selected(); it == nil || it.Title != "a" {
		t.Fatalf("Debt open item = %+v, want title a", it)
	}
}

func TestDebtDonePaneSplitsDoneAndWontfix(t *testing.T) {
	m := press(openTab(t, debtModel(t), tabDebt), "2")
	if got := strings.Join(doneRowIDs(m), " "); got != "debt/2026-09-27-short-ids#item-2" {
		t.Fatalf("Debt Done rows %q", got)
	}
	if it := m.Selected(); it == nil || it.Title != "b" {
		t.Fatalf("Debt Done item = %+v, want title b", it)
	}
	m = press(m, "]")
	if got := strings.Join(doneRowIDs(m), " "); got != "debt/2026-09-27-short-ids#item-3" {
		t.Fatalf("Debt Wontfix rows %q", got)
	}
	if it := m.Selected(); it == nil || it.Title != "c" {
		t.Fatalf("Debt Wontfix item = %+v, want title c", it)
	}
}

// TestDebtFileDoesNotChangeTheOtherTabs is the property behind adding a fifth
// tab: a debt file on the board must never leak into, or take rows away from,
// Specs, Plans, Tasks or Bugs.
func TestDebtFileDoesNotChangeTheOtherTabs(t *testing.T) {
	withDebt := debtModel(t)
	without := withDebt
	plain, err := board.Load(treeCfg(t, map[string]string{
		".acta/plans/2026-09-26-short-ids.md": "---\nid: PLAN-3\nhash: k3f2\n---\n# Short IDs\n",
	}))
	if err != nil {
		t.Fatal(err)
	}
	without.board = plain
	for _, tab := range []int{tabSpecs, tabPlans, tabTasks, tabBugs} {
		a := strings.Join(rowIDs(openTab(t, withDebt, tab)), " ")
		b := strings.Join(rowIDs(openTab(t, without, tab)), " ")
		if a != b {
			t.Fatalf("tab %d changed with the debt file present: got %q, want %q", tab, a, b)
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
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("first selection %s", m.Selected().ID)
	}
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("after j: %s", m.Selected().ID)
	}
	m = press(m, "k")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("after k: %s", m.Selected().ID)
	}
	m = press(m, "k")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("k at the top should stay: %s", m.Selected().ID)
	}
	m = press(m, "ctrl+d")
	if m.openRows()[m.cursor()].id != groupRowID {
		t.Fatalf("ctrl+d should jump a page: %s", m.openRows()[m.cursor()].id)
	}
	m = press(m, "ctrl+u")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
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
		t.Fatalf("G should land on the last row: %v", rowIDs(m))
	}
	m = press(m, "j")
	if m.openRows()[m.cursor()].id != groupRowID {
		t.Fatal("j past the end should stay on the last row")
	}
	m = press(m, "g", "k")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
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
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("j and k should still move: %v", m.Selected())
	}
}

func TestClickOnARowSelectsItAndFocusesItsPane(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	// A row takes three lines: the title, the dim meta line and a blank one.
	m = click(m, 2, g.open.y+1+2*rowLines)
	if m.focus != paneOpen || m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// The blank line under a row belongs to no row, so it only takes focus.
	m = click(m, 2, g.open.y+1+2)
	if m.Selected().ID != "specs/2026-09-22-beta" {
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
	if m.focus != paneOpen || len(rowIDs(m)) != 3 {
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
	if m.focus != paneDetail || m.Selected().ID != "specs/2026-09-20-alpha" {
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
	if m.focus != paneOpen || m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("the wheel should not move off the first row: %v", m.Selected())
	}
	m = wheel(m, 2, g.open.y+2, false)
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("the wheel should move down: %v", m.Selected())
	}
	m = wheel(m, 2, g.open.y+2, false)
	m = wheel(m, 2, g.open.y+2, true)
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("the wheel should move back up: %v", m.Selected())
	}
	// The pane under the pointer wins over the one that has the focus.
	m = press(newModel(t), "3")
	m = wheel(m, 2, g.open.y+2, false)
	if m.focus != paneOpen || m.Selected().ID != "specs/2026-09-22-beta" {
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
	if m.Selected().ID != "specs/2026-09-18-weird" {
		t.Fatalf("j moved to %v", m.Selected())
	}
	m = click(m, 2, g.open.y+1)
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-22-beta" {
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
	if m.help || m.Selected().ID != "specs/2026-09-22-beta" {
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
	if other.popup == nil || other.popup.field != "status" || other.popup.idx != 2 {
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
	// 120 columns, the normal width: pane [1] must still fit all five tab
	// names ("Specs ─ Plans ─ Tasks ─ Bugs ─ Debt") at the left column's
	// usual size, not a wider one carved out just for the tab title.
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
	// All five names must still be in the drawn title at this normal width.
	if title := strings.Split(plain(m.View()), "\n")[g.open.y]; !strings.Contains(title, "Debt") {
		t.Fatalf("pane [1] title at 120 columns is missing Debt: %q", title)
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
	// The tab boxes sit where the drawn title puts the names, so a click lands
	// on the word it points at. The place comes from the view, not from a
	// number written down here, which is how the two drifted apart before.
	title := strings.Split(plain(m.View()), "\n")[g.open.y]
	for i, name := range tabNames {
		want := cellAt(t, title, name)
		if got := g.open.tabs[i]; got.x != want || got.w != len(name) {
			t.Fatalf("pane [1] tab %q sits at %+v, but the title draws it at %d: %q", name, got, want, title)
		}
	}
	if len(g.open.tabs) != len(tabNames) || len(g.done.tabs) != 2 || len(g.detail.tabs) != 0 {
		t.Fatalf("pane [1] tabs %+v pane [2] tabs %+v pane [3] tabs %+v", g.open.tabs, g.done.tabs, g.detail.tabs)
	}
}

func TestSelectionIsPerTab(t *testing.T) {
	m := press(newModel(t), "j", "]", "]", "]", "[", "[", "[")
	if m.Selected().ID != "specs/2026-09-22-beta" {
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
	if m.Selected().ID != "plans/2026-09-21-alpha" {
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
	if got := rowIDs(press(newModel(t), "]")); len(got) != 3 {
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
	m := press(newModel(t), "j") // beta
	next, _ := m.Update(reloadMsg{b: b})
	m = next.(Model)
	if m.Selected().ID != "specs/2026-09-22-beta" {
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
	next, _ = m.Update(reloadMsg{b: smaller})
	m = next.(Model)
	if m.Selected() == nil || m.Selected().ID != "specs/2026-09-22-beta" {
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

// clickWidths gives the widths from 60 to 200 columns where the left column
// changes size, plus both ends of the range. The tab boxes and the drawn names
// both come from that column, so these are every title shape a window can take.
func clickWidths() []int {
	var out []int
	for w := 60; w <= 200; w++ {
		if w == 60 || clamp(w*3/10, 28, 48) != clamp((w-1)*3/10, 28, 48) {
			out = append(out, w)
		}
	}
	if out[len(out)-1] != 200 {
		out = append(out, 200)
	}
	return out
}

// cellAt gives the screen column where name starts in a drawn line. A border
// cell like the corner is one column but three bytes, and the mouse counts
// columns, so the place is counted in cells and not in bytes.
func cellAt(t *testing.T, line, name string) int {
	t.Helper()
	i := strings.Index(line, name)
	if i < 0 {
		t.Fatalf("the line does not draw %q: %q", name, line)
	}
	return utf8.RuneCountInString(line[:i])
}

// drawnLetter gives the screen row and column of the first or the last letter
// of name in the title of a list pane, read back from the view, so this test
// cannot pass with a number the view does not draw. ok is false when the title
// is too narrow to draw that name at all.
func drawnLetter(m Model, p pane, name, end string) (y, x int, ok bool) {
	b := m.geometry().open
	if p == paneDone {
		b = m.geometry().done
	}
	line := strings.Split(plain(m.View()), "\n")[b.y]
	i := strings.Index(line, name)
	if i < 0 {
		return b.y, 0, false
	}
	x = utf8.RuneCountInString(line[:i])
	if end == "last" {
		x += len(name) - 1
	}
	return b.y, x, true
}

// openTab puts pane [1] on the tab with that number.
func openTab(t *testing.T, m Model, tab int) Model {
	t.Helper()
	m = press(m, "1")
	for i := 0; i < len(tabNames) && m.tab != tab; i++ {
		m = press(m, "]")
	}
	if m.tab != tab {
		t.Fatalf("could not open %q", tabNames[tab])
	}
	return m
}

// openTabBefore puts pane [1] on the tab before that number, so a click that
// lands on nothing cannot pass by leaving the tab where it already was.
func openTabBefore(t *testing.T, m Model, tab int) Model {
	return openTab(t, m, (tab+len(tabNames)-1)%len(tabNames))
}

// doneTabBefore puts pane [2] on another finished tab than that number. A pane
// with a single tab stays where it is, and the click is still checked against
// what the title draws.
func doneTabBefore(t *testing.T, m Model, tab int) Model {
	t.Helper()
	m = press(m, "2")
	other := (tab + 1) % len(m.doneTabNames())
	if other == tab {
		return m
	}
	for i := 0; i < 2 && m.doneTab != other; i++ {
		m = press(m, "]")
	}
	if m.doneTab != other {
		t.Fatalf("could not open finished tab %d", other)
	}
	return m
}

// TestClickLandsOnEveryDrawnTabName is the property behind the mouse: the first
// and the last letter of every tab name the title draws, in pane [1] and pane
// [2], at every window size from 60 to 200 columns, must switch to that tab; a
// click on the dashes between two names must switch nothing. The letters come
// from the drawn title, so a click box that drifts from the view fails here.
func TestClickLandsOnEveryDrawnTabName(t *testing.T) {
	widths := clickWidths()
	clicked := map[int]bool{}
	m := sized(newModel(t), widths[0], 40)

	for _, w := range widths {
		for tab := range tabNames {
			for _, end := range []string{"first", "last"} {
				m = sized(m, w, 40)
				m = openTabBefore(t, m, tab)
				y, x, ok := drawnLetter(m, paneOpen, tabNames[tab], end)
				if !ok {
					continue // a narrow title drops the names that do not fit
				}
				if p, _, idx := m.hit(x, y); p != paneOpen || idx != tab {
					t.Fatalf("at %d columns the %s letter of %q maps to pane [%d] tab %d, want tab %d", w, end, tabNames[tab], p+1, idx, tab)
				}
				if got := click(m, x, y).tab; got != tab {
					t.Fatalf("at %d columns a click on the %s letter of %q gave tab %d, want %d", w, end, tabNames[tab], got, tab)
				}
				clicked[tab] = true
			}
		}
	}
	for tab := range tabNames {
		if !clicked[tab] {
			t.Errorf("no width from 60 to 200 drew %q, so it was never clicked", tabNames[tab])
		}
	}

	// Pane [2] follows the tab of pane [1]: Done and Dropped for Specs, Plans
	// and Tasks, Fixed and Wontfix for Bugs. Tasks has Done alone.
	for _, w := range widths {
		for tab := range tabNames {
			m = sized(m, w, 40)
			m = openTab(t, m, tab)
			for i, name := range m.doneTabNames() {
				for _, end := range []string{"first", "last"} {
					m = doneTabBefore(t, m, i)
					y, x, ok := drawnLetter(m, paneDone, name, end)
					if !ok {
						t.Fatalf("with %q open, at %d columns pane [2] does not draw %q", tabNames[tab], w, name)
					}
					if p, _, idx := m.hit(x, y); p != paneDone || idx != i {
						t.Fatalf("with %q open, at %d columns the %s letter of %q maps to pane [%d] tab %d, want tab %d", tabNames[tab], w, end, name, p+1, idx, i)
					}
					if got := click(m, x, y).doneTab; got != i {
						t.Fatalf("with %q open, at %d columns a click on the %s letter of %q gave finished tab %d, want %d", tabNames[tab], w, end, name, got, i)
					}
				}
			}
		}
	}

	// The dashes between two names belong to no tab.
	for _, w := range widths {
		m = sized(m, w, 40)
		m = openTabBefore(t, m, tabBugs)
		before := m.tab
		for _, name := range tabNames[:len(tabNames)-1] {
			y, last, ok := drawnLetter(m, paneOpen, name, "last")
			if !ok {
				continue
			}
			x := last + 1 // the separator cell right after the name, dash or plain space
			if _, _, idx := m.hit(x, y); idx != -1 {
				t.Fatalf("at %d columns the dash after %q maps to tab %d", w, name, idx)
			}
			if got := click(m, x, y).tab; got != before {
				t.Fatalf("at %d columns a click on the dash after %q moved to tab %d", w, name, got)
			}
		}
		for _, tab := range []int{tabSpecs, tabPlans, tabBugs} {
			m = sized(openTab(t, m, tab), w, 40)
			m = doneTabBefore(t, m, 1)
			before := m.doneTab
			y, last, _ := drawnLetter(m, paneDone, m.doneTabNames()[0], "last")
			x := last + 1
			if _, _, idx := m.hit(x, y); idx != -1 {
				t.Fatalf("at %d columns the dash after %q maps to finished tab %d", w, m.doneTabNames()[0], idx)
			}
			if got := click(m, x, y).doneTab; got != before {
				t.Fatalf("at %d columns a click on the dash after %q moved to finished tab %d", w, m.doneTabNames()[0], got)
			}
		}
	}
}

// TestTabTitleAlwaysShowsTheOpenTab is the Task 8 regression: at 120 columns
// the pane [1] title used to drop the open tab's own name off the end, so no
// tab looked selected and a click on its old spot fell into the detail pane.
// This checks, at every width the review named, for every open tab, that the
// title always keeps that tab's name, and that any other name still drawn
// clicks to the right tab.
func TestTabTitleAlwaysShowsTheOpenTab(t *testing.T) {
	m := newModel(t)
	for _, w := range []int{60, 80, 100, 120, 140, 200} {
		for tab := range tabNames {
			m = sized(openTab(t, m, tab), w, 40)
			title := strings.Split(plain(m.View()), "\n")[m.geometry().open.y]
			if !strings.Contains(title, tabNames[tab]) {
				t.Fatalf("at %d columns with %q open, the title drops the open tab: %q", w, tabNames[tab], title)
			}
			if w == 120 {
				for _, name := range tabNames {
					if !strings.Contains(title, name) {
						t.Errorf("at 120 columns the title is missing %q: %q", name, title)
					}
				}
			}
			for other := range tabNames {
				for _, end := range []string{"first", "last"} {
					y, x, ok := drawnLetter(m, paneOpen, tabNames[other], end)
					if !ok {
						continue // a narrow title drops names other than the open one
					}
					if got := click(m, x, y).tab; got != other {
						t.Fatalf("at %d columns a click on the %s letter of %q gave tab %d, want %d", w, end, tabNames[other], got, other)
					}
				}
			}
		}
	}
}

// TestDonePaneShowsTheMostRecentlyCommittedFirst is the order of pane [2]: the
// last commit on the file decides, and the file-name date is the fallback when
// git has no commit for the file.
func TestDonePaneShowsTheMostRecentlyCommittedFirst(t *testing.T) {
	day := func(date string) int64 {
		ts, err := time.Parse("2006-01-02", date)
		if err != nil {
			t.Fatal(err)
		}
		return ts.Unix()
	}
	item := func(id, date string, commit int64) *board.Item {
		return &board.Item{ID: id, Kind: board.KindStory, Title: id, Date: date, Status: "done", CommitAt: commit}
	}
	for _, tc := range []struct {
		name  string
		items []*board.Item
		want  []string
	}{
		{
			// The repro from the plan: the file named 09-01 was committed
			// today, the one named 09-20 last a week ago.
			name:  "the last commit beats the file name",
			items: []*board.Item{item("specs/2026-09-20-late", "2026-09-20", day("2026-09-20")), item("specs/2026-09-01-early", "2026-09-01", day("2026-09-27"))},
			want:  []string{"specs/2026-09-01-early", "specs/2026-09-20-late"},
		},
		{
			name:  "the file name decides when git has no commit",
			items: []*board.Item{item("specs/2026-09-01-early", "2026-09-01", 0), item("specs/2026-09-20-late", "2026-09-20", 0)},
			want:  []string{"specs/2026-09-20-late", "specs/2026-09-01-early"},
		},
		{
			name:  "a commit and a file name compare on the same line",
			items: []*board.Item{item("specs/2026-09-20-late", "2026-09-20", 0), item("specs/2026-09-01-early", "2026-09-01", day("2026-09-27"))},
			want:  []string{"specs/2026-09-01-early", "specs/2026-09-20-late"},
		},
		{
			name:  "an older commit loses to a newer file name",
			items: []*board.Item{item("specs/2026-09-20-late", "2026-09-20", 0), item("specs/2026-09-01-early", "2026-09-01", day("2026-09-10"))},
			want:  []string{"specs/2026-09-20-late", "specs/2026-09-01-early"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newModel(t)
			m.board = &board.Board{Items: tc.items}
			m.tab, m.doneTab = tabSpecs, 0
			got := strings.Join(doneRowIDs(m), " ")
			if want := strings.Join(tc.want, " "); got != want {
				t.Fatalf("pane [2] lists %q, want %q", got, want)
			}
		})
	}
}

// statusY gives the screen row of the bottom line, read back from the view so
// a click test cannot pass on a row the view never draws.
func statusY(t *testing.T, m Model) int {
	t.Helper()
	n := len(strings.Split(m.View(), "\n"))
	if n < 1 {
		t.Fatal("the view draws nothing")
	}
	return n - 1
}

// statusX gives the screen column where name starts on the bottom line, so a
// click lands on the drawn link and not on a number the view never painted.
func statusX(t *testing.T, m Model, name string) int {
	t.Helper()
	last := lastLine(m.View())
	i := strings.Index(plain(last), name)
	if i < 0 {
		t.Fatalf("the bottom line shows no %q: %q", name, plain(last))
	}
	return utf8.RuneCountInString(plain(last)[:i])
}

func TestClickOnBottomLineLinksOpensThem(t *testing.T) {
	for _, tc := range []struct {
		name string
		link string
		url  string
	}{
		{"donate", "Donate", "https://ko-fi.com/someone"},
		{"feedback", "Feedback", "https://example.com/bugs"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := sized(clocked(newModel(t), 20, 46), 200, 30)
			m.cfg.Links.Donate = "https://ko-fi.com/someone"
			m.cfg.Links.Feedback = "https://example.com/bugs"
			var got []string
			m.open = func(url string) error {
				got = append(got, url)
				return nil
			}
			click(m, statusX(t, m, tc.link), statusY(t, m))
			if len(got) != 1 || got[0] != tc.url {
				t.Fatalf("opener got %q, want %q", got, tc.url)
			}
		})
	}
}

func TestClickElsewhereOnBottomLineOpensNothing(t *testing.T) {
	m := sized(clocked(newModel(t), 20, 46), 200, 30)
	m.cfg.Links.Donate = "https://ko-fi.com/someone"
	var got []string
	m.open = func(url string) error {
		got = append(got, url)
		return nil
	}
	y := statusY(t, m)
	before := plain(lastLine(m.View()))
	pipe := strings.Index(before, "|")
	if pipe < 0 {
		t.Fatalf("the bottom line has no divider: %q", before)
	}
	click(m, 0, y)
	click(m, len([]rune(before[:pipe]))-1, y)
	if len(got) != 0 {
		t.Fatalf("opener got %q, want nothing", got)
	}
}

func TestClicksMatchDrawnWordsAtEveryWidth(t *testing.T) {
	for _, donate := range []string{"", "https://ko-fi.com/someone"} {
		for w := 30; w <= 200; w++ {
			m := sized(clocked(newModel(t), 20, 46), w, 30)
			m.cfg.Links.Donate = donate
			m.cfg.Links.Feedback = "https://example.com/bugs"
			raw := lastLine(m.View())
			// The truth is what the terminal draws: each hyperlink wrapper
			// carries its url next to its visible words.
			segs := linkSegments(raw)
			flat := plain(raw)
			type span struct {
				start, end int
				url        string
			}
			var spans []span
			base := 0
			for _, seg := range segs {
				if seg.text == "" {
					continue
				}
				i := strings.Index(flat[base:], seg.text)
				if i < 0 {
					t.Fatalf("donate %q at %d columns: drawn %q not found in %q", donate, w, seg.text, flat)
				}
				start := utf8.RuneCountInString(flat[:base+i])
				spans = append(spans, span{start, start + utf8.RuneCountInString(seg.text), seg.url})
				base += i + len(seg.text)
			}
			if donate != "" {
				found := false
				for _, s := range spans {
					if s.url == donate {
						found = true
					}
				}
				if !found {
					t.Fatalf("donate %q at %d columns: no drawn link opens it: %q", donate, w, flat)
				}
			}
			found := false
			for _, s := range spans {
				if s.url == "https://example.com/bugs" {
					found = true
				}
			}
			if !found {
				t.Fatalf("at %d columns: no drawn link opens feedback: %q", w, flat)
			}
			y := statusY(t, m)
			for x := range w {
				want := ""
				for _, s := range spans {
					if x >= s.start && x < s.end {
						want = s.url
					}
				}
				var got []string
				m.open = func(url string) error {
					got = append(got, url)
					return nil
				}
				click(m, x, y)
				if want == "" && len(got) != 0 {
					t.Fatalf("donate %q at %d columns: cell %d opened %q, want nothing (line %q)", donate, w, x, got, flat)
				}
				if want != "" && (len(got) != 1 || got[0] != want) {
					t.Fatalf("donate %q at %d columns: cell %d opened %q, want %q (line %q)", donate, w, x, got, want, flat)
				}
			}
		}
	}
}

func TestNormalizeVersion(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"v0.3.0", "v0.3.0"},
		{"", "dev"},
		{"(devel)", "dev"},
	} {
		if got := normalizeVersion(tc.in); got != tc.want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// splitModel flips one bug to fixing, so every tab of the fixture holds both
// an in-progress and a not-started item for the split tests.
func splitModel(t *testing.T, keys ...string) Model {
	t.Helper()
	m := newModel(t)
	m.board.Get("bugs/2026-09-26-open").Status = "fixing"
	m = press(m, keys...)
	return m
}

// openGroupIDs reads the item ids of pane [1] with the divider and group rows
// dropped, so a test sees only the item order.
func openGroupIDs(m Model) (doing, rest []string) {
	for _, r := range m.openRows() {
		if r.group || r.divider {
			continue
		}
		if inProgress(m.board.Get(r.id)) {
			doing = append(doing, r.id)
		} else {
			rest = append(rest, r.id)
		}
	}
	return doing, rest
}

func TestInProgressRowsComeFirstInEveryTab(t *testing.T) {
	for tab, keys := range map[int][]string{
		tabSpecs: nil,
		tabPlans: {"1", "]"},
		tabTasks: {"1", "]", "]"},
		tabBugs:  {"1", "]", "]", "]"},
	} {
		m := splitModel(t, keys...)
		if m.tab != tab {
			t.Fatalf("tab %d not open, got %d", tab, m.tab)
		}
		doing, rest := openGroupIDs(m)
		if len(doing) == 0 || len(rest) == 0 {
			t.Fatalf("tab %d holds doing %v rest %v", tab, doing, rest)
		}
		rows := ids(m.openRows())
		if rows[0] != doing[0] || rows[len(doing)] != dividerRowID {
			t.Fatalf("tab %d rows %v", tab, rows)
		}
		for i, id := range doing {
			if rows[i] != id {
				t.Fatalf("tab %d rows %v", tab, rows)
			}
		}
		for i, id := range rest {
			if rows[len(doing)+1+i] != id {
				t.Fatalf("tab %d rows %v", tab, rows)
			}
		}
	}
}

func TestStartedTaskWithNoTicksSortsAsInProgress(t *testing.T) {
	m := newModel(t)
	lonely := m.board.Get("plans/2026-09-23-lonely#task-1")
	if lonely.Done != 0 {
		t.Fatalf("the lonely task should have no ticked box, got %d", lonely.Done)
	}
	lonely.Started = true
	lonely.Status = "doing"
	m = openTab(t, m, tabTasks)
	rows := ids(m.openRows())
	// Both tasks are doing now, so no divider sits between them.
	if len(rows) != 2 || rows[0] != lonely.ID {
		t.Fatalf("the started task should lead, rows %v", rows)
	}
}

func TestUnknownStatusSortsAsNotStarted(t *testing.T) {
	m := splitModel(t)
	rows := ids(m.openRows())
	// The synthetic split board holds only known statuses; the fixture's
	// "bogus" story must also sit below the divider, never above it.
	m2 := newModel(t)
	found := false
	below := false
	for _, r := range m2.openRows() {
		if r.divider {
			below = true
			continue
		}
		if r.id == "specs/2026-09-18-weird" {
			found = true
			if !below {
				t.Fatalf("the bogus story sits above the divider: %v", rows)
			}
		}
	}
	if !found {
		t.Fatal("the bogus story is missing")
	}
}

func TestDividerAbsentWhenAGroupIsEmpty(t *testing.T) {
	// Keeping the items picked from the loaded board means Get still finds
	// them; only the Items slice is swapped.
	keep := func(m Model, want func(*board.Item) bool) Model {
		var items []*board.Item
		for _, it := range m.board.Items {
			if it.Kind == board.KindStory && !it.Legacy && want(it) {
				items = append(items, it)
			}
		}
		// Get no longer finds them, but this test only reads the row ids
		// and the divider flag, never the items behind them.
		m.board = &board.Board{Items: items}
		return m
	}
	full := newModel(t)
	for _, tc := range []struct {
		name string
		want func(*board.Item) bool
		n    int
	}{
		{"only in-progress", func(it *board.Item) bool { return it.Status == "in-progress" }, 1},
		{"only not-started", func(it *board.Item) bool { return it.Status != "in-progress" }, 3},
		{"neither", func(it *board.Item) bool { return false }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := keep(full, tc.want)
			rows := m.openRows()
			if len(rows) != tc.n {
				t.Fatalf("rows %v, want %d", ids(rows), tc.n)
			}
			for _, r := range rows {
				if r.divider {
					t.Fatalf("no divider with one group empty: %v", ids(rows))
				}
			}
		})
	}
	// Both groups present keeps exactly one divider.
	m := splitModel(t)
	n := 0
	for _, r := range m.openRows() {
		if r.divider {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want exactly one divider, got %d in %v", n, ids(m.openRows()))
	}
}

func TestDividerAbsentInSearch(t *testing.T) {
	m := splitModel(t, "/", "d", "o", "i", "n", "g")
	for _, r := range m.openRows() {
		if r.divider {
			t.Fatalf("search should not split: %v", ids(m.openRows()))
		}
	}
}

func TestCursorSkipsTheDivider(t *testing.T) {
	at := func(m Model) row {
		return m.openRows()[m.cursor()]
	}
	for _, k := range []string{"j", "k", "g", "G", "ctrl+d", "ctrl+u"} {
		m := press(splitModel(t), "j", k)
		if r := at(m); r.divider {
			t.Fatalf("%s stopped on the divider", k)
		}
		// Only the legacy group row holds no item; everywhere else
		// something is selected.
		if m.Selected() == nil && !at(m).group {
			t.Fatalf("%s selected nothing", k)
		}
	}
	// Walk every row down and back up: the divider is never the cursor.
	m := press(splitModel(t), "g")
	for i := 0; i < len(m.openRows()); i++ {
		if r := at(m); r.divider {
			t.Fatalf("j stopped on the divider at row %d", i)
		}
		m = press(m, "j")
	}
	m = press(splitModel(t), "G")
	for i := 0; i < len(m.openRows()); i++ {
		if r := at(m); r.divider {
			t.Fatalf("k stopped on the divider at row %d", i)
		}
		m = press(m, "k")
	}
	// The selection still moves over both groups.
	m = press(splitModel(t), "g")
	var seenDoing, seenTodo bool
	for i := 0; i < len(m.openRows()); i++ {
		if sel := m.Selected(); sel != nil {
			if inProgress(sel) {
				seenDoing = true
			} else {
				seenTodo = true
			}
		}
		m = press(m, "j")
	}
	if !seenDoing || !seenTodo {
		t.Fatalf("the cursor should cross both groups, doing %v todo %v", seenDoing, seenTodo)
	}
}

func TestClickOnTheDividerKeepsTheSelection(t *testing.T) {
	m := sized(splitModel(t), 120, 40)
	g := m.geometry()
	div := -1
	for i, r := range m.openRows() {
		if r.divider {
			div = i
			break
		}
	}
	if div < 0 {
		t.Fatal("no divider to click")
	}
	before := m.openRows()[m.cursor()].id
	after := click(m, 2, g.open.y+1+div*rowLines)
	if got := after.openRows()[after.cursor()].id; got != before {
		t.Fatalf("a click on the divider moved from %s to %s", before, got)
	}
	if after.focus != paneOpen {
		t.Fatalf("a click on the divider moved the focus to pane %d", after.focus)
	}
}

func TestEnterFocusesDetailFromBothListPanes(t *testing.T) {
	// Pane [1], every tab.
	tabs := map[int][]string{
		tabSpecs: nil,
		tabPlans: {"]"},
		tabTasks: {"]", "]"},
		tabBugs:  {"]", "]", "]"},
	}
	for tab, keys := range tabs {
		m := press(newModel(t), keys...)
		m.tab = tab
		id := m.Selected().ID
		next, cmd := m.Update(key("enter"))
		m = next.(Model)
		if cmd != nil {
			t.Fatalf("tab %d: enter opened the editor", tab)
		}
		if m.focus != paneDetail {
			t.Fatalf("tab %d: enter left the focus on pane %d", tab, m.focus)
		}
		if m.Selected() == nil || m.Selected().ID != id {
			t.Fatalf("tab %d: enter moved from %s to %v", tab, id, m.Selected())
		}
	}
	// Pane [2].
	m := press(newModel(t), "2")
	id := m.Selected().ID
	next, cmd := m.Update(key("enter"))
	m = next.(Model)
	if cmd != nil {
		t.Fatal("enter in pane [2] opened the editor")
	}
	if m.focus != paneDetail || m.last != paneDone {
		t.Fatalf("enter in pane [2]: focus %d last %d", m.focus, m.last)
	}
	if m.Selected() == nil || m.Selected().ID != id {
		t.Fatalf("enter in pane [2] moved from %s to %v", id, m.Selected())
	}
}

func TestEnterOnTheGroupRowStillToggles(t *testing.T) {
	m := press(newModel(t), "G", "enter")
	if !m.groupOpen {
		t.Fatal("enter on the group row should open it")
	}
	if m.focus == paneDetail {
		t.Fatal("enter on the group row should not focus the detail")
	}
	m = press(m, "enter")
	if m.groupOpen {
		t.Fatal("enter on the group row should close it")
	}
}

func TestEnterOnABranchItemWarnsAndFocuses(t *testing.T) {
	main := treeCfg(t, map[string]string{".acta/bugs/2026-09-20-main.md": "# Main bug\n\n## Symptom\nx\n"})
	branch := board.Tree{Cfg: main, Branch: "feat-x", Files: map[string][]byte{
		".acta/bugs/2026-09-25-branch.md": []byte("# Branch bug\n\n## Symptom\ny\n"),
	}}
	b, err := board.LoadTrees(main, []board.Tree{branch})
	if err != nil {
		t.Fatal(err)
	}
	m := New(main, b, true)
	m.render = func(md string, _ int) string { return md }
	m = press(sized(m, 120, 40), "]", "]", "]")
	next, cmd := m.Update(key("enter"))
	m = next.(Model)
	if cmd != nil {
		t.Fatal("enter on a branch item should not open the editor")
	}
	if !strings.Contains(m.status, "git worktree add") {
		t.Fatalf("enter on a branch item should warn: %q", m.status)
	}
	if m.focus != paneDetail {
		t.Fatalf("enter on a branch item should focus the detail, got pane %d", m.focus)
	}
}

func TestEOpensTheEditorFromEveryPane(t *testing.T) {
	// Pane [1], every tab.
	for tab, keys := range map[int][]string{
		tabSpecs: nil,
		tabPlans: {"]"},
		tabTasks: {"]", "]"},
		tabBugs:  {"]", "]", "]"},
	} {
		m := press(newModel(t), keys...)
		m.tab = tab
		// Tasks refuse the editor: they take status from checkboxes.
		if tab == tabTasks {
			m = press(m, "e")
			if m.Selected() == nil {
				t.Fatalf("tab %d: nothing selected", tab)
			}
			continue
		}
		if _, cmd := m.Update(key("e")); cmd == nil {
			t.Fatalf("tab %d: e opened no editor", tab)
		}
	}
	// Pane [2].
	if _, cmd := press(newModel(t), "2").Update(key("e")); cmd == nil {
		t.Fatal("e in pane [2] opened no editor")
	}
	// Pane [3], from both list panes.
	for _, keys := range [][]string{{"3"}, {"2", "3"}} {
		if _, cmd := press(newModel(t), keys...).Update(key("e")); cmd == nil {
			t.Fatalf("e in pane [3] after %v opened no editor", keys)
		}
	}
}

func TestEOnABranchItemWarns(t *testing.T) {
	main := treeCfg(t, map[string]string{".acta/bugs/2026-09-20-main.md": "# Main bug\n\n## Symptom\nx\n"})
	branch := board.Tree{Cfg: main, Branch: "feat-x", Files: map[string][]byte{
		".acta/bugs/2026-09-25-branch.md": []byte("# Branch bug\n\n## Symptom\ny\n"),
	}}
	b, err := board.LoadTrees(main, []board.Tree{branch})
	if err != nil {
		t.Fatal(err)
	}
	m := New(main, b, true)
	m.render = func(md string, _ int) string { return md }
	m = press(sized(m, 120, 40), "]", "]", "]")
	next, cmd := m.Update(key("e"))
	m = next.(Model)
	if cmd != nil {
		t.Fatal("e on a branch item should not open the editor")
	}
	if !strings.Contains(m.status, "git worktree add") {
		t.Fatalf("e on a branch item should warn: %q", m.status)
	}
}

func TestEscInDetailReturnsToTheListPane(t *testing.T) {
	// [1] -> [3] -> [1].
	m := press(newModel(t), "enter")
	if m.focus != paneDetail || m.last != paneOpen {
		t.Fatalf("enter: focus %d last %d", m.focus, m.last)
	}
	m = press(m, "esc")
	if m.focus != paneOpen {
		t.Fatalf("esc: focus %d", m.focus)
	}
	// [2] -> [3] -> [2].
	m = press(newModel(t), "2", "enter")
	if m.focus != paneDetail || m.last != paneDone {
		t.Fatalf("enter: focus %d last %d", m.focus, m.last)
	}
	m = press(m, "esc")
	if m.focus != paneDone {
		t.Fatalf("esc: focus %d", m.focus)
	}
}

func TestEscInDetailKeepsTheSelection(t *testing.T) {
	m := press(newModel(t), "j", "enter")
	id := m.Selected().ID
	m = press(m, "esc")
	if m.Selected() == nil || m.Selected().ID != id {
		t.Fatalf("esc moved from %s to %v", id, m.Selected())
	}
}
