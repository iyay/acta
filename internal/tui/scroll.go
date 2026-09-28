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

// scrollbar gives the column of cells down the right wall of a pane that
// says where the pane sits inside its content: one cell per inner row, a
// thumb where the window is and a track everywhere else. Nothing is drawn
// when the content fits, because then all of it is on screen already.
func scrollbar(total, visible, first, h int) []string {
	if total <= visible || h <= 0 {
		return nil
	}
	// The thumb keeps the size of the window, so a pane that shows a tenth
	// of its content gets a thumb of a tenth of the column.
	thumb := max(1, h*visible/total)
	pos := 0
	if last := total - visible; last > 0 {
		pos = clamp(first, 0, last) * (h - thumb) / last
	}
	bar := make([]string, h)
	for i := range bar {
		bar[i] = "░"
	}
	for i := pos; i < pos+thumb; i++ {
		bar[i] = "█"
	}
	return bar
}

// count is what a pane writes in its bottom border: the line on screen now
// out of all the lines the pane holds.
func count(first, total int) string {
	return fmt.Sprintf("%d/%d", first+1, total)
}

// boxOf measures one box the way geometry lays them out, also on a narrow
// terminal that shows only one of them: the model needs a height to clamp an
// offset with, whether or not the box is on screen.
func (m Model) boxOf(p pane) box {
	b := m.geometry().at(p)
	if b.h == 0 {
		return m.box(p, 0, 0, m.width, max(0, m.height-1))
	}
	return b
}

// fitOf is how many lines pane p has room for right now.
func (m Model) fitOf(p pane) int {
	if p == paneDetail {
		return m.boxOf(p).inner
	}
	return m.boxOf(p).rows
}

// linesOf is how many lines pane p holds in all, as wide as it is now.
func (m Model) linesOf(p pane) int {
	return m.linesAt(p, m.textOf(p, m.boxOf(p)))
}

// textOf is the width a pane draws its words at, and the width its content is
// measured at. A pane with more content than lines gives one cell to the
// scrollbar, so a body that wraps has fewer cells and more lines. The model
// and the view both measure here, or the end of a body never lands on screen.
func (m Model) textOf(p pane, b box) int {
	inner := b.textW()
	if inner > 1 && m.linesAt(p, inner-1) > b.inner {
		return inner - 1
	}
	return inner
}

// linesAt is how many lines pane p holds when its text is w cells wide: the
// rows of a list, or the body of the detail.
func (m Model) linesAt(p pane, w int) int {
	if p == paneDetail {
		return len(m.detailLines(w))
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
		return []string{faint.Render("nothing here")}
	}
	cur := cursorOf(rows, *sel, *idx)
	out := make([]string, 0, b.rows*rowLines)
	for i := range b.rows {
		n := b.first + i
		if n >= len(rows) {
			break
		}
		it := m.board.Get(rows[n].id)
		brush := faint
		switch {
		case n == cur:
			brush = selected
		case inProgress(it):
			// Work in progress wears the accent, dimmed; the rest stay plain.
			brush = work
		}
		out = append(out, brush.Render(pad(m.rowText(rows[n], it, w), w)))
	}
	return out
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
	head := strings.Repeat("  ", r.depth) + name + "  " + it.Title
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
