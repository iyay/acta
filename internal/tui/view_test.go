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

func TestViewShowsTheThreePanes(t *testing.T) {
	m := sized(clocked(newModel(t), 20, 46), 200, 40)
	v := m.View()
	for _, want := range []string{
		"[1]─Specs", "─ Plans", "─ Tasks", "─ Bugs", "─ Debt",
		"[2]─Done", "─ Dropped",
		"[3]─Detail",
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
		{nil, "[1]─Specs", "[2]─"},
		{[]string{"2"}, "[2]─Done", "[1]─"},
		{[]string{"3"}, "[3]─Detail", "[1]─"},
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
	if p == paneDone {
		return g.done
	}
	return g.open
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

// isDivider says whether a drawn line is the rule between the two groups of a
// list pane, which is a full row of dashes and nothing else.
func isDivider(line string) bool {
	return line != "" && strings.Trim(line, "─") == ""
}

// TestEveryRowIsOneLineOnEveryTab walks every tab, both list panes and four
// widths, and reads what each pane actually draws. A pane has to show one row
// per line it has room for, and every screen line inside it has to be that
// pane's row at that position, the divider, or empty space below the last
// row, so a row can never take two lines.
func TestEveryRowIsOneLineOnEveryTab(t *testing.T) {
	for tab := range tabNames {
		for _, p := range []pane{paneOpen, paneDone} {
			for _, w := range []int{40, 60, 120, 200} {
				m := sized(focused(newModel(t), p, tab), w, 40)
				b := paneBox(m, p)
				inner := b.w - 2
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
				for i, ln := range lines {
					n := b.first + i
					if n >= len(rows) {
						if ln != "" {
							t.Errorf("tab %d pane %d width %d: line %d below the last row holds %q", tab, p, w, i, ln)
						}
						continue
					}
					if rows[n].divider {
						if want := strings.Repeat("─", inner); ln != want {
							t.Errorf("tab %d pane %d width %d: the divider line is %q", tab, p, w, ln)
						}
						continue
					}
					if !isRowLine(m, rows[n], ln, inner) {
						t.Errorf("tab %d pane %d width %d: line %d is %q, which is not row %s", tab, p, w, i, ln, rows[n].id)
					}
				}
			}
		}
	}
}

// TestTheDividerHidesWhenAGroupIsEmpty keeps only the stories of one kind of
// work, so the split between the two groups has nothing to separate and the
// rule has to go with it.
func TestTheDividerHidesWhenAGroupIsEmpty(t *testing.T) {
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
			for i, ln := range innerLines(m, paneBox(m, paneOpen)) {
				if isDivider(ln) {
					t.Errorf("line %d draws a divider with one group empty: %q", i, ln)
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
	for tab := range tabNames {
		for _, p := range []pane{paneOpen, paneDone} {
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
	m = sized(m, 200, 40)
	b := paneBox(m, paneOpen)
	lines := innerLines(m, b)
	if got := lines[0]; !strings.HasPrefix(got, "specs/2026-09-20-alpha  Alpha") || !strings.HasSuffix(got, " 2/5 · claude") {
		t.Errorf("the in-progress row is %q, want the name and the title then 2/5 · claude", got)
	}
	if got := lines[1]; got != strings.Repeat("─", b.w-2) {
		t.Errorf("the divider is %q", got)
	}
	if got := lines[2]; got != "specs/2026-09-22-beta  Beta story" {
		t.Errorf("the not-started row is %q, want only the name and the title", got)
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
	m = sized(m, 60, 40)
	b := paneBox(m, paneOpen)
	row := innerLines(m, b)[0]
	if !strings.HasSuffix(row, " 2/5 · claude") {
		t.Errorf("the count and the agent were cut: %q", row)
	}
	if !strings.Contains(row, "…") {
		t.Errorf("a 200-character title should be cut with …: %q", row)
	}
	if got := lipgloss.Width(row); got != b.w-2 {
		t.Errorf("the row is %d cells wide, the pane holds %d", got, b.w-2)
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
		// One j steps past the first row, so the in-progress row is drawn
		// unselected and shows its own brush.
		m := sized(press(newModel(t), "j"), 120, 40)
		b := paneBox(m, paneOpen)
		v := m.View()

		sel := paintedLine(v, b, 2)
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
		if reset < 0 || lipgloss.Width(plain(sel[:reset])) != b.w-2 {
			t.Errorf("the band stops short of the row width %d: %q", b.w-2, sel)
		}

		// Row 0 is in progress and not selected, so it keeps the accent and
		// is dim.
		going := paintedLine(v, b, 0)
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
	m = sized(openTab(t, m, tabDebt), 120, 40)

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
	m := sized(press(newModel(t), "]", "j"), 120, 40)
	g := m.geometry()
	detail := strings.Join(column(m.View(), g.detail.x, g.detail.w), "\n")
	if strings.Contains(detail, "SPEC") {
		t.Error("a plan with no spec should not show a SPEC line")
	}
	if strings.Contains(detail, "WORKTREE") || strings.Contains(detail, "AGENT") {
		t.Error("a line with no value should be left out")
	}
	// The plan above it does link a spec, so its line has to be there.
	m = sized(press(m, "k"), 120, 40)
	g = m.geometry()
	detail = strings.Join(column(m.View(), g.detail.x, g.detail.w), "\n")
	if !strings.Contains(detail, "SPEC") {
		t.Error("a plan with a spec should show its SPEC line")
	}
}

func TestViewDetailShowsTheSectionOfItsOwnItem(t *testing.T) {
	m := sized(press(newModel(t), "]", "]"), 120, 40)
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
	// With in-progress rows first, three j steps reach the weird story.
	m := sized(press(newModel(t), "j", "j", "j"), 120, 60)
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
	if !strings.Contains(v, "[1]─Specs") || !strings.Contains(v, "[3]─Detail") {
		t.Error("the popup should cover the panes, not replace them")
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
	m.focus, m.last, m.tab = f, paneOpen, tab
	m.doneTab = tab % len(m.doneTabNames())
	return m
}

// TestNoRoundedCorners walks every width from 1 to 200, every tab, and the
// focus on each pane, with no popup and with each of the three popups, so no
// frame the screen can draw still holds a rounded corner.
func TestNoRoundedCorners(t *testing.T) {
	for w := 1; w <= 200; w++ {
		for tab := range tabNames {
			for _, f := range []pane{paneOpen, paneDone, paneDetail} {
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
	if m.edge(paneOpen).GetForeground() != accentColor {
		t.Error("the pane with the focus should draw its border in the accent color")
	}
	if m.edge(paneDone).GetForeground() == accentColor {
		t.Error("a pane without the focus should not wear the accent color")
	}
	m = press(m, "2")
	if m.edge(paneDone).GetForeground() != accentColor {
		t.Error("the focus moved, so the accent color should move with it")
	}
	if m.edge(paneOpen).GetForeground() == accentColor {
		t.Error("the pane that lost the focus should be dim")
	}
}

func TestViewFinishedTabsFollowTheOpenTab(t *testing.T) {
	for _, tc := range []struct {
		keys []string
		want string
	}{
		{nil, "[2]─Done ─ Dropped"},
		{[]string{"]", "]", "2"}, "[2]─Done──"},
		{[]string{"]", "]", "]", "2"}, "[2]─Fixed ─ Wontfix"},
		{[]string{"]", "]", "]", "]", "2"}, "[2]─Done ─ Wontfix"},
	} {
		v := sized(press(newModel(t), tc.keys...), 120, 40).View()
		if !strings.Contains(v, tc.want) {
			t.Errorf("after %v pane [2] is missing %q", tc.keys, tc.want)
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
	m := sized(press(splitModel(t), "j", "j"), 200, 40)
	b := paneBox(m, paneOpen)
	above, below, seen := true, false, false
	for i, ln := range innerLines(m, b) {
		if isDivider(ln) {
			above = false
			continue
		}
		cell := paintedLine(m.View(), b, i)
		switch {
		case strings.Contains(ln, "specs/2026-09-20-alpha"):
			seen = true
			if !above {
				t.Errorf("the in-progress row sits below the divider: %q", ln)
			}
			if !wears(cell, 38, 5, 39) {
				t.Errorf("the in-progress row wears no accent color: %q", cell)
			}
		case strings.Contains(ln, "specs/2026-09-22-beta"):
			below = true
			if above {
				t.Errorf("the not-started row sits above the divider: %q", ln)
			}
			if wears(cell, 38, 5, 39) {
				t.Errorf("the not-started row wears the accent color: %q", cell)
			}
		}
	}
	if !seen || !below {
		t.Error("both rows should show")
	}
}

func TestViewDividerIsDim(t *testing.T) {
	withColors(func() {
		testViewDividerIsDim(t)
	})
}

func testViewDividerIsDim(t *testing.T) {
	m := sized(splitModel(t), 200, 40)
	b := paneBox(m, paneOpen)
	found := false
	for i, ln := range innerLines(m, b) {
		if !isDivider(ln) {
			continue
		}
		found = true
		if want := strings.Repeat("─", b.w-2); ln != want {
			t.Errorf("the divider is %d cells, the pane holds %d: %q", lipgloss.Width(ln), b.w-2, ln)
		}
		if !wears(paintedLine(m.View(), b, i), 2) {
			t.Errorf("the divider wears no dim brush: %q", paintedLine(m.View(), b, i))
		}
	}
	if !found {
		t.Error("no divider drawn")
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
