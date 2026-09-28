package tui

import (
	"fmt"
	"sort"

	"github.com/iyay/acta/internal/board"
)

// doneTab is one tab of the Done pane: the name in its title and the status it
// lists. A kind with only one finished status has no second tab, and it never
// shows.
type doneTab struct {
	name   string
	status string
}

// sidebarTab is one tab of a sidebar pane: the name in its title, the kind of
// item it lists, and the finished tabs the Done pane shows while this tab has
// the focus.
type sidebarTab struct {
	name string
	kind board.Kind
	done []doneTab
}

// sidebarPane is one stacked box on the left. Its tabs are the kinds it shows;
// Active and Done have none, so their title is all they draw.
type sidebarPane struct {
	title string
	tabs  []sidebarTab
}

// sidebar is the whole left column, top to bottom. Index i is key i+1. It is an
// array, so len(sidebar) is a constant the pane numbers are built from.
var sidebar = [...]sidebarPane{
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
	// paneActive is the top box: every item whose work has begun.
	paneActive = pane(0)
	// paneSpecs, panePlans and paneBugs are the three kind boxes, which the
	// tests read by name; TestSidebarTitles pins them to the table.
	paneSpecs = pane(1)
	panePlans = pane(2)
	paneBugs  = pane(3)
	// paneDone is the bottom box: the finished items of the box that had the
	// focus last.
	paneDone = pane(len(sidebar) - 1)
	// paneDetail is the box on the right, which sits outside the table and is
	// key 0.
	paneDetail = pane(len(sidebar))
)

// boxes is every box on screen: the sidebar column and the detail box, so tab
// walks the whole ring and the numbers are read from the table.
const boxes = len(sidebar) + 1

// paneKey is the number a box wears in its title: the sidebar boxes count from
// 1 down the column, and the detail box is 0, so the reader always presses the
// number the box shows.
func paneKey(p pane) string {
	if p == paneDetail {
		return "─[0]─"
	}
	return fmt.Sprintf("─[%d]─", p+1)
}

// tabOf gives the tab a pane has open, and false when it has no tabs at all:
// Active, Done and the detail box have none.
func (m Model) tabOf(p pane) (sidebarTab, bool) {
	if p >= paneDone {
		return sidebarTab{}, false
	}
	tabs := sidebar[p].tabs
	if len(tabs) == 0 {
		return sidebarTab{}, false
	}
	return tabs[clamp(m.tab[p], 0, len(tabs)-1)], true
}

// doneTabOf gives the finished tab the Done pane has open for the pane the
// Done pane follows, and false when that pane has no kind of its own.
func (m Model) doneTabOf() (doneTab, bool) {
	tab, ok := m.tabOf(m.follows)
	if !ok {
		return doneTab{}, false
	}
	return tab.done[clamp(m.tab[paneDone], 0, len(tab.done)-1)], true
}

// doneTabNames gives the tab names the Done pane shows for the tab the pane it
// follows has open. Active has no kind, so after it Done shows no tabs.
func (m Model) doneTabNames() []string {
	tab, ok := m.tabOf(m.follows)
	if !ok {
		return nil
	}
	var out []string
	for _, d := range tab.done {
		if d.name != "" {
			out = append(out, d.name)
		}
	}
	return out
}

// tabsOf gives the names a pane draws in its title: the tabs it has, or the
// title of the box itself when it has none.
func (m Model) tabsOf(p pane) []string {
	switch p {
	case paneDetail:
		return nil
	case paneActive:
		return []string{sidebar[p].title}
	case paneDone:
		if names := m.doneTabNames(); len(names) > 0 {
			return names
		}
		return []string{sidebar[p].title}
	}
	out := make([]string, len(sidebar[p].tabs))
	for i, tab := range sidebar[p].tabs {
		out[i] = tab.name
	}
	return out
}

// onTab gives the tab a pane has open, so its title never drops that name even
// when the pane is too narrow to draw every tab.
func (m Model) onTab(p pane) int {
	if p == paneDone {
		return clamp(m.tab[paneDone], 0, max(0, len(m.doneTabNames())-1))
	}
	if p >= paneDone || len(sidebar[p].tabs) == 0 {
		return 0
	}
	return clamp(m.tab[p], 0, len(sidebar[p].tabs)-1)
}

// rowsOf gives the rows a pane shows: the items in progress, the items of its
// tab, or the finished items it follows.
func (m Model) rowsOf(p pane) []row {
	switch p {
	case paneActive:
		return m.activeRows()
	case paneDone:
		return m.doneRows()
	}
	return m.openRows(p)
}

// searchRows gives a searching reader every item whose text matches, whatever
// pane asked, so one search never depends on which box the focus is in.
func (m Model) searchRows() []row {
	return toRows(m.board.Search(m.query), 0)
}

// activeRows gives the Active pane every item whose work has begun, of every
// kind, in the order the board lists them. inProgress is the one rule that
// decides, so an item that is only open, raw or todo can never get in.
func (m Model) activeRows() []row {
	if m.query != "" {
		return m.searchRows()
	}
	var going []*board.Item
	for _, it := range m.board.Items {
		if !it.Legacy && inProgress(it) {
			going = append(going, it)
		}
	}
	return toRows(going, 0)
}

// openRows gives a pane the items of its tab that are not finished, in the
// order the board lists them. The files outside the root folder have no tab of
// their own, so the Specs tab ends with a single row that opens them.
func (m Model) openRows(p pane) []row {
	if m.query != "" {
		return m.searchRows()
	}
	tab, ok := m.tabOf(p)
	if !ok {
		return nil
	}
	if tab.kind == board.KindPlan {
		return m.treeRows(m.board.List(board.KindPlan, false))
	}
	rows := toRows(m.board.List(tab.kind, false), 0)
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

// toggleExpand gives the focused sidebar pane the room of the whole column,
// or takes the room back when it already has it. The detail box sits outside
// the column, so z there does nothing. The pane that grows shows more rows, so
// its offset goes back inside what it can really show.
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

// doneRows gives the Done pane the finished items of the tab the sidebar pane
// that had the focus last has open, the ones touched last at the top: the date
// of the last commit on the file decides, and the date in the file name is the
// fallback for a file git has no commit for. After the Active pane, which has
// no kind of its own, it lists every closed item in that same order.
func (m Model) doneRows() []row {
	var finished []*board.Item
	if tab, ok := m.tabOf(m.follows); ok {
		d, _ := m.doneTabOf()
		for _, it := range m.board.List(tab.kind, true) {
			if it.Status == d.status {
				finished = append(finished, it)
			}
		}
	} else {
		for _, it := range m.board.Items {
			if !it.Legacy && board.Closed(it.Status) {
				finished = append(finished, it)
			}
		}
	}
	sort.SliceStable(finished, func(i, j int) bool {
		return finished[i].SortTime() > finished[j].SortTime()
	})
	if tab, ok := m.tabOf(m.follows); ok && tab.kind == board.KindPlan {
		return m.treeRows(finished)
	}
	return toRows(finished, 0)
}
