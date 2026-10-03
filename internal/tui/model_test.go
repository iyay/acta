package tui

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/testguard"
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
	testguard.Watch()
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
	m.dcache = &detailCache{}
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
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
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
// notch away from the user. The real program scrolls on the next frame tick,
// so the helper hands that tick over at once.
func wheel(m Model, x, y int, up bool) Model {
	return wheelTick(wheelOnly(m, x, y, up))
}

// wheelOnly turns the wheel one notch and does not send the frame tick, so a
// test can stack several notches into one frame.
func wheelOnly(m Model, x, y int, up bool) Model {
	button := tea.MouseButtonWheelDown
	if up {
		button = tea.MouseButtonWheelUp
	}
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: button})
	return next.(Model)
}

// wheelTick is the frame tick that applies the notches gathered so far.
func wheelTick(m Model) Model {
	next, _ := m.Update(wheelTickMsg{})
	return next.(Model)
}

func ids(rows []row) []string {
	var out []string
	for _, r := range rows {
		out = append(out, r.id)
	}
	return out
}

// rowIDs are the rows of the box the screen shows, so a test reads the same
// list the reader does.
func rowIDs(m Model) []string { return ids(m.rowsOf(m.listPane())) }

func doneRowIDs(m Model) []string { return ids(m.doneRows()) }

// TestEachTabHoldsItsOwnItems reads the rows every tab of the bar lists, so
// no two tabs can ever show the same kind of item.
func TestEachTabHoldsItsOwnItems(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		tab  int
		keys []string
		rows string
	}{
		{tabScratches, nil, "scratch/2026-09-28-idea-brainstorm scratch/2026-09-28-idea-raw"},
		{tabBugs, nil, "specs/2026-09-15-really-bug bugs/2026-09-26-open bugs/2026-09-28-lag"},
		{tabDebts, nil, "debt/2026-09-27-orphan-debt#item-1"},
		{tabSpecs, nil, "specs/2026-09-17-broken specs/2026-09-18-weird specs/2026-09-20-alpha specs/2026-09-22-beta specs/2026-09-28-from-scratch-design " + groupRowID},
		{tabPlans, nil, "plans/2026-09-21-alpha plans/2026-09-23-lonely"},
		{tabActivities, nil, "plans/2026-09-21-alpha plans/2026-09-21-alpha#task-2"},
	} {
		m := press(newModel(t), append([]string{tabKey(tc.tab)}, tc.keys...)...)
		if got := strings.Join(rowIDs(m), " "); got != tc.rows {
			t.Errorf("tab %s holds %q, want %q", topTabs[tc.tab].name, got, tc.rows)
		}
	}
}

// A plan row folds open into its tasks, and those task rows are the only
// other rows the Plans tab lists, so opening a plan is what puts a task on
// the screen.
func TestThePlansTabShowsTasksUnderAnOpenPlan(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabPlans), " ")
	if got := strings.Join(rowIDs(m), " "); got != "plans/2026-09-21-alpha plans/2026-09-21-alpha#task-1 plans/2026-09-21-alpha#task-2 plans/2026-09-23-lonely" {
		t.Fatalf("the opened Plans tree holds %q", got)
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
	t.Parallel()

	m := press(debtModel(t), tabKey(tabDebts))
	if got := strings.Join(rowIDs(m), " "); got != "debt/2026-09-27-short-ids#item-1" {
		t.Fatalf("Debt open rows %q, want only the open line", got)
	}
	if it := m.Selected(); it == nil || it.Title != "a" {
		t.Fatalf("Debt open item = %+v, want title a", it)
	}
}

func TestDebtDonePaneSplitsDoneAndWontfix(t *testing.T) {
	t.Parallel()

	m := press(debtModel(t), tabKey(tabDebts), "tab")
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

// TestDebtFileDoesNotChangeTheOtherTabs is the property behind the Debt tab:
// a debt file on the board must never leak into, or take rows away from,
// Specs, Plans or Bugs.
func TestDebtFileDoesNotChangeTheOtherTabs(t *testing.T) {
	t.Parallel()

	withDebt := debtModel(t)
	without := withDebt
	plain, err := board.Load(treeCfg(t, map[string]string{
		".acta/plans/2026-09-26-short-ids.md": "---\nid: PLAN-3\nhash: k3f2\n---\n# Short IDs\n",
	}))
	if err != nil {
		t.Fatal(err)
	}
	without.board = plain
	for _, i := range []int{tabScratches, tabBugs, tabSpecs, tabPlans, tabActivities} {
		a := strings.Join(rowIDs(press(withDebt, tabKey(i))), " ")
		b := strings.Join(rowIDs(press(without, tabKey(i))), " ")
		if a != b {
			t.Fatalf("tab %s changed with the debt file present: got %q, want %q", topTabs[i].name, a, b)
		}
	}
}

// TestDonePaneHoldsTheFinishedItemsOfTheOpenTab walks every tab of the bar
// and both of its Done sub-tabs, so the Done pane of one tab can never show
// the finished items of another.
func TestDonePaneHoldsTheFinishedItemsOfTheOpenTab(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		tab  int
		done int
		rows string
	}{
		{tabSpecs, 0, "specs/2026-09-16-finished"},
		{tabSpecs, 1, "specs/2026-09-19-dropped"},
		{tabScratches, 0, "scratch/2026-09-28-idea-used"},
		{tabScratches, 1, "scratch/2026-09-28-idea-dropped"},
		{tabPlans, 0, "plans/2026-09-16-finished plans/2026-09-25-crash-fix plans/2026-09-26-dash-tasks plans/2026-09-27-orphan plans/2026-09-28-dotted-tasks"},
		{tabBugs, 0, "bugs/2026-09-24-crash"},
		{tabBugs, 1, ""},
		// The shared fixture's own debt file carries no finished line, so
		// the Debts tab has none here. The Debts sub-tabs are read from
		// their own board in TestDebtDonePaneSplitsDoneAndWontfix.
		{tabDebts, 0, ""},
		{tabDebts, 1, ""},
	} {
		m := press(newModel(t), tabKey(tc.tab), "tab")
		if tc.done > 0 {
			m = press(m, "]")
		}
		if got := strings.Join(doneRowIDs(m), " "); got != tc.rows {
			t.Errorf("tab %s sub-tab %d holds %q, want %q", topTabs[tc.tab].name, tc.done, got, tc.rows)
		}
	}
}

// A plan that is done lists its tasks in the Done pane as a tree too, so
// opening a finished plan shows the tasks it finished with.
func TestTheDonePlansTreeHoldsTheTasksOfAnOpenPlan(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabPlans), "tab", "G")
	shut := doneRowIDs(m)
	if len(shut) < 2 {
		t.Fatal("the fixture has no finished plans, so this test proves nothing")
	}
	opened := shut[len(shut)-1]
	want := append(append([]string{}, shut[:len(shut)-1]...), append([]string{opened}, m.board.Get(opened).Children...)...)
	if got := doneRowIDs(press(m, " ")); strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("the opened Done tree holds %q, want %q", got, want)
	}
}

func TestMoveKeysInListPanes(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs))
	if m.Selected().ID != "specs/2026-09-17-broken" {
		t.Fatalf("first selection %s", m.Selected().ID)
	}
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-18-weird" {
		t.Fatalf("after j: %s", m.Selected().ID)
	}
	m = press(m, "k")
	if m.Selected().ID != "specs/2026-09-17-broken" {
		t.Fatalf("after k: %s", m.Selected().ID)
	}
	m = press(m, "k")
	if m.Selected().ID != "specs/2026-09-17-broken" {
		t.Fatalf("k at the top should stay: %s", m.Selected().ID)
	}
	m = press(m, "ctrl+d")
	if m.openRows()[m.cursor()].id != groupRowID {
		t.Fatalf("ctrl+d should jump a page: %s", m.openRows()[m.cursor()].id)
	}
	m = press(m, "ctrl+u")
	if m.Selected().ID != "specs/2026-09-17-broken" {
		t.Fatalf("ctrl+u should step back a page: %s", m.Selected().ID)
	}
	// The same keys move the Done box.
	m = press(m, tabKey(tabPlans), "tab", "j", "j")
	if m.Selected() == nil || m.Selected().Status != "done" {
		t.Fatalf("the Done box did not move: %v", m.Selected())
	}
}

func TestTopAndBottomKeysInListPanes(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs))
	m = press(m, "G")
	if m.openRows()[m.cursor()].id != groupRowID {
		t.Fatalf("G should land on the last row: %v", rowIDs(m))
	}
	m = press(m, "j")
	if m.openRows()[m.cursor()].id != groupRowID {
		t.Fatal("j past the end should stay on the last row")
	}
	m = press(m, "g", "k")
	if m.Selected().ID != "specs/2026-09-17-broken" {
		t.Fatal("k at the top should stay on the first row")
	}
}

func TestScrollKeysInPaneDetail(t *testing.T) {
	t.Parallel()

	m := press(longModel(t), tabKey(tabPlans), "shift+tab")
	if m.off[paneDetail] != 0 {
		t.Fatalf("the detail starts at the top: %d", m.off[paneDetail])
	}
	m = press(m, "ctrl+d")
	if m.off[paneDetail] != pageLines {
		t.Fatalf("ctrl+d scrolled to %d", m.off[paneDetail])
	}
	m = press(m, "ctrl+u", "ctrl+u")
	if m.off[paneDetail] != 0 {
		t.Fatalf("scroll went below 0: %d", m.off[paneDetail])
	}
	m = press(m, "j", "j")
	if m.off[paneDetail] != 2 {
		t.Fatalf("j should scroll one line: %d", m.off[paneDetail])
	}
	m = press(m, "g")
	if m.off[paneDetail] != 0 {
		t.Fatalf("g should go back to the top: %d", m.off[paneDetail])
	}
	m = press(m, "G")
	if m.off[paneDetail] != m.lastOff(paneDetail) || m.off[paneDetail] == 0 {
		t.Fatalf("G should go to the bottom of the body: %d", m.off[paneDetail])
	}
}

func TestScrollKeysDoNothingInListPanes(t *testing.T) {
	t.Parallel()

	m := press(longModel(t), tabKey(tabPlans), "shift+tab", "ctrl+d", "esc")
	if m.off[paneDetail] != pageLines {
		t.Fatalf("leaving the detail box should keep the body where it was: %d", m.off[paneDetail])
	}
	// A key in a list picks another item, and another item starts the body
	// at its own top. It never scrolls the body the way a key in the detail box
	// does.
	m = press(m, "ctrl+d")
	if m.off[paneDetail] != 0 {
		t.Fatalf("a list pane should not scroll the body: %d", m.off[paneDetail])
	}
	m = press(m, "ctrl+u", "j", "k")
	if m.off[paneDetail] != 0 {
		t.Fatalf("another item should start the body at the top: %d", m.off[paneDetail])
	}
	if m.cursor() != 0 {
		t.Fatalf("j and k should still move: row %d", m.cursor())
	}
}

