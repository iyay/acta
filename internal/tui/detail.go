package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"

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

// footLines is how tall the detail footer is: the rule and the date line.
const footLines = 2

// rule is the line that sets the header and the footer off from the middle.
func (m Model) rule(w int) string {
	return m.styles.faint.Render(strings.Repeat("─", max(1, w)))
}

// dot is the brush of a status dot: green when done, the first frame of the
// pulse while the work is under way, grey while it waits.
func (s styles) dot(mark string) lipgloss.Style {
	switch mark {
	case dotDone:
		return s.done
	case dotGoing:
		return s.pulse[0]
	}
	return s.waiting
}

// paintDates colors the names in the date line, so the dates stand out from
// the words around them. Each part is "name date", split by " · ".
func paintDates(s styles, line string) string {
	parts := strings.Split(line, " · ")
	for i, p := range parts {
		if name, rest, ok := strings.Cut(p, " "); ok {
			parts[i] = s.footLabel.Render(name) + " " + rest
		}
	}
	return strings.Join(parts, " · ")
}

// detailCache keeps the last parts of the detail box. The box is asked for its
// lines several times on every key and wheel notch, and building them renders
// the whole markdown body each time. The model is copied on every update, so
// the cache sits behind a pointer that all the copies share.
type detailCache struct {
	board *board.Board
	item  *board.Item
	width int
	head  []string
	mid   []string
	foot  string
}

// detailParts gives the header, the scrolling part and the date line of the
// detail box, built again only when the item, the width or the board changed.
// A reload makes a new board, so its items are new too and the old lines are
// never shown for them.
//
// With no item under the cursor the message comes from the list instead, and
// the key above cannot see the list, so that box is built every time. It is
// one short line with no markdown in it, so it costs nothing to build again.
func (m Model) detailParts(w int) (head, mid []string, foot string) {
	c := m.dcache
	it := m.Selected()
	if c == nil || it == nil {
		return m.buildDetailParts(w)
	}
	if c.mid != nil && c.board == m.board && c.item == it && c.width == w {
		return c.head, c.mid, c.foot
	}
	head, mid, foot = m.buildDetailParts(w)
	*c = detailCache{board: m.board, item: it, width: w, head: head, mid: mid, foot: foot}
	return head, mid, foot
}

// buildDetailParts splits the detail in three: the header that stays on top,
// the part that scrolls, and the date line that stays at the bottom. Labels are
// upper case, padded to one width, with the colons in one column, and a line
// with no value is left out. With no item on show there is only a message,
// so there is no header and no footer.
func (m Model) buildDetailParts(w int) (head, mid []string, foot string) {
	it := m.Selected()
	if it == nil {
		if len(m.board.Items) == 0 {
			return nil, cut("this repo has no .acta/ yet.\n\npress n to write the first bug, or let the agent plugin create specs and plans.", w), ""
		}
		if len(m.listOf()) == 0 {
			return nil, []string{m.styles.faint.Render("No items")}, ""
		}
		return nil, []string{m.styles.faint.Render("enter opens the group")}, ""
	}
	title := it.Title
	if it.Kind == board.KindDebtItem {
		// A note is too long for one header line, so it goes in the middle.
		title = ""
	}
	fields := []struct{ label, value string }{
		{"ID", idText(it)},
		{kindLabel(it.Kind), title},
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
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		line := truncate(expandTabs(fmt.Sprintf("%-*s: %s", width, f.label, f.value)), w)
		// The id is painted first, because the label then goes on the plain
		// text that is left over.
		if f.label == "ID" {
			line = m.paintID(line, it, lipgloss.NewStyle())
		}
		// A label cut off by a narrow pane is left as it is.
		if f.label != "" && strings.HasPrefix(line, f.label) {
			line = m.styles.label.Render(f.label) + line[len(f.label):]
		}
		head = append(head, line)
	}
	head = append(head, m.rule(w))
	for _, p := range it.Problems {
		mid = append(mid, m.styles.problem.Render(truncate(expandTabs("! "+p), w)))
	}
	if it.Kind == board.KindDebtItem {
		// The list pane already names the other notes, so the middle is this
		// note alone, drawn like the body of a spec so both look the same.
		// The renderer drops text that looks like an HTML tag, such as
		// <uid>, so those signs are escaped first. It can also leave a line
		// wider than the pane, and fit would cut it, so Hardwrap breaks it.
		note := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(expandTabs(it.Title))
		for _, ln := range strings.Split(xansi.Hardwrap(m.render(note, w), w, true), "\n") {
			mid = append(mid, fit(ln, w))
		}
		return head, mid, paintDates(m.styles, dateLine(it, w))
	}
	mid = append(mid, m.workLines(it, w)...)
	for _, ln := range strings.Split(m.render(expandTabs(it.Body), w), "\n") {
		mid = append(mid, fit(ln, w))
	}
	return head, mid, paintDates(m.styles, dateLine(it, w))
}

