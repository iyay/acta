package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
)

// tabBox is where one tab name sits in the top border of a pane, in screen
// cells, the way the mouse counts them.
type tabBox struct{ x, w int }

// box is where one pane sits on the screen and what it holds. first and rows
// belong to the list panes; the detail box scrolls by lines and counts its own.
type box struct {
	x, y, w, h  int // the border rectangle
	inner       int
	first, rows int      // first row on screen, and how many rows fit inside
	tabs        []tabBox // where each tab name sits in the top border
}

// textW is how many cells a pane has for its words: the width without the two
// walls. A pane that scrolls gives one of those cells to the scrollbar.
func (b box) textW() int { return max(0, b.w-2) }

// geom is the whole screen measured in cells from its top left corner. The
// view and the mouse both read it, so a click lands where it looks.
type geom struct {
	wide   bool // false below 60 columns, where only the focused box shows
	leftW  int  // width of the left column
	side   []box
	detail box
	full   box // the focused box stretched over the whole screen when not wide
}

// at gives the box a pane is drawn in. Below 60 columns only the focused pane
// reaches the screen, so the others answer an empty box.
func (g geom) at(p pane) box {
	if p == paneDetail {
		return g.detail
	}
	if int(p) < len(g.side) {
		return g.side[p]
	}
	return box{}
}

// One small palette, readable on a dark and on a light terminal. The accent
// marks the pane with the focus and the tab that is on; the dim brush paints
// what is secondary. The plain text color is left to the terminal.
var (
	accentColor = lipgloss.AdaptiveColor{Light: "25", Dark: "39"}
	accent      = lipgloss.NewStyle().Foreground(accentColor)
	// work marks a row whose work has begun. Dimmed, because the selected
	// row is the one that should catch the eye.
	work = lipgloss.NewStyle().Foreground(accentColor).Faint(true)
	// faint paints every row the cursor is not on.
	faint = lipgloss.NewStyle().Faint(true)
	// selected paints the row the cursor is on: a dark band across the whole
	// row with bright text on it, never reversed video, so the words stay
	// readable wherever the band falls.
	selected = lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "255"}).
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"})
)

// hints is what the left of the status line says when nothing else is going on.
const hints = "? help"

// helpLines is the key map the ? popup shows, grouped by box.
const helpLines = `0 1 2 3 4 5 tab   move between the boxes
] [              next / previous tab
z                expand the focused pane
j k g G          move a list, scroll the detail
ctrl+d ctrl+u    page down and up
enter            focus the detail on the row
e                open the row in the editor
t s n            set a value, new bug
esc              back to the list, close this help
/ r q            search, reload, quit
?                close this help`

// View draws the whole screen: the boxes, the status line, and any popup on
// top of them.
func (m Model) View() string {
	g := m.geometry()
	h := max(3, m.height-1)
	var body string
	if g.wide {
		// A box with no room draws nothing, so it takes no line either and
		// the boxes below it keep the place the geometry gave them.
		var column []string
		for p := range g.side {
			if drawn := m.paneView(pane(p), g.side[p]); drawn != "" {
				column = append(column, drawn)
			}
		}
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.JoinVertical(lipgloss.Left, column...), m.paneView(paneDetail, g.detail))
	} else {
		body = m.paneView(m.focus, g.full)
	}
	lines := strings.Split(body, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for i, ln := range lines {
		lines[i] = fit(ln, m.width)
	}
	body = strings.Join(lines, "\n")
	body = m.cover(body)
	return body + "\n" + fit(m.statusLine(), m.width)
}

// box measures one pane at the given rectangle and works out which of its
// rows the screen shows, so the view and the mouse count the same ones.
func (m Model) box(p pane, x, y, w, h int) box {
	b := box{x: x, y: y, w: w, h: h, inner: max(0, h-2), tabs: tabX(p, x, w, m.tabsOf(p), m.onTab(p))}
	if p == paneDetail {
		return b
	}
	b.rows = b.inner / rowLines
	rows, _, _ := m.slotOf(p)
	b.first = firstOf(m.off[p], len(rows), b.rows)
	return b
}