func TestClickOnARowSelectsItAndFocusesItsPane(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs))
	g := m.geometry()
	// A row takes one line, so the third line of the pane is the third row.
	m = click(m, 2, g.side[paneList].y+1+2)
	if m.focus != paneList || m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// The same works in the Done box, where the second row is the crash fix.
	m = press(newModel(t), tabKey(tabPlans), "tab")
	g = m.geometry()
	m = click(m, 2, g.side[paneDone].y+1+1)
	if m.focus != paneDone || m.Selected().ID != "plans/2026-09-25-crash-fix" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
}

// TestAClickOnTheFirstAndLastVisibleRowLandsOnThatRow reads the rows the pane
// shows and clicks where each of the two ends is drawn, so a click cannot land
// on the neighbour of the row the eye sees.
func TestAClickOnTheFirstAndLastVisibleRowLandsOnThatRow(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		keys []string
		p    pane
	}{
		{"open", []string{tabKey(tabSpecs)}, paneList},
		{"done", []string{tabKey(tabPlans), "tab"}, paneDone},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := sized(press(newModel(t), tc.keys...), 120, 40)
			b := paneBox(m, tc.p)
			rows, _, _ := m.slotOf(tc.p)
			shown := rows[b.first:min(len(rows), b.first+b.rows)]
			// The folded group row holds no item, so the ends are the
			// first and the last item the pane draws.
			first, last := -1, -1
			for i, r := range shown {
				if r.group {
					continue
				}
				if first < 0 {
					first = i
				}
				last = i
			}
			if first < 0 || first == last {
				t.Fatalf("the pane shows %d item rows, want at least two to click", len(shown))
			}
			for name, at := range map[string]int{"first": first, "last": last} {
				after := click(m, 2, b.y+1+at)
				if after.focus != tc.p {
					t.Errorf("a click on the %s row focused pane %d, want %d", name, after.focus, tc.p)
				}
				if got := after.Selected(); got == nil || got.ID != shown[at].id {
					t.Errorf("a click on the %s row selected %v, want %s", name, got, shown[at].id)
				}
			}
		})
	}
}

// TestClickOnATabNameSwitchesTab reads the names the Done pane draws and
// clicks the second one, so a click on a finished sub-tab name lands on that
// sub-tab. The List box has one name, so a click there only takes the focus.
func TestClickOnATabNameSwitchesTab(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabBugs), "tab")
	g := m.geometry()
	done := g.side[paneDone]
	m = click(m, done.x+done.tabs[1].x+1, done.y)
	if m.focus != paneDone || m.done != 1 {
		t.Fatalf("focus %d done sub-tab %d", m.focus, m.done)
	}
	// Clicking the dash before a name belongs to no tab.
	m = click(m, done.x+done.tabs[1].x-1, done.y)
	if m.done != 1 {
		t.Fatalf("a click on the dash moved the sub-tab to %d", m.done)
	}
	// The List box of a tab has one name, so a click on it changes nothing
	// but the focus.
	list := g.side[paneList]
	m = press(m, "[")
	m = click(m, list.x+list.tabs[0].x+1, list.y)
	if m.focus != paneList || m.done != 0 {
		t.Fatalf("a click on the List name gave focus %d and sub-tab %d", m.focus, m.done)
	}
}

func TestClickInsideAPaneOnlyFocusesIt(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs))
	g := m.geometry()
	// Below the last row of the List box is its own border, so use the wide
	// gutter of the left column instead.
	m = click(m, g.side[paneList].x+g.side[paneList].w-1, g.side[paneList].y+g.side[paneList].h-1)
	if m.focus != paneList {
		t.Fatalf("focus %d", m.focus)
	}
	m = click(m, 119, 5)
	if m.focus != paneDetail || m.Selected().ID != "specs/2026-09-17-broken" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// A click on the status line does nothing at all.
	m = click(m, 10, 39)
	if m.focus != paneDetail {
		t.Fatalf("focus %d", m.focus)
	}
}

func TestKeysContinueFromAClickedRow(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs))
	g := m.geometry()
	m = click(m, 2, g.side[paneList].y+1)
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-18-weird" {
		t.Fatalf("j moved to %v", m.Selected())
	}
	m = click(m, 2, g.side[paneList].y+1+1)
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("j did not start from the clicked row: %v", m.Selected())
	}
}

func TestKeyADoesNothing(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	// The cursor is a slice, so it is copied out before the keys: a shared
	// slice would hide a change instead of showing one.
	beforeSel := slices.Clone(m.sel)
	before := m
	after := press(m, "a", "a")
	if strings.Join(rowIDs(after), " ") != strings.Join(rowIDs(before), " ") ||
		strings.Join(doneRowIDs(after), " ") != strings.Join(doneRowIDs(before), " ") ||
		after.focus != before.focus || after.top != before.top || !slices.Equal(after.sel, beforeSel) ||
		after.Selected().ID != before.Selected().ID || after.status != before.status {
		t.Fatalf("a changed the screen: %+v", after)
	}
}

func TestHelpSwallowsKeysUntilItCloses(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), "?")
	if !m.help {
		t.Fatal("? should open the help")
	}
	before := strings.Join(rowIDs(m), " ")
	m = press(m, "j", "k", "]", "3", "/", "a", "t", "n")
	if !m.help || m.focus != paneList || m.top != tabActivities || m.done != 0 || m.searching || m.slug != nil || m.popup != nil {
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
	m = press(m, tabKey(tabSpecs), "?", "esc", "j")
	if m.help || m.Selected().ID != "specs/2026-09-18-weird" {
		t.Fatalf("the keys should work again once the help is closed: %+v", m)
	}
}

func TestAMinuteTickMovesTheClock(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	// The Done box has the focus, so s works on the finished plan it selected.
	m := press(newModel(t), tabKey(tabPlans), "tab", "s")
	if m.popup == nil || m.popup.field != "status" {
		t.Fatalf("the Done pane popup %+v", m.popup)
	}
	if m.popup.idx != 3 {
		t.Fatalf("the popup should start on the row's own status done, got %d", m.popup.idx)
	}
	// The same key with a List box focused works on that box's row instead.
	other := press(newModel(t), tabKey(tabSpecs), "j", "j", "s")
	if other.popup == nil || other.popup.field != "status" || other.popup.idx != 2 {
		t.Fatalf("the Specs box popup %+v", other.popup)
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

// TestGeometryPlacesThePanes reads the screen a reader sees: the tab bar
// owns the top line, the two boxes of a kind tab share the lines under it,
// the detail box sits beside them, and Activities draws one box where a kind
// tab draws two.
func TestGeometryPlacesThePanes(t *testing.T) {
	t.Parallel()

	// 120 columns, the normal width, on the Plans tab: two boxes on the left
	// and the detail box on the right, all starting under the three lines of
	// the tab box.
	m := press(sized(newModel(t), 120, 40), tabKey(tabPlans))
	g := m.geometry()
	if !g.wide || g.leftW != 40 {
		t.Fatalf("wide %v leftW %d", g.wide, g.leftW)
	}
	want := [][4]int{{0, barRows, 40, 18}, {0, barRows + 18, 40, 18}}
	for p, w := range want {
		if got := g.side[pane(p)]; got.x != w[0] || got.y != w[1] || got.w != w[2] || got.h != w[3] {
			t.Fatalf("pane %d %+v, want %v", p, got, w)
		}
	}
	if g.detail.x != 40 || g.detail.y != barRows || g.detail.w != 80 || g.detail.h != 36 {
		t.Fatalf("the detail box %+v", g.detail)
	}
	// The border takes two lines and a row one, so a box shows as many rows
	// as it has room for.
	if g.side[paneList].inner != 16 || g.side[paneList].rows != 16 || g.side[paneList].first != 0 {
		t.Fatalf("the List box holds %+v", g.side[paneList])
	}
	if g.side[paneDone].inner != 16 || g.side[paneDone].rows != 16 || g.side[paneDone].first != 0 {
		t.Fatalf("the Done box holds %+v", g.side[paneDone])
	}
	// Both names of the Done pane are still in the drawn title at this
	// normal width.
	if title := strings.Split(plain(m.View()), "\n")[g.side[paneDone].y]; !strings.Contains(title, "Dropped") {
		t.Fatalf("the Done title at 120 columns is missing Dropped: %q", title)
	}
	// Activities has no Done pane, so it draws one box where Plans draws
	// two, and that box has the whole body height.
	a := sized(newModel(t), 120, 40).geometry()
	if len(a.side) != 1 || a.side[0].y != barRows || a.side[0].h != 36 {
		t.Fatalf("the Activities box is %+v, want one box of the full height", a.side)
	}
	// The left column is a third of the width, the way lazygit sizes its
	// side panels, and never under 28 columns.
	for _, w := range []int{60, 80, 100, 120, 150, 200, 240} {
		if got := sized(newModel(t), w, 40).geometry().leftW; got != max(w/3, 28) {
			t.Errorf("at %d columns the left column is %d", w, got)
		}
	}
	// Below 60 columns only the focused pane is on screen, full width.
	n := sized(press(newModel(t), tabKey(tabSpecs), "tab"), 40, 20).geometry()
	if n.wide || n.full.w != 40 || n.full.h != 16 {
		t.Fatalf("narrow screen %+v", n)
	}
	if n.side[paneList].w != 0 || n.detail.w != 0 {
		t.Fatal("a narrow screen should only draw the focused pane")
	}
	if n.full.x != 0 || n.full.y != barRows || n.full.w != 40 || n.full.h != 16 {
		t.Fatalf("the focused pane should take the whole screen: %+v", n.full)
	}
	// The click boxes sit where the drawn title puts the names, so a click
	// lands on the word it points at. The place comes from the view, not
	// from a number written down here.
	for p, names := range map[pane][]string{paneList: m.tabsOf(paneList), paneDone: m.tabsOf(paneDone)} {
		title := strings.Split(plain(m.View()), "\n")[g.side[p].y]
		for i, name := range names {
			at := cellAt(t, title, name)
			if got := g.side[p].tabs[i]; got.x != at || got.w != len(name) {
				t.Fatalf("pane %d name %q sits at %+v, but the title draws it at %d: %q", p, name, got, at, title)
			}
		}
		if got := len(g.side[p].tabs); got != len(names) {
			t.Fatalf("pane %d has %d click boxes, want %d", p, got, len(names))
		}
	}
	if len(g.detail.tabs) != 0 {
		t.Fatalf("the detail box has %d click boxes, want none", len(g.detail.tabs))
	}
}

// The model measures a pane to clamp an offset with, also on a narrow screen
// where only the focused pane is drawn. That measurement has to be the box
// under the tab box and no taller, or the list scrolls past the lines the
// screen has and the reader cannot see the row the cursor is on.
func TestNarrowScreenMeasuresTheBoxUnderTheTabBox(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		for _, size := range [][2]int{{20, 5}, {40, 8}, {40, 20}, {59, 40}, {59, 12}} {
			w, h := size[0], size[1]
			m := press(sized(newModel(t), w, h), tabKey(i))
			g := m.geometry()
			if g.wide {
				t.Fatalf("%d columns is wide, this test is about the narrow screen", w)
			}
			for _, p := range m.panes() {
				b := m.boxOf(p)
				if b.y != barRows {
					t.Errorf("tab %s at %dx%d: the model measures pane %d on line %d, want %d", topTabs[i].name, w, h, p, b.y, barRows)
				}
				if b.h != g.full.h {
					t.Errorf("tab %s at %dx%d: the model measures pane %d %d lines, the drawn box has %d", topTabs[i].name, w, h, p, b.h, g.full.h)
				}
				if got := m.fitOf(p); got != g.full.rows {
					t.Errorf("tab %s at %dx%d: the model gives pane %d %d rows, the drawn box has %d", topTabs[i].name, w, h, p, got, g.full.rows)
				}
			}
			// Walking to the end of a long list has to leave the last row on
			// screen, because the pane says that many rows fit.
			full := press(sized(longModel(t), w, h), tabKey(i), "G")
			drawn := paneRows(full, full.focus)
			if len(drawn) != full.fitOf(full.focus) {
				t.Fatalf("tab %s at %dx%d: the pane draws %d rows and says %d fit", topTabs[i].name, w, h, len(drawn), full.fitOf(full.focus))
			}
			if len(drawn) == 0 {
				// A box one line high is its title bar and no rows at all, so
				// there is no last line to read.
				continue
			}
			rows, _, _ := full.slotOf(full.focus)
			if len(rows) == 0 {
				continue
			}
			if last := rows[len(rows)-1]; !isThatRow(full, last, drawn[len(drawn)-1]) {
				t.Errorf("tab %s at %dx%d: the last line holds %q, want %q", topTabs[i].name, w, h, drawn[len(drawn)-1], headOf(full, last))
			}
		}
	}
}

// TestSelectionIsPerTab walks every tab of the bar with its own cursor, so
// opening another tab can never move the row one tab had selected.
func TestSelectionIsPerTab(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		tab  int
		keys []string
		want string
	}{
		{tabSpecs, []string{"j"}, "specs/2026-09-18-weird"},
		{tabBugs, []string{"j"}, "bugs/2026-09-26-open"},
		{tabPlans, []string{"tab", "j"}, "plans/2026-09-25-crash-fix"},
		{tabScratches, []string{"j"}, "scratch/2026-09-28-idea-raw"},
		{tabDebts, nil, "debt/2026-09-27-orphan-debt#item-1"},
		{tabActivities, []string{"j"}, "plans/2026-09-21-alpha#task-2"},
	} {
		m := press(newModel(t), append([]string{tabKey(tc.tab)}, tc.keys...)...)
		// A walk over every other tab must not move this one.
		for i := range topTabs {
			if i == tc.tab {
				continue
			}
			m = press(m, tabKey(i), tabKey(tc.tab))
		}
		if got := m.Selected(); got == nil || got.ID != tc.want {
			t.Errorf("tab %s lost its selection: %v, want %s", topTabs[tc.tab].name, got, tc.want)
		}
	}
}

