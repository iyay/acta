package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
)

// Every box keeps its own offset: the first line it shows. A list scrolls the
// same way the detail does, so one rule, one column of numbers and one
// scrollbar serve every box.

// firstOf is the first line a pane shows: its offset pulled back inside the
// range the pane can really show, so an offset from before a resize, or from
// content that shrank, can never leave a pane scrolled past its own end.
func firstOf(off, total, visible int) int {
	return clamp(off, 0, max(0, total-visible))
}

// scrollbar says, line by line down the right wall of a pane, which lines the
// thumb covers: true where the window is, false everywhere else. A pane whose
// content fits has no window to point at, so every line is false.
func scrollbar(total, visible, first, h int) []bool {
	bar := make([]bool, max(0, h))
	if total <= visible || h <= 0 {
		return bar
	}
	// The thumb keeps the size of the window, so a pane that shows a tenth
	// of its content gets a thumb of a tenth of the wall.
	thumb := max(1, h*visible/total)
	pos := 0
	if last := total - visible; last > 0 {
		pos = clamp(first, 0, last) * (h - thumb) / last
	}
	for i := pos; i < pos+thumb; i++ {
		bar[i] = true
	}
	return bar
}

// itemCount is what a sidebar pane writes in its bottom border: the item the
// cursor is on, counting from one, out of all the items the pane holds. A
// pane that holds nothing has no item to count, so it reads 0 of 0.
func itemCount(selected, total int) string {
	if total == 0 {
		return "0 of 0"
	}
	return fmt.Sprintf("%d of %d", selected, total)
}

// boxOf measures one box the way geometry lays them out, also on a narrow
// terminal that shows only one of them: the model needs a height to clamp an
// offset with, whether or not the box is on screen.
func (m Model) boxOf(p pane) box {
	b := m.geometry().at(p)
	if b.h == 0 {
		return m.box(p, 0, barRows, m.width, max(0, m.height-barRows-1))
	}
	return b
}

// fitOf is how many lines pane p has room for right now. The detail box
// counts only its middle when its header and footer stick.
func (m Model) fitOf(p pane) int {
	b := m.boxOf(p)
	if p == paneDetail {
		_, fit := m.detailScroll(b.textW(), b.inner)
		return fit
	}
	return b.rows
}

// linesOf is how many lines pane p holds in all, as wide as it is now.
func (m Model) linesOf(p pane) int {
	return m.linesAt(p, m.boxOf(p).textW())
}

// linesAt is how many lines pane p holds when its text is w cells wide: the
// rows of a list, or the body of the detail.
func (m Model) linesAt(p pane, w int) int {
	if p == paneDetail {
		total, _ := m.detailScroll(w, m.boxOf(p).inner)
		return total
	}
	rows, _, _ := m.slotOf(p)
	return len(rows)
}

// lastOff is how far pane p can scroll: the last line of its content that
// still leaves the window full.
func (m Model) lastOff(p pane) int {
	return max(0, m.linesOf(p)-m.fitOf(p))
}

// scrollPane moves one pane by lines and stops at both ends, so no pane ever
// scrolls past its own content.
func (m *Model) scrollPane(p pane, lines int) {
	m.off[p] = clamp(m.off[p]+lines, 0, m.lastOff(p))
}

// keepVisible slides a list pane so the row under the cursor is on screen,
// and leaves the offset where it is when that row already is. The detail box has
// no cursor of its own, so it keeps whatever place it was left in.
func (m *Model) keepVisible(p pane) {
	if p == paneDetail {
		return
	}
	rows, sel, idx := m.slotOf(p)
	cur := cursorOf(rows, *sel, *idx)
	fit := m.fitOf(p)
	if cur < 0 || fit < 1 {
		return
	}
	off := firstOf(m.off[p], len(rows), fit)
	m.off[p] = clamp(off, cur+1-fit, cur)
}

// clampOff puts an offset back inside what its pane can show, which a resize
// can take away, and keeps the row under the cursor on screen.
func (m *Model) clampOff(p pane) {
	if p == paneDetail {
		m.off[p] = firstOf(m.off[p], m.linesOf(p), m.fitOf(p))
		return
	}
	m.keepVisible(p)
}

