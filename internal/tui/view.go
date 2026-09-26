package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"pm-board/internal/board"
)

// One small palette, readable on a dark and on a light terminal. The accent
// marks the pane with the focus and the tab that is on; the dim brush paints
// what is secondary. The plain text color is left to the terminal.
var (
	accentColor = lipgloss.AdaptiveColor{Light: "25", Dark: "39"}
	accent      = lipgloss.NewStyle().Foreground(accentColor)
	faint       = lipgloss.NewStyle().Faint(true)
	selected    = lipgloss.NewStyle().Reverse(true)
)

// hints is what the left of the status line says when nothing else is going on.
const hints = "? help"

// helpLines is the key map the ? popup shows, grouped by pane.
const helpLines = `1 2 3 tab       move between the panes
] [             next / previous tab
j k g G         move a list, scroll the detail
ctrl+d ctrl+u   page down and up
enter t s n     edit, set, new bug
/ r q           search, reload, quit
? esc           close this help`

// View draws the whole screen: the panes, the status line, and any popup on
// top of them.
func (m Model) View() string {
	g := m.geometry()
	h := max(3, m.height-1)
	var body string
	if g.wide {
		left := lipgloss.JoinVertical(lipgloss.Left, m.paneView(paneOpen, g.open), m.paneView(paneDone, g.done))
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, m.paneView(paneDetail, g.detail))
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
	if pop := m.popupBox(); pop != "" {
		body = m.cover(body, pop, h)
	}
	return body + "\n" + fit(m.statusLine(), m.width)
}

// paneView draws one pane: its border with the title inside the top line, the
// rows it holds, and the bottom line, which carries the line count of the
// detail when that pane has the focus.
func (m Model) paneView(p pane, b box) string {
	if b.w < 2 || b.h < 2 {
		return ""
	}
	edge := m.edge(p)
	inner := b.w - 2
	var content []string
	foot := ""
	if p == paneDetail {
		content, foot = m.detailView(inner, b.inner)
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
		rows = append(rows, edge.Render("│")+pad(line, inner)+edge.Render("│"))
	}
	rows = append(rows, paneBottom(b, edge, foot))
	return strings.Join(rows, "\n")
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
	on := m.tab
	if p == paneDone {
		on = m.doneTab
	}
	pieces := titlePieces(p, names)
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

// segment is one piece of a border title and the brush it is painted with.
// The dashes between two names are a segment of their own, marked as sep, so
// they can be dropped together with the name they follow.
type segment struct {
	text  string
	style lipgloss.Style
	sep   bool
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
		return edge.Render("╭") + fit(edge.Render(plainSegs(segs)), inner) + edge.Render("╮")
	}
	var b strings.Builder
	b.WriteString(edge.Render("╭"))
	for _, s := range segs {
		b.WriteString(s.style.Render(s.text))
	}
	if fill := inner - segWidth(segs); fill > 0 {
		b.WriteString(edge.Render(strings.Repeat("─", fill)))
	}
	b.WriteString(edge.Render("╮"))
	return b.String()
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

// paneBottom draws the bottom line of a pane, with the line count of the
// detail when there is one.
func paneBottom(b box, edge lipgloss.Style, foot string) string {
	inner := b.w - 2
	text := ""
	if foot != "" {
		text = fit(" "+foot+" ", inner)
	}
	fill := strings.Repeat("─", max(0, inner-lipgloss.Width(text)))
	return edge.Render("╰") + edge.Render(text) + edge.Render(fill) + edge.Render("╯")
}

// listView draws the rows of a list pane. Every row takes two lines, the title
// and the dim meta line, plus a blank line under it, so a click on either line
// lands on the same row.
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
		title, meta := m.rowText(rows[n])
		title, meta = truncate(title, w), truncate(meta, w)
		switch {
		case n == cur:
			title, meta = selected.Render(title), selected.Render(meta)
		case p == paneOpen:
			title, meta = accent.Render(title), faint.Render(meta)
		default:
			meta = faint.Render(meta)
		}
		out = append(out, title, meta, "")
	}
	return out
}

// rowText gives the two lines of one row: the short ID, or the file path when
// the file has no ID yet, with the title; and a dim line with the status, the
// progress and the agent when one is working on it.
func (m Model) rowText(r row) (string, string) {
	if r.group {
		arrow := "▸"
		if m.groupOpen {
			arrow = "▾"
		}
		return fmt.Sprintf("%s untyped (%d)", arrow, len(m.board.Untyped(false))), ""
	}
	it := m.board.Get(r.id)
	if it == nil {
		return r.id, ""
	}
	name := it.ShortID
	if name == "" {
		name = it.ID
	}
	var parts []string
	if it.Status != "" {
		parts = append(parts, it.Status)
	}
	if it.Total > 0 {
		parts = append(parts, fmt.Sprintf("%d/%d", it.Done, it.Total))
	}
	if it.Agent != "" {
		parts = append(parts, it.Agent)
	}
	return strings.Repeat("  ", r.depth) + name + "  " + it.Title, "  " + strings.Join(parts, " · ")
}