// hit says what sits under a mouse cell: the pane, the row of it and the tab
// in its title. A cell on none of them answers the focused pane with no row
// and no tab, which leaves the screen where it was.
func (m Model) hit(x, y int) (pane, int, int) {
	g := m.geometry()
	type target struct {
		p pane
		b box
	}
	var boxes []target
	if g.wide {
		for p := range g.side {
			boxes = append(boxes, target{pane(p), g.side[p]})
		}
		boxes = append(boxes, target{paneDetail, g.detail})
	} else {
		boxes = []target{{m.focus, g.full}}
	}
	for _, e := range boxes {
		if x < e.b.x || x >= e.b.x+e.b.w || y < e.b.y || y >= e.b.y+e.b.h {
			continue
		}
		if y == e.b.y {
			for i, t := range e.b.tabs {
				if x >= t.x && x < t.x+t.w {
					return e.p, -1, i
				}
			}
		}
		return e.p, e.b.rowAt(y), -1
	}
	return m.focus, -1, -1
}

// rowAt gives the row under a screen line, or -1 for the border, the title and
// the space below the last row.
func (b box) rowAt(y int) int {
	off := y - b.y - 1
	if off < 0 || off >= b.rows {
		return -1
	}
	return b.first + off
}

// paneView draws one pane: its border with the title inside the top line, the
// rows it holds, the scrollbar down the right wall when there is more content
// than rows, and the bottom line, which a sidebar pane closes with the item
// count of what it holds.
func (m Model) paneView(p pane, b box) string {
	if b.w < 2 || b.h < 1 {
		return ""
	}
	edge := m.edge(p)
	// A box with a single line has room for its title bar and nothing else,
	// so a short screen still says which box is which. A box with no line at
	// all draws nothing, so it takes no line either.
	if b.h == 1 {
		return m.paneTop(p, b, edge)
	}
	// The words of a pane that overflows get one cell less, because the
	// scrollbar takes that cell. The model measures the content at this same
	// width, so the line count on screen is the line count the offset counts.
	inner := m.textOf(p, b)
	total := m.linesAt(p, inner)
	first := b.first
	if p == paneDetail {
		first = firstOf(m.off[p], total, b.inner)
	}
	bar := scrollbar(total, b.inner, first, b.inner)
	// A pane too narrow to spare a cell has nowhere to put the scrollbar, so
	// it keeps that cell.
	if len(bar) == 0 || inner >= b.textW() {
		bar = nil
	}
	// Every sidebar pane counts the items it holds and the one under the
	// cursor. The detail box counts lines, not items, so it writes nothing
	// and keeps its scrollbar.
	foot := ""
	if p != paneDetail {
		rows, sel, idx := m.slotOf(p)
		foot = itemCount(cursorOf(rows, *sel, *idx)+1, len(rows))
	}
	var content []string
	if p == paneDetail {
		content = m.detailView(inner, first, b.inner)
	} else {
		content = m.listView(p, inner, b)
	}
	rows := make([]string, 0, b.h)
	rows = append(rows, m.paneTop(p, b, edge))
	for i := range b.inner {
		line := ""
		if i < len(content) {
			line = content[i]
		}
		cell := ""
		if i < len(bar) {
			cell = faint.Render(bar[i])
		}
		rows = append(rows, edge.Render("│")+pad(line, inner)+cell+edge.Render("│"))
	}
	rows = append(rows, paneBottom(b, edge, foot))
	return strings.Join(rows, "\n")
}

// segment is one piece of a border title and the brush it is painted with.
// The dashes between two names are a segment of their own, marked as sep, so
// they can be dropped together with the name they follow.
type segment struct {
	text  string
	style lipgloss.Style
	sep   bool
}

func segWidth(segs []segment) int {
	w := 0
	for _, s := range segs {
		w += lipgloss.Width(s.text)
	}
	return w
}

func plainSegs(segs []segment) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.text)
	}
	return b.String()
}