// listView draws the rows of a list box. Every row takes one line, so a click
// on any cell of a row lands on that row.
func (m Model) listView(p pane, w int, b box) []string {
	rows, sel, idx := m.slotOf(p)
	if len(rows) == 0 {
		return []string{m.styles.faint.Render("nothing here")}
	}
	cur := cursorOf(rows, *sel, *idx)
	out := make([]string, 0, b.rows*rowLines)
	for i := range b.rows {
		n := b.first + i
		if n >= len(rows) {
			break
		}
		it := m.board.Get(rows[n].id)
		r := rows[n]
		text := pad(m.rowText(r, it, w), w)
		if n == cur {
			out = append(out, m.styles.selected.Render(text))
			continue
		}
		base := lipgloss.NewStyle()
		done := it != nil && dotOf(it) == dotDone
		if done {
			// Done work reads green, so the eye skips it.
			base = m.styles.done
		}
		// A done plan keeps its + or - mark in the plain foreground, so the
		// reader can still see at a glance whether the plan is open. Only the
		// title after the mark says the work is done. The mark and the space
		// behind it leave the row, so the id keeps the color of its kind and
		// sits in the green part just as it does on any other done row.
		mark := ""
		if done && r.tree && r.depth == 0 {
			if i := strings.IndexByte(text, ' '); i >= 0 {
				mark, text = text[:i+1], text[i+1:]
			}
		}
		line := lipgloss.NewStyle().Render(mark) + m.paintID(text, it, base)
		if r.tree && r.depth > 0 && it != nil && dotOf(it) == dotGoing {
			// The tree dot of work under way is the pulse dot, so the view can
			// swap it for the frame of the moment. The mark comes before the
			// title, so the first dot on the line is the mark.
			line = strings.Replace(line, dotGoing, m.styles.goingDot, 1)
		}
		out = append(out, line)
	}
	return out
}

// paintID draws a row with its id in the color of its kind and the rest in
// base. A row with no item, or whose id was cut off, is all base.
func (m Model) paintID(text string, it *board.Item, base lipgloss.Style) string {
	return m.styles.paintID(text, it, base)
}

// paintID is the same for the code that holds only brushes, so a list row and
// a line of work in the detail read alike. A bold base keeps the id bold, so
// the line the reader is on stays bold from the dot to the agent.
func (s styles) paintID(text string, it *board.Item, base lipgloss.Style) string {
	if it == nil {
		return base.Render(text)
	}
	name := it.ShortID
	if name == "" {
		name = it.ID
	}
	i := strings.Index(text, name)
	if i < 0 {
		return base.Render(text)
	}
	brush := s.kind(it.Kind).Bold(base.GetBold())
	return base.Render(text[:i]) + brush.Render(name) + base.Render(text[i+len(name):])
}

// rowText gives the one line of a row: the short ID, or the file path when the
// file has no ID yet, with the title. A row whose work has begun ends with the
// ticked boxes and the agent, placed at the right end. No status word ever
// reaches a row, because the count and the agent already say what kind of
// work it is. The title gives way first, so the count and the agent always
// fit whole.
func (m Model) rowText(r row, it *board.Item, w int) string {
	if r.group {
		arrow := "▸"
		if m.groupOpen {
			arrow = "▾"
		}
		return truncate(fmt.Sprintf("%s untyped (%d)", arrow, len(m.board.Untyped(false))), w)
	}
	if it == nil {
		return truncate(r.id, w)
	}
	name := it.ShortID
	if name == "" {
		name = it.ID
	}
	lead := strings.Repeat("  ", r.depth)
	if r.tree {
		lead += m.treeMark(r, it) + " "
	}
	head := lead + name + "  " + it.Title
	tally := progressText(it)
	if !inProgress(it) || tally == "" {
		return truncate(head, w)
	}
	if it.Agent != "" {
		tally += " · " + it.Agent
	}
	// The count sits one cell off the right wall, and a title too long for
	// what is left is cut with ….
	room := w - lipgloss.Width(tally) - 1
	if room < len("…") {
		return truncate(head, w)
	}
	return truncate(head, room) + strings.Repeat(" ", w-lipgloss.Width(truncate(head, room))-lipgloss.Width(tally)) + tally
}

// treeMark is what a tree row starts with: + on a shut plan, - on an open
// one, and the status dot on a task, so the reader sees what opens and what
// is done without reading the detail.
func (m Model) treeMark(r row, it *board.Item) string {
	if r.depth > 0 {
		return dotOf(it)
	}
	if m.isOpen(it.ID) {
		return "-"
	}
	return "+"
}
