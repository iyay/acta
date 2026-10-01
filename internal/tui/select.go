package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

// drag is text picked with the mouse. It lives in screen cells, so it only
// means something until the screen moves.
type drag struct {
	on     bool // from a press on words until the highlight clears
	held   bool // the button is still down, so the end follows the pointer
	area   rect // the words of the pane the press landed in
	ax, ay int  // the cell the press landed on
	ex, ey int  // the cell the pointer is on now
}

// rect is a block of screen cells, both ends included.
type rect struct{ x0, y0, x1, y1 int }

func (r rect) has(x, y int) bool { return x >= r.x0 && x <= r.x1 && y >= r.y0 && y <= r.y1 }

// span is the cells of one screen row a drag covers, both ends included.
type span struct{ y, x0, x1 int }

// textArea is where a box keeps its words: inside the walls, below the title
// line and above the bottom line. The scrollbar is the right wall, so it is
// left out too. A box with no room for words gives false.
func (b box) textArea() (rect, bool) {
	r := rect{b.x + 1, b.y + 1, b.x + b.w - 2, b.y + b.h - 2}
	return r, r.x1 >= r.x0 && r.y1 >= r.y0
}

// shown says there is something to paint or copy: the end left the anchor.
func (d drag) shown() bool { return d.on && (d.ax != d.ex || d.ay != d.ey) }

// to moves the end to the pointer. It stays inside the pane the press was
// in, so a drag never picks up another pane or a wall.
func (d drag) to(x, y int) drag {
	d.ex = clamp(x, d.area.x0, d.area.x1)
	d.ey = clamp(y, d.area.y0, d.area.y1)
	return d
}

// spans gives the rows a drag covers, top to bottom, the way a terminal
// picks text: the first row from the start on, the middle rows whole, and
// the last row up to the end. A drag up or left works the same as down.
func (d drag) spans() []span {
	if !d.shown() {
		return nil
	}
	sx, sy, ex, ey := d.ax, d.ay, d.ex, d.ey
	if ey < sy || (ey == sy && ex < sx) {
		sx, sy, ex, ey = ex, ey, sx, sy
	}
	out := make([]span, 0, ey-sy+1)
	for y := sy; y <= ey; y++ {
		x0, x1 := d.area.x0, d.area.x1
		if y == sy {
			x0 = sx
		}
		if y == ey {
			x1 = ex
		}
		out = append(out, span{y, x0, x1})
	}
	return out
}

// dragText is the picked text the way the screen shows it: no colors, and
// no spaces at the end of a row.
func dragText(frame string, d drag) string {
	lines := strings.Split(frame, "\n")
	var rows []string
	for _, s := range d.spans() {
		if s.y >= len(lines) {
			break
		}
		cut := xansi.Strip(xansi.Cut(lines[s.y], s.x0, s.x1+1))
		rows = append(rows, strings.TrimRight(cut, " "))
	}
	return strings.Join(rows, "\n")
}

// paintDrag lays the band over the picked cells. The words under it drop
// their own colors, so the band reads the same wherever it falls.
func paintDrag(frame string, d drag, brush lipgloss.Style) string {
	sp := d.spans()
	if len(sp) == 0 {
		return frame
	}
	lines := strings.Split(frame, "\n")
	for _, s := range sp {
		if s.y >= len(lines) {
			break
		}
		ln := lines[s.y]
		mid := xansi.Strip(xansi.Cut(ln, s.x0, s.x1+1))
		lines[s.y] = xansi.Cut(ln, 0, s.x0) + brush.Render(mid) + xansi.Cut(ln, s.x1+1, xansi.StringWidth(ln))
	}
	return strings.Join(lines, "\n")
}

// anchorAt starts a drag on the cell of a press. A press on a wall, a title
// line, a bottom line or the status line holds no words, so it starts none.
func (m Model) anchorAt(x, y int) drag {
	g := m.geometry()
	p, _, _ := m.hit(x, y)
	b := g.full
	if g.wide {
		b = g.at(p)
	}
	a, ok := b.textArea()
	if !ok || !a.has(x, y) {
		return drag{}
	}
	return drag{on: true, held: true, area: a, ax: x, ay: y, ex: x, ey: y}
}

// copyDrag copies what the drag picked. The text comes from a screen drawn
// with no band, so the band never ends up in it.
func (m *Model) copyDrag() tea.Cmd {
	bare := *m
	bare.drag = drag{}
	return m.copyPicked(dragText(bare.draw(), m.drag))
}

// copyPicked puts text on the clipboard and says so the way y does: the same
// short toast, so the picked text stays off the status line. Text that is
// only spaces copies nothing and says nothing.
func (m *Model) copyPicked(text string) tea.Cmd {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if err := m.clip(text); err != nil {
		m.status = "copy failed: " + err.Error()
		return nil
	}
	m.status = "copied to clipboard"
	return nil
}
