package tui

import (
	"fmt"
	"path/filepath"
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

func TestViewShowsTheTabBarAndTheDetail(t *testing.T) {
	t.Parallel()

	m := press(sized(clocked(newModel(t), 20, 46), 200, 40), tabKey(tabPlans))
	v := m.View()
	// A row is drawn as a plain row, an id in its kind color and a title
	// after it, so the words are read without the color codes in between.
	plainV := plain(v)
	for _, want := range []string{
		" 1 Scratches  2 Bugs  3 Debts  4 Specs  5 Plans  6 Activities",
		"─Open", "─Done ─ Dropped", "─Detail",
		"basic · live · 2026-09-27 20:46",
		"plans/2026-09-21-alpha  Alpha plan",
		"ID        : plans/2026-09-21-alpha",
		"FILE      : .acta/plans/2026-09-21-alpha.md",
		"Write the failing test",
	} {
		if !strings.Contains(plainV, want) {
			t.Errorf("view is missing %q", want)
		}
	}
	// The tab box holds the first three lines of the screen and the status
	// line the last, so the boxes live between them.
	lines := strings.Split(v, "\n")
	if !strings.Contains(plain(lines[1]), "5 Plans") {
		t.Errorf("the names line is %q, want the tab box with Plans open", plain(lines[1]))
	}
	for _, line := range lines {
		if w := lipgloss.Width(line); w > 200 {
			t.Fatalf("line wider than the terminal (%d): %q", w, line)
		}
	}
}

// namesIn gives the words a line of the tab box draws, the cells between its
// two walls.
func namesIn(line string) string {
	return strings.TrimSuffix(strings.TrimPrefix(plain(line), "│"), "│")
}

// The tab bar is the one place the reader sees which tab is open, so a narrow
// screen has to drop the names farthest from the open tab rather than cut the
// open name away, and it must never show half a name of another tab.
func TestTabBarKeepsTheOpenTab(t *testing.T) {
	t.Parallel()

	for i, tab := range topTabs {
		m := sized(press(newModel(t), tabKey(i)), 200, 40)
		open := fmt.Sprintf("%d %s", i+1, tab.name)
		for w := 2; w <= 200; w++ {
			inner := w - 2
			bar := namesIn(strings.Split(sized(m, w, 40).View(), "\n")[1])
			if inner >= len(open)+1 {
				if !strings.Contains(bar, open) {
					t.Fatalf("%s at %d columns: %q has no %q", tab.name, w, bar, open)
				}
			} else if want := " " + open; len(bar) != min(inner, len(want)) || !strings.HasPrefix(want, bar) {
				// Too narrow even for the name, so the line holds as much of
				// it as the box allows, and nothing else.
				t.Fatalf("%s at %d columns: %q, want the first %d cells of %q", tab.name, w, bar, min(inner, len(want)), want)
			}
			// A cut name reads as a word of its own, so every name on the
			// line has to carry the number that opens it and be a whole one.
			for j, other := range topTabs {
				name := fmt.Sprintf("%d %s", j+1, other.name)
				if strings.Contains(bar, other.name) && !strings.Contains(bar, name) {
					t.Fatalf("%s at %d columns: %q shows %q without the number in front of it, so it is a piece of a name", tab.name, w, bar, other.name)
				}
			}
		}
	}
}

// The bar has to say which names go missing first: the one that sits farthest
// from the open tab, and on a tie the right one, so a name the reader is
// working in always outlives the ones around it.
func TestTabBarDropsTheNameFarthestFromTheOpenTab(t *testing.T) {
	t.Parallel()

	for i, tab := range topTabs {
		m := sized(press(newModel(t), tabKey(i)), 200, 40)
		for w := 2; w <= 200; w++ {
			bar := namesIn(strings.Split(sized(m, w, 40).View(), "\n")[1])
			drop := make([]bool, len(topTabs))
			for widthOfBar(drop) > w-2 {
				// dropOrder walks the tabs from the right, skipping the open
				// one, so the first match at the widest distance is the one
				// on the right, which is the one a tie has to drop.
				farthest, at := -1, -1
				for _, j := range dropOrder(len(topTabs), i) {
					if drop[j] {
						continue
					}
					if d := max(j-i, i-j); d > farthest {
						farthest, at = d, j
					}
				}
				if at < 0 {
					break
				}
				drop[at] = true
			}
			for j, other := range topTabs {
				if j == i {
					continue
				}
				if got := strings.Contains(bar, other.name); got != !drop[j] {
					t.Fatalf("%s at %d columns: %q holds %q = %t, want %t", tab.name, w, bar, other.name, got, !drop[j])
				}
			}
		}
	}
}

// widthOfBar counts the cells the names line takes when the tabs in drop are
// missing, the same way the view counts them.
func widthOfBar(drop []bool) int {
	w := 1
	for i, tab := range topTabs {
		if drop[i] {
			continue
		}
		if w > 1 {
			w += 2
		}
		// Every name carries the number of the key that opens it, so the
		// numbers take cells of their own.
		w += len(strconv.Itoa(i+1)) + 1 + len(tab.name)
	}
	return w
}

// closedBox says what is wrong with the top three lines of a screen, so every
// test that reads the tab box reads it the same way. Nothing wrong comes back
// as no error: one box, closed on both sides, as wide as the screen.
func closedBox(lines []string, w int) error {
	if len(lines) < barRows {
		return fmt.Errorf("the screen has %d lines, want a box of %d", len(lines), barRows)
	}
	// The two border lines hold nothing but their own dashes, so a box that
	// something else was drawn over shows it. The names line in between holds
	// the tab names, so only its walls are read.
	for k, corners := range [][3]string{{"┌", "─", "┐"}, {"│", "", "│"}, {"└", "─", "┘"}} {
		line := plain(lines[k])
		if got := lipgloss.Width(line); got != w {
			return fmt.Errorf("line %d is %d cells wide, want %d", k, got, w)
		}
		if !strings.HasPrefix(line, corners[0]) || !strings.HasSuffix(line, corners[2]) {
			return fmt.Errorf("line %d is %q, want a box of %d cells", k, line, w)
		}
		if corners[1] != "" && strings.Trim(line, corners[0]+corners[1]+corners[2]) != "" {
			return fmt.Errorf("line %d is %q, want nothing between its corners but dashes", k, line)
		}
	}
	return nil
}

// TestTabsSitInABoxOfTheirOwn reads the top three lines at several widths
// and on every tab, so no width and no open tab can break the box.
func TestTabsSitInABoxOfTheirOwn(t *testing.T) {
	t.Parallel()

	for _, w := range []int{40, 60, 120, 200} {
		for i := range topTabs {
			m := press(sized(newModel(t), w, 30), tabKey(i))
			lines := strings.Split(m.View(), "\n")
			if err := closedBox(lines, w); err != nil {
				t.Errorf("w=%d tab %d: %v", w, i, err)
			}
			mid := plain(lines[1])
			if strings.ContainsAny(mid, "[]") {
				t.Errorf("w=%d tab %d: names still wear brackets %q", w, i, mid)
			}
			if open := fmt.Sprintf("%d %s", i+1, topTabs[i].name); !strings.Contains(mid, open) {
				t.Errorf("w=%d: the open tab %q is missing from %q", w, open, mid)
			}
			if w >= 120 {
				for j, tab := range topTabs {
					if name := fmt.Sprintf("%d %s", j+1, tab.name); !strings.Contains(mid, name) {
						t.Errorf("w=%d: %q is missing from %q", w, name, mid)
					}
				}
			}
			g := m.geometry()
			first := g.full
			if g.wide {
				first = g.side[paneList]
			}
			if first.y != barRows {
				t.Errorf("w=%d tab %d: the first pane starts on line %d, want %d", w, i, first.y, barRows)
			}
		}
	}
}

// TestTheOpenTabIsTheOnlyBrightName checks the brush of every name, so the
// band can never sit on two tabs or on the wrong one, and so every other name
// is plain text in the color of its own kind.
func TestTheOpenTabIsTheOnlyBrightName(t *testing.T) {
	withColors(func() {
		for i := range topTabs {
			m := press(sized(newModel(t), 160, 40), tabKey(i))
			mid := strings.Split(m.View(), "\n")[1]
			for j, tab := range topTabs {
				name := fmt.Sprintf("%d %s", j+1, tab.name)
				color := m.styles.tabColor(tab.kind)
				band := lipgloss.NewStyle().Bold(true).Foreground(m.styles.bandFG).Background(color)
				if on := strings.Contains(mid, band.Render(name)); on != (j == i) {
					t.Errorf("tab %d open: %q in the band = %v", i, tab.name, on)
				}
				if j != i && !strings.Contains(mid, lipgloss.NewStyle().Foreground(color).Render(name)) {
					t.Errorf("tab %d open: %q is not plain text in its kind color", i, tab.name)
				}
			}
		}
	})
}

// TestTabsWearTheirKindColors reads every tab name in both themes with each
// tab open, so no tab can go faint or lose the color of its kind.
func TestTabsWearTheirKindColors(t *testing.T) {
	withTrueColor(func() {
		for _, name := range []string{"tokyo-night", "terminal"} {
			for open := range topTabs {
				m := press(actModel(t).WithTheme(name, true), tabKey(open))
				line := m.barLine(make([]bool, len(topTabs)))
				for i, tb := range topTabs {
					label := fmt.Sprintf("%d %s", i+1, tb.name)
					want := lipgloss.NewStyle().Foreground(m.styles.tabColor(tb.kind)).Render(label)
					if i == open {
						want = lipgloss.NewStyle().Bold(true).Foreground(m.styles.bandFG).
							Background(m.styles.tabColor(tb.kind)).Render(label)
					}
					if !strings.Contains(line, want) {
						t.Errorf("%s, open %d: tab %q is not drawn as %q in %q", name, open, label, want, line)
					}
				}
				if sgrHas(line, "2") {
					t.Errorf("%s, open %d: a tab is faint: %q", name, open, line)
				}
			}
		}
	})
}

// TestStatusLineColorsProjectAndLive reads the words that tell something at a
// glance: the project wears the accent, live wears green, and paused does not.
func TestStatusLineColorsProjectAndLive(t *testing.T) {
	withTrueColor(func() {
		m := sized(actModel(t).WithTheme("tokyo-night", true), 200, 40)
		line := m.statusLine()
		project := filepath.Base(m.cfg.RepoRoot)
		if !strings.Contains(line, m.styles.accent.Render(project)) {
			t.Errorf("project %q is not in the accent: %q", project, plain(line))
		}
		if !strings.Contains(line, m.styles.live.Render("live")) {
			t.Errorf("live is not green: %q", plain(line))
		}
		m.manual = true
		if strings.Contains(m.statusLine(), m.styles.live.Render("paused")) {
			t.Error("paused is green")
		}
	})
}

// The box has to close and hold its width on every screen the terminal can
// report. A width too narrow for the two walls has no box to close, so every
// line is a plain run of dashes, and the panes below still start under three
// drawn lines.
func TestTheTabBoxHoldsItsWidthOnEveryScreen(t *testing.T) {
	t.Parallel()

	// One subtest per width, so the sizes spread over every core instead of
	// one long loop holding the whole package up.
	for _, w := range sweep(0, 200, []int{0, 1, 2, 3, 5}) {
		t.Run(fmt.Sprintf("w%d", w), func(t *testing.T) {
			t.Parallel()
			for i := range topTabs {
				m := press(sized(newModel(t), 200, 30), tabKey(i))
				rows := m.tabRows(w)
				if len(rows) != barRows {
					t.Fatalf("w=%d tab %d: %d rows, want %d", w, i, len(rows), barRows)
				}
				for k, row := range rows {
					if got := lipgloss.Width(row); got != w {
						t.Errorf("w=%d tab %d row %d: %d cells wide, %q", w, i, k, got, plain(row))
					}
					if w < 2 && strings.Trim(plain(row), "─") != "" {
						t.Errorf("w=%d tab %d row %d: %q, want only dashes, there is no room for a corner", w, i, k, plain(row))
					}
				}
				if w >= 2 {
					if err := closedBox(rows, w); err != nil {
						t.Errorf("w=%d tab %d: %v", w, i, err)
					}
				}
			}
		})
	}
}

// Edge sizes of the layout. Under -short the size sweeps read only these,
// so a daily run is fast. The full run still walks every size.
var (
	edgeWidths  = []int{1, 2, 5, 6, 7, 29, 30, 31, 59, 60, 61, 93, 94, 159, 160, 161, 200}
	edgeHeights = []int{10, 11, 24, 40, 60}
)

func sweep(lo, hi int, short []int) []int {
	var out []int
	if testing.Short() {
		for _, v := range short {
			if v >= lo && v <= hi {
				out = append(out, v)
			}
		}
		return out
	}
	for v := lo; v <= hi; v++ {
		out = append(out, v)
	}
	return out
}

// These two check the size helper itself, because the slow tests above it
// would quietly walk too few sizes if the helper picked the wrong ones. A full
// run has to walk every size, and a short run only the edge sizes above.
func TestSweepFullRangeWithoutShort(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("checks the full range")
	}
	got := sweep(3, 6, []int{4})
	if !slices.Equal(got, []int{3, 4, 5, 6}) {
		t.Fatalf("sweep(3, 6) = %v, want every value", got)
	}
}