// detailView gives the lines of the detail box from the one its offset points at,
// so the body scrolls and the model never needs to know how long it is.
func (m Model) detailView(w, first, h int) []string {
	lines := m.detailLines(w)
	if first >= len(lines) {
		return nil
	}
	lines = lines[first:]
	if len(lines) > h {
		lines = lines[:h]
	}
	return lines
}

// statusPiece is one word group of the right side of the bottom line: the
// words on screen and the link they open, empty when the words open nothing.
type statusPiece struct {
	text string
	url  string
	x    int
	w    int
}

// statusPieces lays out the right side of the bottom line: the project, the
// watch mode and the date with time, then a divider, then the links and the
// version. A narrow line drops the project first, then the mode, but the date
// and time stay always, so the clock never leaves the screen. The view draws
// these pieces and the mouse reads its boxes off the drawn line, so a click
// and the drawn words cannot drift apart.
func (m Model) statusPieces() []statusPiece {
	info := []string{filepath.Base(m.cfg.RepoRoot), m.mode(), m.now.Format("2006-01-02 15:04")}
	var links []statusPiece
	if m.cfg.Links.Donate != "" {
		links = append(links, statusPiece{text: "Donate", url: m.cfg.Links.Donate})
	}
	if m.cfg.Links.Feedback != "" {
		links = append(links, statusPiece{text: "Feedback", url: m.cfg.Links.Feedback})
	}
	tail := append(links, statusPiece{text: m.version})
	return joinStatus(info, tail, m.width)
}

// joinStatus picks how many info words fit: it drops the project first, then
// the mode, and keeps the date, the links and the version always. The left
// side takes its cells first, so the drop matches the drawn line.
func joinStatus(info []string, tail []statusPiece, w int) []statusPiece {
	for len(info) > 1 {
		out := append(infoPieces(info), tail...)
		if lipgloss.Width(statusText(out))+lipgloss.Width(hints) <= w {
			return out
		}
		info = info[1:]
	}
	return append(infoPieces(info), tail...)
}

func infoPieces(info []string) []statusPiece {
	out := make([]statusPiece, 0, len(info))
	for _, t := range info {
		out = append(out, statusPiece{text: t})
	}
	return out
}

// statusText draws the pieces the way the bottom line shows them: the info
// words joined by " · ", then " | ", then the links and the version joined by
// two spaces. Links carry an OSC 8 wrapper so cmd+click opens them.
func statusText(pieces []statusPiece) string {
	nInfo := 0
	for _, p := range pieces {
		if p.url != "" {
			break
		}
		nInfo++
	}
	words := make([]string, 0, len(pieces))
	for _, p := range pieces {
		word := p.text
		if p.url != "" {
			word = osc8(p.url, p.text)
		}
		words = append(words, word)
	}
	right := strings.Join(words[:nInfo], " · ")
	if len(words) > nInfo {
		if right != "" {
			right += " | "
		}
		right += strings.Join(words[nInfo:], "  ")
	}
	return right
}

// osc8 wraps text in a terminal hyperlink, so cmd+click opens the url in
// terminals that read them. Plain text stays readable where links do not work.
func osc8(url, text string) string {
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}

// statusLine draws the bottom line: the help hint, the search box or the last
// message on the left, and the project, the watch mode, the date and time,
// the links and the version on the right.
func (m Model) statusLine() string {
	left, dim := hints, true
	switch {
	case m.searching || m.query != "":
		left, dim = "/"+m.query, false
	case m.status != "":
		left, dim = m.status, false
	}
	pieces := m.statusPieces()
	right := statusText(pieces)
	// The date and the links are what the line always tells, so the left
	// gives way first when the window is narrow.
	room := m.width - lipgloss.Width(right) - lipgloss.Width(left)
	if room < 0 {
		if m.width-lipgloss.Width(right) <= 0 {
			return fit(right, m.width)
		}
		left = fit(left, m.width-lipgloss.Width(right))
		room = m.width - lipgloss.Width(right) - lipgloss.Width(left)
		if room < 0 {
			room = 0
		}
	}
	if dim {
		left = faint.Render(left)
	}
	gap := ""
	if room > 0 {
		gap = strings.Repeat(" ", room)
	}
	return left + gap + faint.Render(right)
}

