package tui

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// withColors pins a color profile while the test draws, because lipgloss
// drops every color code without a terminal. The rest of the package reads
// the plain text back, so the profile is restored right after.
func withColors(f func()) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	f()
}

func sized(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

// ansi strips the color codes a style leaves behind, so a test can read the
// screen the way a person does. It also strips the OSC 8 hyperlink wrappers
// the bottom line puts around its links, leaving the link words behind.
var ansi = regexp.MustCompile("\x1b\\[[0-9;]*[A-Za-z]")

var osc = regexp.MustCompile("\x1b\\]8;;[^\x1b\\\\]*\x1b\\\\")

func plain(s string) string { return ansi.ReplaceAllString(osc.ReplaceAllString(s, ""), "") }

// body cuts the pane walls off a screen line and drops the padding on the
// right, so a test sees only what the pane holds.
func body(line string) string {
	b := plain(line)
	b = strings.TrimPrefix(b, "│")
	b = strings.TrimSuffix(b, "│")
	return strings.TrimRight(b, " ")
}

// column cuts one pane out of the whole screen, counted in cells.
func column(v string, x, w int) []string {
	var out []string
	for _, ln := range strings.Split(v, "\n") {
		r := []rune(plain(ln))
		if x >= len(r) {
			continue
		}
		out = append(out, string(r[x:min(len(r), x+w)]))
	}
	return out
}

// clocked pins the clock of the model, so a test can read the time it draws.
func clocked(m Model, hh, mm int) Model {
	m.now = time.Date(2026, 9, 27, hh, mm, 0, 0, time.UTC)
	return m
}

// linesFor finds the screen line that holds a piece of text.
func lineFor(t *testing.T, v, want string) ([]string, int) {
	t.Helper()
	lines := strings.Split(v, "\n")
	for i, ln := range lines {
		if strings.Contains(ln, want) {
			return lines, i
		}
	}
	t.Fatalf("no line holds %q", want)
	return nil, -1
}

func TestViewShowsTheSidebarAndTheDetail(t *testing.T) {
	m := sized(clocked(press(newModel(t), "2", "j", "j"), 20, 46), 200, 40)
	v := m.View()
	for _, want := range []string{
		"[1]─Active", "[2]─Specs ─ Scratchpad", "[3]─Plans ─ Tasks",
		"[4]─Bugs ─ Debt", "[5]─Done", "[0]─Detail",
		"basic · live · 2026-09-27 20:46",
		"specs/2026-09-20-alpha  Alpha story",
		"ID        : specs/2026-09-20-alpha",
		"FILE      : .acta/specs/2026-09-20-alpha.md",
		"Some text.",
	} {
		if !strings.Contains(v, want) {
			t.Errorf("view is missing %q", want)
		}
	}
	for _, line := range strings.Split(v, "\n") {
		if w := lipgloss.Width(line); w > 200 {
			t.Fatalf("line wider than the terminal (%d): %q", w, line)
		}
	}
}

func TestViewNeverOverflowsAnyWindow(t *testing.T) {
	for w := 30; w <= 200; w++ {
		for h := 10; h <= 60; h++ {
			base := sized(newModel(t), w, h)
			for name, m := range map[string]Model{
				"open":   base,
				"detail": press(base, "3"),
				"help":   press(base, "?"),
			} {
				for _, line := range strings.Split(sized(m, w, h).View(), "\n") {
					if got := lipgloss.Width(line); got > w {
						t.Fatalf("%dx%d %s: line is %d cells wide, the window is %d: %q", w, h, name, got, w, line)
					}
				}
			}
		}
	}
}

func TestViewNarrowShowsOnlyTheFocusedPane(t *testing.T) {
	for _, tc := range []struct {
		keys []string
		want string
		gone string
	}{
		{nil, "[1]─Active", "[2]─"},
		{[]string{"2"}, "[2]─Specs", "[1]─"},
		{[]string{"5"}, "[5]─Done", "[1]─"},
		{[]string{"0"}, "[0]─Detail", "[1]─"},
	} {
		m := sized(press(newModel(t), tc.keys...), 40, 20)
		v := m.View()
		if !strings.Contains(v, tc.want) {
			t.Errorf("after %v the view is missing %q", tc.keys, tc.want)
		}
		if strings.Contains(v, tc.gone) {
			t.Errorf("after %v the view still shows %q", tc.keys, tc.gone)
		}
	}
}

// paneBox gives the box pane p is drawn in. Below 60 columns only the pane
// with the focus reaches the screen, so a test that walks a pane focuses it
// first and reads the same box the screen shows.
func paneBox(m Model, p pane) box {
	g := m.geometry()
	if !g.wide {
		return g.full
	}
	return g.at(p)
}

// innerLines gives the lines drawn inside a pane, one per screen row, with
// the walls and the padding cut off so a test reads only the rows.
func innerLines(m Model, b box) []string {
	col := column(m.View(), b.x, b.w)
	out := make([]string, 0, b.inner)
	for i := range b.inner {
		y := b.y + 1 + i
		if y >= len(col) {
			break
		}
		out = append(out, body(col[y]))
	}
	return out
}

// cutTo cuts text to w cells with …, the same cut a row draws, so a test can
// say what a row should look like in a pane that is too narrow for the words.
func cutTo(text string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(text) <= w {
		return text
	}
	r := []rune(text)
	for len(r) > 0 && lipgloss.Width(string(r)) > w-1 {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// cells keeps the first w cells of a drawn line, so the scrollbar cell on the
// right wall is left out of what the row itself says.
func cells(text string, w int) string {
	if w <= 0 {
		return ""
	}
	for len(text) > 0 && lipgloss.Width(text) > w {
		text = text[:len(text)-1]
	}
	return text
}

// rowCount gives the words a row ends with when its work is in progress: the
// ticked boxes, then the agent when one is known.
func rowCount(it *board.Item) string {
	if it.Total == 0 {
		return ""
	}
	out := fmt.Sprintf("%d/%d", it.Done, it.Total)
	if it.Agent != "" {
		out += " · " + it.Agent
	}
	return out
}

// isRowLine says whether a drawn line is the row of the pane and nothing
// else: the words of the row, whole or cut to the cells the pane has, with the
// count of work in progress at the right end. A row that spilled onto a second
// line, a blank line between rows, or a meta line under the title all fail
// here.
func isRowLine(m Model, r row, line string, w int) bool {
	if r.group {
		return strings.Contains(line, "untyped")
	}
	it := m.board.Get(r.id)
	head := r.id
	if it != nil {
		name := it.ShortID
		if name == "" {
			name = it.ID
		}
		head = strings.Repeat("  ", r.depth) + name + "  " + it.Title
	}
	// Work in progress keeps its count whole at the right end, and the rest
	// of the line is the head cut to the cells the pane has left. The row is
	// padded out to the wall, and the padding is not part of it.
	rest := strings.TrimRight(line, " ")
	if it != nil && inProgress(it) {
		if tail := rowCount(it); tail != "" {
			if !strings.HasSuffix(rest, " "+tail) {
				return false
			}
			rest = strings.TrimRight(strings.TrimSuffix(rest, " "+tail), " ")
		}
	}
	return rest == cutTo(head, lipgloss.Width(rest))
}

// lineOfRow gives the drawn line that holds an item, so a test reads the row
// it means and not the line number it guessed.
func lineOfRow(t *testing.T, lines []string, id string) string {
	t.Helper()
	return lineWith(t, lines, id)
}

// lineWith gives the drawn line that holds word, the same way lineOfRow finds
// a row by its id, so a test can look for the tail of a row as well.
func lineWith(t *testing.T, lines []string, word string) string {
	t.Helper()
	for _, ln := range lines {
		if strings.Contains(ln, word) {
			return ln
		}
	}
	t.Fatalf("no drawn row holds %s: %q", word, lines)
	return ""
}

// TestEveryRowIsOneLineOnEveryTab walks every tab, both list panes and four
// widths, and reads what each pane actually draws. A pane has to show one row
// per line it has room for, and every screen line inside it has to be that
// pane's row at that position, the divider, or empty space below the last
// row, so a row can never take two lines.
func TestEveryRowIsOneLineOnEveryTab(t *testing.T) {
	for tab := range 2 {
		for _, p := range []pane{paneSpecs, paneDone} {
			for _, w := range []int{40, 60, 120, 200} {
				m := sized(focused(newModel(t), p, tab), w, 40)
				b := paneBox(m, p)
				inner := b.textW()
				rows, _, _ := m.slotOf(p)
				lines := innerLines(m, b)
				if len(lines) < b.rows {
					t.Fatalf("tab %d pane %d width %d: the pane drew %d of its %d rows", tab, p, w, len(lines), b.rows)
				}
				if b.inner > 0 && b.rows != b.inner {
					t.Fatalf("tab %d pane %d width %d: the pane shows %d rows in %d lines, want one row per line", tab, p, w, b.rows, b.inner)
				}
				if len(rows) == 0 {
					if lines[0] != "nothing here" {
						t.Errorf("tab %d pane %d width %d: an empty pane shows %q", tab, p, w, lines[0])
					}
					continue
				}
				// The words of a box fill every cell between the walls, so a
				// row runs right up to the scrollbar on the right wall.
				for i, ln := range lines {
					n := b.first + i
					if n >= len(rows) {
						if strings.TrimSpace(ln) != "" {
							t.Errorf("tab %d pane %d width %d: line %d below the last row holds %q", tab, p, w, i, ln)
						}
						continue
					}
					if !isRowLine(m, rows[n], cells(ln, inner), inner) {
						t.Errorf("tab %d pane %d width %d: line %d is %q, which is not row %s", tab, p, w, i, ln, rows[n].id)
					}
				}
			}
		}
	}
}

// TestAPaneWithOneKindOfWorkDrawsNoRule keeps only the stories of one kind of
// work, so whatever else the pane drew between its rows, every line left is a
// row of that work.
func TestAPaneWithOneKindOfWorkDrawsNoRule(t *testing.T) {
	full := newModel(t)
	for _, tc := range []struct {
		name string
		keep func(*board.Item) bool
	}{
		{"only in progress", func(it *board.Item) bool { return inProgress(it) }},
		{"only not started", func(it *board.Item) bool { return !inProgress(it) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var items []*board.Item
			for _, it := range full.board.Items {
				if it.Kind == board.KindStory && !it.Legacy && tc.keep(it) {
					items = append(items, it)
				}
			}
			m := sized(full, 200, 40)
			m.board = &board.Board{Items: items}
			for i, ln := range innerLines(m, paneBox(m, paneSpecs)) {
				if ln != "" && strings.Trim(ln, "─") == "" {
					t.Errorf("line %d draws a rule where a row belongs: %q", i, ln)
				}
			}
		})
	}
}

// statusWords are the words a list pane must never paint. A row says what
// kind of work it is by its count and its agent, not by naming a status.
var statusWords = []string{"todo", "doing", "open", "done", "approved", "dropped", "fixed", "wontfix", "in-progress", "fixing", "draft"}

// saysWord reports whether a drawn line carries w as a whole word, so the
// digits of a count like 2/5 and the tail of a file name never read as a hit.
func saysWord(line, w string) bool {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(w) + `\b`).MatchString(line)
}

// boardSays reports whether a file on the board really holds w, so a status
// word inside a file name or a title is not read as the view painting one.
func boardSays(m Model, w string) bool {
	word := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(w) + `\b`)
	for _, it := range m.board.Items {
		if word.MatchString(it.ID) || word.MatchString(it.Title) {
			return true
		}
	}
	return false
}

// TestNoListPanePaintsAStatusWord reads every line both list panes draw on
// every tab. A status word on screen has to come from a file name or a title
// the board really holds, which leaves no room for the view painting one.
func TestNoListPanePaintsAStatusWord(t *testing.T) {
	for tab := range 2 {
		for _, p := range []pane{paneSpecs, paneDone} {
			m := sized(focused(newModel(t), p, tab), 200, 40)
			for i, line := range innerLines(m, paneBox(m, p)) {
				for _, w := range statusWords {
					if saysWord(line, w) && !boardSays(m, w) {
						t.Errorf("tab %d pane %d line %d paints the status %q: %q", tab, p, i, w, line)
					}
				}
			}
		}
	}
}

// TestInProgressRowEndsWithItsCountAndAgent gives one item a count and an
// agent the way the board holds them, then reads the rows. The count and the
// agent sit at the right end of the in-progress row, and a row that is not in
// progress holds nothing but its name and its title.
func TestInProgressRowEndsWithItsCountAndAgent(t *testing.T) {
	m := newModel(t)
	it := m.board.Get("specs/2026-09-20-alpha")
	it.Done, it.Total, it.Agent = 2, 5, "claude"
	m = press(sized(m, 200, 40), "2")
	b := paneBox(m, paneSpecs)
	lines := innerLines(m, b)
	if got := cells(lineOfRow(t, lines, "specs/2026-09-20-alpha"), b.textW()); !strings.HasPrefix(strings.TrimSpace(got), "specs/2026-09-20-alpha  Alpha") ||
		!strings.HasSuffix(strings.TrimSpace(got), "2/5 · claude") {
		t.Errorf("the in-progress row is %q, want the name and the title then 2/5 · claude", got)
	}
	if got := lineOfRow(t, lines, "specs/2026-09-28-from-scratch-design"); strings.Contains(got, "claude") {
		t.Errorf("the not-started row is %q, want only the name and the title", got)
	}
	// No line between the two rows is a rule, because there is no rule left.
	for i, ln := range lines {
		if ln != "" && strings.Trim(ln, "─ ") == "" {
			t.Errorf("line %d is a rule: %q", i, ln)
		}
	}
}

// TestALongTitleIsCutSoTheCountFits gives the in-progress row a title far
// longer than the pane. The title gives way first, so the count and the agent
// stay whole at the right end.
func TestALongTitleIsCutSoTheCountFits(t *testing.T) {
	m := newModel(t)
	it := m.board.Get("specs/2026-09-20-alpha")
	it.Done, it.Total, it.Agent = 2, 5, "claude"
	it.Title = strings.Repeat("x", 200)
	m = press(sized(m, 60, 40), "2")
	b := paneBox(m, paneSpecs)
	row := cells(lineWith(t, innerLines(m, b), "2/5 · claude"), b.textW())
	if !strings.HasSuffix(strings.TrimSpace(row), "2/5 · claude") {
		t.Errorf("the count and the agent were cut: %q", row)
	}
	if !strings.Contains(row, "…") {
		t.Errorf("a 200-character title should be cut with …: %q", row)
	}
	if got := lipgloss.Width(row); got != b.textW() {
		t.Errorf("the row is %d cells wide, the pane holds %d", got, b.textW())
	}
}

// sgrParams gives the numbers inside every color code of a drawn line, so a
// test can ask which brushes it wears without counting escape characters.
func sgrParams(line string) [][]int {
	var out [][]int
	for _, m := range regexp.MustCompile("\x1b\\[([0-9;]*)m").FindAllStringSubmatch(line, -1) {
		var ps []int
		for _, p := range strings.Split(m[1], ";") {
			n, err := strconv.Atoi(p)
			if err != nil {
				ps = nil
				break
			}
			ps = append(ps, n)
		}
		out = append(out, ps)
	}
	return out
}

// wears reports whether one color code of a drawn line holds every number in
// want, which is how a test names a brush.
func wears(line string, want ...int) bool {
	for _, ps := range sgrParams(line) {
		all := true
		for _, n := range want {
			if !slices.Contains(ps, n) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

// paintedLine cuts the row of a pane out of a raw screen line. The cuts go by
// display cells and skip both walls, so the row comes out with its color codes
// and without the accent the border of the focused pane wears.
func paintedLine(v string, b box, i int) string {
	lines := strings.Split(v, "\n")
	y := b.y + 1 + i
	if y < 0 || y >= len(lines) || b.w < 3 {
		return ""
	}
	cell := xansi.TruncateLeft(xansi.Truncate(lines[y], b.x+1+b.w-2, ""), b.x+1, "")
	// The walls of a pane wear the border brush, and the cut keeps the code
	// each wall opened with. The row is the one run between them, from the
	// reset after the left wall to the reset that ends it.
	wall := strings.Index(cell, "\x1b[0m")
	if wall < 0 {
		return cell
	}
	cell = cell[wall+len("\x1b[0m"):]
	if end := strings.Index(cell, "\x1b[0m"); end >= 0 {
		cell = cell[:end+len("\x1b[0m")]
	}
	return cell
}

// TestTheSelectedRowWearsASubtleBand reads the brushes of the selected row
// and of the rows around it. The cursor is a dark band across the whole row
// with bright text, never reversed video; every other row is dim and carries
// no background, and the in-progress one keeps the accent.
func TestTheSelectedRowWearsASubtleBand(t *testing.T) {
	withColors(func() {
		// The rows are the board's own order, so the in-progress spec is the
		// third one. One j puts the cursor on the row above it, and the
		// in-progress row is then drawn unselected with its own brush.
		m := sized(press(newModel(t), "2", "j"), 120, 40)
		b := paneBox(m, paneSpecs)
		v := m.View()

		sel := paintedLine(v, b, 1)
		if wears(sel, 7) {
			t.Errorf("the selected row is reverse video: %q", sel)
		}
		if wears(sel, 2) {
			t.Errorf("the selected row is faint: %q", sel)
		}
		if !wears(sel, 48, 5, 236) {
			t.Errorf("the selected row has no background 236: %q", sel)
		}
		reset := strings.LastIndex(sel, "\x1b[0m")
		if want := b.textW(); reset < 0 || lipgloss.Width(plain(sel[:reset])) != want {
			t.Errorf("the band stops short of the row width %d: %q", want, sel)
		}

		// The in-progress row is not selected, so it keeps the accent and
		// is dim.
		going := paintedLine(v, b, 2)
		if !wears(going, 2) || !wears(going, 38, 5, 39) {
			t.Errorf("an unselected in-progress row should be dim and keep the accent: %q", going)
		}
		if wears(going, 48, 5, 236) {
			t.Errorf("an unselected row wears a background: %q", going)
		}
		for _, i := range []int{3, 4} {
			row := paintedLine(v, b, i)
			if !wears(row, 2) {
				t.Errorf("row %d is not dim: %q", i, row)
			}
			if wears(row, 48, 5, 236) {
				t.Errorf("row %d wears a background: %q", i, row)
			}
		}
	})
}

func TestViewDetailHeaderAlignsItsColons(t *testing.T) {
	var colon = regexp.MustCompile(`^([A-Z]+) +: `)
	m := sized(press(clocked(newModel(t), 20, 46), "j"), 120, 40)
	g := m.geometry()
	var at []int
	first := ""
	for _, seg := range column(m.View(), g.detail.x, g.detail.w) {
		b := strings.TrimRight(strings.TrimPrefix(strings.TrimSuffix(seg, "│"), "│"), " ")
		if !colon.MatchString(b) {
			continue
		}
		if first == "" {
			first = b
		}
		at = append(at, strings.Index(b, ":"))
	}
	if len(at) < 3 {
		t.Fatalf("the detail header should hold several labels, found %d", len(at))
	}
	for _, c := range at {
		if c != at[0] {
			t.Fatalf("the colons do not line up: %v", at)
		}
	}
	if !strings.HasPrefix(first, "ID") {
		t.Errorf("the first label should be ID, got %q", first)
	}
}

func TestViewDebtDetail(t *testing.T) {
	cfg := treeCfg(t, map[string]string{
		".acta/plans/2026-09-26-short-ids.md": "---\nid: PLAN-3\nhash: k3f2\n---\n# Short IDs\n",
		".acta/debt/2026-09-27-short-ids.md":  "---\nid: DEBT-1\nhash: t9qe\nparent: plans/2026-09-26-short-ids\n---\n# Review NOTEs: Short IDs\n\n- [ ] a\n- [x] b\n- [-] c\n",
	})
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	// 120 columns, the normal width, must still draw all five tab names.
	m = sized(openTab(t, m, paneBugs, 1), 120, 40)

	v := m.View()
	// The gap between names can be the wide " ─ " or, once five names need
	// to fit, a plain space, so look for the names themselves in order.
	if i, j := strings.Index(v, "Bugs"), strings.Index(v, "Debt"); i < 0 || j < i {
		t.Fatalf("Debt should come after Bugs in the tab title: %q", v)
	}

	g := m.geometry()
	var colon = regexp.MustCompile(`^([A-Z]+) +: `)
	var at []int
	var labels []string
	for _, seg := range column(v, g.detail.x, g.detail.w) {
		row := strings.TrimRight(strings.TrimPrefix(strings.TrimSuffix(seg, "│"), "│"), " ")
		if !colon.MatchString(row) {
			continue
		}
		at = append(at, strings.Index(row, ":"))
		labels = append(labels, row)
	}
	for _, c := range at {
		if c != at[0] {
			t.Fatalf("the detail colons do not line up: %v (%v)", at, labels)
		}
	}
	detail := strings.Join(labels, "\n")
	for _, want := range []string{`DEBT\s+: a$`, `STATUS\s+: open$`, `FROM\s+: PLAN-3 . Short IDs$`} {
		if ok, _ := regexp.MatchString("(?m)"+want, detail); !ok {
			t.Errorf("detail header is missing %q, got:\n%s", want, detail)
		}
	}

}

func TestViewDetailLeavesEmptyLinesOut(t *testing.T) {
	// The first plan of the list links no spec, so no SPEC line is drawn.
	m := sized(press(newModel(t), "3"), 120, 40)
	g := m.geometry()
	detail := strings.Join(column(m.View(), g.detail.x, g.detail.w), "\n")
	if strings.Contains(detail, "SPEC") {
		t.Error("a plan with no spec should not show a SPEC line")
	}
	if strings.Contains(detail, "WORKTREE") || strings.Contains(detail, "AGENT") {
		t.Error("a line with no value should be left out")
	}
	// The plan below it does link a spec, so its line has to be there.
	m = sized(press(m, "j"), 120, 40)
	g = m.geometry()
	detail = strings.Join(column(m.View(), g.detail.x, g.detail.w), "\n")
	if !strings.Contains(detail, "SPEC") {
		t.Error("a plan with a spec should show its SPEC line")
	}
}

func TestViewDetailShowsTheSectionOfItsOwnItem(t *testing.T) {
	m := sized(press(newModel(t), "3", "]", "j"), 120, 40)
	v := plain(m.View())
	if !strings.Contains(v, "plans/2026-09-21-alpha#task-2") {
		t.Error("the task row is missing")
	}
	if !strings.Contains(v, "TASK      : Second step") {
		t.Error("the detail should name the kind and the title")
	}
	if strings.Contains(v, "First step") {
		t.Error("the detail of a task should not hold another section")
	}
}
func TestViewShowsProblemsOfTheSelectedItem(t *testing.T) {
	// The rows are in board order, so three j steps reach the weird story.
	m := sized(press(newModel(t), "2", "j", "j", "j"), 120, 60)
	if !strings.Contains(plain(m.View()), "! ") {
		t.Error("the problems of the selected item are missing")
	}
	if !strings.Contains(plain(m.View()), "untyped (1)") {
		t.Error("the row that holds the files outside .acta/ is missing")
	}
}

func TestViewStatusLineShowsHelpAndClock(t *testing.T) {
	m := sized(clocked(newModel(t), 20, 46), 100, 30)
	last := lastLine(m.View())
	if got := lipgloss.Width(last); got != 100 {
		t.Errorf("the status line is %d cells wide, want 100", got)
	}
	if !strings.HasPrefix(plain(last), "? help") {
		t.Errorf("the left should show only ? help, got %q", plain(last))
	}
	if !strings.Contains(last, "basic · live · 2026-09-27 20:46") {
		t.Errorf("the project, mode or date is missing: %q", plain(last))
	}
	if !strings.Contains(last, "Feedback") {
		t.Errorf("the feedback link is missing: %q", plain(last))
	}
	if strings.Contains(plain(last), "Donate") {
		t.Errorf("Donate shows with no url set: %q", plain(last))
	}
	if strings.Contains(plain(last), "pmb") {
		t.Errorf("the line must not name pmb: %q", plain(last))
	}
	m.manual = true
	if !strings.Contains(m.View(), "basic · paused · 2026-09-27 20:46") {
		t.Error("watching off should read paused")
	}
	m.status = "acta: x status done committed"
	m = sized(m, 100, 30)
	if !strings.Contains(m.View(), "committed") {
		t.Error("the last message should take the place of the hints")
	}
	m.searching, m.query = true, "alpha"
	if !strings.Contains(plain(m.View()), "/alpha") {
		t.Error("the search box should take the place of the hints")
	}
	for w := 30; w <= 200; w++ {
		last := lastLine(sized(clocked(newModel(t), 20, 46), w, 20).View())
		if got := lipgloss.Width(last); got > w {
			t.Fatalf("at %d columns the status line is %d cells wide", w, got)
		}
		if !strings.Contains(plain(last), "2026-09-27 20:46") {
			t.Fatalf("at %d columns the date and time are missing: %q", w, plain(last))
		}
		if strings.Contains(plain(last), "pmb") {
			t.Fatalf("at %d columns the line names pmb: %q", w, plain(last))
		}
	}
}

func TestViewStatusLineShowsDonateWhenSet(t *testing.T) {
	m := sized(clocked(newModel(t), 20, 46), 200, 30)
	m.cfg.Links.Donate = "https://ko-fi.com/someone"
	last := lastLine(m.View())
	if !strings.Contains(plain(last), "Donate") {
		t.Errorf("Donate missing with the url set: %q", plain(last))
	}
	if !strings.Contains(plain(last), "Feedback") {
		t.Errorf("Feedback missing: %q", plain(last))
	}
	m.cfg.Links.Donate = ""
	if strings.Contains(plain(lastLine(m.View())), "Donate") {
		t.Errorf("Donate shows with no url set: %q", plain(lastLine(m.View())))
	}
}

func TestViewStatusLineDropsProjectThenStatus(t *testing.T) {
	wide := plain(lastLine(sized(clocked(newModel(t), 20, 46), 200, 30).View()))
	if !strings.Contains(wide, "basic · live · 2026-09-27 20:46") {
		t.Fatalf("a wide line keeps all three, got %q", wide)
	}
	var sawProject, sawBareMode, sawDateOnly, sawLeftGone bool
	for w := 199; w >= 30; w-- {
		line := plain(lastLine(sized(clocked(newModel(t), 20, 46), w, 20).View()))
		if !strings.Contains(line, "2026-09-27 20:46") {
			t.Fatalf("at %d columns the date and time are gone: %q", w, line)
		}
		switch {
		case strings.Contains(line, "basic"):
			sawProject = true
		case strings.Contains(line, "live") && !sawBareMode:
			if sawDateOnly {
				t.Fatalf("at %d columns the status came back after the date stood alone: %q", w, line)
			}
			sawBareMode = true
		case !strings.Contains(line, "live"):
			sawDateOnly = true
		}
		if !strings.Contains(line, "? help") {
			sawLeftGone = true
		}
	}
	if !sawProject {
		t.Error("no width kept the project")
	}
	if !sawBareMode {
		t.Error("no width dropped the project but kept the status")
	}
	if !sawDateOnly {
		t.Error("no width dropped the status but kept the date")
	}
	if !sawLeftGone {
		t.Error("no width gave up the left side")
	}
}

func TestViewHelpPopupCoversThePanes(t *testing.T) {
	m := sized(press(clocked(newModel(t), 20, 46), "?"), 100, 30)
	v := m.View()
	if !strings.Contains(v, "Keys") || !strings.Contains(v, "new bug") {
		t.Error("the help popup is missing")
	}
	if !strings.Contains(v, "[1]─Active") || !strings.Contains(v, "[0]─Detail") {
		t.Error("the popup should cover the boxes, not replace them")
	}
	if !strings.HasSuffix(plain(lastLine(v)), "2026-09-27 20:46 | Feedback  dev") {
		t.Error("the popup should not hide the status line")
	}
	if strings.Contains(press(m, "?").View(), "Keys") {
		t.Error("? should close the help popup")
	}
}

// focused puts the focus on a pane and picks the tab of pane [1], so a test
// can walk every tab with the focus on either side of the screen.
func focused(m Model, f pane, tab int) Model {
	m.focus, m.last = f, paneSpecs
	if f < paneDone {
		m.tab[f] = tab % max(1, len(sidebar[f].tabs))
	}
	m.tab[paneSpecs] = tab % max(1, len(sidebar[paneSpecs].tabs))
	m.tab[paneDone] = tab % max(1, len(m.doneTabNames()))
	m.follows = paneSpecs
	return m
}

// TestNoRoundedCorners walks every width from 1 to 200, every tab, and the
// focus on each pane, with no popup and with each of the three popups, so no
// frame the screen can draw still holds a rounded corner.
func TestNoRoundedCorners(t *testing.T) {
	for w := 1; w <= 200; w++ {
		for tab := range len(sidebar[paneSpecs].tabs) {
			for _, f := range []pane{paneSpecs, paneDone, paneDetail} {
				m := sized(focused(newModel(t), f, tab), w, 40)
				for _, frame := range []struct {
					what string
					v    string
				}{
					{"plain", m.View()},
					{"help", press(m, "?").View()},
					{"value", press(m, "s").View()},
					{"slug", press(m, "n").View()},
				} {
					if strings.ContainsAny(plain(frame.v), "╭╮╰╯") {
						t.Fatalf("width %d, tab %d, focus %d, %s: rounded corner in the frame", w, tab, f, frame.what)
					}
				}
			}
		}
	}
}

// TestPopupChangesOnlyItsBox reads the screen with the help popup and the
// screen without it cell by cell. Every cell outside the box has to be the
// same, so the panes behind the popup stay on screen instead of being wiped.
func TestPopupChangesOnlyItsBox(t *testing.T) {
	withColors(func() {
		for _, w := range []int{60, 120, 200} {
			for _, h := range []int{3, 10, 24, 40, 60} {
				m := sized(newModel(t), w, h)
				pop := press(m, "?")
				before := strings.Split(plain(m.View()), "\n")
				after := strings.Split(plain(pop.View()), "\n")
				x0, y0, bw, bh := popupRect(strings.Split(pop.popupBox(), "\n"), pop.width, pop.height)
				if bh == 0 {
					t.Fatalf("%dx%d: the help popup drew no box", w, h)
				}
				outside := 0
				for y := range min(len(before), len(after)) {
					b, a := []rune(before[y]), []rune(after[y])
					for x := range min(len(b), len(a)) {
						if x >= x0 && x < x0+bw && y >= y0 && y < y0+bh {
							continue
						}
						if b[x] != a[x] {
							outside++
						}
					}
				}
				t.Logf("%dx%d: %d cells outside the popup box changed", w, h, outside)
				if outside != 0 {
					t.Errorf("%dx%d: %d cells outside the popup box changed", w, h, outside)
				}
			}
		}
	})
}

// atCell gives the first n cells of a plain line, whole runes only, so a test
// can say what a splice has to leave on either side of the box.
func atCell(s string, n int) string {
	var b strings.Builder
	for _, r := range s {
		if lipgloss.Width(b.String()+string(r)) > n {
			break
		}
		b.WriteRune(r)
	}
	return b.String()
}

// TestSpliceKeepsWideRunesWhole puts a box over a line of wide runes at every
// spot it can land. When both edges sit between runes, the line outside the
// box has to come out exactly as it went in; when an edge lands inside a wide
// rune, the rune has to stay whole and the line has to keep its width.
func TestSpliceKeepsWideRunesWhole(t *testing.T) {
	line := "日本語 abc テキスト"
	want := lipgloss.Width(line)
	for x0 := range want {
		for w := 1; x0+w <= want; w++ {
			got := plain(splice(line, strings.Repeat("x", w), x0, w))
			if !utf8.ValidString(got) {
				t.Fatalf("splice at %d width %d cut a rune in half: %q", x0, w, got)
			}
			if lipgloss.Width(got) != want {
				t.Fatalf("splice at %d width %d: the line is %d cells wide, want %d", x0, w, lipgloss.Width(got), want)
			}
			left, cutRight := atCell(line, x0), atCell(line, x0+w)
			if lipgloss.Width(left) != x0 || lipgloss.Width(cutRight) != x0+w {
				continue
			}
			right := strings.TrimPrefix(line, cutRight)
			if got != left+strings.Repeat("x", w)+right {
				t.Fatalf("splice at %d width %d: %q", x0, w, got)
			}
		}
	}
}

func TestViewValuePopupAndSlug(t *testing.T) {
	m := sized(press(newModel(t), "]", "]", "]", "s"), 100, 30)
	if !strings.Contains(m.View(), "wontfix") {
		t.Error("the value popup options are not shown")
	}
	m = sized(press(newModel(t), "n", "a", "b"), 100, 30)
	if !strings.Contains(m.View(), "new bug slug: ab") {
		t.Error("the slug prompt is not shown")
	}
}

func TestViewFocusedPaneWearsTheAccent(t *testing.T) {
	m := newModel(t)
	if m.edge(paneActive).GetForeground() != accentColor {
		t.Error("the box with the focus should draw its border in the accent color")
	}
	if m.edge(paneDone).GetForeground() == accentColor {
		t.Error("a box without the focus should not wear the accent color")
	}
	m = press(m, "5")
	if m.edge(paneDone).GetForeground() != accentColor {
		t.Error("the focus moved, so the accent color should move with it")
	}
	if m.edge(paneActive).GetForeground() == accentColor {
		t.Error("the box that lost the focus should be dim")
	}
}

func TestViewFinishedTabsFollowTheLastSidebarPane(t *testing.T) {
	for _, tc := range []struct {
		keys []string
		want string
	}{
		{[]string{"2", "5"}, "[5]─Done ─ Dropped"},
		{[]string{"3", "5"}, "[5]─Done ─"},
		{[]string{"4", "5"}, "[5]─Fixed ─ Wontfix"},
		{[]string{"4", "]", "5"}, "[5]─Done ─ Wontfix"},
	} {
		v := sized(press(newModel(t), tc.keys...), 120, 40).View()
		if !strings.Contains(v, tc.want) {
			t.Errorf("after %v the Done box is missing %q", tc.keys, tc.want)
		}
	}
}

func TestViewEmptyRepo(t *testing.T) {
	b, err := board.Load(config.Default(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	m := sized(New(config.Default(t.TempDir()), b, true), 100, 30)
	if !strings.Contains(m.View(), "no .acta/ yet") {
		t.Error("empty repo message missing")
	}
}

func lastLine(v string) string {
	lines := strings.Split(v, "\n")
	return lines[len(lines)-1]
}

func TestViewInProgressRowsWearTheAccent(t *testing.T) {
	withColors(func() {
		testViewInProgressRowsWearTheAccent(t)
	})
}

func testViewInProgressRowsWearTheAccent(t *testing.T) {
	// Step twice, so neither the in-progress row nor the not-started one
	// shows the selected brush: both must show their own brush instead.
	m := sized(press(splitModel(t), "2", "j"), 200, 40)
	b := paneBox(m, paneSpecs)
	seenGoing, seenWaiting := false, false
	for i, ln := range innerLines(m, b) {
		cell := paintedLine(m.View(), b, i)
		switch {
		case strings.Contains(ln, "specs/2026-09-20-alpha"):
			seenGoing = true
			if !wears(cell, 38, 5, 39) {
				t.Errorf("the in-progress row wears no accent color: %q", cell)
			}
		case strings.Contains(ln, "specs/2026-09-22-beta"):
			seenWaiting = true
			if wears(cell, 38, 5, 39) {
				t.Errorf("the not-started row wears the accent color: %q", cell)
			}
		}
	}
	if !seenGoing || !seenWaiting {
		t.Error("both rows should show")
	}
}

func TestViewHasNoSectionLabels(t *testing.T) {
	m := sized(splitModel(t), 200, 40)
	for _, ln := range strings.Split(plain(m.View()), "\n") {
		low := strings.ToLower(body(ln))
		if strings.HasPrefix(low, "in progress") || strings.HasPrefix(low, "not started") {
			t.Errorf("a section label leaked into the list: %q", ln)
		}
	}
}

func TestViewHelpListsTheNewKeys(t *testing.T) {
	m := sized(press(newModel(t), "?"), 120, 40)
	v := plain(m.View())
	for _, want := range []string{"enter", "focus the detail", "e", "open the row", "esc", "back to the list"} {
		if !strings.Contains(v, want) {
			t.Errorf("the help shows no %q:\n%s", want, v)
		}
	}
	if strings.Contains(v, "enter t s n     edit, set, new bug") {
		t.Error("the old enter line is still there")
	}
}

// TestZTogglesAndFocusRestores walks the whole life of the expand key: z gives
// the focused pane the room, z takes it back, moving to another box takes it
// back, and z on the detail box does nothing at all.
func TestZTogglesAndFocusRestores(t *testing.T) {
	m := sized(newModel(t), 120, 40)
	if m.expanded != -1 {
		t.Fatalf("a screen with nothing expanded reads %d", m.expanded)
	}
	// Pane [3] has the focus, z expands it and z puts the room back.
	m = press(m, "3", "z")
	if m.expanded != int(panePlans) {
		t.Fatalf("z on pane [3] expanded %d, want %d", m.expanded, panePlans)
	}
	m = press(m, "z")
	if m.expanded != -1 {
		t.Fatalf("a second z should give the room back, it reads %d", m.expanded)
	}
	// A tab of the same pane is not another box, so the room stays.
	m = press(m, "z", "]")
	if m.expanded != int(panePlans) {
		t.Fatalf("a tab change should keep the pane expanded, it reads %d", m.expanded)
	}
	// Moving the focus to another box gives the room back, by key and by tab.
	for _, keys := range [][]string{{"4"}, {"tab"}, {"shift+tab"}, {"0"}, {"1"}} {
		m = press(sized(newModel(t), 120, 40), append([]string{"3", "z"}, keys...)...)
		if m.expanded != -1 {
			t.Errorf("%v should give the room back, it reads %d", keys, m.expanded)
		}
	}
	// The detail box has no room to give, so z there does nothing.
	m = press(sized(newModel(t), 120, 40), "3", "z", "0", "z")
	if m.expanded != -1 {
		t.Fatalf("z on the detail box expanded %d, want nothing", m.expanded)
	}
	// The pane that has the room is the one that grew, and the others keep 3
	// lines on a screen tall enough to give them.
	wide := press(sized(newModel(t), 120, 40), "3", "z")
	hs := wide.leftHeights(wide.height - 1)
	if hs[panePlans] <= 3 {
		t.Errorf("the expanded pane has %d lines, want more than the 3 of the others: %v", hs[panePlans], hs)
	}
	for p, h := range hs {
		if p != int(panePlans) && h != 3 {
			t.Errorf("pane %d has %d lines next to the expanded one, want 3", p+1, h)
		}
	}
}

// TestCounterShowsSelectedItemNotLine reads what every box writes in its
// bottom border: the item under the cursor out of the items the pane holds,
// never the line on screen, and nothing at all on the detail box.
func TestCounterShowsSelectedItemNotLine(t *testing.T) {
	withColors(func() {
		// Forty open plans, more than the box can show. The cursor walks down
		// while the list stays at its top, so a counter reading the line on
		// screen would keep saying 1.
		m := press(longModel(t), keyOf(panePlans))
		for i, want := range []string{"1 of 40", "2 of 40", "3 of 40"} {
			if i > 0 {
				m = press(m, "j")
			}
			if m.off[panePlans] != 0 {
				t.Fatalf("the list scrolled to line %d, the counter cannot be a line number", m.off[panePlans])
			}
			if got := footOf(t, m, panePlans); got != want {
				t.Errorf("pane [3] with item %d writes %q, want %q", i+1, got, want)
			}
		}
		// The counter sits at the right end of the border, the way lazygit
		// puts it, with the corner as the last thing the line holds.
		if line := bottomLine(t, m, panePlans); !strings.HasSuffix(line, " 3 of 40 ┘") {
			t.Errorf("the bottom border of pane [3] reads %q, want the counter at its right end", line)
		}
		// Every sidebar pane of the plain fixture writes the item under the
		// cursor out of the items it holds, on each of its tabs.
		for p := pane(0); p < paneDetail; p++ {
			for tab := range max(1, len(sidebar[p].tabs)) {
				// A screen of its own for every box, because the keys write
				// into the cursors the model shares.
				one := press(sized(newModel(t), 160, 50), keyOf(p))
				for range tab {
					one = press(one, "]")
				}
				rows, _, _ := one.slotOf(p)
				if len(rows) == 0 {
					continue
				}
				stepped := press(one, "j")
				rows, sel, idx := stepped.slotOf(p)
				want := itemCount(cursorOf(rows, *sel, *idx)+1, len(rows))
				if got := footOf(t, stepped, p); got != want {
					t.Errorf("pane %d on tab %d writes %q, want %q", p+1, tab, got, want)
				}
			}
		}
		// A search that finds nothing empties every pane, and an empty pane
		// still counts, as 0 of 0.
		empty := sized(newModel(t), 160, 50)
		empty.query = "nothing in the board matches this"
		if rows, _, _ := empty.slotOf(panePlans); len(rows) != 0 {
			t.Fatalf("the search left %d rows in pane [3]", len(rows))
		}
		if got := footOf(t, empty, panePlans); got != "0 of 0" {
			t.Errorf("an empty pane writes %q, want %q", got, "0 of 0")
		}
		// The detail box keeps its scrollbar and writes no counter, with an
		// item under it and with nothing under it.
		for _, d := range []Model{press(sized(newModel(t), 160, 50), "3", "0"), empty} {
			if got := footOf(t, d, paneDetail); strings.Contains(got, " of ") {
				t.Errorf("the detail box writes a counter: %q", got)
			}
		}
	})
}

// thumbFixture gives every pane more rows than any of the two screens below
// can show, and the detail body a whole screenful of lines, so all six boxes
// have a thumb to draw: thirty of every kind, work under way for the Active
// pane, and finished ones for the Done pane.
func thumbFixture(t *testing.T) Model {
	t.Helper()
	body := "# Plan\n" + strings.Repeat("\nA line of the body.\n", 60) +
		"\n### Task 1: First step\n\n- [ ] **Step 1: Do it**\n"
	files := map[string]string{}
	for i := range 30 {
		n := fmt.Sprint(i)
		files[".acta/specs/2026-09-20-open-"+n+".md"] = "# Open spec\n"
		files[".acta/specs/2026-09-19-done-"+n+".md"] = "---\nstatus: done\n---\n# Done spec\n"
		files[".acta/specs/2026-09-18-going-"+n+".md"] = "---\nstatus: in-progress\n---\n# Going spec\n"
		files[".acta/scratch/2026-09-20-idea-"+n+".md"] = "---\nstatus: brainstorming\n---\n# Idea\n"
		files[".acta/plans/2026-09-20-open-"+n+".md"] = body
		files[".acta/plans/2026-09-19-done-"+n+".md"] = "---\nstatus: done\n---\n" + body
		files[".acta/bugs/2026-09-20-open-"+n+".md"] = "# Open bug\n"
		files[".acta/bugs/2026-09-19-fixed-"+n+".md"] = "---\nstatus: fixed\n---\n# Fixed bug\n"
		files[".acta/debt/2026-09-20-debt-"+n+".md"] = "# Review NOTEs: debt\n\n- [ ] a thing to do\n"
		files[".acta/debt/2026-09-19-debt-"+n+".md"] = "---\nstatus: done\n---\n# Review NOTEs: debt\n\n- [x] a thing done\n"
	}
	cfg := treeCfg(t, files)
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return m
}

// paneLine gives the drawn line of pane p at the body line i, as plain text.
func paneLine(v string, b box, i int) string {
	lines := strings.Split(v, "\n")
	y := b.y + 1 + i
	if y < 0 || y >= len(lines) {
		return ""
	}
	r := []rune(plain(lines[y]))
	if b.x >= len(r) {
		return ""
	}
	return string(r[b.x:min(len(r), b.x+b.w)])
}

// rawPaneLine gives the drawn line of pane p at the body line i with its
// colors, so a test can ask which brush a cell wears.
func rawPaneLine(v string, b box, i int) string {
	lines := strings.Split(v, "\n")
	y := b.y + 1 + i
	if y < 0 || y >= len(lines) {
		return ""
	}
	return xansi.TruncateLeft(xansi.Truncate(lines[y], b.x+b.w, ""), b.x, "")
}

// scrollTo puts pane p at the top, in the middle or at the end of its content.
func scrollTo(m Model, p pane, at string) Model {
	if p == paneDetail {
		m = press(m, "0")
	}
	switch at {
	case "top":
		return press(m, "g")
	case "middle":
		return press(m, "ctrl+d", "ctrl+d")
	default:
		return press(m, "G")
	}
}

// paneTabs is how many tabs a pane has to walk: the two kind boxes show two
// each, the rest show one, and the detail box sits outside the table.
func paneTabs(p pane) int {
	if p >= paneDone {
		return 1
	}
	return max(1, len(sidebar[p].tabs))
}

// TestThumbSitsOnTheBorderNotInside walks every pane, both screens and all
// three places in the content, and reads the drawn cells. No pane spends a
// cell of its own on a scrollbar: the line is as wide as the pane, the cell
// left of the right wall is content, and the thumb is that wall and nothing
// else.
func TestThumbSitsOnTheBorderNotInside(t *testing.T) {
	withColors(func() {
		for _, size := range [][2]int{{80, 30}, {160, 50}} {
			for p := pane(0); p <= paneDetail; p++ {
				for tab := range paneTabs(p) {
					for _, at := range []string{"top", "middle", "end"} {
						checkThumbOnTheBorder(t, p, tab, at, size[0], size[1])
					}
				}
			}
		}
	})
}

func checkThumbOnTheBorder(t *testing.T, p pane, tab int, at string, w, h int) {
	t.Helper()
	m := press(sized(thumbFixture(t), w, h), keyOf(p))
	for range tab {
		m = press(m, "]")
	}
	if p == paneDone {
		// Done reads the pane that had the focus last, so it is read with a
		// kind pane behind it and never with Active, which has none.
		m = press(press(m, keyOf(paneSpecs)), keyOf(paneDone))
	}
	if p == paneDetail {
		// The detail box shows the body of the item under the cursor, so a
		// list pane picks one first.
		m = press(press(m, keyOf(panePlans)), keyOf(paneDetail))
	}
	m = scrollTo(m, p, at)
	if rows, _, _ := m.slotOf(m.listPane()); p != paneDetail && len(rows) == 0 {
		t.Fatalf("pane %d tab %d holds no rows to scroll", p+1, tab)
	}
	b := paneBox(m, p)
	if b.inner < 1 {
		t.Fatalf("pane %d tab %d has no room to draw", p+1, tab)
	}
	v := m.View()
	// The thumb wears the brush the pane's own wall wears, so the accent on
	// the pane with the focus and the dim brush on the others.
	wall := m.edge(p)
	other := accent
	if m.focus == p {
		other = faint
	}
	thumbs := 0
	firstThumb, lastThumb := -1, -1
	for i := range b.inner {
		ln := paneLine(v, b, i)
		r := []rune(ln)
		if len(r) != b.w {
			t.Errorf("pane %d tab %d %dx%d at the %s: line %d is %d cells, the pane is %d: %q",
				p+1, tab, w, h, at, i, len(r), b.w, ln)
			continue
		}
		if string(r[0]) != "│" {
			t.Errorf("pane %d tab %d %dx%d at the %s: line %d starts with %q, want the left wall",
				p+1, tab, w, h, at, i, string(r[0]))
		}
		last := string(r[len(r)-1])
		if last != "│" && last != "┃" {
			t.Errorf("pane %d tab %d %dx%d at the %s: line %d ends on %q, want a wall cell",
				p+1, tab, w, h, at, i, last)
		}
		// The cell left of the right wall is the last cell of the pane's own
		// content, never a bar of its own.
		if inner := string(r[len(r)-2]); strings.ContainsAny(inner, "░█┃") {
			t.Errorf("pane %d tab %d %dx%d at the %s: line %d spends an inner cell on %q",
				p+1, tab, w, h, at, i, inner)
		}
		raw := rawPaneLine(v, b, i)
		if !strings.Contains(raw, wall.Render(last)) {
			t.Errorf("pane %d tab %d %dx%d at the %s: line %d draws its wall cell %q in another brush: %q",
				p+1, tab, w, h, at, i, last, raw)
		}
		if strings.Contains(raw, other.Render(last)) {
			t.Errorf("pane %d tab %d %dx%d at the %s: line %d draws its wall cell %q in the other pane's brush: %q",
				p+1, tab, w, h, at, i, last, raw)
		}
		if last != "┃" {
			continue
		}
		thumbs++
		if firstThumb < 0 {
			firstThumb = i
		}
		lastThumb = i
	}
	if strings.ContainsAny(plain(v), "░█") {
		t.Errorf("pane %d tab %d %dx%d at the %s: the view still draws a track or a block",
			p+1, tab, w, h, at)
	}
	if thumbs < 1 {
		t.Errorf("pane %d tab %d %dx%d at the %s: %d lines of content, no thumb on the wall",
			p+1, tab, w, h, at, m.linesOf(p))
		return
	}
	switch at {
	case "top":
		if firstThumb != 0 {
			t.Errorf("pane %d tab %d %dx%d: the thumb starts on line %d, want the first", p+1, tab, w, h, firstThumb)
		}
	case "middle":
		// A wall with room to move the thumb off both ends has to move it:
		// half way down, the thumb cannot still sit on the first line. A
		// wall three lines high with a one-line thumb rounds to the first
		// line, because that is where the middle of the content lands.
		if b.inner-thumbs < 3 {
			break
		}
		if firstThumb <= 0 || lastThumb >= b.inner-1 {
			t.Errorf("pane %d tab %d %dx%d: the thumb sits on lines %d to %d of %d, want it inside",
				p+1, tab, w, h, firstThumb, lastThumb, b.inner)
		}
	default:
		if lastThumb != b.inner-1 {
			t.Errorf("pane %d tab %d %dx%d: the thumb ends on line %d, want the last of %d",
				p+1, tab, w, h, lastThumb, b.inner)
		}
	}
}

// TestThumbOnlyWhenThereIsSomethingToScroll reads the plain fixture on a screen
// tall enough to hold all of it. Nothing overflows, so no pane draws a thumb.
func TestThumbOnlyWhenThereIsSomethingToScroll(t *testing.T) {
	for _, size := range [][2]int{{80, 30}, {160, 50}, {200, 120}} {
		m := press(sized(newModel(t), size[0], size[1]), keyOf(panePlans))
		m = press(m, keyOf(paneDetail))
		v := m.View()
		for p := pane(0); p < paneDetail; p++ {
			// A pane whose content fits has no window to point at, so it
			// draws no thumb. One that overflows has to draw one, or the
			// reader cannot tell how far down the pane is.
			overflows := m.linesOf(p) > m.fitOf(p)
			thumb := slices.Contains(wallCells(m, p), "┃")
			if thumb != overflows {
				t.Errorf("%dx%d pane %d: %d lines of content in %d rows draws a thumb: %v",
					size[0], size[1], p+1, m.linesOf(p), m.fitOf(p), thumb)
			}
		}
		if m.linesOf(paneDetail) <= m.fitOf(paneDetail) && slices.Contains(wallCells(m, paneDetail), "┃") {
			t.Errorf("%dx%d: the detail box draws a thumb for a body of %d lines in %d rows",
				size[0], size[1], m.linesOf(paneDetail), m.fitOf(paneDetail))
		}
		if strings.ContainsAny(plain(v), "░█") {
			t.Errorf("%dx%d: the view still draws a track or a block", size[0], size[1])
		}
	}
}
