package tui

import (
	"errors"
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

// rowIDs are the rows of the box the screen shows, so a test reads the same
// list the reader does.
func rowIDs(m Model) []string { return ids(m.rowsOf(m.listPane())) }

func doneRowIDs(m Model) []string { return ids(m.doneRows()) }

func TestTabKeyFocusesTheNextPane(t *testing.T) {
	m := newModel(t)
	if m.focus != paneActive {
		t.Fatalf("the screen starts on pane %d", m.focus)
	}
	for _, want := range []pane{1, 2, 3, 4, paneDetail, paneActive} {
		m = press(m, "tab")
		if m.focus != want {
			t.Fatalf("tab gave focus %d, want %d", m.focus, want)
		}
	}
}

func TestShiftTabFocusesThePreviousPane(t *testing.T) {
	m := newModel(t)
	for _, want := range []pane{paneDetail, paneDone, 3, 2, 1, paneActive} {
		m = press(m, "shift+tab")
		if m.focus != want {
			t.Fatalf("shift+tab gave focus %d, want %d", m.focus, want)
		}
	}
}

func TestNumberKeysFocusThatPane(t *testing.T) {
	m := newModel(t)
	for _, s := range []string{"1", "2", "3", "4", "5", "0"} {
		m = press(m, s)
		if m.focus != paneOfKey(s) {
			t.Fatalf("%s gave focus %d", s, m.focus)
		}
	}
}

func TestTabsCycleAndWrap(t *testing.T) {
	// Every pane with tabs walks its own two tabs and comes back around.
	for p := 1; p < int(paneDone); p++ {
		m := press(newModel(t), keyOf(pane(p)))
		for _, want := range []int{1, 0, 1} {
			m = press(m, "]")
			if m.tab[p] != want {
				t.Errorf("pane %d: ] gave tab %d, want %d", p, m.tab[p], want)
			}
		}
		for _, want := range []int{0, 1, 0} {
			m = press(m, "[")
			if m.tab[p] != want {
				t.Errorf("pane %d: [ gave tab %d, want %d", p, m.tab[p], want)
			}
		}
	}
	// The tab really changed: the Specs pane shows the specs, not the ideas.
	m := press(newModel(t), "2")
	if got := strings.Join(rowIDs(m), " "); !strings.HasPrefix(got, "specs/2026-09-28-from-scratch-design") {
		t.Fatalf("Specs rows %q", got)
	}
	if got := strings.Join(rowIDs(press(m, "]")), " "); !strings.HasPrefix(got, "scratch/2026-09-28-idea-brainstorm") {
		t.Fatalf("Scratchpad rows %q", got)
	}
}

func TestDonePaneTabsFollowTheLastSidebarPane(t *testing.T) {
	for _, tc := range []struct {
		keys  []string
		names string
	}{
		{[]string{"2"}, "Done Dropped"},
		{[]string{"2", "]"}, "Specced Dropped"},
		{[]string{"3"}, "Done Dropped"},
		{[]string{"3", "]"}, "Done"},
		{[]string{"4"}, "Fixed Wontfix"},
		{[]string{"4", "]"}, "Done Wontfix"},
		{[]string{"1"}, ""},
	} {
		m := press(newModel(t), append(tc.keys, "5")...)
		if got := strings.Join(m.doneTabNames(), " "); got != tc.names {
			t.Errorf("after %v the finished tabs are %q, want %q", tc.keys, got, tc.names)
		}
	}
}

func TestDonePaneTabsCycleAndWrap(t *testing.T) {
	m := press(newModel(t), "2", "5")
	if m.tab[paneDone] != 0 {
		t.Fatalf("done tab %d", m.tab[paneDone])
	}
	m = press(m, "]")
	if m.tab[paneDone] != 1 {
		t.Fatalf("] gave done tab %d", m.tab[paneDone])
	}
	m = press(m, "]")
	if m.tab[paneDone] != 0 {
		t.Fatalf("the finished tabs should wrap, got %d", m.tab[paneDone])
	}
	m = press(m, "[")
	if m.tab[paneDone] != 1 {
		t.Fatalf("[ gave done tab %d", m.tab[paneDone])
	}
	// A task is only ever done, so there is no second tab to move to.
	m = press(m, "3", "]", "5", "]", "]", "]")
	if m.tab[paneDone] != 0 {
		t.Fatalf("Tasks should stay on Done, got %d", m.tab[paneDone])
	}
	// Coming from a pane whose tab has two finished tabs, the one without a
	// second drops back to its only tab.
	m = press(newModel(t), "3", "]", "5", "]", "4", "[", "5")
	if m.tab[2] != 1 || m.tab[paneDone] != 0 {
		t.Fatalf("Tasks should be back on Done, tab %d done %d", m.tab[2], m.tab[paneDone])
	}
}

func TestPaneDetailHasNoTabs(t *testing.T) {
	m := press(newModel(t), "3", "]", "5", "]", "0", "]", "[", "]")
	if m.tab[2] != 1 || m.tab[paneDone] != 0 {
		t.Fatalf("the detail box moved the tabs: tab %d done %d", m.tab[2], m.tab[paneDone])
	}
}

func TestEachTabHoldsItsOwnItems(t *testing.T) {
	for _, tc := range []struct {
		key  string
		tabs []string
		rows string
	}{
		{"2", nil, "specs/2026-09-28-from-scratch-design specs/2026-09-22-beta specs/2026-09-20-alpha specs/2026-09-18-weird specs/2026-09-17-broken " + groupRowID},
		{"2", []string{"]"}, "scratch/2026-09-28-idea-brainstorm scratch/2026-09-28-idea-raw"},
		{"3", nil, "plans/2026-09-23-lonely plans/2026-09-21-alpha"},
		{"3", []string{"]"}, "plans/2026-09-23-lonely#task-1 plans/2026-09-21-alpha#task-2"},
		{"4", nil, "bugs/2026-09-28-lag bugs/2026-09-26-open specs/2026-09-15-really-bug"},
		{"4", []string{"]"}, "debt/2026-09-27-orphan-debt#item-1"},
	} {
		m := press(newModel(t), append([]string{tc.key}, tc.tabs...)...)
		if got := strings.Join(rowIDs(m), " "); got != tc.rows {
			t.Errorf("pane %s tab %v holds %q, want %q", tc.key, tc.tabs, got, tc.rows)
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
	m := openTab(t, debtModel(t), 3, 1)
	if got := strings.Join(rowIDs(m), " "); got != "debt/2026-09-27-short-ids#item-1" {
		t.Fatalf("Debt open rows %q, want only the open line", got)
	}
	if it := m.Selected(); it == nil || it.Title != "a" {
		t.Fatalf("Debt open item = %+v, want title a", it)
	}
}

func TestDebtDonePaneSplitsDoneAndWontfix(t *testing.T) {
	m := press(openTab(t, debtModel(t), 3, 1), "5")
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
	for _, tc := range []struct{ p, tab int }{{1, 0}, {2, 0}, {2, 1}, {3, 0}} {
		a := strings.Join(rowIDs(openTab(t, withDebt, pane(tc.p), tc.tab)), " ")
		b := strings.Join(rowIDs(openTab(t, without, pane(tc.p), tc.tab)), " ")
		if a != b {
			t.Fatalf("pane %d tab %d changed with the debt file present: got %q, want %q", tc.p, tc.tab, a, b)
		}
	}
}

func TestDonePaneHoldsTheFinishedItemsOfTheOpenTab(t *testing.T) {
	m := press(newModel(t), "2", "5")
	if got := strings.Join(doneRowIDs(m), " "); got != "specs/2026-09-16-finished" {
		t.Fatalf("Specs done %q", got)
	}
	m = press(m, "]")
	if got := strings.Join(doneRowIDs(m), " "); got != "specs/2026-09-19-dropped" {
		t.Fatalf("Specs dropped %q", got)
	}
	// The Done pane keeps the finished tab it was on, so step back to Done.
	m = press(m, "3", "5", "[")
	if got := strings.Join(doneRowIDs(m), " "); got != "plans/2026-09-28-dotted-tasks plans/2026-09-27-orphan plans/2026-09-26-dash-tasks plans/2026-09-25-crash-fix plans/2026-09-16-finished" {
		t.Fatalf("Plans done %q", got)
	}
	m = press(m, "3", "]", "5")
	if got := strings.Join(doneRowIDs(m), " "); got != "plans/2026-09-28-dotted-tasks#task-2.1 plans/2026-09-28-dotted-tasks#task-2.2 plans/2026-09-27-orphan#task-1 plans/2026-09-26-dash-tasks#task-F-1 plans/2026-09-26-dash-tasks#task-F-2 plans/2026-09-25-crash-fix#task-F1 plans/2026-09-21-alpha#task-1 plans/2026-09-16-finished#task-1" {
		t.Fatalf("Tasks done %q", got)
	}
	m = press(m, "4", "5")
	if got := strings.Join(doneRowIDs(m), " "); got != "bugs/2026-09-24-crash" {
		t.Fatalf("Bugs fixed %q", got)
	}
	m = press(m, "]")
	if got := doneRowIDs(m); len(got) != 0 {
		t.Fatalf("no bug is wontfix, got %v", got)
	}
}

func TestMoveKeysInListPanes(t *testing.T) {
	m := press(newModel(t), "2")
	if m.Selected().ID != "specs/2026-09-28-from-scratch-design" {
		t.Fatalf("first selection %s", m.Selected().ID)
	}
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("after j: %s", m.Selected().ID)
	}
	m = press(m, "k")
	if m.Selected().ID != "specs/2026-09-28-from-scratch-design" {
		t.Fatalf("after k: %s", m.Selected().ID)
	}
	m = press(m, "k")
	if m.Selected().ID != "specs/2026-09-28-from-scratch-design" {
		t.Fatalf("k at the top should stay: %s", m.Selected().ID)
	}
	m = press(m, "ctrl+d")
	if m.openRows(paneSpecs)[m.cursor()].id != groupRowID {
		t.Fatalf("ctrl+d should jump a page: %s", m.openRows(paneSpecs)[m.cursor()].id)
	}
	m = press(m, "ctrl+u")
	if m.Selected().ID != "specs/2026-09-28-from-scratch-design" {
		t.Fatalf("ctrl+u should step back a page: %s", m.Selected().ID)
	}
	// The same keys move the Done box.
	m = press(m, "3", "5", "j", "j")
	if m.Selected() == nil || m.Selected().Status != "done" {
		t.Fatalf("the Done box did not move: %v", m.Selected())
	}
}

func TestTopAndBottomKeysInListPanes(t *testing.T) {
	m := press(newModel(t), "2")
	m = press(m, "G")
	if m.openRows(paneSpecs)[m.cursor()].id != groupRowID {
		t.Fatalf("G should land on the last row: %v", rowIDs(m))
	}
	m = press(m, "j")
	if m.openRows(paneSpecs)[m.cursor()].id != groupRowID {
		t.Fatal("j past the end should stay on the last row")
	}
	m = press(m, "g", "k")
	if m.Selected().ID != "specs/2026-09-28-from-scratch-design" {
		t.Fatal("k at the top should stay on the first row")
	}
}

func TestScrollKeysInPaneDetail(t *testing.T) {
	m := press(longModel(t), "3", "0")
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
	m := press(longModel(t), "3", "0", "ctrl+d", "3")
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
	m := newModel(t)
	g := m.geometry()
	// A row takes one line, so the third line of the pane is the third row.
	m = click(m, 2, g.side[paneSpecs].y+1+2)
	if m.focus != paneSpecs || m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// The same works in the Done box, where the second row is the orphan plan.
	m = press(newModel(t), "3", "5")
	g = m.geometry()
	m = click(m, 2, g.side[paneDone].y+1+1)
	if m.focus != paneDone || m.Selected().ID != "plans/2026-09-27-orphan" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
}

// TestAClickOnTheFirstAndLastVisibleRowLandsOnThatRow reads the rows the pane
// shows and clicks where each of the two ends is drawn, so a click cannot land
// on the neighbour of the row the eye sees.
func TestAClickOnTheFirstAndLastVisibleRowLandsOnThatRow(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []string
		p    pane
	}{
		{"open", nil, paneSpecs},
		{"done", []string{"3", "5"}, paneDone},
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

func TestClickOnATabNameSwitchesTab(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	plans := g.side[panePlans]
	m = click(m, plans.x+plans.tabs[1].x+1, plans.y)
	if m.tab[panePlans] != 1 {
		t.Fatalf("clicking Tasks gave tab %d", m.tab[panePlans])
	}
	// The click also gave the box the focus, and the tasks it lists showed up.
	if m.focus != panePlans || strings.Join(rowIDs(m), " ") != "plans/2026-09-23-lonely#task-1 plans/2026-09-21-alpha#task-2" {
		t.Fatalf("focus %d rows %v", m.focus, rowIDs(m))
	}
	// Wontfix is the second finished tab of a bug, so click that one.
	m = press(m, "4")
	g = m.geometry()
	done := g.side[paneDone]
	m = click(m, done.x+done.tabs[1].x+1, done.y)
	if m.focus != paneDone || m.tab[paneDone] != 1 {
		t.Fatalf("focus %d done tab %d", m.focus, m.tab[paneDone])
	}
	// Clicking the dash before a name belongs to no tab.
	m = click(m, done.x+done.tabs[1].x-1, done.y)
	if m.tab[paneDone] != 1 {
		t.Fatalf("a click on the dash moved the tab to %d", m.tab[paneDone])
	}
}

func TestClickInsideAPaneOnlyFocusesIt(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	// Below the last row of pane [1] is its own border, so use the wide gutter
	// of the left column instead.
	m = click(m, g.side[paneSpecs].x+g.side[paneSpecs].w-1, g.side[paneSpecs].y+g.side[paneSpecs].h-1)
	if m.focus != paneSpecs {
		t.Fatalf("focus %d", m.focus)
	}
	m = click(m, 119, 5)
	if m.focus != paneDetail || m.Selected().ID != "specs/2026-09-28-from-scratch-design" {
		t.Fatalf("focus %d selection %v", m.focus, m.Selected())
	}
	// A click on the status line does nothing at all.
	m = click(m, 10, 39)
	if m.focus != paneDetail {
		t.Fatalf("focus %d", m.focus)
	}
}

func TestKeysContinueFromAClickedRow(t *testing.T) {
	m := newModel(t)
	g := m.geometry()
	m = click(m, 2, g.side[paneSpecs].y+1)
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("j moved to %v", m.Selected())
	}
	m = click(m, 2, g.side[paneSpecs].y+1+1)
	m = press(m, "j")
	if m.Selected().ID != "specs/2026-09-20-alpha" {
		t.Fatalf("j did not start from the clicked row: %v", m.Selected())
	}
}

func TestKeyADoesNothing(t *testing.T) {
	m := newModel(t)
	// The tab and the cursor are slices, so they are copied out before the
	// keys: a shared slice would hide a change instead of showing one.
	beforeTabs, beforeSel := slices.Clone(m.tab), slices.Clone(m.sel)
	before := m
	after := press(m, "a", "a")
	if strings.Join(rowIDs(after), " ") != strings.Join(rowIDs(before), " ") ||
		strings.Join(doneRowIDs(after), " ") != strings.Join(doneRowIDs(before), " ") ||
		after.focus != before.focus || !slices.Equal(after.tab, beforeTabs) || !slices.Equal(after.sel, beforeSel) ||
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
	if !m.help || m.focus != paneActive || m.tab[paneSpecs] != 0 || m.searching || m.slug != nil || m.popup != nil {
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
	m = press(m, "2", "?", "esc", "j")
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
	// The Done box has the focus, so s works on the finished plan it selected.
	m := press(newModel(t), "3", "5", "s")
	if m.popup == nil || m.popup.field != "status" {
		t.Fatalf("pane [2] popup %+v", m.popup)
	}
	if m.popup.idx != 3 {
		t.Fatalf("the popup should start on the row's own status done, got %d", m.popup.idx)
	}
	// The same key with a kind box focused works on that box's row instead.
	other := press(newModel(t), "2", "j", "j", "s")
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

func TestGeometryPlacesThePanes(t *testing.T) {
	// 120 columns, the normal width: a kind pane must still fit both of its
	// tab names ("Specs ─ Scratchpad") at the left column's usual size, not a
	// wider one carved out just for the tab title.
	m := sized(newModel(t), 120, 40)
	g := m.geometry()
	if !g.wide || g.leftW != 36 {
		t.Fatalf("wide %v leftW %d", g.wide, g.leftW)
	}
	// The five boxes share the height equally, the last one taking the
	// leftover lines, and the last line of the screen is the status line.
	want := [][4]int{{0, 0, 36, 7}, {0, 7, 36, 7}, {0, 14, 36, 7}, {0, 21, 36, 7}, {0, 28, 36, 11}}
	for p, w := range want {
		if got := g.side[pane(p)]; got.x != w[0] || got.y != w[1] || got.w != w[2] || got.h != w[3] {
			t.Fatalf("pane %d %+v, want %v", p, got, w)
		}
	}
	if g.detail.x != 36 || g.detail.y != 0 || g.detail.w != 84 || g.detail.h != 39 {
		t.Fatalf("the detail box %+v", g.detail)
	}
	// The border takes two lines and a row one, so a pane shows as many rows
	// as it has room for.
	if g.side[paneSpecs].inner != 5 || g.side[paneSpecs].rows != 5 || g.side[paneSpecs].first != 0 {
		t.Fatalf("the Specs box holds %+v", g.side[paneSpecs])
	}
	if g.side[paneDone].inner != 9 || g.side[paneDone].rows != 9 || g.side[paneDone].first != 0 {
		t.Fatalf("the Done box holds %+v", g.side[paneDone])
	}
	// Both tab names are still in the drawn title at this normal width.
	if title := strings.Split(plain(m.View()), "\n")[g.side[paneSpecs].y]; !strings.Contains(title, "Scratchpad") {
		t.Fatalf("the Specs title at 120 columns is missing Scratchpad: %q", title)
	}
	// The Done box has room for every finished plan, so the window stays at
	// the top even with the cursor on the last one.
	done := press(m, "3", "5", "G")
	g = done.geometry()
	if g.side[paneDone].first != 0 || g.side[paneDone].rows != 9 {
		t.Fatalf("the Done box should keep every finished plan in sight: %+v", g.side[paneDone])
	}
	// The left column is 30% of the width, held between 28 and 48.
	for _, w := range []int{60, 80, 100, 120, 200} {
		if got := sized(newModel(t), w, 40).geometry().leftW; got != clamp(w*3/10, 28, 48) {
			t.Errorf("at %d columns the left column is %d", w, got)
		}
	}
	// Below 60 columns only the focused pane is on screen, full width.
	n := sized(press(newModel(t), "5"), 40, 20).geometry()
	if n.wide || n.full.w != 40 || n.full.h != 19 {
		t.Fatalf("narrow screen %+v", n)
	}
	if n.side[paneSpecs].w != 0 || n.side[paneDone].w != 0 || n.detail.w != 0 {
		t.Fatal("a narrow screen should only draw the focused pane")
	}
	if n.full.x != 0 || n.full.y != 0 || n.full.w != 40 || n.full.h != 19 {
		t.Fatalf("the focused pane should take the whole screen: %+v", n.full)
	}
	// The tab boxes sit where the drawn title puts the names, so a click lands
	// on the word it points at. The place comes from the view, not from a
	// number written down here, which is how the two drifted apart before.
	for p := pane(1); p < paneDone; p++ {
		title := strings.Split(plain(m.View()), "\n")[g.side[p].y]
		for i, tab := range sidebar[p].tabs {
			want := cellAt(t, title, tab.name)
			if got := g.side[p].tabs[i]; got.x != want || got.w != len(tab.name) {
				t.Fatalf("pane %d tab %q sits at %+v, but the title draws it at %d: %q", p, tab.name, got, want, title)
			}
		}
	}
	for p := pane(1); p < paneDone; p++ {
		if got := len(g.side[p].tabs); got != len(sidebar[p].tabs) {
			t.Fatalf("pane %d has %d click boxes, want %d", p, got, len(sidebar[p].tabs))
		}
	}
	if len(g.side[paneDone].tabs) != len(done.doneTabNames()) || len(g.side[paneActive].tabs) != 1 || len(g.detail.tabs) != 0 {
		t.Fatalf("Active %+v Done %+v detail %+v", g.side[paneActive].tabs, g.side[paneDone].tabs, g.detail.tabs)
	}
}

func TestSelectionIsPerTab(t *testing.T) {
	m := press(newModel(t), "2", "j", "]", "[")
	if m.Selected().ID != "specs/2026-09-22-beta" {
		t.Fatalf("the open tab lost its selection: %v", m.Selected())
	}
	m = press(newModel(t), "3", "5", "j", "]", "[")
	if m.Selected().ID != "plans/2026-09-27-orphan" {
		t.Fatalf("the finished tab lost its selection: %v", m.Selected())
	}
}

func TestSelectedFollowsTheLastFocusedListPane(t *testing.T) {
	m := press(newModel(t), "3", "5", "j", "0")
	if m.Selected().ID != "plans/2026-09-27-orphan" {
		t.Fatalf("the detail should keep showing the Done box: %v", m.Selected())
	}
	// Moving the cursor inside the Done box keeps it the one on show.
	m = press(m, "shift+tab", "j", "0")
	if m.Selected().ID != "plans/2026-09-26-dash-tasks" {
		t.Fatalf("the detail should keep showing the Done box: %v", m.Selected())
	}
	// A kind pane takes over as soon as it has the focus.
	m = press(m, "3")
	if m.Selected().ID != "plans/2026-09-23-lonely" {
		t.Fatalf("the Plans box should show its own row: %v", m.Selected())
	}
}

func TestUntypedGroupRow(t *testing.T) {
	m := press(newModel(t), "2")
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
	if got := rowIDs(press(newModel(t), "3")); len(got) != 2 {
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
	m := press(newModel(t), "3", "]", "s")
	if m.popup != nil || !strings.Contains(m.status, "checkboxes") {
		t.Fatalf("task popup: %v %q", m.popup, m.status)
	}
	m = press(newModel(t), "2", "G", "enter", "j", "t")
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
	m = press(m, "4", "j", "s")
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
	m := press(newModel(t), "4", "t", "esc")
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
	m := press(newModel(t), "2", "j") // the second spec of the newest date
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

// openTab puts a box on the tab with that number, the way a reader gets there.
func openTab(t *testing.T, m Model, p pane, tab int) Model {
	t.Helper()
	m = press(m, keyOf(p))
	for range len(sidebar[p].tabs) {
		if m.tab[p] == tab {
			return m
		}
		m = press(m, "]")
	}
	t.Fatalf("pane %d has no tab %d", p, tab)
	return m
}

// openTabBefore puts a box on the tab before that number, so a click that
// lands on nothing cannot pass by leaving the tab where it already was.
func openTabBefore(t *testing.T, m Model, p pane, tab int) Model {
	t.Helper()
	n := len(sidebar[p].tabs)
	return openTab(t, m, p, (tab+n-1)%n)
}

// doneTabBefore puts the Done box on another finished tab than that number. A
// tab with a single name stays where it is, and the click is still checked
// against what the title draws.
func doneTabBefore(t *testing.T, m Model, tab int) Model {
	t.Helper()
	m = press(m, "5")
	other := (tab + 1) % len(m.doneTabNames())
	if other == tab {
		return m
	}
	for range 2 {
		if m.tab[paneDone] == other {
			break
		}
		m = press(m, "]")
	}
	if m.tab[paneDone] != other {
		t.Fatalf("could not open finished tab %d", other)
	}
	return m
}

// TestClickLandsOnEveryDrawnTabName is the property behind the mouse: the first
// and the last letter of every tab name a title draws, in every kind box and in
// the Done box, at every window size from 60 to 200 columns, must switch to
// that tab; a click on the dashes between two names must switch nothing. The
// letters come from the drawn title, so a click box that drifts from the view
// fails here.
func TestClickLandsOnEveryDrawnTabName(t *testing.T) {
	widths := clickWidths()
	clicked := map[string]bool{}
	m := sized(newModel(t), widths[0], 40)

	for _, w := range widths {
		for p := pane(1); p < paneDone; p++ {
			for tab := range sidebar[p].tabs {
				name := sidebar[p].tabs[tab].name
				for _, end := range []string{"first", "last"} {
					m = sized(m, w, 40)
					m = openTabBefore(t, m, p, tab)
					y, x, ok := drawnLetter(m, p, name, end)
					if !ok {
						continue // a narrow title drops the names that do not fit
					}
					if got, _, idx := m.hit(x, y); got != p || idx != tab {
						t.Fatalf("at %d columns the %s letter of %q maps to box %d tab %d, want tab %d", w, end, name, got+1, idx, tab)
					}
					if got := click(m, x, y).tab[p]; got != tab {
						t.Fatalf("at %d columns a click on the %s letter of %q gave tab %d, want %d", w, end, name, got, tab)
					}
					clicked[name] = true
				}
			}
		}
	}
	for p := pane(1); p < paneDone; p++ {
		for _, tab := range sidebar[p].tabs {
			if !clicked[tab.name] {
				t.Errorf("no width from 60 to 200 drew %q, so it was never clicked", tab.name)
			}
		}
	}

	// The Done box follows the box and tab that had the focus last: Done and
	// Dropped for the specs, Specced and Dropped for the ideas, Fixed and
	// Wontfix for the bugs. Tasks and Active have no second name.
	for _, w := range widths {
		for p := pane(1); p < paneDone; p++ {
			for tab := range sidebar[p].tabs {
				m = sized(m, w, 40)
				m = openTab(t, m, p, tab)
				for i, name := range m.doneTabNames() {
					for _, end := range []string{"first", "last"} {
						m = doneTabBefore(t, m, i)
						y, x, ok := drawnLetter(m, paneDone, name, end)
						if !ok {
							t.Fatalf("with %q open, at %d columns the Done box does not draw %q", sidebar[p].tabs[tab].name, w, name)
						}
						if got, _, idx := m.hit(x, y); got != paneDone || idx != i {
							t.Fatalf("with %q open, at %d columns the %s letter of %q maps to box %d tab %d, want tab %d", sidebar[p].tabs[tab].name, w, end, name, got+1, idx, i)
						}
						if got := click(m, x, y).tab[paneDone]; got != i {
							t.Fatalf("with %q open, at %d columns a click on the %s letter of %q gave finished tab %d, want %d", sidebar[p].tabs[tab].name, w, end, name, got, i)
						}
					}
				}
			}
		}
	}

	// The dashes between two names belong to no tab.
	for _, w := range widths {
		for p := pane(1); p < paneDone; p++ {
			m = sized(m, w, 40)
			m = openTabBefore(t, m, p, 1)
			before := m.tab[p]
			for _, name := range sidebar[p].tabs[:len(sidebar[p].tabs)-1] {
				y, last, ok := drawnLetter(m, p, name.name, "last")
				if !ok {
					continue
				}
				x := last + 1 // the separator cell right after the name, dash or plain space
				if _, _, idx := m.hit(x, y); idx != -1 {
					t.Fatalf("at %d columns the dash after %q maps to tab %d", w, name.name, idx)
				}
				if got := click(m, x, y).tab[p]; got != before {
					t.Fatalf("at %d columns a click on the dash after %q moved to tab %d", w, name.name, got)
				}
			}
		}
		for p := pane(1); p < paneDone; p++ {
			for tab := range sidebar[p].tabs {
				if len(sidebar[p].tabs[tab].done) < 2 {
					continue
				}
				m = sized(openTab(t, m, p, tab), w, 40)
				m = doneTabBefore(t, m, 1)
				before := m.tab[paneDone]
				y, last, _ := drawnLetter(m, paneDone, m.doneTabNames()[0], "last")
				x := last + 1
				if _, _, idx := m.hit(x, y); idx != -1 {
					t.Fatalf("at %d columns the dash after %q maps to finished tab %d", w, m.doneTabNames()[0], idx)
				}
				if got := click(m, x, y).tab[paneDone]; got != before {
					t.Fatalf("at %d columns a click on the dash after %q moved to finished tab %d", w, m.doneTabNames()[0], got)
				}
			}
		}
	}
}

// TestTabTitleAlwaysShowsTheOpenTab is the Task 8 regression: the title of a
// kind box used to drop the open tab's own name off the end, so no tab looked
// selected and a click on its old spot fell into the detail box. This checks,
// at every width the review named, for every open tab, that the title always
// keeps that tab's name, and that any other name still drawn clicks to the
// right tab.
func TestTabTitleAlwaysShowsTheOpenTab(t *testing.T) {
	m := newModel(t)
	for _, w := range []int{60, 80, 100, 120, 140, 200} {
		for p := pane(1); p < paneDone; p++ {
			for tab := range sidebar[p].tabs {
				m = sized(openTab(t, m, p, tab), w, 40)
				title := strings.Split(plain(m.View()), "\n")[m.geometry().at(p).y]
				if !strings.Contains(title, sidebar[p].tabs[tab].name) {
					t.Fatalf("at %d columns with %q open, the title drops the open tab: %q", w, sidebar[p].tabs[tab].name, title)
				}
				if w == 120 {
					for _, other := range sidebar[p].tabs {
						if !strings.Contains(title, other.name) {
							t.Errorf("at 120 columns the title is missing %q: %q", other.name, title)
						}
					}
				}
				for other := range sidebar[p].tabs {
					for _, end := range []string{"first", "last"} {
						y, x, ok := drawnLetter(m, p, sidebar[p].tabs[other].name, end)
						if !ok {
							continue // a narrow title drops names other than the open one
						}
						if got := click(m, x, y).tab[p]; got != other {
							t.Fatalf("at %d columns a click on the %s letter of %q gave tab %d, want %d", w, end, sidebar[p].tabs[other].name, got, other)
						}
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
			m.tab[paneSpecs], m.tab[paneDone] = 0, 0
			m.follows = paneSpecs
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

// openGroupIDs reads the item ids of a kind pane with the group row dropped, so
// a test sees only the item order.
func openGroupIDs(m Model) (going, rest []string) {
	for _, r := range m.openRows(paneSpecs) {
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

func TestEveryTabHoldsItsOpenItemsInBoardOrder(t *testing.T) {
	m := splitModel(t)
	for p := pane(1); p < paneDone; p++ {
		for tab := range sidebar[p].tabs {
			kind := sidebar[p].tabs[tab].kind
			var want []string
			for _, it := range m.board.List(kind, false) {
				want = append(want, it.ID)
			}
			if kind == board.KindStory {
				want = append(want, groupRowID)
			}
			mm := openTab(t, m, p, tab)
			if got := strings.Join(ids(mm.rowsOf(p)), " "); got != strings.Join(want, " ") {
				t.Errorf("pane %d tab %d holds %q, want %q", p, tab, got, strings.Join(want, " "))
			}
		}
	}
}

func TestStartedTaskWithNoTicksCountsAsInProgress(t *testing.T) {
	m := newModel(t)
	lonely := m.board.Get("plans/2026-09-23-lonely#task-1")
	if lonely.Done != 0 {
		t.Fatalf("the lonely task should have no ticked box, got %d", lonely.Done)
	}
	lonely.Started = true
	lonely.Status = "in-progress"
	m = openTab(t, m, panePlans, 1)
	rows := ids(m.openRows(panePlans))
	// Both tasks are under way, and neither one hides behind a rule row.
	if len(rows) != 2 || !slices.Contains(rows, lonely.ID) {
		t.Fatalf("the started task is missing from the Tasks box, rows %v", rows)
	}
}

// A task someone started before ticking a box is work under way, so its row
// wears the accent and ends on its count and its agent, the same as a task
// with a box ticked.
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
		m = press(sized(focused(m, panePlans, 1), 200, 40), "3", "j")
		it := b.Get("plans/2026-09-21-a#task-1")
		if it == nil || it.Status != "in-progress" || it.Done != 0 || it.Total != 2 {
			t.Fatalf("the started task = %+v, want in-progress at 0 of 2", it)
		}
		bx := paneBox(m, panePlans)
		lines := innerLines(m, bx)
		if len(lines) < 3 {
			t.Fatalf("the pane drew %d lines: %q", len(lines), lines)
		}
		if got := plain(lines[0]); !strings.HasSuffix(strings.TrimRight(got, " "), "0/2 · omp") {
			t.Errorf("the started row is %q, want it to end on 0/2 · omp", got)
		}
		if got := plain(lines[1]); !strings.Contains(got, "Two") {
			t.Errorf("line 1 is %q, want the second task", got)
		}
		// The cursor has moved on, so row 0 is drawn like every other
		// in-progress row the reader is not on.
		if row := paintedLine(m.View(), bx, 0); !wears(row, 2) || !wears(row, 38, 5, 39) {
			t.Errorf("the started row should be dim and wear the accent: %q", row)
		}
	})
}

func TestUnknownStatusIsListedAndNotInProgress(t *testing.T) {
	m := press(newModel(t), "2")
	found := false
	for _, r := range m.openRows(paneSpecs) {
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
	// Whatever mix of items a pane holds, the rows are the items themselves.
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
			rows := keep(full, tc.want).openRows(paneSpecs)
			if len(rows) != tc.n {
				t.Fatalf("rows %v, want %d", ids(rows), tc.n)
			}
		})
	}
	// A pane that holds both kinds of work at once still lists only items.
	rows := splitModel(t).openRows(paneSpecs)
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
	m := splitModel(t)
	m = press(m, "/", "A", "l", "p", "h", "a")
	rows := ids(m.openRows(paneSpecs))
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
	m := press(sized(splitModel(t), 120, 40), "2")
	g := m.geometry()
	b := g.at(paneSpecs)
	below := b.y + 1 + b.rows
	if below >= b.y+b.h {
		t.Skip("the pane is full, so there is no space below the rows")
	}
	before := m.openRows(paneSpecs)[m.cursor()].id
	after := click(m, 2, below)
	if got := after.openRows(paneSpecs)[after.cursor()].id; got != before {
		t.Fatalf("a click on the empty space moved from %s to %s", before, got)
	}
	if after.focus != paneSpecs {
		t.Fatalf("a click on the empty space moved the focus to pane %d", after.focus)
	}
}

func TestEnterFocusesDetailFromBothListPanes(t *testing.T) {
	// Every kind box, every tab.
	for p := pane(1); p < paneDone; p++ {
		for tab := range sidebar[p].tabs {
			// Enter on a plan row opens the plan instead, which the tree tests
			// in plantree_test.go pin.
			if sidebar[p].tabs[tab].kind == board.KindPlan {
				continue
			}
			m := openTab(t, newModel(t), p, tab)
			id := m.Selected().ID
			next, cmd := m.Update(key("enter"))
			m = next.(Model)
			if cmd != nil {
				t.Fatalf("pane %d tab %d: enter opened the editor", p, tab)
			}
			if m.focus != paneDetail {
				t.Fatalf("pane %d tab %d: enter left the focus on pane %d", p, tab, m.focus)
			}
			if m.Selected() == nil || m.Selected().ID != id {
				t.Fatalf("pane %d tab %d: enter moved from %s to %v", p, tab, id, m.Selected())
			}
		}
	}
	// The Done box.
	m := press(newModel(t), "4", "5")
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
	m := press(newModel(t), "2", "G", "enter")
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
	m = press(sized(m, 120, 40), "4")
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
	// Every kind box, every tab.
	for p := pane(1); p < paneDone; p++ {
		for tab := range sidebar[p].tabs {
			m := openTab(t, newModel(t), p, tab)
			// Tasks refuse the editor: they take status from checkboxes.
			if p == panePlans && tab == 1 {
				m = press(m, "e")
				if m.Selected() == nil {
					t.Fatalf("pane %d tab %d: nothing selected", p, tab)
				}
				continue
			}
			if _, cmd := m.Update(key("e")); cmd == nil {
				t.Fatalf("pane %d tab %d: e opened no editor", p, tab)
			}
		}
	}
	// The Done box.
	if _, cmd := press(newModel(t), "3", "5").Update(key("e")); cmd == nil {
		t.Fatal("e in the Done box opened no editor")
	}
	// The detail box, from both kinds of list box.
	for _, keys := range [][]string{{"0"}, {"5", "0"}, {"3", "0"}} {
		if _, cmd := press(newModel(t), keys...).Update(key("e")); cmd == nil {
			t.Fatalf("e in the detail box after %v opened no editor", keys)
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
	m = press(sized(m, 120, 40), "4")
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
	// The Specs box -> the detail -> the Specs box.
	m := press(newModel(t), "2", "enter")
	if m.focus != paneDetail || m.last != paneSpecs {
		t.Fatalf("enter: focus %d last %d", m.focus, m.last)
	}
	m = press(m, "esc")
	if m.focus != paneSpecs {
		t.Fatalf("esc: focus %d", m.focus)
	}
	// The Done box -> the detail -> the Done box.
	m = press(newModel(t), "5", "enter")
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