// detailView gives the lines of pane [3] and the line count for its bottom
// border. The body scrolls, and the cut of that scroll happens here, so the
// model never needs to know how long the body is.
func (m Model) detailView(w, h int) ([]string, string) {
	lines := m.detailLines(w)
	foot := ""
	if m.focus == paneDetail && len(lines) > 0 {
		foot = fmt.Sprintf("%d/%d", min(m.scroll+1, len(lines)), len(lines))
	}
	lines = lines[min(m.scroll, max(0, len(lines)-1)):]
	if len(lines) > h {
		lines = lines[:h]
	}
	return lines, foot
}

// detailLines draws the header and the body of the item on show. Labels are
// upper case, padded to one width, with the colons in one column, and a line
// with no value is left out.
func (m Model) detailLines(w int) []string {
	it := m.Selected()
	if it == nil {
		if len(m.board.Items) == 0 {
			return cut("this repo has no .pm/ yet.\n\npress n to write the first bug, or let the agent plugin create specs and plans.", w)
		}
		return []string{faint.Render("enter opens the group")}
	}
	fields := []struct{ label, value string }{
		{"ID", idText(it)},
		{kindLabel(it.Kind), it.Title},
		{"STATUS", it.Status},
		{"REF", it.Ref},
		{"SPEC", m.specText(it)},
		{"WORKTREE", worktreeText(it)},
		{"AGENT", it.Agent},
		{"FILE", m.fileText(it)},
		{"TASKS", progressText(it)},
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
		lines = append(lines, truncate(fmt.Sprintf("%-*s: %s", width, f.label, f.value), w))
	}
	for _, p := range it.Problems {
		lines = append(lines, truncate("! "+p, w))
	}
	lines = append(lines, faint.Render(strings.Repeat("─", max(1, w))))
	for _, ln := range strings.Split(m.render(it.Body, w), "\n") {
		lines = append(lines, fit(ln, w))
	}
	return lines
}

// statusPiece is one word group of the right side of the bottom line: the
// words on screen and the link they open, empty when the words open nothing.
// nInfo says how many leading pieces are info words (project, mode, date),
// so the view and the mouse split info from links at the same place.
type statusPiece struct {
	text  string
	url   string
	nInfo int
	x     int
	w     int
}

// statusPieces lays out the right side of the bottom line: the project, the
// watch mode and the date with time, then a divider, then the links and the
// version. A narrow line drops the project first, then the mode, but the date
// and time stay always, so the clock never leaves the screen. The view draws
// these pieces and the mouse reads their boxes, so a click and the drawn
// words cannot drift apart.
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
	out := joinStatus(info, tail, m.width)
	// Walk the same gaps the text draws, so each click box starts where its
	// words start on screen.
	right := statusText(out)
	x := m.width - lipgloss.Width(right)
	nInfo := 0
	for _, p := range out {
		if p.url != "" {
			break
		}
		nInfo++
	}
	for i := range out {
		if i > 0 {
			if i == nInfo {
				x += lipgloss.Width(" | ")
			} else if i < nInfo {
				x += lipgloss.Width(" · ")
			} else {
				x += 2
			}
		}
		out[i].x, out[i].w, out[i].nInfo = x, lipgloss.Width(out[i].text), nInfo
		x += out[i].w
	}
	return out
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
	m.statusBoxes = boxesOf(pieces)
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

// boxesOf keeps the link boxes of the drawn pieces, so a click and the drawn
// words come from the same numbers.
func boxesOf(pieces []statusPiece) []statusPiece {
	out := make([]statusPiece, 0, len(pieces))
	for _, p := range pieces {
		if p.url != "" {
			out = append(out, p)
		}
	}
	return out
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
	rows = append(rows, accent.Render("╰"+strings.Repeat("─", inner)+"╯"))
	return strings.Join(rows, "\n")
}

// cover puts a box over the middle of the body, so the panes stay on screen
// above and below it.
func (m Model) cover(body, box string, h int) string {
	lines := strings.Split(body, "\n")
	rows := strings.Split(box, "\n")
	top := max(0, (h-len(rows))/2)
	for i, r := range rows {
		if top+i >= len(lines) {
			break
		}
		lines[top+i] = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, fit(r, m.width))
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

// kindLabel is the label of the line that names the kind and the title.
func kindLabel(k board.Kind) string {
	switch k {
	case board.KindPlan:
		return "PLAN"
	case board.KindTask:
		return "TASK"
	case board.KindBug:
		return "BUG"
	}
	return "SPEC"
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