func TestSweepSamplesInsideRangeUnderShort(t *testing.T) {
	t.Parallel()

	if !testing.Short() {
		t.Skip("checks the short list")
	}
	got := sweep(3, 6, []int{1, 4, 6, 9})
	if !slices.Equal(got, []int{4, 6}) {
		t.Fatalf("sweep(3, 6) = %v, want only listed values inside the range", got)
	}
}

func TestViewNeverOverflowsAnyWindow(t *testing.T) {
	t.Parallel()

	// One subtest per width, so the sizes spread over every core instead of
	// one long loop holding the whole package up.
	for _, w := range sweep(30, 200, edgeWidths) {
		t.Run(fmt.Sprintf("w%d", w), func(t *testing.T) {
			t.Parallel()
			for _, h := range sweep(10, 60, edgeHeights) {
				base := sized(newModel(t), w, h)
				for name, m := range map[string]Model{
					"open":   base,
					"detail": press(base, tabKey(tabPlans)),
					"help":   press(base, "?"),
				} {
					for _, line := range strings.Split(sized(m, w, h).View(), "\n") {
						if got := lipgloss.Width(line); got > w {
							t.Fatalf("%dx%d %s: line is %d cells wide, the window is %d: %q", w, h, name, got, w, line)
						}
					}
				}
			}
		})
	}
}