// detailLines is the whole detail as one block: the header, the middle, the
// rule and the date line. A box too short to keep the header on top scrolls
// this.
func (m Model) detailLines(w int) []string {
	head, mid, foot := m.detailParts(w)
	out := append(append([]string(nil), head...), mid...)
	if foot != "" {
		out = append(out, m.rule(w), foot)
	}
	return out
}

// dateLine is the footer of the detail. It names all three dates, and a dash
// stands in for a date that is not set, so the line keeps one shape. A box
// too narrow for the long names takes the short ones, so every date stays
// readable, and only a box too narrow for those gets a cut line.
func dateLine(it *board.Item, w int) string {
	or := func(s string) string {
		if s == "" {
			return "-"
		}
		return s
	}
	c, s, f := or(it.Created), or(it.StartedOn), or(it.Finished)
	long := "created " + c + " · started " + s + " · finished " + f
	if lipgloss.Width(long) <= w {
		return long
	}
	short := "c " + c + " · s " + s + " · f " + f
	if lipgloss.Width(short) <= w {
		return short
	}
	return truncate(short, w)
}

// stickyMid is how many middle lines a detail box h lines tall shows under a
// header of head lines and over the two footer lines. It is 0 when the box has
// no room for the header, the two footer lines and 3 middle lines: then the
// whole detail scrolls as one block, so a short box still shows everything.
func stickyMid(head, h int) int {
	if n := h - head - footLines; n >= 3 {
		return n
	}
	return 0
}

// detailScroll says how many lines the detail scrolls over and how many of
// them the box shows at once: the middle alone when the header and the
// footer stick, the whole block when they do not.
func (m Model) detailScroll(w, h int) (total, fit int) {
	head, mid, foot := m.detailParts(w)
	if n := stickyMid(len(head), h); n > 0 && foot != "" {
		return len(mid), n
	}
	return len(m.detailLines(w)), h
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
// every plan under it, a debt file its lines. An item with no work at all
// lists nothing and leaves no empty line.
func (m Model) workLines(it *board.Item, w int) []string {
	switch it.Kind {
	case board.KindTask:
		return m.stepLines(it, w)
	case board.KindDebt:
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
		out = append(out, m.paintID(truncate(expandTabs(shortRef(it)+"  "+it.Title), w), it, lipgloss.NewStyle()))
		out = append(out, m.taskLines(it, w)...)
	}
	return out
}

// debtLines are the checklist lines of a debt file, one line each.
func (m Model) debtLines(file *board.Item, w int) []string {
	var out []string
	for _, id := range file.Children {
		if line := m.board.Get(id); line != nil {
			out = append(out, workLine(m.styles, line, false, w))
		}
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
			mark, brush := dotWaiting, lipgloss.NewStyle()
			switch {
			case s.State != ' ':
				mark, brush = dotDone, m.styles.done
			case going:
				mark, going = dotGoing, false
			}
			line := truncate(expandTabs(mark+" "+s.Text), w)
			if strings.HasPrefix(line, mark) {
				line = m.styles.dot(mark).Render(mark) + brush.Render(line[len(mark):])
			} else {
				line = brush.Render(line)
			}
			out = append(out, line)
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
// finished line reads green so the eye skips it, the ID keeps the color of
// its kind, the line the reader is on is bold, and the dot wears the color of
// its state.
func workLine(s styles, it *board.Item, on bool, w int) string {
	mark, brush := dotOf(it), lipgloss.NewStyle()
	if mark == dotDone {
		// Done work reads green, so the eye skips it.
		brush = s.done
	}
	if on {
		brush = brush.Bold(true)
	}
	text := shortRef(it) + "  " + it.Title
	if inProgress(it) {
		text += "  " + progressText(it)
		if it.Agent != "" {
			text += " · " + it.Agent
		}
	}
	line := truncate(expandTabs(mark+" "+text), w)
	if !strings.HasPrefix(line, mark) {
		return brush.Render(line)
	}
	return s.dot(mark).Render(mark) + s.paintID(line[len(mark):], it, brush)
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