// statusLineBoxes gives the click boxes of the bottom line by reading the
// line the view draws: each OSC 8 wrapper carries its url next to its words,
// so a click lands on the drawn words even when the right side is cut to
// fit a narrow window.
func (m Model) statusLineBoxes() []statusPiece {
	raw := m.statusLine()
	flat := stripCodes(raw)
	var out []statusPiece
	base := 0
	for _, seg := range linkSegments(raw) {
		i := strings.Index(flat[base:], seg.text)
		if i < 0 {
			continue
		}
		x := lipgloss.Width(flat[:base+i])
		out = append(out, statusPiece{text: seg.text, url: seg.url, x: x, w: lipgloss.Width(seg.text)})
		base += i + len(seg.text)
	}
	return out
}

// linkSegment is one hyperlink the bottom line draws: the words on screen
// next to the url they open.
type linkSegment struct {
	text string
	url  string
}

// linkSegments reads the hyperlinks out of a drawn line. Each opener holds
// the url, the words run until the closer, so even a cut word keeps the
// cells still on screen.
func linkSegments(raw string) []linkSegment {
	var out []linkSegment
	for {
		open := strings.Index(raw, "\x1b]8;;")
		if open < 0 {
			return out
		}
		raw = raw[open+len("\x1b]8;;"):]
		end := strings.Index(raw, "\x1b\\")
		if end < 0 {
			return out
		}
		url := raw[:end]
		raw = raw[end+len("\x1b\\"):]
		end = strings.Index(raw, "\x1b]8;;\x1b\\")
		if end < 0 {
			return out
		}
		if url != "" {
			out = append(out, linkSegment{text: raw[:end], url: url})
		}
		raw = raw[end+len("\x1b]8;;\x1b\\"):]
	}
}

// stripCodes drops the color and hyperlink codes from a drawn line, leaving
// the words a person sees. The mouse counts cells on what is left.
func stripCodes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], "\x1b]8;;") {
			end := strings.Index(s[i:], "\x1b\\")
			if end < 0 {
				break
			}
			i += end + len("\x1b\\")
			continue
		}
		if strings.HasPrefix(s[i:], "\x1b[") {
			end := strings.IndexAny(s[i:], "ABCDEFGHJKSTfmnsu")
			if end < 0 {
				break
			}
			i += end + 1
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// mode says whether the board follows the files as they change.
func (m Model) mode() string {
	if m.manual {
		return "paused"
	}
	return "live"
}

// edge is the brush of a pane border: the accent on the pane with the focus,
// dim on the others, so the focus is plain to see.
func (m Model) edge(p pane) lipgloss.Style {
	if m.focus == p {
		return accent
	}
	return faint
}

// popupBox draws what sits on top of the panes: the help screen, the picker
// for a value, or the slug prompt of a new bug. It is empty when nothing is
// open.
func (m Model) popupBox() string {
	switch {
	case m.help:
		return m.boxView("Keys", helpLines)
	case m.popup != nil:
		var b strings.Builder
		for i, o := range m.popup.options {
			if i == m.popup.idx {
				b.WriteString(selected.Render("> "+o) + "\n")
			} else {
				b.WriteString("  " + o + "\n")
			}
		}
		b.WriteString("\n" + faint.Render("j k to move, enter to save, esc to cancel"))
		return m.boxView("Set "+m.popup.field, b.String())
	case m.slug != nil:
		return m.boxView("New bug", "new bug slug: "+*m.slug+"█\n\n"+
			faint.Render("lower case words joined by -, enter to open the editor, esc to cancel"))
	}
	return ""
}

// boxView draws a centered box with a title inside its top border.
func (m Model) boxView(title, content string) string {
	w := clamp(m.width*2/3, 24, 60)
	if w > m.width {
		w = m.width
	}
	if w < 6 {
		return ""
	}
	inner := w - 2
	rows := []string{topLine(w, accent, []segment{{text: "─", style: accent}, {text: title, style: accent.Bold(true)}})}
	for _, ln := range strings.Split(content, "\n") {
		rows = append(rows, accent.Render("│")+pad(fit(ln, inner), inner)+accent.Render("│"))
	}
	rows = append(rows, accent.Render("└"+strings.Repeat("─", inner)+"┘"))
	return strings.Join(rows, "\n")
}

// cover puts the popup over the middle of the body and keeps the panes it
// hides on either side of the box, so only the box itself changes on screen.
func (m Model) cover(body string) string {
	rows := strings.Split(m.popupBox(), "\n")
	x0, y0, w, _ := popupRect(rows, m.width, m.height)
	if w == 0 {
		return body
	}
	lines := strings.Split(body, "\n")
	for i, r := range rows {
		if y0+i < len(lines) {
			lines[y0+i] = splice(lines[y0+i], r, x0, w)
		}
	}
	return strings.Join(lines, "\n")
}

// idText gives the ID line of the detail: the number ID, the permanent hash
// and the path, with the empty pieces left out.
func idText(it *board.Item) string {
	var parts []string
	for _, s := range []string{it.ShortID, it.Hash, it.ID} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " · ")
}