func TestViewNarrowShowsOnlyTheFocusedPane(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		keys []string
		want string
		gone string
	}{
		{nil, "─Tasks", "─Done"},
		{[]string{tabKey(tabSpecs)}, "─Open", "─Done"},
		{[]string{tabKey(tabSpecs), "tab"}, "─Done ─ Dropped", "─Open"},
		{[]string{tabKey(tabSpecs), "shift+tab"}, "─Detail", "─Open"},
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
		head = strings.Repeat("  ", r.depth)
		if r.tree {
			head += m.treeMark(r, it) + " "
		}
		head += name + "  " + it.Title
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
	t.Parallel()

	for i := range topTabs {
		for _, p := range press(newModel(t), tabKey(i)).panes() {
			for _, w := range []int{40, 60, 120, 200} {
				m := sized(press(newModel(t), tabKey(i)), w, 40)
				m.focusPane(p)
				tab := topTabs[i].name
				b := paneBox(m, p)
				inner := b.textW()
				rows, _, _ := m.slotOf(p)
				lines := innerLines(m, b)
				if len(lines) < b.rows {
					t.Fatalf("tab %s pane %d width %d: the pane drew %d of its %d rows", tab, p, w, len(lines), b.rows)
				}
				if b.inner > 0 && b.rows != b.inner {
					t.Fatalf("tab %s pane %d width %d: the pane shows %d rows in %d lines, want one row per line", tab, p, w, b.rows, b.inner)
				}
				if len(rows) == 0 {
					if lines[0] != "nothing here" {
						t.Errorf("tab %s pane %d width %d: an empty pane shows %q", tab, p, w, lines[0])
					}
					continue
				}
				// The words of a box fill every cell between the walls, so a
				// row runs right up to the scrollbar on the right wall.
				for n, ln := range lines {
					row := b.first + n
					if row >= len(rows) {
						if strings.TrimSpace(ln) != "" {
							t.Errorf("tab %s pane %d width %d: line %d below the last row holds %q", tab, p, w, n, ln)
						}
						continue
					}
					if !isRowLine(m, rows[row], cells(ln, inner), inner) {
						t.Errorf("tab %s pane %d width %d: line %d is %q, which is not row %s", tab, p, w, n, ln, rows[row].id)
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
	t.Parallel()

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
			m := press(sized(full, 200, 40), tabKey(tabSpecs))
			m.board = &board.Board{Items: items}
			for i, ln := range innerLines(m, paneBox(m, paneList)) {
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
	t.Parallel()

	for i := range topTabs {
		for _, p := range press(newModel(t), tabKey(i)).panes() {
			m := sized(press(newModel(t), tabKey(i)), 200, 40)
			m.focusPane(p)
			for n, line := range innerLines(m, paneBox(m, p)) {
				for _, w := range statusWords {
					if saysWord(line, w) && !boardSays(m, w) {
						t.Errorf("tab %s pane %d line %d paints the status %q: %q", topTabs[i].name, p, n, w, line)
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
	t.Parallel()

	m := newModel(t)
	it := m.board.Get("specs/2026-09-20-alpha")
	it.Done, it.Total, it.Agent = 2, 5, "claude"
	m = press(sized(m, 200, 40), tabKey(tabSpecs))
	b := paneBox(m, paneList)
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
	t.Parallel()

	m := newModel(t)
	it := m.board.Get("specs/2026-09-20-alpha")
	it.Done, it.Total, it.Agent = 2, 5, "claude"
	it.Title = strings.Repeat("x", 200)
	m = press(sized(m, 60, 40), tabKey(tabSpecs))
	b := paneBox(m, paneList)
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
// want, which is how a test names a brush. A 24-bit color is skipped: its
// middle number is the word "truecolor", not a brush anyone asked for.
func wears(line string, want ...int) bool {
	for _, ps := range sgrParams(line) {
		if truecolor(ps) {
			continue
		}
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

// truecolor says whether a color code carries 24-bit color: 38 or 48, then
// the number 2, then red, green and blue.
func truecolor(ps []int) bool {
	return len(ps) == 5 && (ps[0] == 38 || ps[0] == 48) && ps[1] == 2
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
		m := sized(press(newModel(t), tabKey(tabSpecs), "j"), 120, 40)
		b := paneBox(m, paneList)
		v := m.View()

		sel := paintedLine(v, b, 1)
		if wears(sel, 7) {
			t.Errorf("the selected row is reverse video: %q", sel)
		}
		if wears(sel, 2) {
			t.Errorf("the selected row is faint: %q", sel)
		}
		if !wears(sel, 48, 5, 23) {
			t.Errorf("the selected row has no background 23: %q", sel)
		}
		reset := strings.LastIndex(sel, "\x1b[0m")
		if want := b.textW(); reset < 0 || lipgloss.Width(plain(sel[:reset])) != want {
			t.Errorf("the band stops short of the row width %d: %q", want, sel)
		}

		// The in-progress row is not selected, so it keeps the accent and
		// is not faint, so a row under way reads at full brightness.
		going := paintedLine(v, b, 2)
		if wears(going, 2) || !wears(going, 38, 5, 111) {
			t.Errorf("an unselected in-progress row should keep the accent without faint: %q", going)
		}
		if wears(going, 48, 5, 23) {
			t.Errorf("an unselected row wears a background: %q", going)
		}
		for _, i := range []int{3, 4} {
			row := paintedLine(v, b, i)
			if sgrHas(row, "2") {
				t.Errorf("row %d is faint: %q", i, plain(row))
			}
			if wears(row, 48, 5, 23) {
				t.Errorf("row %d wears a background: %q", i, row)
			}
		}
	})
}

func TestViewDetailHeaderAlignsItsColons(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	// The Debts tab of the bar lists the debt items of the board.
	m = sized(press(m, tabKey(tabDebts)), 120, 40)

	v := m.View()
	// The tab bar names every tab in order, so Debts comes after Bugs and
	// before the tab that is open.
	if i, j := strings.Index(v, "Bugs"), strings.Index(v, "Debts"); i < 0 || j < i {
		t.Fatalf("Debts should come after Bugs in the tab bar: %q", v)
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
	for _, want := range []string{`DEBT\s+: a$`, `STATUS\s+: open$`, `FROM\s+: PLN-0003 . Short IDs$`} {
		if ok, _ := regexp.MatchString("(?m)"+want, detail); !ok {
			t.Errorf("detail header is missing %q, got:\n%s", want, detail)
		}
	}

}

func TestViewDetailLeavesEmptyLinesOut(t *testing.T) {
	t.Parallel()

	// Oldest first, so the first plan of the list links a spec, and its SPEC
	// line is drawn.
	m := sized(press(newModel(t), tabKey(tabPlans)), 120, 40)
	g := m.geometry()
	detail := strings.Join(column(m.View(), g.detail.x, g.detail.w), "\n")
	if !strings.Contains(detail, "SPEC") {
		t.Error("a plan with a spec should show its SPEC line")
	}
	if strings.Contains(detail, "WORKTREE") || strings.Contains(detail, "AGENT") {
		t.Error("a line with no value should be left out")
	}
	// The plan below it links no spec, so no SPEC line is drawn.
	m = sized(press(m, "j"), 120, 40)
	g = m.geometry()
	detail = strings.Join(column(m.View(), g.detail.x, g.detail.w), "\n")
	if strings.Contains(detail, "SPEC") {
		t.Error("a plan with no spec should not show a SPEC line")
	}
}

func TestViewDetailShowsTheSectionOfItsOwnItem(t *testing.T) {
	t.Parallel()

	// Activities lists the one task of the fixture that is under way under the
	// plan it belongs to, so one j step reaches the task and the detail of a
	// task holds its own section and nothing else.
	m := sized(press(newModel(t), tabKey(tabActivities), "j"), 120, 40)
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
	t.Parallel()

	// The rows run oldest file date first, so one j step reaches the weird
	// story, the one with a bad status.
	m := sized(press(newModel(t), tabKey(tabSpecs), "j"), 120, 60)
	if !strings.Contains(plain(m.View()), "! ") {
		t.Error("the problems of the selected item are missing")
	}
	if !strings.Contains(plain(m.View()), "untyped (1)") {
		t.Error("the row that holds the files outside .acta/ is missing")
	}
}

func TestViewStatusLineShowsHelpAndClock(t *testing.T) {
	t.Parallel()

	m := sized(clocked(newModel(t), 20, 46), 100, 30)
	last := lastLine(m.View())
	if got := lipgloss.Width(last); got != 100 {
		t.Errorf("the status line is %d cells wide, want 100", got)
	}
	if !strings.HasPrefix(plain(last), "? help") {
		t.Errorf("the left should show only ? help, got %q", plain(last))
	}
	if !strings.Contains(plain(last), "basic · live · 2026-09-27 20:46") {
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
	if !strings.Contains(plain(m.View()), "basic · paused · 2026-09-27 20:46") {
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	m := sized(press(clocked(newModel(t), 20, 46), "?"), 100, 30)
	v := m.View()
	if !strings.Contains(v, "Keys") || !strings.Contains(v, "new bug") {
		t.Error("the help popup is missing")
	}
	if !strings.Contains(v, "─Tasks") || !strings.Contains(v, "─Detail") {
		t.Error("the popup should cover the boxes, not replace them")
	}
	if !strings.HasSuffix(plain(lastLine(v)), "2026-09-27 20:46 | Feedback  dev") {
		t.Error("the popup should not hide the status line")
	}
	// The help names every key of the new layout.
	help := plain(v)
	for _, want := range []string{"1-6", "←", "→", "tab shift+tab", "[ ]", "space enter"} {
		if !strings.Contains(help, want) {
			t.Errorf("the help shows no %q:\n%s", want, help)
		}
	}
	if strings.Contains(press(m, "?").View(), "Keys") {
		t.Error("? should close the help popup")
	}
}

// TestNoRoundedCorners walks every width from 1 to 200, every tab of the bar,
// and the focus on each of its panes, with no popup and with each of the three
// popups, so no frame the screen can draw still holds a rounded corner.
func TestNoRoundedCorners(t *testing.T) {
	t.Parallel()

	// The list of boxes of a tab does not read the width, so it is built once
	// per tab instead of once per width.
	focuses := make([][]pane, len(topTabs))
	for i := range topTabs {
		focuses[i] = append(press(newModel(t), tabKey(i)).panes(), paneDetail)
	}
	// One subtest per width, so the sizes spread over every core instead of
	// one long loop holding the whole package up.
	for _, w := range sweep(1, 200, edgeWidths) {
		t.Run(fmt.Sprintf("w%d", w), func(t *testing.T) {
			t.Parallel()
			for i := range topTabs {
				for _, f := range focuses[i] {
					m := press(sized(newModel(t), w, 40), tabKey(i))
					m.focusPane(f)
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
							t.Fatalf("width %d, tab %s, focus %d, %s: rounded corner in the frame", w, topTabs[i].name, f, frame.what)
						}
					}
				}
			}
		})
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

// A popup sits over the panes, so the tab box above them stays whole whenever
// the screen is tall enough to hold the popup whole. A screen too short for
// that has to choose, and it keeps the tab box, because half a box over the
// tab bar helps nobody.
func TestAPopupNeverCoversTheTabBox(t *testing.T) {
	t.Parallel()

	for _, w := range []int{60, 120} {
		for h := 4; h <= 24; h++ {
			for _, open := range []string{"?", "t", "s", "n"} {
				pop := press(sized(newModel(t), w, h), open)
				lines := strings.Split(plain(pop.View()), "\n")
				box := len(strings.Split(pop.popupBox(), "\n"))
				if err := closedBox(lines, w); err != nil && box > h {
					t.Errorf("%dx%d popup %q: %v", w, h, open, err)
				}
			}
		}
	}
}

// A popup has to be whole on screen: a box the screen is tall enough for has
// its top border, its bottom border and every row between them drawn, and the
// last row may sit on the last screen line when the box is as tall as the
// screen. Every popup the keys open is read at every height from its own row
// count up to 50, at the three widths the plan names.
func TestAPopupIsWholeOnTheScreen(t *testing.T) {
	t.Parallel()

	for _, w := range []int{50, 80, 160} {
		for _, open := range []string{"?", "t", "s", "n"} {
			rows := plainLines(strings.Split(press(sized(newModel(t), w, 50), open).popupBox(), "\n"))
			if len(rows) < 2 {
				t.Fatalf("width %d, popup %q drew no box", w, open)
			}
			for h := len(rows); h <= 50; h++ {
				lines := plainLines(strings.Split(press(sized(newModel(t), w, h), open).View(), "\n"))
				y0 := boxOn(lines, rows[0])
				if y0 < 0 {
					t.Errorf("%dx%d popup %q: the top border of the box is not on the screen", w, h, open)
					continue
				}
				for i, r := range rows {
					if y0+i < len(lines) && strings.Contains(lines[y0+i], r) {
						continue
					}
					t.Errorf("%dx%d popup %q: row %d of the box, %q, is not on the screen", w, h, open, i, r)
				}
			}
		}
	}
}

// boxOn is the screen line that holds this row of a popup box, or -1 when no
// line does. A test finds the box on the screen instead of asking the model
// where it meant to put it, so a box the model pushed off the screen is caught.
func boxOn(lines []string, row string) int {
	for i, ln := range lines {
		if strings.Contains(ln, row) {
			return i
		}
	}
	return -1
}

// TestPopupDimsTheBackground opens every popup the keys can open, at both
// sizes, under both color profiles. Every cell outside the box is the dim
// style over its own plain text, every line of the box is the box the model
// draws, and esc gives back the exact screen that was there before.
func TestPopupDimsTheBackground(t *testing.T) {
	// The brush the view paints the screen behind a popup with, spelled out
	// here so this test checks the color the plan names and not the one the
	// view happens to use today: tokyo-night slot 8, never faint.
	dim := func() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("#414868")) }
	withColors(func() { checkPopupDim(t, dim()) })
	withTrueColor(func() { checkPopupDim(t, dim()) })
}

// checkPopupDim opens every popup at both sizes and checks that every cell
// behind the box is the dim brush over its own plain text, and that esc gives
// back the exact screen from before.
func checkPopupDim(t *testing.T, dim lipgloss.Style) {
	t.Helper()
	for _, open := range []string{"?", "t", "s", "n"} {
		for _, size := range [][2]int{{80, 30}, {160, 50}} {
			m := press(sized(newModel(t), size[0], size[1]), tabKey(tabBugs))
			before := m.View()
			pop := press(m, open)
			rows := strings.Split(pop.popupBox(), "\n")
			x0, y0, w, h := popupRect(rows, pop.width, pop.height)
			if w == 0 {
				t.Errorf("popup %q at %dx%d never opened", open, size[0], size[1])
				continue
			}
			// The theme background is under every line and after every
			// reset. Take it off again, because what this test reads is
			// the brush on the text and the frame has its own test.
			frame := strings.TrimSuffix(pop.styles.paintFrame(""), "\x1b[K")
			after := strings.Split(pop.View(), "\n")
			for y, drawn := range after {
				ln := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(drawn, frame), "\x1b[K"), frame, "")
				parts := []string{ln}
				if y >= y0 && y < y0+h && y < len(after)-1 {
					i := strings.Index(ln, rows[y-y0])
					if i < 0 {
						t.Errorf("popup %q at %dx%d line %d: the box row is not drawn the way popupBox draws it", open, size[0], size[1], y)
						continue
					}
					head, tail := ln[:i], ln[i+len(rows[y-y0]):]
					if lipgloss.Width(head) != x0 || lipgloss.Width(rows[y-y0]) != w {
						t.Errorf("popup %q at %dx%d line %d: the box sits at column %d and is %d cells wide, want %d and %d", open, size[0], size[1], y, lipgloss.Width(head), lipgloss.Width(rows[y-y0]), x0, w)
					}
					parts = []string{head, tail}
				}
				for _, seg := range parts {
					if got, want := seg, dim.Render(plain(seg)); got != want {
						t.Errorf("popup %q at %dx%d line %d: the background keeps its own color\n got %q\nwant %q", open, size[0], size[1], y, got, want)
					}
				}
			}
			if closed := press(pop, "esc").View(); closed != before {
				t.Errorf("popup %q at %dx%d: the view is not byte-equal to the one before it opened", open, size[0], size[1])
			}
		}
	}
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
	t.Parallel()

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
	t.Parallel()

	m := sized(press(newModel(t), tabKey(tabBugs), "tab", "s"), 100, 30)
	if !strings.Contains(m.View(), "wontfix") {
		t.Error("the value popup options are not shown")
	}
	m = sized(press(newModel(t), "n", "a", "b"), 100, 30)
	if !strings.Contains(m.View(), "new bug slug: ab") {
		t.Error("the slug prompt is not shown")
	}
}

func TestViewFocusedPaneWearsTheAccent(t *testing.T) {
	t.Parallel()

	// A tab with both boxes, so the accent really moves from one to the other.
	m := press(newModel(t), tabKey(tabBugs))
	accentColor := m.styles.accentColor
	if m.edge(paneList).GetForeground() != accentColor {
		t.Error("the box with the focus should draw its border in the accent color")
	}
	if m.edge(paneDone).GetForeground() == accentColor {
		t.Error("a box without the focus should not wear the accent color")
	}
	m = press(m, "tab")
	if m.edge(paneDone).GetForeground() != accentColor {
		t.Error("the focus moved, so the accent color should move with it")
	}
	if m.edge(paneList).GetForeground() == accentColor {
		t.Error("the box that lost the focus should be dim")
	}
}

// TestViewDonePanesShowTheirOwnSubTabs reads the title of the Done pane of
// every tab of the bar, so no tab can ever draw the sub-tabs of another.
func TestViewDonePanesShowTheirOwnSubTabs(t *testing.T) {
	t.Parallel()

	for i := range topTabs {
		if len(topTabs[i].done) == 0 {
			continue
		}
		var names []string
		for _, sub := range topTabs[i].done {
			names = append(names, sub.name)
		}
		title := "─" + strings.Join(names, " ─ ")
		for d := range topTabs[i].done {
			keys := []string{tabKey(i), "tab"}
			if d > 0 {
				keys = append(keys, "]")
			}
			v := sized(press(newModel(t), keys...), 120, 40).View()
			if !strings.Contains(v, title) {
				t.Errorf("tab %s is missing the Done title %q", topTabs[i].name, title)
			}
			if got := strings.Join(press(press(newModel(t), keys...), "j").tabsOf(paneDone), " "); got != strings.Join(names, " ") {
				t.Errorf("tab %s draws the Done sub-tabs %q, want %q", topTabs[i].name, got, strings.Join(names, " "))
			}
		}
	}
}

func TestViewEmptyRepo(t *testing.T) {
	t.Parallel()

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
	m := sized(press(splitModel(t), tabKey(tabSpecs), "j"), 200, 40)
	b := paneBox(m, paneList)
	seenGoing, seenWaiting := false, false
	for i, ln := range innerLines(m, b) {
		cell := paintedLine(m.View(), b, i)
		switch {
		case strings.Contains(ln, "specs/2026-09-20-alpha"):
			seenGoing = true
			if !wears(cell, 38, 5, 111) {
				t.Errorf("the in-progress row wears no accent color: %q", cell)
			}
		case strings.Contains(ln, "specs/2026-09-22-beta"):
			seenWaiting = true
			if wears(cell, 38, 5, 111) {
				t.Errorf("the not-started row wears the accent color: %q", cell)
			}
		}
	}
	if !seenGoing || !seenWaiting {
		t.Error("both rows should show")
	}
}

func TestViewHasNoSectionLabels(t *testing.T) {
	t.Parallel()

	m := sized(splitModel(t), 200, 40)
	for _, ln := range strings.Split(plain(m.View()), "\n") {
		low := strings.ToLower(body(ln))
		if strings.HasPrefix(low, "in progress") || strings.HasPrefix(low, "not started") {
			t.Errorf("a section label leaked into the list: %q", ln)
		}
	}
}

func TestViewHelpListsTheNewKeys(t *testing.T) {
	t.Parallel()

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
// the focused box the room, z takes it back, moving to another box takes it
// back, and z on the detail box does nothing at all.
func TestZTogglesAndFocusRestores(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 40)
	if m.expanded != -1 {
		t.Fatalf("a screen with nothing expanded reads %d", m.expanded)
	}
	// The List box of the Plans tab has the focus, z expands it and z puts
	// the room back.
	m = press(m, tabKey(tabPlans), "z")
	if m.expanded != int(paneList) {
		t.Fatalf("z on the List box expanded %d, want %d", m.expanded, paneList)
	}
	m = press(m, "z")
	if m.expanded != -1 {
		t.Fatalf("a second z should give the room back, it reads %d", m.expanded)
	}
	// A Done sub-tab of the same box is not another box, so the room stays.
	m = press(sized(newModel(t), 120, 40), tabKey(tabPlans), "tab", "z", "]")
	if m.expanded != int(paneDone) {
		t.Fatalf("a sub-tab change should keep the box expanded, it reads %d", m.expanded)
	}
	// Moving the focus to another box gives the room back, by key and by tab.
	for _, keys := range [][]string{{tabKey(tabBugs)}, {"tab"}, {"shift+tab"}, {tabKey(tabScratches)}} {
		m = press(sized(newModel(t), 120, 40), append([]string{tabKey(tabPlans), "z"}, keys...)...)
		if m.expanded != -1 {
			t.Errorf("%v should give the room back, it reads %d", keys, m.expanded)
		}
	}
	// The detail box has no room to give, so z there does nothing.
	m = press(sized(newModel(t), 120, 40), tabKey(tabPlans), "z", "shift+tab", "z")
	if m.expanded != -1 {
		t.Fatalf("z on the detail box expanded %d, want nothing", m.expanded)
	}
	// The box that has the room is the one that grew, and the other keeps 3
	// lines on a screen tall enough to give it.
	wide := press(sized(newModel(t), 120, 40), tabKey(tabPlans), "z")
	hs := wide.leftHeights(wide.height - 2)
	if hs[paneList] <= 3 {
		t.Errorf("the expanded box has %d lines, want more than the 3 of the other: %v", hs[paneList], hs)
	}
	for p, h := range hs {
		if p != int(paneList) && h != 3 {
			t.Errorf("box %d has %d lines next to the expanded one, want 3", p+1, h)
		}
	}
}

// TestCounterShowsSelectedItemNotLine reads what every box writes in its
// bottom border: the item under the cursor out of the items the box holds,
// never the line on screen, and nothing at all on the detail box.
func TestCounterShowsSelectedItemNotLine(t *testing.T) {
	withColors(func() {
		// Forty open plans, more than the box can show. The cursor walks down
		// while the list stays at its top, so a counter reading the line on
		// screen would keep saying 1.
		m := press(longModel(t), tabKey(tabPlans))
		for i, want := range []string{"1 of 40", "2 of 40", "3 of 40"} {
			if i > 0 {
				m = press(m, "j")
			}
			if m.off[paneList] != 0 {
				t.Fatalf("the list scrolled to line %d, the counter cannot be a line number", m.off[paneList])
			}
			if got := footOf(t, m, paneList); got != want {
				t.Errorf("the List box with item %d writes %q, want %q", i+1, got, want)
			}
		}
		// The counter sits at the right end of the border, the way lazygit
		// puts it, with the corner as the last thing the line holds.
		if line := bottomLine(t, m, paneList); !strings.HasSuffix(line, " 3 of 40 ┘") {
			t.Errorf("the bottom border of the List box reads %q, want the counter at its right end", line)
		}
		// Every box of every tab of the plain fixture writes the item under
		// the cursor out of the items it holds.
		for i := range topTabs {
			for _, p := range press(newModel(t), tabKey(i)).panes() {
				// A screen of its own for every box, because the keys write
				// into the cursors the model shares.
				one := press(sized(newModel(t), 160, 50), tabKey(i))
				one.focusPane(p)
				rows, _, _ := one.slotOf(p)
				if len(rows) == 0 {
					continue
				}
				stepped := press(one, "j")
				rows, sel, idx := stepped.slotOf(p)
				want := itemCount(cursorOf(rows, *sel, *idx)+1, len(rows))
				if got := footOf(t, stepped, p); got != want {
					t.Errorf("tab %s box %d writes %q, want %q", topTabs[i].name, p+1, got, want)
				}
			}
		}
		// A search that finds nothing empties every box, and an empty box
		// still counts, as 0 of 0.
		empty := press(sized(newModel(t), 160, 50), tabKey(tabPlans))
		empty.query = "nothing in the board matches this"
		if rows, _, _ := empty.slotOf(paneList); len(rows) != 0 {
			t.Fatalf("the search left %d rows in the List box", len(rows))
		}
		if got := footOf(t, empty, paneList); got != "0 of 0" {
			t.Errorf("an empty box writes %q, want %q", got, "0 of 0")
		}
		// The detail box keeps its scrollbar and writes no counter, with an
		// item under it and with nothing under it.
		for _, d := range []Model{press(sized(newModel(t), 160, 50), tabKey(tabPlans), "shift+tab"), empty} {
			if got := footOf(t, d, paneDetail); strings.Contains(got, " of ") {
				t.Errorf("the detail box writes a counter: %q", got)
			}
		}
	})
}

// thumbFixture gives every box more rows than either screen below can show,
// and the detail body a whole screenful of lines, so every box has a thumb to
// draw: thirty of every kind, work under way for the Activities tab, and
// finished ones for the Done boxes.
func thumbFixture(t *testing.T) Model {
	t.Helper()
	body := "# Plan\n" + strings.Repeat("\nA line of the body.\n", 60) +
		"\n### Task 1: First step\n\n- [ ] **Step 1: Do it**\n"
	files := map[string]string{}
	// Sixty of every kind, so a list still has a window between its top and
	// its end once the boxes of the tab bar layout are a whole screen tall.
	for i := range 60 {
		n := fmt.Sprint(i)
		files[".acta/specs/2026-09-20-open-"+n+".md"] = "# Open spec\n"
		files[".acta/specs/2026-09-19-done-"+n+".md"] = "---\nstatus: done\n---\n# Done spec\n"
		files[".acta/specs/2026-09-18-going-"+n+".md"] = "---\nstatus: in-progress\n---\n# Going spec\n"
		files[".acta/scratch/2026-09-20-idea-"+n+".md"] = "---\nstatus: brainstorming\n---\n# Idea\n"
		files[".acta/plans/2026-09-20-open-"+n+".md"] = body
		// A plan with a ticked box gives the Activities tab a task that is
		// under way, which is the only kind of row that tab lists.
		files[".acta/plans/2026-09-17-going-"+n+".md"] = "# Going plan\n\n### Task 1: Under way\n\n- [x] a step\n- [ ] another step\n"
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
	switch at {
	case "top":
		return press(m, "g")
	case "middle":
		// A list scrolls with its cursor, so the window lands inside the
		// content once the cursor has walked back far enough from the end
		// to leave the last window.
		return press(m, "G", "ctrl+u", "ctrl+u", "ctrl+u", "ctrl+u", "ctrl+u")
	default:
		return press(m, "G")
	}
}

// TestThumbSitsOnTheBorderNotInside walks every box of every tab, both
// screens and all three places in the content, and reads the drawn cells. No
// box spends a cell of its own on a scrollbar: the line is as wide as the box,
// the cell left of the right wall is content, and the thumb is that wall and
// nothing else.
func TestThumbSitsOnTheBorderNotInside(t *testing.T) {
	withColors(func() {
		sizes := [][2]int{{80, 30}, {160, 50}}
		if testing.Short() {
			// One size keeps the check on every tab and pane. The full run adds the wide one.
			sizes = sizes[:1]
		}
		// The group returns only when every subtest in it is done, so all of
		// them run while the colors are on.
		t.Run("group", func(t *testing.T) {
			for _, size := range sizes {
				for i := range topTabs {
					for _, p := range append(press(newModel(t), tabKey(i)).panes(), paneDetail) {
						t.Run(fmt.Sprintf("%dx%d/%s/%d", size[0], size[1], topTabs[i].name, p), func(t *testing.T) {
							t.Parallel()
							for _, at := range []string{"top", "middle", "end"} {
								checkThumbOnTheBorder(t, i, p, at, size[0], size[1])
							}
						})
					}
				}
			}
		})
	})
}

func checkThumbOnTheBorder(t *testing.T, tab int, p pane, at string, w, h int) {
	t.Helper()
	m := press(sized(thumbFixture(t), w, h), tabKey(tab))
	// The detail box shows a plan body, the only content on this board long
	// enough to scroll, so it is read from the Plans tab. Oldest first puts
	// the short "going" plans ahead of the long "open" ones, so jump to the
	// last row to land on one with a body worth scrolling.
	if p == paneDetail {
		m = press(m, tabKey(tabPlans), "G")
	}
	m.focusPane(p)
	// A Done sub-tab no item of the board reaches has nothing to scroll, so
	// the next one that holds rows is read instead.
	for range len(m.doneTabNames()) {
		if rows, _, _ := m.slotOf(paneDone); len(rows) > 0 {
			break
		}
		m = press(m, "]")
	}
	// A Done pane that no sub-tab of the tab fills has nothing to scroll.
	if p == paneDone {
		if rows, _, _ := m.slotOf(paneDone); len(rows) == 0 {
			return
		}
	}
	m = scrollTo(m, p, at)
	if rows, _, _ := m.slotOf(m.listPane()); p != paneDetail && len(rows) == 0 {
		t.Fatalf("tab %s box %d holds no rows to scroll", topTabs[tab].name, p+1)
	}
	b := paneBox(m, p)
	if b.inner < 1 {
		t.Fatalf("tab %s box %d has no room to draw", topTabs[tab].name, p+1)
	}
	v := m.View()
	// The thumb wears the brush the pane's own wall wears, so the accent on
	// the pane with the focus and the dim brush on the others.
	wall := m.edge(p)
	other := m.styles.accent
	if m.focus == p {
		other = m.styles.faint
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
		// line, because that is where the middle of the content lands, and
		// content one page long has no window between its two ends at all.
		if b.inner-thumbs < 3 || m.lastOff(p) <= pageLines {
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

// TestThumbOnlyWhenThereIsSomethingToScroll reads the plain fixture on a
// screen tall enough to hold all of it. Nothing overflows, so no box draws a
// thumb.
func TestThumbOnlyWhenThereIsSomethingToScroll(t *testing.T) {
	t.Parallel()

	for _, size := range [][2]int{{80, 30}, {160, 50}, {200, 120}} {
		m := press(sized(newModel(t), size[0], size[1]), tabKey(tabPlans), "shift+tab")
		v := m.View()
		for _, p := range m.panes() {
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

// Every list row puts its title at the same column, whatever number the id
// carries. A row that is one digit wider pushes its title one cell right, so
// the eye has to find the title again on every line. The old format let that
// happen: SCRATCH-3 and SCRATCH-13 were two widths. The new one pads every
// number to four digits, so the column holds for one digit and for four.
func TestListRowsLineUp(t *testing.T) {
	m := detailModel(t, treeCfg(t, map[string]string{
		// Old ids on the left, new on the right: both forms have to land on
		// the same width, so an unmigrated file still lines up.
		".acta/scratch/2026-09-20-a.md": "---\nid: SCRATCH-3\nhash: aa1b\nstatus: raw\n---\n# One\n",
		".acta/scratch/2026-09-21-b.md": "---\nid: SCR-1234\nhash: bb2c3d4\nstatus: raw\n---\n# Two\n",
		".acta/plans/2026-09-22-c.md":   "---\nid: PLAN-3\nhash: cc5d\n---\n# Three\n",
		".acta/plans/2026-09-23-d.md":   "---\nid: PLN-1234\nhash: dd6e7f80\n---\n# Four\n",
		".acta/specs/2026-09-24-e.md":   "---\nid: SPEC-3\nhash: ee9a\n---\n# Five\n",
		".acta/specs/2026-09-25-f.md":   "---\nid: SPC-1234\nhash: ffb0c1d2e\n---\n# Six\n",
		".acta/bugs/2026-09-26-g.md":    "---\nid: BUG-3\nhash: gg2e\n---\n# Seven\n\n## Symptom\nx\n",
		".acta/bugs/2026-09-27-h.md":    "---\nid: BUG-1234\nhash: hh3f4a5b6\n---\n# Eight\n\n## Symptom\ny\n",
		".acta/debt/2026-09-28-i.md":    "---\nid: DEBT-3\nhash: ii7c8\n---\n# Nine\n\n- [ ] first note\n",
		".acta/debt/2026-09-29-j.md":    "---\nid: DBT-1234\nhash: jj9d0e1f2a\n---\n# Ten\n\n- [ ] second note\n",
	}))
	for _, c := range []struct {
		kind, short string
		ids         []string
	}{
		{"scratch", "SCR", []string{"scratch/2026-09-20-a", "scratch/2026-09-21-b"}},
		{"plan", "PLN", []string{"plans/2026-09-22-c", "plans/2026-09-23-d"}},
		{"spec", "SPC", []string{"specs/2026-09-24-e", "specs/2026-09-25-f"}},
		{"bug", "BUG", []string{"bugs/2026-09-26-g", "bugs/2026-09-27-h"}},
		{"debt item", "DBT", []string{"debt/2026-09-28-i#item-1", "debt/2026-09-29-j#item-1"}},
	} {
		col, idCol := -1, -1
		for _, id := range c.ids {
			it := m.board.Get(id)
			if it == nil {
				t.Fatalf("the board holds no %s", id)
			}
			// The row shows the id in the new form, so a wide old number
			// still takes the same room as a new one.
			if want := c.short + "-"; !strings.HasPrefix(it.ShortID, want) {
				t.Errorf("%s %s: short id %q does not start with %q", c.kind, id, it.ShortID, want)
			}
			// Every id of a kind is one width: the number is padded, and a
			// debt item adds its own 2-digit number on top of that.
			if idCol == -1 {
				idCol = lipgloss.Width(it.ShortID)
			}
			if got := lipgloss.Width(it.ShortID); got != idCol {
				t.Errorf("%s %s: short id %q is %d cells, want %d like the rest of the kind",
					c.kind, id, it.ShortID, got, idCol)
			}
			row := m.rowText(row{id: id}, it, 80)
			if !strings.HasPrefix(row, it.ShortID+"  ") {
				t.Errorf("%s %s: row %q does not start with %q", c.kind, id, row, it.ShortID+"  ")
			}
			at := strings.Index(row, it.Title)
			if col == -1 {
				col = at
			}
			if at != col {
				t.Errorf("%s %s: row %q puts its title at %d, want %d", c.kind, id, row, at, col)
			}
		}
	}
}

// drewNew calls View once and says whether it really drew. It puts a marker
// in the kept frame: a reused frame hands the marker back, a new draw does not.
func drewNew(m Model) bool {
	before := m.frame.s
	m.frame.s = "\x00sentinel"
	got := m.View()
	drew := got != "\x00sentinel"
	if !drew {
		m.frame.s = before
	}
	return drew
}

func TestViewReusesTheFrameForANotchThatOnlyGathers(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m.View()
	m = wheelOnly(m, b.x+1, b.y+2, false)
	if drewNew(m) {
		t.Error("a notch that only adds to the delta drew the screen again")
	}
}

func TestViewReusesTheFrameForIgnoredMouseAndEmptyTick(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	m.View()
	// A wheel over the list, while the detail box has the focus, is ignored.
	lb := scrollBox(m, paneList)
	m = wheelOnly(m, lb.x+1, lb.y+2, false)
	if drewNew(m) {
		t.Error("a wheel over a pane without the focus drew the screen again")
	}
	m = wheelTick(m)
	if drewNew(m) {
		t.Error("a tick with nothing gathered drew the screen again")
	}
	next, _ := m.Update(tea.MouseMsg{X: 1, Y: 1, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
	if drewNew(next.(Model)) {
		t.Error("a mouse release drew the screen again")
	}
}

func TestViewDrawsAgainAfterEveryChange(t *testing.T) {
	t.Parallel()

	base := paneModel(t, paneDetail)
	b := scrollBox(base, paneDetail)
	cases := map[string]func(Model) Model{
		"a tick that scrolls": func(m Model) Model { return wheelTick(wheelOnly(m, b.x+1, b.y+2, false)) },
		"a key":               func(m Model) Model { return press(m, "j") },
		"a click on a row":    func(m Model) Model { lb := scrollBox(m, paneList); return click(m, lb.x+1, lb.y+2) },
		"a reload":            func(m Model) Model { return reloaded(m) },
		"a resize":            func(m Model) Model { return sized(m, m.width-1, m.height) },
		"a clock tick": func(m Model) Model {
			next, _ := m.Update(clockMsg(time.Now()))
			return next.(Model)
		},
		"a flush to another pane": func(m Model) Model {
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m.focus = paneList
			lb := scrollBox(m, paneList)
			return wheelOnly(m, lb.x+1, lb.y+2, false)
		},
	}
	for name, change := range cases {
		m := base
		m.frame = &frameCache{}
		m.View()
		m = wheelOnly(m, b.x+1, b.y+2, false) // leave same set, so the change must clear it
		m = change(m)
		if !drewNew(m) {
			t.Errorf("%s: View gave the old frame back", name)
		}
	}
}
