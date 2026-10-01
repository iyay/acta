package tui

import (
	"maps"
	"slices"

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

// paneKey is what a box wears at the start of its title. No box has a key of
// its own any more: tab and enter reach the detail.
func paneKey(pane) string { return "─" }

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
		// Activities has no kind, so its box lists tasks and not the open
		// items every other tab shows.
		if topTabs[m.top].kind == "" {
			return []string{"Tasks"}
		}
		return []string{"Open"}
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

// activityRows gives Activities the tasks under way, each under its head: the
// plan it belongs to, or the bug that plan fixes. Heads and the tasks under
// a head keep the order of the pane. A task whose plan is gone has no head,
// so it stands on its own row at the end.
func (m Model) activityRows() []row {
	if m.query != "" {
		return m.searchRows()
	}
	newest := m.newest[paneList]
	under := map[string][]*board.Item{}
	var heads, loose []*board.Item
	for _, t := range m.board.List(board.KindTask, true) {
		if !inProgress(t) {
			continue
		}
		h := m.headOf(t)
		if h == nil {
			loose = append(loose, t)
			continue
		}
		if _, seen := under[h.ID]; !seen {
			heads = append(heads, h)
		}
		under[h.ID] = append(under[h.ID], t)
	}
	var out []row
	for _, h := range ordered(heads, newest) {
		out = append(out, row{id: h.ID, tree: true})
		if !m.isOpen(h.ID) {
			continue
		}
		for _, t := range ordered(under[h.ID], newest) {
			out = append(out, row{id: t.ID, depth: 1, tree: true})
		}
	}
	return append(out, toRows(ordered(loose, newest), 0)...)
}

// headOf is the row a task under way sits under on Activities: its plan, or
// the bug that plan fixes. It goes one level up only, so a spec above a plan
// never becomes a head.
func (m Model) headOf(t *board.Item) *board.Item {
	plan := m.board.Get(t.PlanID)
	if plan == nil {
		return nil
	}
	if p := m.board.Get(plan.SpecID); p != nil && p.Kind == board.KindBug {
		return p
	}
	return plan
}

// isOpen says whether a tree head shows its tasks. Plans start shut and
// Activities groups start open, so each tab keeps the set that differs from
// how it starts.
func (m Model) isOpen(id string) bool {
	if topTabs[m.top].kind == "" {
		return !m.shutActs[id]
	}
	return m.openPlans[id]
}

// setOpen opens or shuts a tree head. It builds a new map each time, so an
// older copy of the model keeps the tree it drew.
func (m *Model) setOpen(id string, open bool) {
	acts := topTabs[m.top].kind == ""
	src := m.openPlans
	if acts {
		src = m.shutActs
	}
	next := make(map[string]bool, len(src)+1)
	maps.Copy(next, src)
	if acts {
		next[id] = !open
		m.shutActs = next
	} else {
		next[id] = open
		m.openPlans = next
	}
	m.keepVisible(m.listPane())
}

// openRows gives the List box the items of the open tab that are not
// finished, by the date in the file name, oldest first unless the pane was
// flipped. Plans come as a tree. The files outside the root folder have no
// tab of their own, so the Specs tab ends with a single row that opens them.
func (m Model) openRows() []row {
	if m.query != "" {
		return m.searchRows()
	}
	tab := topTabs[m.top]
	items := ordered(m.board.List(tab.kind, false), m.newest[paneList])
	// Bugs and debt are where the user picks what to fix next, so the most
	// urgent ones come first.
	if tab.kind == board.KindBug || tab.kind == board.KindDebtItem {
		items = byPriority(items)
	}
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

// toggleExpand gives the focused list box the room of the whole column, or
// takes the room back when it already has it. The detail box has a column of
// its own to lose, so z there zooms it over the whole body instead. The box
// that grows shows more rows, so its offset goes back inside what it can
// really show.
func (m *Model) toggleExpand() {
	if m.expanded == int(m.focus) {
		m.expanded = -1
	} else {
		m.expanded = int(m.focus)
	}
	m.clampOff(m.focus)
}

// doneRows gives the Done box the finished items of the open tab's sub-tab,
// by the date in the file name, oldest first unless the pane was flipped.
// Plans come as a tree.
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
	finished = ordered(finished, m.newest[paneDone])
	if tab.tree {
		return m.treeRows(finished)
	}
	return toRows(finished, 0)
}
