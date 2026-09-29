package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
)

// The dots of the list between the header and the body: work under way wears
// the accent, work not begun and work finished wear nothing, so a long list
// reads at a glance.
const (
	dotGoing   = "●"
	dotWaiting = "○"
	dotDone    = "✓"
)

// detailLines draws the header, the list of work and the body of the item on
// show. Labels are upper case, padded to one width, with the colons in one
// column, and a line with no value is left out.
func (m Model) detailLines(w int) []string {
	it := m.Selected()
	if it == nil {
		if len(m.board.Items) == 0 {
			return cut("this repo has no .acta/ yet.\n\npress n to write the first bug, or let the agent plugin create specs and plans.", w)
		}
		if len(m.listOf()) == 0 {
			return []string{m.styles.faint.Render("No items")}
		}
		return []string{m.styles.faint.Render("enter opens the group")}
	}
	fields := []struct{ label, value string }{
		{"ID", idText(it)},
		{kindLabel(it.Kind), it.Title},
		{"STATUS", it.Status},
		{"AUTHOR", it.Author},
		{"FROM", m.fromText(it)},
		{"REF", it.Ref},
		{"SPEC", m.specText(it)},
		linkField("CLOSES", m.closesText(it.Closes)),
		linkField("CLOSED BY", m.closesText(it.ClosedBy)),
		{"WORKTREE", worktreeText(it)},
		{"AGENT", it.Agent},
		{"FILE", m.fileText(it)},
		{tasksLabel(it.Kind), progressText(it)},
		{"FIXED", it.FixedIn},
	}
	width := 0
	for _, f := range fields {
		width = max(width, len(f.label))
	}
	width += 2
	var lines []string
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		lines = append(lines, truncate(expandTabs(fmt.Sprintf("%-*s: %s", width, f.label, f.value)), w))
	}
	for _, p := range it.Problems {
		lines = append(lines, truncate(expandTabs("! "+p), w))
	}
	lines = append(lines, m.styles.faint.Render(strings.Repeat("─", max(1, w))))
	lines = append(lines, m.workLines(it, w)...)
	for _, ln := range strings.Split(m.render(expandTabs(it.Body), w), "\n") {
		lines = append(lines, fit(ln, w))
	}
	return lines
}

// linkField is a header line for a link the item may not have. An item with
// no link carries no label either, so the colons of the lines that are drawn
// stay in the column the other labels give them.
func linkField(label, value string) struct{ label, value string } {
	if value == "" {
		label = ""
	}
	return struct{ label, value string }{label, value}
}

// closesText names the items this one closes, or the ones that close it, each
// by its short id when the file carries one and by its path when it does not.
// No link gives no text, so the field drops out of the header.
func (m Model) closesText(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if it := m.board.Get(id); it != nil && it.ShortID != "" {
			id = it.ShortID
		}
		out = append(out, id)
	}
	return strings.Join(out, ", ")
}

// workLines are the lines of work an item holds, one line each. What a kind
// lists: a plan its tasks, a task its steps, a spec or a bug the tasks of
// every plan under it, a debt file or debt item the lines of the debt file.
// An item with no work at all lists nothing and leaves no empty line.
func (m Model) workLines(it *board.Item, w int) []string {
	switch it.Kind {
	case board.KindTask:
		return m.stepLines(it, w)
	case board.KindDebt, board.KindDebtItem:
		return m.debtLines(it, w)
	case board.KindPlan:
		return m.taskLines(it, w)
	case board.KindScratch:
		return m.scratchLines(it, w)
	}
	return m.planLines(it, w)
}

// scratchLines are the specs that came out of an idea, one line each. An
// idea is where a spec starts, so its detail names the specs it produced the
// way a plan names the spec it carries.
func (m Model) scratchLines(it *board.Item, w int) []string {
	var out []string
	for _, id := range it.Children {
		if s := m.board.Get(id); s != nil {
			out = append(out, workLine(m.styles, s, false, w))
		}
	}
	return out
}

// taskLines are the tasks of a plan, in file order.
func (m Model) taskLines(plan *board.Item, w int) []string {
	var out []string
	for _, id := range plan.Children {
		if t := m.board.Get(id); t != nil {
			out = append(out, workLine(m.styles, t, false, w))
		}
	}
	return out
}

