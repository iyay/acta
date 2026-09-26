package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"pm-board/internal/board"
)

var (
	faint    = lipgloss.NewStyle().Faint(true)
	bold     = lipgloss.NewStyle().Bold(true)
	selected = lipgloss.NewStyle().Reverse(true)
	pane     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
)

const helpText = `pmb keys

  1 2 3, tab     switch tab
  j k, arrows    move
  g G            top, bottom
  ctrl-d ctrl-u  scroll the detail pane
  enter          open in $EDITOR (a task opens at its heading)
  /              search, esc clears
  a              active / all
  t s            set type / status
  n              new bug
  r              reload
  ?              close this help
  q              quit`

// View draws the whole screen.
func (m Model) View() string {
	if m.help {
		lines := strings.Split(helpText, "\n")
		for i, ln := range lines {
			lines[i] = truncate(ln, m.width)
		}
		return strings.Join(lines, "\n")
	}
	leftW := clamp(m.width*2/5, 12, max(12, m.width-8))
	rightW := max(8, m.width-leftW)
	bodyH := max(5, m.height-4)

	left := pane.Width(leftW - 2).Height(bodyH - 2).Render(m.listView(leftW-2, bodyH-2))
	right := pane.Width(rightW - 2).Height(bodyH - 2).Render(m.detailView(rightW-2, bodyH-2))
	return lipgloss.JoinVertical(lipgloss.Left,
		m.header(),
		m.tabsView(),
		lipgloss.JoinHorizontal(lipgloss.Top, left, right),
		m.footer(),
	)
}

func (m Model) header() string {
	filter, mode := "active", "live"
	if m.showAll {
		filter = "all"
	}
	if m.manual {
		mode = "manual"
	}
	return bold.Render(truncate(fmt.Sprintf("pmb · %s · %s · %s", filepath.Base(m.cfg.RepoRoot), filter, mode), m.width))
}

func (m Model) tabsView() string {
	var plain, parts []string
	for i, name := range tabNames {
		plain = append(plain, fmt.Sprintf("[%d] %s %d", i+1, name, len(m.board.List(tabKinds[i], m.showAll))))
		label := plain[i]
		if tab(i) == m.tab && m.query == "" {
			label = selected.Render(label)
		}
		parts = append(parts, label)
	}
	full := strings.Join(parts, "  ")
	if lipgloss.Width(full) <= m.width {
		return full
	}
	return truncate(strings.Join(plain, "  "), m.width)
}

func (m Model) listView(w, h int) string {
	rows := m.rows()
	if len(rows) == 0 {
		return faint.Render("nothing here")
	}
	cur := m.cursor()
	start := max(0, cur-h+1)
	var lines []string
	for i := start; i < len(rows) && len(lines) < h; i++ {
		line := truncate(m.rowLabel(rows[i]), w)
		if i == cur {
			line = selected.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m Model) rowLabel(r row) string {
	if r.group {
		arrow := "▸"
		if m.groupOpen {
			arrow = "▾"
		}
		return fmt.Sprintf("%s untyped (%d)", arrow, len(m.board.Untyped(m.showAll)))
	}
	it := m.board.Get(r.id)
	label := strings.Repeat("  ", r.depth) + icon(it) + " "
	if it.Ref != "" && it.Kind != board.KindTask {
		label += it.Ref + " "
	}
	label += it.Title
	if it.Kind == board.KindTask {
		if p := m.board.Get(it.Parent); p != nil {
			label += " · " + p.Title
		}
	}
	return label + "  " + faint.Render(short(it))
}

func (m Model) detailView(w, h int) string {
	if m.popup != nil {
		return m.popupView()
	}
	if m.slug != nil {
		return "new bug slug: " + *m.slug + "█\n\n" + faint.Render("lower case words joined by -, enter to open the editor, esc to cancel")
	}
	it := m.Selected()
	if it == nil {
		if len(m.board.Items) == 0 {
			return "this repo has no .pm/ yet.\n\npress n to write the first bug, or let the agent plugin create specs and plans."
		}
		return faint.Render("enter opens the group")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s · %s\n", strings.ToUpper(string(it.Kind)), bold.Render(it.Title))
	status := it.Status
	if it.StatusSource == "frontmatter" {
		status += " (set by hand)"
	}
	fmt.Fprintf(&b, "%s · %d/%d", status, it.Done, it.Total)
	if it.Kind != board.KindTask {
		b.WriteString(" tasks")
	}
	b.WriteString("\n")
	if it.Ref != "" {
		fmt.Fprintf(&b, "ref    %s\n", it.Ref)
	}
	if p := m.board.Get(it.Parent); p != nil {
		fmt.Fprintf(&b, "parent %s\n", p.Title)
	}
	fmt.Fprintf(&b, "id     %s\n", it.ID)
	if rel, err := filepath.Rel(m.cfg.RepoRoot, it.Path); err == nil {
		fmt.Fprintf(&b, "file   %s\n", rel)
	}
	if it.FixedIn != "" {
		fmt.Fprintf(&b, "fixed  %s\n", it.FixedIn)
	}
	for _, p := range it.Problems {
		fmt.Fprintf(&b, "! %s\n", p)
	}
	if len(it.Children) > 0 {
		b.WriteString("\nTasks\n")
		for _, id := range it.Children {
			c := m.board.Get(id)
			fmt.Fprintf(&b, " %s %s %s  %d/%d\n", icon(c), c.TaskNum, c.Title, c.Done, c.Total)
		}
	}
	b.WriteString(faint.Render(strings.Repeat("─", max(1, w))) + "\n")
	b.WriteString(m.render(it.Body, w))

	lines := strings.Split(b.String(), "\n")
	start := min(m.scroll, max(0, len(lines)-1))
	lines = lines[start:]
	if len(lines) > h {
		lines = lines[:h]
	}
	for i, ln := range lines {
		lines[i] = truncate(ln, w)
	}
	return strings.Join(lines, "\n")
}

func (m Model) popupView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "set %s\n\n", m.popup.field)
	for i, o := range m.popup.options {
		if i == m.popup.idx {
			b.WriteString(selected.Render("> "+o) + "\n")
		} else {
			b.WriteString("  " + o + "\n")
		}
	}
	b.WriteString("\n" + faint.Render("j k to move, enter to save, esc to cancel"))
	return b.String()
}

func (m Model) footer() string {
	if m.searching || m.query != "" {
		return truncate("/"+m.query, m.width)
	}
	if m.status != "" {
		return truncate(m.status, m.width)
	}
	return faint.Render(truncate("1 2 3 tab · j k move · enter edit · / search · a all · t type · s status · n bug · ? help · q quit", m.width))
}

func icon(it *board.Item) string {
	if len(it.Problems) > 0 {
		return "!"
	}
	switch it.Status {
	case "done", "fixed", "dropped", "wontfix":
		return "●"
	case "draft", "approved", "todo", "open":
		return "○"
	default:
		return "◐"
	}
}

func short(it *board.Item) string {
	if it.Total > 0 {
		return fmt.Sprintf("%d/%d", it.Done, it.Total)
	}
	return it.Status
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
