package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

// geometry measures the screen. The left column is 30% of the width, held
// between 28 and 48 columns, and below 60 columns only the focused box is on
// screen because two columns of a small terminal fit nothing.
func (m Model) geometry() geom {
	// The top line is the tab bar and the last line the status line, so the
	// boxes share the rest. A terminal with fewer rows gets the rows it has.
	bodyH := max(0, m.height-2)
	panes := m.panes()
	g := geom{wide: m.width >= 60, leftW: clamp(m.width*3/10, 28, 48), side: make([]box, len(panes))}
	y := 1
	for p, h := range m.leftHeights(bodyH) {
		g.side[p] = m.box(pane(p), 0, y, g.leftW, h)
		y += h
	}
	// A box of zero width is not on screen, which is how a narrow terminal
	// leaves all but the focused one out.
	g.detail = m.box(paneDetail, g.leftW, 1, max(0, m.width-g.leftW), bodyH)
	if g.wide {
		return g
	}
	g.full = m.box(m.focus, 0, 1, m.width, bodyH)
	// Only the focused box is on screen now, so the others are zero boxes and
	// neither the view nor the mouse finds them.
	for p := range g.side {
		g.side[p] = box{}
	}
	g.detail = box{}
	return g
}

// split gives n boxes a share of h lines: the same number each, and the last
// box whatever is left, so the column always ends exactly where the detail box
// does. A screen too short to give a box its two border lines still splits
// without losing a line.
func split(h, n int) []int {
	out := make([]int, n)
	if h < 0 || n < 1 {
		return out
	}
	each := h / n
	for p := range out[:n-1] {
		out[p] = each
	}
	out[n-1] = h - each*(n-1)
	return out
}

// leftHeights gives the boxes of the open tab h lines, one height each.
// Nothing expanded means an even share, the last box taking what is left
// over, so the column ends where the detail box ends. An expanded box takes
// what the others do not, and the others keep room lines each, or only their
// title bar when the screen is too short for room lines. A box can end up with
// no lines at all on a screen that has almost none, which is what an expanded
// box leaves the others with, and it then draws nothing.
func (m Model) leftHeights(h int) []int {
	n := len(m.panes())
	others := n - 1
	if m.expanded < 0 || m.expanded >= n {
		return split(h, n)
	}
	out := make([]int, n)
	each := expandedRoom
	if h < expandedRoom*others+expandedRoom {
		each = 1
	}
	for p := range out {
		if p != m.expanded {
			out[p] = each
		}
	}
	out[m.expanded] = h - each*others
	return out
}

// expandedRoom is how many lines a box keeps while another one has the room:
// its top border, one row and its bottom border.
const expandedRoom = 3

// paneTop draws the top line of a pane. The title sits inside the border,
// lazygit style: the pane number, then the tabs. The tab that is on is bold in
// the accent color and the others stay dim. The pieces come from titlePieces,
// the same helper the mouse counts its click boxes with.
func (m Model) paneTop(p pane, b box, edge lipgloss.Style) string {
	names := m.tabsOf(p)
	if p == paneDetail {
		names = []string{"Detail"}
	}
	on := m.onTab(p)
	pieces := titlePieces(p, names, on, max(0, b.w-2))
	segs := make([]segment, 0, len(pieces))
	for _, piece := range pieces {
		if piece.tab < 0 {
			segs = append(segs, segment{text: piece.text, style: edge, sep: piece.sep})
			continue
		}
		style := faint
		switch {
		case p == paneDetail:
			if m.focus == p {
				style = accent.Bold(true)
			}
		case piece.tab == on:
			style = accent.Bold(true)
		}
		segs = append(segs, segment{text: piece.text, style: style})
	}
	return topLine(b.w, edge, segs)
}

// topLine draws the top line of a box of width w: the left corner, the pieces,
// then dashes up to the right corner. Pieces that do not fit are dropped from
// the right, and a piece that alone is too wide is cut, so the line never
// grows past w.
func topLine(w int, edge lipgloss.Style, segs []segment) string {
	if w < 2 {
		return edge.Render(strings.Repeat("─", max(0, w)))
	}
	inner := w - 2
	for len(segs) > 1 && segWidth(segs) > inner {
		segs = segs[:len(segs)-1]
	}
	for len(segs) > 1 && segs[len(segs)-1].sep {
		segs = segs[:len(segs)-1]
	}
	if segWidth(segs) > inner {
		return edge.Render("┌") + fit(edge.Render(plainSegs(segs)), inner) + edge.Render("┐")
	}
	var b strings.Builder
	b.WriteString(edge.Render("┌"))
	for _, s := range segs {
		b.WriteString(s.style.Render(s.text))
	}
	if fill := inner - segWidth(segs); fill > 0 {
		b.WriteString(edge.Render(strings.Repeat("─", fill)))
	}
	b.WriteString(edge.Render("┐"))
	return b.String()
}

// paneBottom draws the bottom line of a pane, with whatever the pane writes
// there at the right end, next to the corner. A sidebar pane writes its
// counter there, lazygit style.
func paneBottom(b box, edge lipgloss.Style, foot string) string {
	inner := b.w - 2
	text := ""
	if foot != "" {
		text = fit(" "+foot+" ", inner)
	}
	fill := strings.Repeat("─", max(0, inner-lipgloss.Width(text)))
	return edge.Render("└") + edge.Render(fill) + edge.Render(text) + edge.Render("┘")
}

// popupRect says where a box of these rows is drawn on a screen of this size:
// the left column, the first row, the width and the height. The screen and
// the tests both read it, so what the box covers and what a test checks
// cannot drift apart.
func popupRect(rows []string, width, height int) (x, y, w, h int) {
	if len(rows) == 0 || rows[0] == "" {
		return 0, 0, 0, 0
	}
	w = lipgloss.Width(rows[0])
	body := max(3, height-1)
	return max(0, (width-w)/2), max(0, (body-len(rows))/2), w, len(rows)
}

// splice puts one row of a box over the cells from x0 to x0+w of a line and
// keeps the cells on either side. The cuts go by display width. A wide rune
// cannot be cut in half, so one on an edge of the box is dropped whole and
// the cell it leaves behind becomes a space.
func splice(line, row string, x0, w int) string {
	head := pad(xansi.Truncate(line, x0, ""), x0)
	tail := xansi.TruncateLeft(line, x0+w, "")
	// A wide rune on the edge stays whole, so the tail can be one cell too
	// long. One more cell of cut drops that rune, and the cell it left behind
	// becomes the space that goes in front of the tail.
	gap := max(0, x0+w+lipgloss.Width(tail)-lipgloss.Width(line))
	return head + pad(row, w) + strings.Repeat(" ", gap) + xansi.TruncateLeft(line, x0+w+gap, "")
}