func TestSelectedFollowsTheLastFocusedListPane(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabPlans), "tab", "j", "tab")
	if m.Selected().ID != "plans/2026-09-25-crash-fix" {
		t.Fatalf("the detail should keep showing the Done box: %v", m.Selected())
	}
	// Moving the cursor inside the Done box keeps it the one on show.
	m = press(m, "shift+tab", "j", "tab")
	if m.Selected().ID != "plans/2026-09-26-dash-tasks" {
		t.Fatalf("the detail should keep showing the Done box: %v", m.Selected())
	}
	// The List box takes over as soon as it has the focus.
	m = press(m, "tab")
	if m.Selected().ID != "plans/2026-09-21-alpha" {
		t.Fatalf("the Plans box should show its own row: %v", m.Selected())
	}
}

func TestUntypedGroupRow(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs))
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
	if got := rowIDs(press(newModel(t), tabKey(tabPlans))); len(got) != 2 {
		t.Fatalf("plans rows %v", got)
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	// A task takes its status from its checkboxes, so s on a task row says
	// so instead of opening the picker. Open a plan to reach a task row.
	m := press(newModel(t), tabKey(tabPlans), " ", "j", "s")
	if m.popup != nil || !strings.Contains(m.status, "checkboxes") {
		t.Fatalf("task popup: %v %q", m.popup, m.status)
	}
	m = press(newModel(t), tabKey(tabSpecs), "G", "enter", "j", "t")
	if m.popup != nil || !strings.Contains(m.status, "legacy") {
		t.Fatalf("legacy popup: %v %q", m.popup, m.status)
	}
}

