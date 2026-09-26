package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

// geometry measures the screen. The left column is 30% of the width, held
// between 28 and 48 columns, and below 60 columns only the focused pane is on
// screen because two columns of a small terminal fit nothing.
func (m Model) geometry() geom {
	// The last line belongs to the status line, so the panes share the rest.
	// A terminal with fewer rows than that gets the rows it really has.
	bodyH := max(0, m.height-1)
	top, bottom := split(bodyH)
	g := geom{wide: m.width >= 60, leftW: clamp(m.width*3/10, 28, 48)}
	rest := max(0, m.width-g.leftW)
	g.open = m.box(paneOpen, 0, 0, g.leftW, top)
	g.done = m.box(paneDone, 0, top, g.leftW, bottom)
	// A pane of zero width is not on screen, which is how a narrow terminal
	// leaves two of the three out.
	g.detail = m.box(paneDetail, g.leftW, 0, rest, bodyH)
	if g.wide {
		return g
	}
	g.full = m.box(m.focus, 0, 0, m.width, bodyH)
	// Only the focused pane is on screen now, so the other two are zero boxes
	// and neither the view nor the mouse finds them.
	g.open, g.done, g.detail = box{}, box{}, box{}
	return g
}

// split divides the height between the two left panes: about two thirds for
// the open one, and never fewer than one row each. Both rows always add up to
// h, so the left column ends exactly where the detail pane does.
func split(h int) (int, int) {
	if h < 2 {
		return h, 0
	}
	top := clamp(h*2/3, 1, h-1)
	return top, h - top
}

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

// paneBottom draws the bottom line of a pane, with the line count of the
// detail when there is one.
func paneBottom(b box, edge lipgloss.Style, foot string) string {
	inner := b.w - 2
	text := ""
	if foot != "" {
		text = fit(" "+foot+" ", inner)
	}
	fill := strings.Repeat("─", max(0, inner-lipgloss.Width(text)))
	return edge.Render("└") + edge.Render(text) + edge.Render(fill) + edge.Render("┘")
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