// shortRef names an item the short way: its number ID, else its hash, else the
// path ID.
func shortRef(it *board.Item) string {
	switch {
	case it.ShortID != "":
		return it.ShortID
	case it.Hash != "":
		return it.Hash
	}
	return it.ID
}

// specText names the spec or bug a plan carries, and is empty for the items
// that link none.
func (m Model) specText(it *board.Item) string {
	if it.SpecID == "" {
		return ""
	}
	if s := m.board.Get(it.SpecID); s != nil {
		return shortRef(s) + " · " + s.Title
	}
	return it.SpecID
}

// worktreeText names the branch an item came from, and says so when that
// worktree is not checked out on disk.
func worktreeText(it *board.Item) string {
	switch {
	case it.Worktree == "":
		return ""
	case !it.OnDisk:
		return it.Worktree + " (not checked out)"
	}
	return it.Worktree
}

// fileText gives the path of the file, seen from the repo root.
func (m Model) fileText(it *board.Item) string {
	if !it.OnDisk {
		return ""
	}
	if rel, err := filepath.Rel(m.cfg.RepoRoot, it.Path); err == nil {
		return rel
	}
	return it.Path
}

// progressText gives the ticked boxes of an item, or nothing when it has none.
func progressText(it *board.Item) string {
	if it.Total == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", it.Done, it.Total)
}

// cut splits text into lines and clamps each one to w cells.
func cut(text string, w int) []string {
	lines := strings.Split(text, "\n")
	for i, ln := range lines {
		lines[i] = truncate(ln, w)
	}
	return lines
}

// fit cuts a line to w cells, and counts the cells the way the terminal does,
// so a line that carries colors is never cut in the middle of a code.
func fit(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	return lipgloss.NewStyle().MaxWidth(w).Render(s)
}

// pad fills a line out to w cells, so the pane walls stay in one column.
func pad(s string, w int) string {
	if gap := w - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return fit(s, w)
}

// truncate cuts a line to w cells so panes never wrap and break the layout.
func truncate(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > w-1 {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// newRenderer renders markdown with glamour and keeps the result, because
// drawing a long spec on every key press is slow.
func newRenderer(dark bool) func(string, int) string {
	style := "light"
	if dark {
		style = "dark"
	}
	cache := map[string]string{}
	return func(md string, width int) string {
		key := fmt.Sprintf("%d\x00%s", width, md)
		if out, ok := cache[key]; ok {
			return out
		}
		out := md
		if r, err := glamour.NewTermRenderer(glamour.WithStandardStyle(style), glamour.WithWordWrap(width)); err == nil {
			if s, err := r.Render(md); err == nil {
				out = strings.TrimRight(s, "\n")
			}
		}
		cache[key] = out
		return out
	}
}