func TestStatusPopupSetsValue(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	var got []string
	m.setValue = func(id, field, value string) (write.Outcome, error) {
		got = []string{id, field, value}
		return write.Outcome{Committed: true}, nil
	}
	m = press(m, tabKey(tabBugs), "j", "s")
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
	t.Parallel()

	m := press(newModel(t), tabKey(tabBugs), "t", "esc")
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
	t.Parallel()

	cfg, b := fixture(t)
	m := press(newModel(t), tabKey(tabSpecs), "j") // the second spec of the oldest date
	next, _ := m.Update(reloadMsg{b: b})
	m = next.(Model)
	if m.Selected().ID != "specs/2026-09-18-weird" {
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
	t.Parallel()

	next, _ := newModel(t).Update(WatchFailed(errors.New("too many files")))
	m := next.(Model)
	if !m.manual || !strings.Contains(m.status, "press r") {
		t.Fatalf("manual %v status %q", m.manual, m.status)
	}
}

func TestNewBugSlugInput(t *testing.T) {
	t.Parallel()

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
		if w == 60 || max(w/3, 28) != max((w-1)/3, 28) {
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
	b := m.geometry().at(p)
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

// doneTabBefore puts the Done pane on another finished sub-tab than that one,
// so a click that lands on nothing cannot pass by leaving the sub-tab where it
// already was.
func doneTabBefore(t *testing.T, m Model, i int) Model {
	t.Helper()
	n := len(m.doneTabNames())
	other := (i + 1) % n
	for range n {
		if m.done == other {
			break
		}
		m = press(m, "]")
	}
	if m.done != other {
		t.Fatalf("could not open the finished sub-tab %d", other)
	}
	return m
}

// TestClickLandsOnEveryDrawnTabName is the property behind the mouse: the first
// and the last letter of every finished sub-tab name the Done pane draws, of
// every tab of the bar, at every window size from 60 to 200 columns, must
// switch to that sub-tab; a click on the dashes between two names must switch
// nothing. The letters come from the drawn title, so a click box that drifts
// from the view fails here.
func TestClickLandsOnEveryDrawnTabName(t *testing.T) {
	t.Parallel()

	widths := clickWidths()
	clicked := map[string]bool{}

	for i := range topTabs {
		if len(topTabs[i].done) == 0 {
			continue
		}
		for _, w := range widths {
			for d, name := range topTabs[i].done {
				for _, end := range []string{"first", "last"} {
					m := press(sized(newModel(t), w, 40), tabKey(i), "tab")
					m = doneTabBefore(t, m, d)
					y, x, ok := drawnLetter(m, paneDone, name.name, end)
					if !ok {
						t.Fatalf("at %d columns the Done pane of %s does not draw %q", w, topTabs[i].name, name.name)
					}
					if got, _, idx := m.hit(x, y); got != paneDone || idx != d {
						t.Fatalf("at %d columns the %s letter of %q maps to box %d sub-tab %d, want sub-tab %d", w, end, name.name, got+1, idx, d)
					}
					if got := click(m, x, y).done; got != d {
						t.Fatalf("at %d columns a click on the %s letter of %q gave sub-tab %d, want %d", w, end, name.name, got, d)
					}
					clicked[name.name] = true
				}
			}
			// The top box has one name, so a click on it only takes the
			// focus and changes no sub-tab.
			m := press(sized(newModel(t), w, 40), tabKey(i), "tab")
			m = press(m, "]")
			name := m.tabsOf(paneList)[0]
			y, x, ok := drawnLetter(m, paneList, name, "first")
			if !ok {
				t.Fatalf("at %d columns the top pane of %s draws no %s name", w, topTabs[i].name, name)
			}
			got := click(m, x, y)
			if got.focus != paneList || got.done != 1 {
				t.Fatalf("a click on the top name of %s gave focus %d and sub-tab %d, want the top box on %d", topTabs[i].name, got.focus, got.done, 1)
			}
		}
	}
	for i := range topTabs {
		for _, d := range topTabs[i].done {
			if !clicked[d.name] {
				t.Errorf("no width from 60 to 200 drew %q, so it was never clicked", d.name)
			}
		}
	}

	// The dashes between two names belong to no sub-tab.
	for i := range topTabs {
		if len(topTabs[i].done) < 2 {
			continue
		}
		for _, w := range widths {
			m := press(sized(newModel(t), w, 40), tabKey(i), "tab")
			m = doneTabBefore(t, m, 1)
			before := m.done
			y, last, ok := drawnLetter(m, paneDone, m.doneTabNames()[0], "last")
			if !ok {
				t.Fatalf("with %q open, at %d columns the Done pane draws no %q", topTabs[i].name, w, m.doneTabNames()[0])
			}
			x := last + 1 // the separator cell right after the name, dash or plain space
			if _, _, idx := m.hit(x, y); idx != -1 {
				t.Fatalf("at %d columns the dash after %q maps to sub-tab %d", w, m.doneTabNames()[0], idx)
			}
			if got := click(m, x, y).done; got != before {
				t.Fatalf("at %d columns a click on the dash after %q moved to sub-tab %d", w, m.doneTabNames()[0], got)
			}
		}
	}
}

// TestTabTitleAlwaysShowsTheOpenTab is the Task 8 regression, still true for
// the Done pane: a title must never drop the open sub-tab's own name off the
// end, so no sub-tab looks unselected and a click on its old spot falls into
// the list. This checks it for every tab, at every width the review named.
func TestTabTitleAlwaysShowsTheOpenTab(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		if len(topTabs[i].done) == 0 {
			continue
		}
		for _, w := range []int{60, 80, 100, 120, 140, 200} {
			for d := range len(topTabs[i].done) {
				m := press(sized(newModel(t), w, 40), tabKey(i), "tab")
				m = doneTabBefore(t, m, d)
				title := strings.Split(plain(m.View()), "\n")[m.geometry().at(paneDone).y]
				if name := topTabs[i].done[d].name; !strings.Contains(title, name) {
					t.Fatalf("at %d columns with %q open, the title drops the open sub-tab: %q", w, name, title)
				}
				for other := range len(topTabs[i].done) {
					for _, end := range []string{"first", "last"} {
						y, x, ok := drawnLetter(m, paneDone, topTabs[i].done[other].name, end)
						if !ok {
							continue // a narrow title drops names other than the open one
						}
						if got := click(m, x, y).done; got != other {
							t.Fatalf("at %d columns a click on the %s letter of %q gave sub-tab %d, want %d", w, end, topTabs[i].done[other].name, got, other)
						}
					}
				}
			}
		}
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

// openGroupIDs reads the item ids of the open list with the group row
// dropped, so a test sees only the item order.
func openGroupIDs(m Model) (going, rest []string) {
	for _, r := range m.openRows() {
		if r.group {
			continue
		}
		if inProgress(m.board.Get(r.id)) {
			going = append(going, r.id)
		} else {
			rest = append(rest, r.id)
		}
	}
	return going, rest
}

// TestEveryTabHoldsItsOpenItemsInFileDateOrder checks each tab against the
// board: the pane lists exactly the open items of its own kind, oldest file
// date first.
func TestEveryTabHoldsItsOpenItemsInFileDateOrder(t *testing.T) {
	t.Parallel()

	m := splitModel(t)
	for i := range topTabs {
		kind := topTabs[i].kind
		if kind == "" {
			// Activities lists the tasks under way under a head, so the rows
			// that are a task or stand on their own are the ones to check.
			var want []string
			for _, it := range m.board.List(board.KindTask, true) {
				if inProgress(it) {
					want = append(want, it.ID)
				}
			}
			slices.Sort(want)
			var got []string
			for _, r := range press(m, tabKey(i)).rowsOf(paneList) {
				if r.depth == 1 || !r.tree {
					got = append(got, r.id)
				}
			}
			slices.Sort(got)
			if strings.Join(got, " ") != strings.Join(want, " ") {
				t.Errorf("tab %s holds %q, want %q", topTabs[i].name, got, want)
			}
			continue
		}
		var want []string
		for _, it := range ordered(m.board.List(kind, false), false) {
			want = append(want, it.ID)
		}
		if kind == board.KindStory {
			want = append(want, groupRowID)
		}
		mm := press(m, tabKey(i))
		if got := strings.Join(ids(mm.rowsOf(paneList)), " "); got != strings.Join(want, " ") {
			t.Errorf("tab %s holds %q, want %q", topTabs[i].name, got, strings.Join(want, " "))
		}
	}
}

func TestStartedTaskWithNoTicksCountsAsInProgress(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	lonely := m.board.Get("plans/2026-09-23-lonely#task-1")
	if lonely.Done != 0 {
		t.Fatalf("the lonely task should have no ticked box, got %d", lonely.Done)
	}
	lonely.Started = true
	lonely.Status = "in-progress"
	// Oldest first puts alpha before lonely, so step down to lonely before
	// opening it.
	m = press(m, tabKey(tabPlans), "j", " ")
	rows := ids(m.openRows())
	// Both tasks are under way, and neither one hides behind a rule row.
	if len(rows) != 3 || !slices.Contains(rows, lonely.ID) {
		t.Fatalf("the started task is missing from the Plans tree, rows %v", rows)
	}
	// A task under way is work in progress, so Activities lists it too.
	if got := ids(press(m, tabKey(tabActivities)).rowsOf(paneList)); !slices.Contains(got, lonely.ID) {
		t.Errorf("Activities does not list the started task: %v", got)
	}
}

// A task someone started before ticking a box is work under way, so its row
// ends on its count and its agent, the same as a task with a box ticked. The
// words read in the plain foreground: work under way is no longer the accent,
// and it is not faint either.
func TestAStartedTaskKeepsTheAccentAndItsCountAndAgent(t *testing.T) {
	withColors(func() {
		cfg := treeCfg(t, map[string]string{
			".acta/plans/2026-09-21-a.md": "# A plan\n\n### Task 1: One\n- [ ] a\n- [ ] b\n\n### Task 2: Two\n- [ ] c\n",
			".acta/.agents.json":          `{"plans/2026-09-21-a#task-1": {"agent": "omp", "at": "2026-09-26T09:00:00+07:00", "started": true}}`,
		})
		b, err := board.Load(cfg)
		if err != nil {
			t.Fatal(err)
		}
		m := New(cfg, b, true)
		m.render = func(md string, _ int) string { return md }
		m = press(sized(m, 200, 40), tabKey(tabPlans), " ", "j", "j")
		it := b.Get("plans/2026-09-21-a#task-1")
		if it == nil || it.Status != "in-progress" || it.Done != 0 || it.Total != 2 {
			t.Fatalf("the started task = %+v, want in-progress at 0 of 2", it)
		}
		bx := paneBox(m, paneList)
		lines := innerLines(m, bx)
		if len(lines) < 4 {
			t.Fatalf("the pane drew %d lines: %q", len(lines), lines)
		}
		if got := plain(lines[1]); !strings.HasSuffix(strings.TrimRight(got, " "), "0/2 · omp") {
			t.Errorf("the started row is %q, want it to end on 0/2 · omp", got)
		}
		if got := plain(lines[2]); !strings.Contains(got, "Two") {
			t.Errorf("line 2 is %q, want the second task", got)
		}
		// The cursor has moved on, so the started task is drawn like every
		// other in-progress row the reader is not on: plain, neither faint
		// nor the accent.
		if row := paintedLine(m.View(), bx, 1); wears(row, 2) || wears(row, 38, 5, 111) {
			t.Errorf("the started row should read plain, not faint and not the accent: %q", row)
		}
	})
}

func TestUnknownStatusIsListedAndNotInProgress(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs))
	found := false
	for _, r := range m.openRows() {
		if r.id != "specs/2026-09-18-weird" {
			continue
		}
		found = true
		if inProgress(m.board.Get(r.id)) {
			t.Error("a story with a status nobody knows is not work in progress")
		}
	}
	if !found {
		t.Fatal("the bogus story is missing from the Specs box")
	}
}

func TestNoPaneEverDrawsARuleRow(t *testing.T) {
	t.Parallel()

	// Whatever mix of items a list holds, the rows are the items themselves.
	keep := func(m Model, want func(*board.Item) bool) Model {
		var items []*board.Item
		for _, it := range m.board.Items {
			if it.Kind == board.KindStory && !it.Legacy && want(it) {
				items = append(items, it)
			}
		}
		// Get no longer finds them, but this test only reads the row ids.
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
		{"only not-started", func(it *board.Item) bool { return it.Status != "in-progress" }, 4},
		{"neither", func(*board.Item) bool { return false }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows := keep(press(full, tabKey(tabSpecs)), tc.want).openRows()
			if len(rows) != tc.n {
				t.Fatalf("rows %v, want %d", ids(rows), tc.n)
			}
		})
	}
	// A pane that holds both kinds of work at once still lists only items.
	rows := splitModel(t, tabKey(tabSpecs)).openRows()
	var going, waiting bool
	for _, r := range rows {
		it := full.board.Get(r.id)
		if it == nil {
			continue
		}
		if inProgress(it) {
			going = true
		} else {
			waiting = true
		}
	}
	if !going || !waiting {
		t.Fatalf("this case needs a pane that mixes both, got %v", ids(rows))
	}
	for _, r := range rows {
		if r.id != groupRowID && full.board.Get(r.id) == nil {
			t.Fatalf("the pane drew a row that is neither an item nor the group row: %v", ids(rows))
		}
	}
}

func TestSearchListsEveryMatch(t *testing.T) {
	t.Parallel()

	m := splitModel(t)
	m = press(m, tabKey(tabSpecs), "/", "A", "l", "p", "h", "a")
	rows := ids(m.openRows())
	if len(rows) == 0 {
		t.Fatal("the search found nothing")
	}
	for _, id := range rows {
		if m.board.Get(id) == nil {
			t.Fatalf("the search listed %q, which is no item", id)
		}
	}
}

func TestClickOnTheSpaceBelowTheRowsKeepsTheSelection(t *testing.T) {
	t.Parallel()

	m := press(sized(splitModel(t), 120, 40), tabKey(tabSpecs))
	g := m.geometry()
	b := g.at(paneList)
	below := b.y + 1 + b.rows
	if below >= b.y+b.h {
		t.Skip("the pane is full, so there is no space below the rows")
	}
	before := m.openRows()[m.cursor()].id
	after := click(m, 2, below)
	if got := after.openRows()[after.cursor()].id; got != before {
		t.Fatalf("a click on the empty space moved from %s to %s", before, got)
	}
	if after.focus != paneList {
		t.Fatalf("a click on the empty space moved the focus to pane %d", after.focus)
	}
}

// TestEnterFocusesDetailFromBothListPanes walks every tab of the bar and both
// of its list panes, so enter reaches the detail from anywhere a reader can
// put the focus.
func TestEnterFocusesDetailFromBothListPanes(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		for _, p := range press(newModel(t), tabKey(i)).panes() {
			// Enter on a plan row opens the plan instead, and a task row is
			// read from the tree, so both are covered by the tree tests in
			// plantree_test.go. A box that lists nothing has no row to move
			// from either.
			m := press(newModel(t), tabKey(i))
			m.focusPane(p)
			if rows, _, _ := m.slotOf(p); len(rows) == 0 {
				continue
			}
			if it := m.Selected(); it != nil && it.Kind == board.KindPlan {
				continue
			}
			if it := m.Selected(); it != nil && it.Kind == board.KindTask {
				// A task row has no detail of its own to open into either,
				// so it is read from the Activities tab below.
				continue
			}
			id := m.Selected().ID
			next, cmd := m.Update(key("enter"))
			m = next.(Model)
			if cmd != nil {
				t.Errorf("tab %s pane %d: enter opened the editor", topTabs[i].name, p)
			}
			if m.focus != paneDetail {
				t.Errorf("tab %s pane %d: enter left the focus on pane %d", topTabs[i].name, p, m.focus)
			}
			if m.last != p {
				t.Errorf("tab %s pane %d: enter left the detail showing pane %d", topTabs[i].name, p, m.last)
			}
			if m.Selected() == nil || m.Selected().ID != id {
				t.Errorf("tab %s pane %d: enter moved from %s to %v", topTabs[i].name, p, id, m.Selected())
			}
		}
	}
}

