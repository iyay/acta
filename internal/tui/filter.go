package tui

import (
	"slices"
	"strings"

	"github.com/iyay/acta/internal/board"
)

// listFilter is what f and F picked on one tab. An empty value lets every
// item through.
type listFilter struct {
	priority string // a level from board.Priorities, noPriority, or ""
	status   string
}

// noPriority is the filter value that keeps only items with no priority.
const noPriority = "none"

// priorityChoices is the round f walks, All first.
var priorityChoices = append(append([]string{""}, board.Priorities...), noPriority)

// hasPriority says whether the items of a tab carry a priority at all.
func hasPriority(k board.Kind) bool {
	return k == board.KindBug || k == board.KindDebtItem
}

// statusChoices is the round F walks on a tab, All first. The List box shows
// only open items, so the closed statuses would always give an empty list.
// The Plans tree is picked by the status of its tasks.
func statusChoices(k board.Kind) []string {
	if k == board.KindPlan {
		k = board.KindTask
	}
	out := []string{""}
	for _, s := range board.Allowed(k) {
		if !board.Closed(s) {
			out = append(out, s)
		}
	}
	return out
}

// next gives the value after cur in a round, back to the first after the last.
func next(round []string, cur string) string {
	return round[(slices.Index(round, cur)+1)%len(round)]
}

// cycleFilter moves the priority filter (f) or the status filter (F) of the
// open tab one step. A tab with no priority treats f as F, so the reader
// does not have to remember which key works where.
func (m *Model) cycleFilter(priority bool) {
	kind := topTabs[m.top].kind
	if kind == "" {
		m.status = "Activity has no filter"
		return
	}
	f := m.filters[m.top]
	if priority && hasPriority(kind) {
		f.priority = next(priorityChoices, f.priority)
	} else {
		round := statusChoices(kind)
		if len(round) < 3 {
			m.status = topTabs[m.top].name + " has only one open status, so there is nothing to filter"
			return
		}
		f.status = next(round, f.status)
	}
	m.filters[m.top] = f
	m.moveTo(0)
}

// filterOn says the open tab has a filter set.
func (m Model) filterOn() bool {
	return m.filters[m.top] != listFilter{}
}

// filterRows keeps the rows of the List box whose item passes the filter of
// the open tab. A plan stays when a task under it passes.
func (m Model) filterRows(rows []row) []row {
	if !m.filterOn() {
		return rows
	}
	f := m.filters[m.top]
	return keepRows(rows, func(r row) bool {
		it := m.board.Get(r.id)
		if it == nil {
			return false
		}
		if f.priority == noPriority && it.Priority != "" {
			return false
		}
		if f.priority != "" && f.priority != noPriority && it.Priority != f.priority {
			return false
		}
		return f.status == "" || it.Status == f.status
	})
}

// filterLabel is what the List box title adds for the filters of the open
// tab, like " [H · fixing]", or "" when none is set.
func (m Model) filterLabel() string {
	f := m.filters[m.top]
	var parts []string
	switch f.priority {
	case "":
	case noPriority:
		parts = append(parts, noPriority)
	default:
		parts = append(parts, strings.ToUpper(f.priority[:1]))
	}
	if f.status != "" {
		parts = append(parts, f.status)
	}
	if len(parts) == 0 {
		return ""
	}
	return " [" + strings.Join(parts, " · ") + "]"
}

// filterHints names the filter keys that do something on the open tab.
func (m Model) filterHints() []string {
	kind := topTabs[m.top].kind
	var out []string
	if hasPriority(kind) {
		out = append(out, "Priority filter: f")
	}
	if kind != "" && len(statusChoices(kind)) > 2 {
		out = append(out, "Status filter: F")
	}
	return out
}