// planLines are the tasks of every plan under a spec or a bug, one plain line
// naming each plan above its own tasks, because a plan is a file of its own
// and the reader has to know whose tasks they are.
func (m Model) planLines(parent *board.Item, w int) []string {
	var out []string
	for _, it := range m.board.Items {
		if it.Kind != board.KindPlan || it.SpecID != parent.ID {
			continue
		}
		out = append(out, truncate(expandTabs(shortRef(it)+"  "+it.Title), w))
		out = append(out, m.taskLines(it, w)...)
	}
	return out
}

// debtLines are the checklist lines of a debt file. The line on show is the
// bright one, the others are dim, so the reader knows which NOTE is open.
func (m Model) debtLines(it *board.Item, w int) []string {
	file, on := it, it.ID
	if it.Kind == board.KindDebtItem {
		if file = m.board.Get(it.Parent); file == nil {
			return nil
		}
	}
	var out []string
	for _, id := range file.Children {
		line := m.board.Get(id)
		if line == nil {
			continue
		}
		out = append(out, workLine(m.styles, line, id == on, w))
	}
	return out
}

// stepLines are the checklist boxes of a task. A step has no id of its own and
// no count, so its line is the dot and the words; the first box still open in
// a task under way is the step the reader is on.
func (m Model) stepLines(it *board.Item, w int) []string {
	going := inProgress(it)
	var out []string
	for _, sec := range board.Parse([]byte(it.Body)).Tasks {
		for _, s := range sec.Steps {
			mark, brush := dotWaiting, m.styles.faint
			switch {
			case s.State != ' ':
				mark = dotDone
			case going:
				mark, brush, going = dotGoing, m.styles.accent, false
			}
			out = append(out, brush.Render(truncate(expandTabs(mark+" "+s.Text), w)))
		}
	}
	return out
}

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

// workLine is one line of the list: the dot, the short ID, the title, and for
// work under way its count and its agent, the same tail a list row wears. A
// line the reader is on stays bright; the rest are dim, dot included.
func workLine(s styles, it *board.Item, on bool, w int) string {
	mark, brush := dotOf(it), s.faint
	if mark == dotGoing {
		brush = s.work
	}
	if on {
		brush = lipgloss.NewStyle()
	}
	text := shortRef(it) + "  " + it.Title
	if inProgress(it) {
		text += "  " + progressText(it)
		if it.Agent != "" {
			text += " · " + it.Agent
		}
	}
	return brush.Render(truncate(expandTabs(mark+" "+text), w))
}

// tasksLabel names the line that counts the work: a task counts its steps, so
// it says SUBTASKS, and everything else counts its tasks.
func tasksLabel(k board.Kind) string {
	if k == board.KindTask {
		return "SUBTASKS"
	}
	return "TASKS"
}

// kindLabel is the label of the line that names the kind and the title.
func kindLabel(k board.Kind) string {
	switch k {
	case board.KindPlan:
		return "PLAN"
	case board.KindTask:
		return "TASK"
	case board.KindBug:
		return "BUG"
	case board.KindDebtItem:
		return "DEBT"
	}
	return "SPEC"
}

// fromText names the plan a debt item's review came from, read straight off
// the debt file's own frontmatter, because the board only keeps that link
// long enough to check it, not to hand it back later.
func (m Model) fromText(it *board.Item) string {
	if it.Kind != board.KindDebtItem {
		return ""
	}
	raw, err := os.ReadFile(it.Path)
	if err != nil {
		return ""
	}
	planID, _ := board.Parse(raw).Front["parent"].(string)
	if planID == "" {
		return ""
	}
	if p := m.board.Get(planID); p != nil {
		return shortRef(p) + " · " + p.Title
	}
	return planID
}

// expandTabs turns every tab into spaces, so a line is as wide as the screen
// will draw it. lipgloss counts a tab as nothing while the terminal draws it
// up to the next stop of eight, and a line that comes out shorter than it
// looks spills over the wall of its pane.
func expandTabs(s string) string {
	if !strings.Contains(s, "\t") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 8)
	col := 0
	for _, r := range s {
		switch {
		case r == '\n':
			col = 0
		case r == '\t':
			spaces := 8 - col%8
			b.WriteString(strings.Repeat(" ", spaces))
			col += spaces
			continue
		default:
			col += lipgloss.Width(string(r))
		}
		b.WriteRune(r)
	}
	return b.String()
}