func TestEnterOnTheGroupRowStillToggles(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabSpecs), "G", "enter")
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
	t.Parallel()

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
	m = press(sized(m, 120, 40), tabKey(tabBugs), "j") // oldest first: the branch bug of 09-25 is the second row
	next, cmd := m.Update(key("enter"))
	m = next.(Model)
	if extra := notToast(cmd); len(extra) > 0 {
		t.Fatalf("enter on a branch item opened something: %T", extra[0])
	}
	if !strings.Contains(m.status, "git worktree add") {
		t.Fatalf("enter on a branch item should warn: %q", m.status)
	}
	if m.focus != paneDetail {
		t.Fatalf("enter on a branch item should focus the detail, got pane %d", m.focus)
	}
}

// TestEOpensTheEditorFromEveryPane walks every tab of the bar and both of
// its list panes, plus the detail box, so no box the reader can reach refuses
// the editor.
func TestEOpensTheEditorFromEveryPane(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		for _, p := range press(newModel(t), tabKey(i)).panes() {
			m := press(newModel(t), tabKey(i))
			m.focusPane(p)
			it := m.Selected()
			if it == nil {
				continue
			}
			if _, cmd := m.Update(key("e")); cmd == nil {
				t.Errorf("tab %s box %d: e opened no editor", topTabs[i].name, p+1)
			}
		}
		// The detail box of this tab, from a list pane that has the focus.
		if _, cmd := press(newModel(t), tabKey(i), "shift+tab").Update(key("e")); cmd == nil {
			t.Errorf("tab %s: e in the detail box opened no editor", topTabs[i].name)
		}
	}
}

func TestEOnABranchItemWarns(t *testing.T) {
	t.Parallel()

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
	m = press(sized(m, 120, 40), tabKey(tabBugs), "j") // oldest first: the branch bug of 09-25 is the second row
	next, cmd := m.Update(key("e"))
	m = next.(Model)
	if extra := notToast(cmd); len(extra) > 0 {
		t.Fatalf("e on a branch item opened something: %T", extra[0])
	}
	if !strings.Contains(m.status, "git worktree add") {
		t.Fatalf("e on a branch item should warn: %q", m.status)
	}
}

// TestEscInDetailReturnsToTheListPane walks out of the detail box from both
// list panes of a tab, so esc always lands back on the list the reader came
// from.
func TestEscInDetailReturnsToTheListPane(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone} {
		m := press(newModel(t), tabKey(tabSpecs))
		m.focusPane(p)
		m = press(m, "enter")
		if m.focus != paneDetail || m.last != p {
			t.Fatalf("enter from pane %d: focus %d last %d", p, m.focus, m.last)
		}
		m = press(m, "esc")
		if m.focus != p {
			t.Fatalf("esc from pane %d: focus %d", p, m.focus)
		}
	}
}

func TestEscInDetailKeepsTheSelection(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), "j", "enter")
	id := m.Selected().ID
	m = press(m, "esc")
	if m.Selected() == nil || m.Selected().ID != id {
		t.Fatalf("esc moved from %s to %v", id, m.Selected())
	}
}

// TestZeroChangesNothing walks every tab of the bar and every box of every
// tab, and checks the 0 key leaves the screen exactly as it found it, so no
// reader can wander into a dead key.
func TestZeroChangesNothing(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		for _, p := range append(press(newModel(t), tabKey(i)).panes(), paneDetail) {
			m := press(newModel(t), tabKey(i))
			m.focusPane(p)
			before := sized(m, 160, 40).View()
			got := press(m, "0")
			if got.focus != m.focus || got.top != m.top {
				t.Errorf("tab %s box %d: 0 moved the focus to %d", topTabs[i].name, p, got.focus)
			}
			if after := sized(got, 160, 40).View(); after != before {
				t.Errorf("tab %s box %d: 0 redrew the screen", topTabs[i].name, p)
			}
		}
	}
}

// TestTabRingStillReachesTheDetailFromEveryListBox walks the ring from every
// list box of every tab, so dropping the 0 key never strands a reader who
// wants the detail.
func TestTabRingStillReachesTheDetailFromEveryListBox(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		for _, p := range press(newModel(t), tabKey(i)).panes() {
			m := press(newModel(t), tabKey(i))
			m.focusPane(p)
			reached := false
			for _, k := range []string{"tab", "shift+tab"} {
				if press(m, k).focus == paneDetail {
					reached = true
				}
			}
			if !reached {
				t.Errorf("tab %s box %d: neither tab nor shift+tab reaches the detail", topTabs[i].name, p)
			}
		}
	}
}

// TestEnterOpensTheDetailOnAPlainRow checks the second way into the detail
// still works: on a row that is not a tree row, enter leaves the list alone
// and moves the focus to the detail.
func TestEnterOpensTheDetailOnAPlainRow(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		for _, p := range press(newModel(t), tabKey(i)).panes() {
			m := press(newModel(t), tabKey(i))
			m.focusPane(p)
			rows := m.rowsOf(p)
			row := -1
			for k, r := range rows {
				if !r.tree && !r.group && m.board.Get(r.id) != nil {
					row = k
					break
				}
			}
			if row < 0 {
				continue
			}
			m.moveTo(row)
			before := rowIDs(m)
			m = press(m, "enter")
			if m.focus != paneDetail || m.last != p {
				t.Errorf("tab %s box %d: enter on a plain row gave focus %d last %d", topTabs[i].name, p, m.focus, m.last)
			}
			if after := rowIDs(m); !slices.Equal(after, before) {
				t.Errorf("tab %s box %d: enter on a plain row changed the list to %v", topTabs[i].name, p, after)
			}
		}
	}
}

// TestNoTitleOrHelpNamesKeyZero reads every tab's screen and the help text, so
// no key map or box title sends a reader after a key that is gone.
func TestNoTitleOrHelpNamesKeyZero(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		if v := plain(press(sized(newModel(t), 160, 40), tabKey(i)).View()); strings.Contains(v, "[0]") {
			t.Errorf("tab %s still draws [0]", topTabs[i].name)
		}
	}
	for _, ln := range strings.Split(helpLines, "\n") {
		if strings.HasPrefix(ln, "0 ") {
			t.Errorf("help still lists 0: %q", ln)
		}
	}
}

// activityFiles is a board with work under a spec plan, under a bug plan,
// and a plan with no work begun.
func activityFiles() map[string]string {
	return map[string]string{
		".acta/specs/2026-09-20-s.md": "---\nid: SPC-0001\n---\n# Spec S\n",
		".acta/bugs/2026-09-21-b.md":  "---\nid: BUG-0002\n---\n# Bug B\n",
		".acta/plans/2026-09-22-p.md": "---\nid: PLN-0003\nparent: specs/2026-09-20-s\n---\n# Plan P\n\n### Task 1: Going\n- [x] a\n- [ ] b\n\n### Task 2: Waiting\n- [ ] c\n\n### Task 3: Also going\n- [x] d\n- [ ] e\n",
		".acta/plans/2026-09-23-q.md": "---\nid: PLN-0004\nparent: bugs/2026-09-21-b\n---\n# Plan Q\n\n### Task 1: Fixing\n- [x] f\n- [ ] g\n",
		".acta/plans/2026-09-24-r.md": "---\nid: PLN-0005\n---\n# Plan R\n\n### Task 1: Not begun\n- [ ] h\n",
	}
}

// actModel opens Activities over activityFiles.
func actModel(t *testing.T) Model {
	t.Helper()
	cfg := treeCfg(t, activityFiles())
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return press(sized(m, 160, 40), tabKey(tabActivities))
}

// A task under way sits under the plan it belongs to, or under the bug that
// plan fixes, one level up only, and every other task is left out.
func TestActivitiesGroupsTasksUnderTheirParent(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	want := []string{
		"bugs/2026-09-21-b",
		"plans/2026-09-23-q#task-1",
		"plans/2026-09-22-p",
		"plans/2026-09-22-p#task-1",
		"plans/2026-09-22-p#task-3",
	}
	rows := m.rowsOf(paneList)
	if got := ids(rows); !slices.Equal(got, want) {
		t.Fatalf("rows %q, want %q", got, want)
	}
	for _, r := range rows {
		it := m.board.Get(r.id)
		switch {
		case !r.tree:
			t.Errorf("%s is not a tree row", r.id)
		case r.depth == 0 && it.Kind == board.KindTask:
			t.Errorf("task %s sits at the top level", r.id)
		case r.depth == 1 && !inProgress(it):
			t.Errorf("%s is listed but not in progress", r.id)
		case r.depth > 1:
			t.Errorf("%s goes deeper than one level", r.id)
		}
	}
	// The heads keep the order of the pane, and the task of the bug stays
	// under the bug, so the plan under the spec lands on top when the pane is
	// flipped.
	flipped := []string{
		"plans/2026-09-22-p",
		"plans/2026-09-22-p#task-1",
		"plans/2026-09-22-p#task-3",
		"bugs/2026-09-21-b",
		"plans/2026-09-23-q#task-1",
	}
	if got := ids(press(m, "o").rowsOf(paneList)); !slices.Equal(got, flipped) {
		t.Errorf("newest first holds %q, want %q", got, flipped)
	}
}

// enter shuts one Activities group and opens it again, and leaves the other
// groups and the Plans tab as they were.
func TestEnterShutsOneActivitiesGroup(t *testing.T) {
	t.Parallel()

	m := press(actModel(t), "enter")
	want := []string{"bugs/2026-09-21-b", "plans/2026-09-22-p", "plans/2026-09-22-p#task-1", "plans/2026-09-22-p#task-3"}
	if got := ids(m.rowsOf(paneList)); !slices.Equal(got, want) {
		t.Fatalf("after enter on the bug: %q, want %q", got, want)
	}
	if m.focus == paneDetail {
		t.Error("enter on a head moved the focus to the detail")
	}
	if got := len(press(m, "enter").rowsOf(paneList)); got != 5 {
		t.Errorf("a second enter left %d rows, want 5", got)
	}
	// A task row is not a head: enter opens it in the detail. The last row of
	// the board is a task, not a head.
	if got := press(actModel(t), "G", "enter"); got.focus != paneDetail {
		t.Error("enter on a task row did not open the detail")
	}
	// A group the reader shut on Activities leaves the Plans tab alone: no
	// plan opens there on its own.
	for _, r := range press(press(m, "j", "enter"), tabKey(tabPlans)).rowsOf(paneList) {
		if r.depth > 0 {
			t.Errorf("the Plans tab opened %s on its own", r.id)
		}
	}
}

// A task whose plan is not on the board has no head, so it stands on its own
// row at the end instead of going missing.
func TestATaskWithNoPlanStandsOnItsOwnRow(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.board.Get("plans/2026-09-22-p#task-1").PlanID = ""
	rows := m.rowsOf(paneList)
	last := rows[len(rows)-1]
	if last.id != "plans/2026-09-22-p#task-1" {
		t.Fatalf("the last row is %q, want the task with no plan", last.id)
	}
	if last.depth != 0 || last.tree {
		t.Errorf("the task with no plan is a row of depth %d tree %v, want a plain row at the top level", last.depth, last.tree)
	}
}

func TestHAndLFoldActivitiesGroups(t *testing.T) {
	t.Parallel()

	m := press(actModel(t), "j") // the task under the bug
	back := press(m, "h")
	want := []string{"bugs/2026-09-21-b", "plans/2026-09-22-p", "plans/2026-09-22-p#task-1", "plans/2026-09-22-p#task-3"}
	if got := ids(back.rowsOf(paneList)); !slices.Equal(got, want) {
		t.Fatalf("h on a task row: %q, want %q", got, want)
	}
	if it := back.Selected(); it == nil || it.ID != "bugs/2026-09-21-b" {
		t.Errorf("h left the cursor on %v, want the bug", it)
	}
	if got := len(press(back, "l").rowsOf(paneList)); got != 5 {
		t.Errorf("l on the shut bug left %d rows, want 5", got)
	}
	// l on a task row under a head does not open anything, and the group it
	// sits under keeps its tasks.
	open := press(actModel(t), "j", "j", "j", "j")
	if got := rowIDs(press(open, "l")); len(got) != 5 {
		t.Errorf("l on a task row left %q, want all 5 rows", got)
	}
	// h on a task row shuts the head right above it, not the first head of
	// the list: the last task row sits under the plan, not under the bug.
	shut := press(open, "h")
	if got, want := ids(shut.rowsOf(paneList)), []string{"bugs/2026-09-21-b", "plans/2026-09-23-q#task-1", "plans/2026-09-22-p"}; !slices.Equal(got, want) {
		t.Errorf("h on the last task row: %q, want %q", got, want)
	}
	if it := shut.Selected(); it == nil || it.ID != "plans/2026-09-22-p" {
		t.Errorf("h on the last task row left the cursor on %v, want the plan above it", it)
	}
	if !strings.Contains(helpLines, "h l") {
		t.Errorf("help does not list h l: %q", helpLines)
	}
}

// A list that is not a tree has no head to fold, so h and l leave both the
// rows and the cursor alone on it.
func TestHAndLLeaveAFlatListAlone(t *testing.T) {
	t.Parallel()

	bugs := press(newModel(t), tabKey(tabBugs))
	if len(rowIDs(bugs)) == 0 {
		t.Fatal("the fixture has no open bug, so this test proves nothing")
	}
	for _, k := range []string{"h", "l"} {
		got := press(bugs, k)
		if !slices.Equal(rowIDs(got), rowIDs(bugs)) || got.cursor() != bugs.cursor() {
			t.Errorf("%s changed the Bugs list: %q cursor %d, want %q cursor %d", k, rowIDs(got), got.cursor(), rowIDs(bugs), bugs.cursor())
		}
		if s, want := foldState(got), foldState(bugs); s != want {
			t.Errorf("%s on a flat list folded a head: %s, want %s", k, s, want)
		}
	}
	// The same list after moving off the first row, so the cursor is not
	// already sitting on the row the key would pick.
	second := press(bugs, "j")
	for _, k := range []string{"h", "l"} {
		if got := press(second, k); got.cursor() != second.cursor() {
			t.Errorf("%s moved the cursor on the second bug row to %d, want %d", k, got.cursor(), second.cursor())
		}
	}
}

// leakedLines gives the lines the after frame has that the before frame did
// not, narrowed to the ones holding needle. A board prints its ids as row
// text, so a copy is clean only when it adds no line that holds them.
func leakedLines(before, after, needle string) []string {
	old := make(map[string]bool)
	for _, ln := range strings.Split(before, "\n") {
		old[ln] = true
	}
	var out []string
	for _, ln := range strings.Split(after, "\n") {
		if strings.Contains(ln, needle) && !old[ln] {
			out = append(out, ln)
		}
	}
	return out
}

// y hands the clipboard the id the design names for the row under the cursor,
// and the status line says the copy landed, not what it was.
func TestYCopiesTheIDOfTheRow(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		keys []string
		want string
	}{
		{"bug head", nil, "BUG-0002"},
		{"task", []string{"j"}, "PLN-0004.01"},
		{"plan head", []string{"j", "j"}, "PLN-0003"},
		{"spec head", []string{tabKey(tabSpecs)}, "SPC-0001"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var got []string
			m := actModel(t)
			m.clip = func(s string) error { got = append(got, s); return nil }
			moved := press(m, c.keys...)
			before := plain(moved.View())
			after := press(moved, "y")
			if len(got) != 1 || got[0] != c.want {
				t.Fatalf("clipboard got %q, want %q", got, c.want)
			}
			if after.status != "copied to clipboard" {
				t.Errorf("status %q, want %q", after.status, "copied to clipboard")
			}
			view := plain(after.View())
			if !strings.Contains(view, "copied to clipboard") {
				t.Errorf("the screen does not show %q", "copied to clipboard")
			}
			// The list prints the id as row text, so the toast is free of it
			// only when the copy puts it on no line of its own.
			if leak := leakedLines(before, view, c.want); len(leak) != 0 {
				t.Errorf("the copy put %q on screen: %q", c.want, leak)
			}
		})
	}
}

// A row with no number of its own copies the id the file does have: the path
// for a file written without an id line, the hash when it carries one.
func TestYCopiesTheIDOfARowWithNoNumber(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ file, body, want string }{
		{".acta/plans/2026-09-30-noid.md", "# Plan N\n\n### Task 1: Work\n- [ ] a\n", "plans/2026-09-30-noid"},
		{".acta/plans/2026-09-30-hash.md", "---\nhash: k3f2\n---\n# Plan H\n\n### Task 1: Work\n- [ ] a\n", "PLN-k3f2"},
	} {
		cfg := treeCfg(t, map[string]string{c.file: c.body})
		b, err := board.Load(cfg)
		if err != nil {
			t.Fatal(err)
		}
		m := press(sized(New(cfg, b, true), 160, 40), tabKey(tabPlans))
		if m.Selected() == nil {
			t.Fatalf("%s: the Plans list has no row under the cursor: %q", c.file, rowIDs(m))
		}
		var got []string
		m.clip = func(s string) error { got = append(got, s); return nil }
		after := press(m, "y")
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%s: copied %q, want %q", c.file, got, c.want)
		}
		if after.status != "copied to clipboard" {
			t.Errorf("%s: status %q", c.file, after.status)
		}
	}
}

// The status line always tells how the copy went: the error when it failed,
// and "nothing selected" when the cursor sits on a row with no item, where the
// clipboard is never touched at all.
func TestYSaysWhyTheCopyFailed(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.clip = func(string) error { return errors.New("no clipboard") }
	if got := press(m, "y").status; got != "copy failed: no clipboard" {
		t.Errorf("failed copy: status %q", got)
	}

	called := false
	stub := func(string) error { called = true; return nil }
	// The group row of the Specs list holds legacy files, not one item, so
	// there is no id to copy from it.
	group := press(newModel(t), tabKey(tabSpecs), "G")
	if group.Selected() != nil {
		t.Fatalf("the last Specs row is an item, so this proves nothing: %v", group.Selected())
	}
	group.clip = stub
	if got := press(group, "y").status; got != "nothing selected" || called {
		t.Errorf("group row: status %q, clip called %v", got, called)
	}

	called = false
	empty := press(newModel(t), "/", "zzzz-no-such-item", "enter")
	if len(rowIDs(empty)) != 0 {
		t.Fatalf("the search matched rows %q, so this proves nothing", rowIDs(empty))
	}
	empty.clip = stub
	if got := press(empty, "y").status; got != "nothing selected" || called {
		t.Errorf("empty list: status %q, clip called %v", got, called)
	}
}

// The id y copies has to be one acta takes back: the copied text goes into
// the board and has to come out as the very item the row stands for. Every
// kind of row is walked, on boards with short ids and on boards without, plus
// the rows that stand for no item at all.
func TestYCopiesAnIDTheBoardTakesBack(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		what string
		m    Model
	}{
		{"scratches", press(newModel(t), tabKey(tabScratches))},
		{"bugs", press(newModel(t), tabKey(tabBugs))},
		{"debts", press(newModel(t), tabKey(tabDebts))},
		{"specs", press(newModel(t), tabKey(tabSpecs))},
		{"the Specs group row", press(newModel(t), tabKey(tabSpecs), "G")},
		{"plans", press(newModel(t), tabKey(tabPlans))},
		{"a plan with its tasks open", press(newModel(t), tabKey(tabPlans), " ")},
		{"activities", actModel(t)},
		{"a plan file with no id", planOf(t, "2026-09-30-noid", "# Plan N\n\n### Task 1: Work\n- [ ] a\n")},
		{"a plan file with only a hash", planOf(t, "2026-09-30-hash", "---\nhash: k3f2\n---\n# Plan H\n\n### Task 1: Work\n- [ ] a\n")},
		{"a list with no rows", press(newModel(t), "/", "zzzz-no-such-item", "enter")},
	} {
		checkCopiedID(t, c.what, c.m)
	}
}

// checkCopiedID puts the cursor on every row a list shows and checks that the
// id y copies is one the board finds again, and that a row standing for no
// item copies nothing at all.
func checkCopiedID(t *testing.T, what string, m Model) {
	t.Helper()

	rows := m.rowsOf(m.listPane())
	if len(rows) == 0 {
		called := false
		m.clip = func(string) error { called = true; return nil }
		if got := press(m, "y").status; got != "nothing selected" || called {
			t.Errorf("%s: a list with no rows gave %q, clip called %v", what, got, called)
		}
		return
	}
	for i, r := range rows {
		at := m
		at.moveTo(i)
		it := at.Selected()
		var got []string
		at.clip = func(s string) error { got = append(got, s); return nil }
		after := press(at, "y")
		if it == nil {
			if len(got) != 0 || after.status != "nothing selected" {
				t.Errorf("%s row %d %q stands for no item, but copied %q, status %q", what, i, r.id, got, after.status)
			}
			continue
		}
		if len(got) != 1 {
			t.Errorf("%s row %d %q: the clipboard got %q", what, i, r.id, got)
			continue
		}
		back := at.board.Get(got[0])
		if back == nil {
			t.Errorf("%s row %d %q: the board does not know the copied id %q", what, i, r.id, got[0])
			continue
		}
		if back.ID != it.ID {
			t.Errorf("%s row %d %q: copied %q, which the board reads back as %s", what, i, r.id, got[0], back.ID)
		}
	}
}

// planOf is the Plans list of a board with one plan file of the given body,
// with that plan opened, so its task rows are on the screen too.
func planOf(t *testing.T, name, body string) Model {
	t.Helper()

	cfg := treeCfg(t, map[string]string{".acta/plans/" + name + ".md": body})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return press(sized(m, 160, 40), tabKey(tabPlans), " ")
}

// The terminal reads the id back out of the OSC 52 code, so the code has to
// decode to the text it was given, whole and for every text.
func TestOSC52CarriesTheTextWhole(t *testing.T) {
	t.Parallel()

	for _, text := range []string{"PLN-0003#task-1", "", "SPC-ünïcode", strings.Repeat("x", 500)} {
		s := osc52(text)
		inner, ok := strings.CutPrefix(s, "\x1b]52;c;")
		inner, ok2 := strings.CutSuffix(inner, "\a")
		if !ok || !ok2 {
			t.Fatalf("%q: not an OSC 52 code: %q", text, s)
		}
		raw, err := base64.StdEncoding.DecodeString(inner)
		if err != nil || string(raw) != text {
			t.Errorf("%q: decodes to %q, %v", text, raw, err)
		}
	}
}

// grabStdout runs f with the screen output written to a file, so a test can
// read what the program sent to the terminal.
func grabStdout(t *testing.T, f func() error) (string, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "out")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = file
	ferr := f()
	os.Stdout = old
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(out), ferr
}

// pbcopy takes the text when the machine has it, and OSC 52 is the way out
// when it is missing or fails, so a copy still lands over SSH and tmux. The
// pbcopy here is a shim writing to a file, so the test never touches the
// clipboard of the machine it runs on. It is not parallel because it moves
// PATH and the screen output of the whole test run.
func TestCopyTextUsesPBCopyAndFallsBackToOSC52(t *testing.T) {
	dir := t.TempDir()
	empty := t.TempDir()
	landed := filepath.Join(dir, "landed")
	// The shim runs with PATH holding the shim folder alone, so the shell
	// finds cat by its full path.
	shim := "#!/bin/sh\n/bin/cat > " + landed + "\n"
	put := func(body string) {
		if err := os.WriteFile(filepath.Join(dir, "pbcopy"), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	read := func() string {
		b, err := os.ReadFile(landed)
		if err != nil {
			if os.IsNotExist(err) {
				return ""
			}
			t.Fatal(err)
		}
		return string(b)
	}

	put(shim)
	t.Setenv("PATH", dir)
	out, err := grabStdout(t, func() error { return copyText("PLN-0003#task-1") })
	if err != nil {
		t.Fatalf("pbcopy copy failed: %v", err)
	}
	if read() != "PLN-0003#task-1" {
		t.Errorf("pbcopy got %q, want the id", read())
	}
	if out != "" {
		t.Errorf("pbcopy took the copy, yet the screen got %q too", out)
	}

	// pbcopy there but broken: the terminal takes over.
	put("#!/bin/sh\nexit 1\n")
	out, err = grabStdout(t, func() error { return copyText("BUG-0002") })
	if err != nil {
		t.Fatalf("broken pbcopy: %v", err)
	}
	if out != osc52("BUG-0002") {
		t.Errorf("broken pbcopy: screen got %q, want the OSC 52 code", out)
	}

	// No pbcopy at all, on a machine like a Linux box over SSH.
	t.Setenv("PATH", empty)
	out, err = grabStdout(t, func() error { return copyText("SPC-0001") })
	if err != nil {
		t.Fatalf("no pbcopy: %v", err)
	}
	if out != osc52("SPC-0001") {
		t.Errorf("no pbcopy: screen got %q, want the OSC 52 code", out)
	}

	// No pbcopy and a screen that will not take the code: the error names
	// both ways that failed, so nothing is swallowed.
	closed, err := os.Create(filepath.Join(empty, "closed"))
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = closed
	cerr := copyText("SPC-0001")
	os.Stdout = old
	if cerr == nil {
		t.Fatal("a copy that reached nothing must say so")
	}
	for _, want := range []string{"pbcopy", "closed"} {
		if !strings.Contains(cerr.Error(), want) {
			t.Errorf("error %q does not name %q", cerr, want)
		}
	}
}

// A good copy starts a timer and the timer hides the message, and a message
// that came later keeps the line.
func TestCopyToastHidesByItself(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.clip = func(string) error { return nil }
	next, cmd := m.Update(key("y"))
	m = next.(Model)
	if cmd == nil {
		t.Fatal("a copy started no timer, so the toast would stay")
	}
	if m.status != "copied to clipboard" {
		t.Fatalf("status %q", m.status)
	}
	next, _ = m.Update(clearStatusMsg{text: "copied to clipboard"})
	if got := next.(Model).status; got != "" {
		t.Fatalf("toast did not hide: %q", got)
	}
}

func TestCopyToastKeepsANewerMessage(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.status = "reload failed: disk"
	next, _ := m.Update(clearStatusMsg{text: "copied to clipboard"})
	if got := next.(Model).status; got != "reload failed: disk" {
		t.Fatalf("a newer message was cleared: %q", got)
	}
}

// Every message on the status line hides by itself, errors too, so the key
// hints always come back.
func TestEveryToastHidesByItself(t *testing.T) {
	t.Parallel()

	broken := actModel(t)
	broken.clip = func(string) error { return errors.New("no clipboard") }

	onTask := press(actModel(t), tabKey(tabPlans), "l", "j")
	if it := onTask.Selected(); it == nil || it.Kind != board.KindTask {
		t.Fatalf("the row under the cursor is %+v, want a task", it)
	}

	cfg := treeCfg(t, map[string]string{})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	empty := New(cfg, b, true)

	cases := []struct {
		name string
		m    Model
		run  func(Model) (tea.Model, tea.Cmd)
	}{
		{"copy failed", broken, func(m Model) (tea.Model, tea.Cmd) { return m.Update(key("y")) }},
		{"editor error", actModel(t), func(m Model) (tea.Model, tea.Cmd) {
			return m.Update(editorDoneMsg{err: errors.New("boom")})
		}},
		{"tasks refuse the popup", onTask, func(m Model) (tea.Model, tea.Cmd) {
			return m.Update(key("s"))
		}},
		{"nothing selected", empty, func(m Model) (tea.Model, tea.Cmd) { return m.Update(key("y")) }},
	}
	for _, c := range cases {
		next, cmd := c.run(c.m)
		got := next.(Model)
		if got.status == "" {
			t.Fatalf("%s: no message on the line", c.name)
		}
		if cmd == nil {
			t.Fatalf("%s: %q started no timer, so it would stay", c.name, got.status)
		}
		cleared, _ := got.Update(clearStatusMsg{text: got.status})
		if s := cleared.(Model).status; s != "" {
			t.Fatalf("%s: the toast did not hide: %q", c.name, s)
		}
	}
}

// The editor runs with the program suspended, so the terminal comes back
// with mouse reporting off and nothing turns it back on. Every editorDoneMsg
// has to ask for the mouse again, or clicks and the wheel stay dead until
// the TUI restarts. A message that never came from the editor must leave the
// mouse alone.
func TestMouseComesBackAfterTheEditor(t *testing.T) {
	t.Parallel()

	want := tea.EnableMouseCellMotion()
	paths := []struct {
		name string
		msg  editorDoneMsg
	}{
		{"no error", editorDoneMsg{}},
		{"a new bug file", editorDoneMsg{newBug: "no-such-bug.md"}},
		{"the editor failed", editorDoneMsg{err: errors.New("boom")}},
	}
	for _, c := range paths {
		_, cmd := actModel(t).Update(c.msg)
		msgs := runNow(cmd, 200*time.Millisecond)
		if !slices.Contains(msgs, want) {
			t.Errorf("%s: the mouse stayed off after the editor, got %v", c.name, msgs)
		}
	}

	// "q" answers with a quit command, so there is a message to look at. A key
	// that returns nothing would make this loop empty and the check below
	// would never run.
	_, cmd := actModel(t).Update(key("q"))
	msgs := runNow(cmd, 200*time.Millisecond)
	if len(msgs) == 0 {
		t.Fatalf("the key press gave back no message, so nothing was checked")
	}
	for _, msg := range msgs {
		if msg == want {
			t.Errorf("a key press with no editor behind it turned the mouse on: %v", msgs)
		}
	}
}

// The search box is not a toast: typing there changes the list, not the
// status line, so no timer is started for it.
func TestSearchTextStartsNoTimer(t *testing.T) {
	t.Parallel()

	m := press(actModel(t), "/")
	next, cmd := m.Update(key("BUG"))
	got := next.(Model)
	if got.query != "BUG" {
		t.Fatalf("the query is %q, the keys never reached the box", got.query)
	}
	if cmd != nil || got.status != "" {
		t.Fatalf("typing in the search started a timer: status %q, cmd %v", got.status, cmd != nil)
	}
}

// A message the model was built with, such as a theme that did not load, was
// never handled by Update, so the program starts the timer for it itself.
func TestTheFirstToastHidesFromTheFirstFrame(t *testing.T) {
	t.Parallel()

	m := newModel(t).WithTheme("nope", true)
	if m.status == "" {
		t.Fatal("a theme that does not load must say so")
	}
	want := clearStatusMsg{text: m.status}
	for _, msg := range runNow(m.Init(), 2*toastFor) {
		if msg == want {
			return
		}
	}
	t.Fatalf("nothing sent %#v for the message the program started with", want)
}

// The timer a toast gets sends the clear message for that exact text, so a
// toast that took the line later is not taken away by an older timer.
func TestToastTimerClearsThatExactMessage(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.clip = func(string) error { return nil }
	next, cmd := m.Update(key("y"))
	status := next.(Model).status
	if status == "" || cmd == nil {
		t.Fatalf("status %q, cmd nil = %v", status, cmd == nil)
	}
	// The toast waits toastFor before it hides, so this test waits with it.
	want := clearStatusMsg{text: status}
	for _, msg := range runNow(cmd, 2*toastFor) {
		if msg == want {
			cleared, _ := next.Update(want)
			if s := cleared.(Model).status; s != "" {
				t.Fatalf("the toast did not hide: %q", s)
			}
			return
		}
	}
	t.Fatalf("no command sent %#v, got %v", want, runNow(cmd, 2*toastFor))
}

// runNow runs a command, or every command in a batch, each on its own
// goroutine, and gives back the messages that arrive inside the wait. A
// command still waiting when the wait is over gives back nothing, so a slow
// toast timer never holds a test up.
func runNow(cmd tea.Cmd, within time.Duration) []tea.Msg {
	if cmd == nil {
		return nil
	}
	got := make(chan tea.Msg, 1)
	go func() { got <- cmd() }()
	var msg tea.Msg
	select {
	case msg = <-got:
	case <-time.After(within):
		return nil
	}
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}
	var out []tea.Msg
	for _, inner := range batch {
		out = append(out, runNow(inner, within)...)
	}
	return out
}

// notToast runs what a key asked for and gives back every message that is
// not the toast timer, so a test can still catch the side effect it means.
func notToast(cmd tea.Cmd) []tea.Msg {
	var out []tea.Msg
	for _, msg := range runNow(cmd, 50*time.Millisecond) {
		if _, clear := msg.(clearStatusMsg); !clear {
			out = append(out, msg)
		}
	}
	return out
}

func TestClearStatusAfterSendsItsText(t *testing.T) {
	t.Parallel()

	if got := clearStatusAfter(0, "copied X")(); got != (clearStatusMsg{text: "copied X"}) {
		t.Fatalf("got %#v", got)
	}
}

// quietModel is a board with no work under way, so the pulse never starts.
func quietModel(t *testing.T) Model {
	t.Helper()
	cfg := treeCfg(t, map[string]string{".acta/bugs/2026-09-21-q.md": "---\nid: BUG-0009\n---\n# Quiet\n"})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return New(cfg, b, true)
}

func TestPulseMovesOnlyWithADotOnScreen(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.frame = &frameCache{dots: true}
	next, cmd := m.Update(pulseMsg{})
	got := next.(Model)
	if cmd == nil || !got.pulsing {
		t.Fatal("a pulse with work under way did not arm the next one")
	}
	if got.pulse != 1 || got.same {
		t.Fatalf("pulse %d same %v, want frame 1 drawn again", got.pulse, got.same)
	}
	got.frame = &frameCache{dots: false}
	next, cmd = got.Update(pulseMsg{})
	quiet := next.(Model)
	if cmd == nil {
		t.Fatal("work under way but no dot on screen stopped the pulse chain")
	}
	if quiet.pulse != 1 || !quiet.same {
		t.Fatalf("with no dot on screen the pulse moved to %d or drew again (same %v)", quiet.pulse, quiet.same)
	}
}

func TestPulseStopsWithNoWork(t *testing.T) {
	t.Parallel()

	m := quietModel(t)
	m.frame = &frameCache{dots: true}
	next, cmd := m.Update(pulseMsg{})
	stopped := next.(Model)
	if cmd != nil || stopped.pulsing {
		t.Fatal("a board with no work under way kept pulsing")
	}
	// The last frame showed a dot, so a tick that stops the chain has to say
	// the screen is the same, or the view draws a frame nobody can see a
	// change in.
	if !stopped.same {
		t.Error("the last tick of a chain that stops left the screen marked as changed")
	}
}

func TestResizeAndReloadArmThePulseOnce(t *testing.T) {
	t.Parallel()

	m := actModel(t)
	m.pulsing = false
	next, cmd := m.Update(tea.WindowSizeMsg{Width: m.width + 1, Height: m.height})
	armed := next.(Model)
	if cmd == nil || !armed.pulsing {
		t.Fatal("a resize with work under way did not arm the pulse")
	}
	next, cmd = armed.Update(tea.WindowSizeMsg{Width: armed.width + 1, Height: armed.height})
	if cmd == nil || fmt.Sprint(cmd()) != fmt.Sprint(tea.ClearScreen()) {
		t.Fatal("a second resize while pulsing must only clear the screen")
	}
	quiet := quietModel(t)
	if _, cmd := quiet.Update(tea.WindowSizeMsg{Width: 99, Height: 30}); cmd == nil || fmt.Sprint(cmd()) != fmt.Sprint(tea.ClearScreen()) {
		t.Fatal("a resize with no work must only clear the screen")
	}
	idle := actModel(t)
	idle.pulsing = false
	next, cmd = idle.Update(reloadMsg{b: idle.board})
	if cmd == nil || !next.(Model).pulsing {
		t.Fatal("a reload with work under way did not arm the pulse")
	}
}

func TestPulseAfterSendsAPulse(t *testing.T) {
	t.Parallel()

	if got := pulseAfter(0)(); got != (pulseMsg{}) {
		t.Fatalf("got %#v", got)
	}
}

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

// TestMarkRefusesAWorktreeAndALegacyRow keeps the keys off the rows whose file
// lives somewhere else, because a write would land in the wrong tree.
func TestMarkRefusesAWorktreeAndALegacyRow(t *testing.T) {
	t.Parallel()

	// A plan that came from a worktree branch, opened to its first task.
	m := press(worktreeTaskModel(t), tabKey(tabPlans), " ", "j")
	it := m.Selected()
	if it == nil || it.Kind != board.KindTask || it.Worktree == "" {
		t.Fatalf("not on a worktree task row: %+v", it)
	}
	called := false
	m.markItem = func(string, bool) (write.Outcome, error) {
		called = true
		return write.Outcome{}, nil
	}
	next, cmd := m.Update(key("+"))
	m = next.(Model)
	if called || len(notToast(cmd)) > 0 {
		t.Fatalf("worktree row: called %v", called)
	}
	if !strings.Contains(m.status, it.Worktree) {
		t.Fatalf("the status does not name the worktree: %q", m.status)
	}

	// A task from a folder the board calls legacy. The board really holds
	// such tasks, but the lists leave them out, so the flag is set here to
	// reach the guard the same way such a row would.
	root := press(newModel(t), tabKey(tabPlans), " ", "j")
	sel := root.Selected()
	if sel == nil || sel.Kind != board.KindTask {
		t.Fatalf("not on a task row: %+v", sel)
	}
	sel.Legacy = true
	called = false
	root.markItem = func(string, bool) (write.Outcome, error) {
		called = true
		return write.Outcome{}, nil
	}
	next, cmd = root.Update(key("-"))
	root = next.(Model)
	if called || len(notToast(cmd)) > 0 {
		t.Fatalf("legacy row: called %v", called)
	}
	if !strings.Contains(root.status, "legacy") {
		t.Fatalf("status %q", root.status)
	}
}

// TestMarkWithNoRowSelected says so instead of writing something unknown.
func TestMarkWithNoRowSelected(t *testing.T) {
	t.Parallel()

	// The Specs group row holds the legacy files, not an item of its own.
	m := press(newModel(t), tabKey(tabSpecs), "G")
	if m.Selected() != nil {
		t.Fatalf("the group row is an item, so this proves nothing: %v", m.Selected())
	}
	called := false
	m.markItem = func(string, bool) (write.Outcome, error) {
		called = true
		return write.Outcome{}, nil
	}
	next, cmd := m.Update(key("+"))
	m = next.(Model)
	if called || len(notToast(cmd)) > 0 {
		t.Fatalf("no row: called %v", called)
	}
	if !strings.Contains(m.status, "nothing selected") {
		t.Fatalf("status %q", m.status)
	}
}

// TestMarkDoesNothingWhileAnOverlayIsOpen walks every box that sits over the
// list, because each one takes the keys until it closes.
func TestMarkDoesNothingWhileAnOverlayIsOpen(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		open []string
	}{
		{"the help", []string{"?"}},
		{"a status popup", []string{tabKey(tabBugs), "j", "s"}},
		{"a type popup", []string{tabKey(tabBugs), "j", "t"}},
		{"the slug prompt", []string{"n"}},
		{"the search", []string{"/"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Start on a task row, so a key that leaks through writes.
			m := press(newModel(t), tabKey(tabPlans), " ", "j")
			if m.Selected() == nil || m.Selected().Kind != board.KindTask {
				t.Fatalf("not on a task row: %+v", m.Selected())
			}
			m = press(m, tc.open...)
			called := false
			m.markItem = func(string, bool) (write.Outcome, error) {
				called = true
				return write.Outcome{}, nil
			}
			for _, k := range []string{"+", "-"} {
				next, cmd := m.Update(key(k))
				m = next.(Model)
				if called || cmd != nil {
					t.Fatalf("%q under %s: called %v cmd %v", k, tc.name, called, cmd)
				}
			}
		})
	}
}

// TestMarkShowsTheError tells the truth when the write fails.
func TestMarkShowsTheError(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabPlans), " ", "j")
	m.markItem = func(string, bool) (write.Outcome, error) {
		return write.Outcome{}, errors.New("disk is full")
	}
	m = press(m, "+")
	if !strings.Contains(m.status, "disk is full") {
		t.Fatalf("status %q", m.status)
	}
}

// TestMarkReloadsTheBoard asks the program to read the files again, so the
// screen shows the boxes the write just ticked.
func TestMarkReloadsTheBoard(t *testing.T) {
	t.Parallel()

	m := press(newModel(t), tabKey(tabPlans), " ", "j")
	m.markItem = func(string, bool) (write.Outcome, error) {
		return write.Outcome{Committed: true}, nil
	}
	_, cmd := m.Update(key("+"))
	if cmd == nil {
		t.Fatal("+ asked for no reload")
	}
	for _, msg := range runNow(cmd, time.Second) {
		if _, ok := msg.(reloadMsg); ok {
			return
		}
	}
	t.Fatal("+ asked for no reload")
}

// worktreeTaskModel holds one plan that came from a worktree branch, opened
// to its task, so the keys meet a row they must not write.
func worktreeTaskModel(t *testing.T) Model {
	t.Helper()
	main := treeCfg(t, nil)
	wt := treeCfg(t, map[string]string{".acta/plans/2026-09-21-a.md": "# A plan\n\n### Task 1: Step\n\n- [ ] do it\n"})
	b, err := board.LoadTrees(main, []board.Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	m := New(main, b, true)
	m.render = func(md string, _ int) string { return md }
	m.dcache = &detailCache{}
	return sized(m, 120, 40)
}

// TestMarkWritesTheFile is the one test that runs the markItem of New, so
// the wiring is not taken on faith. Auto commit is off, so no git repository
// is needed to see the box on disk change.
func TestMarkWritesTheFile(t *testing.T) {
	t.Parallel()

	cfg := treeCfg(t, map[string]string{".acta/plans/2026-09-30-p.md": "# P\n\n### Task 1: Step\n\n- [ ] do it\n"})
	cfg.AutoCommit = false
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	m.dcache = &detailCache{}
	m = press(sized(m, 120, 40), tabKey(tabPlans), " ", "j")
	if it := m.Selected(); it == nil || it.Kind != board.KindTask {
		t.Fatalf("not on a task row: %+v", it)
	}
	path := filepath.Join(cfg.Root, "plans", "2026-09-30-p.md")
	m = press(m, "+")
	if !strings.Contains(m.status, "written") {
		t.Fatalf("status %q", m.status)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "- [x] do it") {
		t.Fatalf("+ did not tick the box:\n%s", body)
	}
	m = press(m, "-")
	body, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "- [ ] do it") {
		t.Fatalf("- did not open the box:\n%s", body)
	}
}
