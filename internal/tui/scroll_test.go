package tui

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
)

// longModel builds a board with more rows than a pane has room for, so all
// three panes overflow: forty plans still open for pane [1], forty finished
// ones for pane [2], and a body of sixty lines so the detail pane overflows
// whatever item is selected.
func longModel(t *testing.T) Model {
	t.Helper()
	body := "# Plan " + strings.Repeat("x", 40) + "\n" + strings.Repeat("\nA line of the body.\n", 60) +
		"\n### Task 1: First step\n\n- [ ] **Step 1: Do it**\n"
	return boardModel(t, body)
}

// boardModel puts one body on every one of the forty open plans and the same
// body ticked on every one of the forty finished plans, so all three panes
// have more rows than they can show.
func boardModel(t *testing.T, body string) Model {
	t.Helper()
	files := map[string]string{}
	done := strings.Replace(body, "- [ ]", "- [x]", 1)
	for i := range 40 {
		files[fmt.Sprintf(".acta/plans/2026-09-20-open-%02d.md", i)] = body
		files[fmt.Sprintf(".acta/plans/2026-09-20-done-%02d.md", i)] = done
	}
	cfg := treeCfg(t, files)
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return sized(m, 120, 40)
}

// wrapModel is the long board with a renderer that wraps the way glamour
// does, so how many lines a body has depends on the width the pane gives it.
// That is the real case: a model that measures the body at one width and a
// view that draws it at another drifts apart.
func wrapModel(t *testing.T) Model {
	t.Helper()
	// A line of exactly two pane widths fits two lines at the width the model
	// measures the body at, and needs three at the width one cell less that the
	// view draws it at. A model and a view that measure differently drift a
	// whole line apart, so the end of the body never lands on screen.
	n := scrollBox(boardModel(t, "# Plan\n"), paneDetail).textW()
	m := boardModel(t, "# Plan\n\n### Task 1: First step\n\n- [ ] **Step 1: Do it**\n\n"+
		strings.Repeat("\n"+strings.Repeat("z", 2*n)+"\n", 20))
	m.render = func(md string, w int) string {
		var out []string
		for _, ln := range strings.Split(md, "\n") {
			r := []rune(ln)
			for w > 0 && len(r) > w {
				out, r = append(out, string(r[:w])), r[w:]
			}
			out = append(out, string(r))
		}
		return strings.Join(out, "\n")
	}
	return m
}

// paneModel is the long board with pane p focused and the Plans tab open,
// where all three panes have rows to show.
func paneModel(t *testing.T, p pane) Model {
	t.Helper()
	// The plans list has rows to walk; the detail box needs one of them
	// selected before it has a body to scroll.
	m := press(longModel(t), keyOf(panePlans))
	if p == paneDetail {
		return press(m, "0")
	}
	return press(m, keyOf(p))
}

// scrollBox is the box pane p is drawn in, the detail pane included; the
// view tests only need the two list panes, so they answer for the open one
// whenever they are asked about anything else.
func scrollBox(m Model, p pane) box {
	g := m.geometry()
	if !g.wide {
		return g.full
	}
	return g.at(p)
}

// paneRows gives the lines pane p draws between its walls, padding included,
// so a test can read the words of a row. The right wall wears the plain bar or
// the scrollbar thumb, and both walls come off here.
func paneRows(m Model, p pane) []string {
	b := scrollBox(m, p)
	col := column(m.View(), b.x, b.w)
	out := make([]string, 0, b.inner)
	for i := range b.inner {
		y := b.y + 1 + i
		if y >= len(col) {
			break
		}
		out = append(out, cutWalls(plain(col[y])))
	}
	return out
}

// cutWalls takes both walls off a drawn line of a pane, whichever glyph each
// one wears.
func cutWalls(line string) string {
	return strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(line, "│"), "┃"), "│")
}

// wallCells gives the right wall of every line a pane draws: the plain bar, or
// the thumb on the lines the pane has scrolled to.
func wallCells(m Model, p pane) []string {
	b := scrollBox(m, p)
	col := column(m.View(), b.x, b.w)
	out := make([]string, 0, b.inner)
	for i := range b.inner {
		y := b.y + 1 + i
		if y >= len(col) {
			break
		}
		r := []rune(plain(col[y]))
		if len(r) == 0 {
			out = append(out, " ")
			continue
		}
		out = append(out, string(r[len(r)-1]))
	}
	return out
}

// thumbRows counts the lines a thumb covers.
func thumbRows(bar []bool) int {
	n := 0
	for _, on := range bar {
		if on {
			n++
		}
	}
	return n
}

// thumbSpan gives the first and the last line the scrollbar thumb covers, or
// two -1 when the pane draws no thumb at all.
func thumbSpan(m Model, p pane) (top, bottom int) {
	top, bottom = -1, -1
	for i, c := range wallCells(m, p) {
		if c != "┃" {
			continue
		}
		if top < 0 {
			top = i
		}
		bottom = i
	}
	return top, bottom
}

// thumbAt gives the line the scrollbar thumb starts on, or -1 when the pane
// draws no thumb at all.
func thumbAt(m Model, p pane) int {
	top, _ := thumbSpan(m, p)
	return top
}

// bottomLine gives the raw bottom border a pane draws, walls and dashes and
// all, so a test can say where on the line the text sits.
func bottomLine(t *testing.T, m Model, p pane) string {
	t.Helper()
	b := scrollBox(m, p)
	col := column(m.View(), b.x, b.w)
	if b.y+b.h-1 >= len(col) {
		t.Fatalf("pane %d draws no bottom border", p)
	}
	return plain(col[b.y+b.h-1])
}

// footOf gives what a pane writes in its bottom border, with the border itself
// cut off: the counter of a sidebar pane, nothing at all on the detail box.
func footOf(t *testing.T, m Model, p pane) string {
	t.Helper()
	return strings.Trim(bottomLine(t, m, p), "─└┘ ")
}

// drawnFirst gives the first line a pane has on screen, walls and padding cut
// off, which is where the reader looks to see what a box is showing.
func drawnFirst(m Model, p pane) string {
	rows := paneRows(m, p)
	if len(rows) == 0 {
		return ""
	}
	return strings.TrimSpace(rows[0])
}

// shows says whether the screen has the content line at index i on top of
// pane p: the item of a list, or the line of a body in the detail box. The
// answer comes from the drawn line, never from the offset, so a pane that
// claims a place it has not drawn is caught here.
func shows(m Model, p pane, i int) bool {
	if p == paneDetail {
		lines := m.detailLines(scrollBox(m, p).textW())
		return i >= 0 && i < len(lines) && drawnFirst(m, p) == strings.TrimSpace(lines[i])
	}
	drawn := paneRows(m, p)
	items, _, _ := m.slotOf(p)
	return i >= 0 && i < len(items) && len(drawn) > 0 && isThatRow(m, items[i], drawn[0])
}

// screenTops gives the line every box has on top, so a test can say what the
// screen shows and not only what the model holds. A box that draws nothing
// reads empty.
func screenTops(m Model) []string {
	out := make([]string, boxes)
	for p := pane(0); int(p) < boxes; p++ {
		out[p] = drawnFirst(m, p)
	}
	return out
}

// TestItemCount is the whole counter rule: the item under the cursor, counting
// from one, out of the items the pane holds, and nothing at all when the pane
// holds no items.
func TestItemCount(t *testing.T) {
	cases := []struct {
		sel, total int
		want       string
	}{{0, 0, "0 of 0"}, {1, 1, "1 of 1"}, {3, 20, "3 of 20"}, {236, 236, "236 of 236"}}
	for _, c := range cases {
		if got := itemCount(c.sel, c.total); got != c.want {
			t.Errorf("itemCount(%d, %d) = %q, want %q", c.sel, c.total, got, c.want)
		}
	}
}

// headOf gives the words a row draws at its start, so a test can tell which
// row a drawn line belongs to.
func headOf(m Model, r row) string {
	it := m.board.Get(r.id)
	if it == nil {
		return r.id
	}
	name := it.ShortID
	if name == "" {
		name = it.ID
	}
	lead := strings.Repeat("  ", r.depth)
	if r.tree {
		lead += m.treeMark(r, it) + " "
	}
	return strings.TrimSpace(lead + name + "  " + it.Title)
}

// isThatRow says whether a drawn line is that row and nothing else, cut to
// the cells it had room for.
func isThatRow(m Model, r row, line string) bool {
	rest := strings.TrimRight(line, " ")
	return rest == cutTo(headOf(m, r), lipgloss.Width(rest))
}

// TestScrollbarRowsAreThumbOnly is the whole thumb rule: no thumb when the
// content fits, a thumb of at least one line, the thumb on the first line at
// the top and on the last line at the end, and a thumb that keeps the size of
// the window the pane shows.
func TestScrollbarRowsAreThumbOnly(t *testing.T) {
	cases := []struct {
		total, visible, first, h int
		want                     string // one char per line: T thumb, . no thumb
	}{
		{10, 10, 0, 10, ".........."}, // fits: no thumb
		{0, 10, 0, 10, ".........."},
		{100, 10, 0, 10, "T........."},
		{100, 10, 90, 10, ".........T"},
		{20, 10, 5, 10, "..TTTTT..."},
		{1000, 5, 500, 5, "..T.."},
	}
	for _, c := range cases {
		rows := scrollbar(c.total, c.visible, c.first, c.h)
		if len(rows) != c.h {
			t.Errorf("scrollbar(%d,%d,%d,%d) has %d lines, want %d", c.total, c.visible, c.first, c.h, len(rows), c.h)
			continue
		}
		got := ""
		for _, on := range rows {
			if on {
				got += "T"
				continue
			}
			got += "."
		}
		if got != c.want {
			t.Errorf("scrollbar(%d,%d,%d,%d) = %s, want %s", c.total, c.visible, c.first, c.h, got, c.want)
		}
	}
}

func TestScrollbarThumbFollowsTheOffset(t *testing.T) {
	cases := []struct {
		total, vis, first   int
		wantTop, wantBottom bool
	}{
		{100, 10, 0, true, false},
		{100, 10, 90, false, true},
	}
	for _, c := range cases {
		bar := scrollbar(c.total, c.vis, c.first, c.vis)
		if len(bar) != c.vis {
			t.Fatalf("bar has %d lines, want %d", len(bar), c.vis)
		}
		if bar[0] != c.wantTop || bar[c.vis-1] != c.wantBottom {
			t.Errorf("%+v: bar %v", c, bar)
		}
	}
	if slices.Contains(scrollbar(5, 10, 0, 10), true) {
		t.Error("a thumb on a pane whose content fits")
	}
}

func TestScrollbarThumbSitsBetweenTheEnds(t *testing.T) {
	// Half way down, the thumb sits inside the column, never on either end.
	bar := scrollbar(100, 10, 45, 10)
	if len(bar) != 10 {
		t.Fatalf("bar has %d lines, want 10", len(bar))
	}
	if bar[0] || bar[9] {
		t.Errorf("the thumb jumped to an end: %v", bar)
	}
	if n := thumbRows(bar); n != 1 {
		t.Errorf("one line should be the thumb, got %d: %v", n, bar)
	}
	// A window that covers half the content gets a thumb of half the column,
	// so the thumb never lies about how much is on screen.
	bar = scrollbar(100, 50, 25, 10)
	if n := thumbRows(bar); n != 5 {
		t.Errorf("half of the content should give a thumb of 5 lines, got %d: %v", n, bar)
	}
	if bar[1] || !bar[2] || !bar[6] || bar[7] {
		t.Errorf("the thumb should sit in the middle of the column: %v", bar)
	}
}

func TestScrollbarSurvivesOddNumbers(t *testing.T) {
	// An offset past the end reads as the end, and one before the start reads
	// as the start: a stale offset cannot leave the thumb off the column.
	for _, c := range []struct {
		first    int
		wantLine int
	}{
		{-5, 0}, {0, 0}, {45, 4}, {90, 9}, {999, 9},
	} {
		bar := scrollbar(100, 10, c.first, 10)
		if len(bar) != 10 {
			t.Fatalf("first %d: bar has %d lines, want 10", c.first, len(bar))
		}
		if n := thumbRows(bar); n != 1 {
			t.Errorf("first %d: one line should be the thumb, got %d: %v", c.first, n, bar)
		}
		if !bar[c.wantLine] {
			t.Errorf("first %d: the thumb should sit on line %d: %v", c.first, c.wantLine, bar)
		}
	}
	// A pane with no line to show still draws a column instead of crashing.
	if bar := scrollbar(10, 0, 0, 10); len(bar) != 10 || !bar[0] {
		t.Errorf("a pane with no visible line drew %v, want a thumb on line 0", bar)
	}
	// Content that exactly fills the pane needs no thumb, and neither does a
	// pane too short to draw one.
	if slices.Contains(scrollbar(10, 10, 0, 10), true) || len(scrollbar(10, 4, 0, 0)) != 0 {
		t.Error("a thumb where there is nothing to scroll")
	}
}

func TestScrollbarShowsOnlyOnOverflow(t *testing.T) {
	// The fixture fits in every box of a tall enough screen, so none of them
	// draws a thumb.
	tall := press(sized(newModel(t), 120, 80), "3")
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		if at := thumbAt(tall, p); at != -1 {
			t.Errorf("pane %d draws a thumb for content that fits, on line %d", p, at)
		}
	}
	// The long board overflows in all three, and each one says so with a
	// thumb sitting at the top of its own wall.
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		// The detail box needs an item under it before it has a body to
		// scroll, so it is read from the same board with one selected.
		m := paneModel(t, p)
		if p == paneDetail {
			m = press(paneModel(t, panePlans), "0")
		}
		if at := thumbAt(m, p); at != 0 {
			t.Errorf("pane %d draws no thumb at the top, thumb on line %d", p, at)
		}
	}
}

func TestWheelScrollsOnlyTheFocusedPane(t *testing.T) {
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		m := paneModel(t, p)
		before := m.Selected()
		shown := screenTops(m)
		b := scrollBox(m, p)
		m = wheel(m, b.x+1, b.y+2, false)
		if m.off[p] != 1 {
			t.Errorf("pane %d: one notch should scroll it by one, off is %d", p, m.off[p])
		}
		for q := range boxes {
			if q != int(p) && drawnFirst(m, pane(q)) != shown[q] {
				t.Errorf("the wheel over pane %d moved pane %d on screen to %q, was %q", p, q, drawnFirst(m, pane(q)), shown[q])
			}
		}
		// The wheel scrolls a pane, it never picks another item.
		if p != paneDetail {
			if got := m.Selected(); got == nil || before == nil || got.ID != before.ID {
				t.Errorf("the wheel over pane %d changed the selection to %v", p, got)
			}
		}
		// What the screen has on top follows the wheel.
		if !shows(m, p, 1) {
			t.Errorf("pane %d: after one notch the top of the box should be its second line, it shows %q", p, drawnFirst(m, p))
		}
		m = wheel(m, b.x+1, b.y+2, true)
		if m.off[p] != 0 {
			t.Errorf("pane %d: one notch back should be at the top, off is %d", p, m.off[p])
		}
		if !shows(m, p, 0) {
			t.Errorf("pane %d: back at the top the box should show its first line, it shows %q", p, drawnFirst(m, p))
		}
		// The wheel stops at both ends, so turning it forever never leaves a
		// line of nothing on screen and never counts past the content.
		for range 3 {
			m = wheel(m, b.x+1, b.y+2, true)
			if m.off[p] != 0 {
				t.Errorf("pane %d: the wheel scrolled above the top, off is %d", p, m.off[p])
			}
		}
		for range 200 {
			m = wheel(m, b.x+1, b.y+2, false)
		}
		if want := m.lastOff(p); m.off[p] != want {
			t.Errorf("pane %d: the wheel scrolled to %d, the content allows %d", p, m.off[p], want)
		}
		if !shows(m, p, m.lastOff(p)) {
			t.Errorf("pane %d: at the end the box should start at line %d, it shows %q", p, m.lastOff(p), drawnFirst(m, p))
		}
		// A body of a plan is mostly empty lines, so the first line on screen
		// may be one of them. A row of a list never is, because every row has
		// words on it.
		blank := 0
		for _, ln := range paneRows(m, p) {
			if strings.TrimSpace(ln) == "" {
				blank++
			}
		}
		if blank == len(paneRows(m, p)) || (p != paneDetail && blank > 0) {
			t.Errorf("pane %d: the wheel left %d of %d lines blank", p, blank, len(paneRows(m, p)))
		}
	}
}

// A pane with less content than it has rows has nowhere to scroll, so every
// key and every notch leaves it at the top. An offset below zero would count
// from a line that does not exist.
func TestAPaneShorterThanItsContentNeverScrolls(t *testing.T) {
	m := sized(newModel(t), 120, 40)
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		if m.lastOff(p) != 0 {
			t.Errorf("pane %d: content that fits should not scroll, lastOff is %d", p, m.lastOff(p))
		}
		b := scrollBox(m, p)
		m = click(m, b.x+1, b.y+2)
		m = press(m, "G", "j", "j", "ctrl+d")
		if m.off[p] != 0 {
			t.Errorf("pane %d: content that fits scrolled to %d", p, m.off[p])
		}
		for range 5 {
			m = wheel(m, b.x+1, b.y+2, false)
		}
		if m.off[p] != 0 {
			t.Errorf("pane %d: the wheel scrolled a pane that fits, off is %d", p, m.off[p])
		}
		// Nothing on screen says the pane moved: no thumb, and the counter
		// still counts the items of the pane, which G left on the last one.
		if p == paneDetail {
			if strings.Contains(footOf(t, m, p), " of ") {
				t.Errorf("pane %d: the detail box writes a counter %q", p, footOf(t, m, p))
			}
		} else if rows, _, _ := m.slotOf(p); len(rows) > 0 {
			if got, want := footOf(t, m, p), itemCount(len(rows), len(rows)); got != want {
				t.Errorf("pane %d: content that fits writes %q, want %q", p, got, want)
			}
		}
		if at := thumbAt(m, p); at != -1 {
			t.Errorf("pane %d: content that fits draws a thumb on line %d", p, at)
		}
		if p == paneDetail {
			lines := m.detailLines(scrollBox(m, p).textW())
			if got, want := strings.TrimRight(paneRows(m, p)[0], " "), plain(strings.TrimRight(lines[0], " ")); got != want {
				t.Errorf("pane %d: the first line on screen is %q, the body starts on %q", p, got, want)
			}
		} else if rows, _, _ := m.slotOf(p); len(rows) > 0 {
			drawn := strings.TrimSpace(paneRows(m, p)[0])
			if !strings.HasPrefix(drawn, headOf(m, rows[0])[:8]) {
				t.Errorf("pane %d: the first line on screen is %q, the content starts on %q", p, drawn, headOf(m, rows[0]))
			}
		}
	}
}

// A wheel belongs to the pane that has the focus, the way the keys do. Over
// any other pane it does nothing at all: no scroll, no focus change, so a
// wheel can never move a pane the user is not looking at.
func TestWheelOverAnUnfocusedPaneDoesNothing(t *testing.T) {
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		for _, q := range []pane{panePlans, paneDone, paneDetail} {
			if q == p {
				continue
			}
			// The pane under the pointer is scrolled away from the top first,
			// so a wheel that ignored the focus would show on screen instead of
			// hiding behind a clamp at zero.
			m := press(paneModel(t, p), keyOf(p), "G")
			m = press(m, keyOf(q))
			sel := m.Selected()
			before := m.off
			shown := screenTops(m)
			if before[p] == 0 || shows(m, p, 0) {
				t.Fatalf("pane %d did not scroll before the wheel over it", p)
			}
			g := m.geometry()
			target := g.at(p)
			if target.w < 2 || target.h < 2 {
				continue
			}
			// Both notches, so neither direction of the wheel can slip past.
			for _, up := range []bool{false, true} {
				m = wheel(m, target.x+1, target.y+2, up)
				if m.focus != q {
					t.Errorf("a wheel over pane %d took the focus off pane %d", p, q)
				}
				if !slices.Equal(m.off, before) {
					t.Errorf("a wheel over unfocused pane %d moved something: %v, want %v", p, m.off, before)
				}
				if !slices.Equal(screenTops(m), shown) {
					t.Errorf("a wheel over unfocused pane %d moved the screen: %v, want %v", p, screenTops(m), shown)
				}
				if got := m.Selected(); got == nil || sel == nil || got.ID != sel.ID {
					t.Errorf("a wheel over unfocused pane %d changed the selection to %v", p, got)
				}
			}
		}
	}
}

// A click takes the focus, so the wheel over the same pane scrolls it right
// after, with no key in between.
func TestWheelAfterAClickScrollsThatPane(t *testing.T) {
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		m := paneModel(t, panePlans)
		b := scrollBox(m, p)
		m = click(m, b.x+1, b.y+2)
		m = wheel(m, b.x+1, b.y+2, false)
		if m.focus != p {
			t.Errorf("pane %d: the click should have taken the focus, focus is %d", p, m.focus)
		}
		if m.off[p] != 1 {
			t.Errorf("pane %d: the wheel after the click should scroll it by one, off is %d", p, m.off[p])
		}
	}
}

func TestKeysActOnTheFocusedPaneOnly(t *testing.T) {
	// The keys of the detail box scroll it and leave both lists where they
	// were, selection and offset.
	m := press(longModel(t), "3", "G")
	before := screenTops(m)
	m = press(m, "0", "ctrl+d", "ctrl+d")
	if m.off[paneDetail] != 2*pageLines {
		t.Fatalf("two pages should scroll the detail to %d, off is %d", 2*pageLines, m.off[paneDetail])
	}
	// The screen says the same: the detail box moved two pages down, the two
	// lists did not move at all.
	after := screenTops(m)
	if !shows(m, paneDetail, 2*pageLines) {
		t.Errorf("the detail box shows %q, it should show its line %d", after[paneDetail], 2*pageLines)
	}
	if after[panePlans] != before[panePlans] || after[paneDone] != before[paneDone] {
		t.Errorf("the keys in the detail box moved a list on screen: %v, want %v", after, before)
	}
	// Coming back to a list finds it exactly where it was left.
	m = press(m, "3")
	if want := m.lastOff(panePlans); m.off[panePlans] != want {
		t.Errorf("the list scrolled to %d, want %d", m.off[panePlans], want)
	}
	if got := screenTops(m)[panePlans]; got != before[panePlans] {
		t.Errorf("the list shows %q, it showed %q before the keys", got, before[panePlans])
	}
	// A key in pane [1] moves pane [1] and no other pane. Another item
	// starts the detail at its own top, which is the one place a key in a
	// list reaches into the detail box.
	m = press(m, "k")
	after = screenTops(m)
	if after[paneDone] != before[paneDone] {
		t.Errorf("k in pane [1] moved pane [2] on screen: %v, want %v", after[paneDone], before[paneDone])
	}
	if !shows(m, paneDetail, 0) {
		t.Errorf("another item should start the detail at its top, the screen shows %q", drawnFirst(m, paneDetail))
	}
	// Walking to the top of a list scrolls it back up, and still no other
	// pane moves.
	m = press(m, "g")
	after = screenTops(m)
	if !shows(m, panePlans, 0) {
		t.Errorf("g should take the list back to its first row, the screen shows %q", after[panePlans])
	}
	if after[paneDone] != before[paneDone] || !shows(m, paneDetail, 0) {
		t.Errorf("g in pane [1] moved another pane on screen: %v", after)
	}
}

func TestListKeepsTheSelectedRowVisible(t *testing.T) {
	m := press(longModel(t), "3")
	rows, _, _ := m.slotOf(panePlans)
	// Walk down until the cursor sits below the fold, then keep walking: the
	// pane has to follow the cursor, not the other way round.
	for m.cursor() < scrollBox(m, panePlans).rows {
		m = press(m, "j")
	}
	below := m.cursor()
	m = press(m, "j")
	if m.cursor() == below {
		t.Fatal("the list should still move down")
	}
	b := scrollBox(m, panePlans)
	cur := m.cursor()
	if cur < b.first || cur >= b.first+b.rows {
		t.Fatalf("row %d is off screen: the pane shows rows %d to %d", cur, b.first, b.first+b.rows-1)
	}
	// The row the pane says is on top is the row that is drawn on top.
	if !shows(m, panePlans, b.first) {
		t.Fatalf("the top of the pane shows %q, the pane starts at row %d of %d", drawnFirst(m, panePlans), b.first, len(rows))
	}
	drawn := paneRows(m, panePlans)
	if !isThatRow(m, rows[b.first], drawn[0]) {
		t.Errorf("the top line holds %q, the first row is %q", drawn[0], headOf(m, rows[b.first]))
	}
	if !isThatRow(m, rows[cur], drawn[cur-b.first]) {
		t.Errorf("the selected row %d is not drawn where the cursor is", cur)
	}
	// Walking back up brings the cursor home and keeps it on screen.
	m = press(m, "k")
	if m.cursor() != below {
		t.Errorf("k should step back to row %d, is %d", below, m.cursor())
	}
	if m.cursor() < scrollBox(m, panePlans).first {
		t.Error("k moved the cursor off the top of the pane")
	}
}

func TestSelectingAnotherItemPutsTheDetailBackAtTheTop(t *testing.T) {
	m := press(longModel(t), "3", "0")
	for m.off[paneDetail] < 5 {
		m = press(m, "j")
	}
	if m.off[paneDetail] == 0 {
		t.Fatal("the detail never scrolled")
	}
	if !shows(m, paneDetail, m.off[paneDetail]) {
		t.Errorf("the detail shows %q at offset %d", drawnFirst(m, paneDetail), m.off[paneDetail])
	}
	m = press(m, "3", "j")
	if m.off[paneDetail] != 0 {
		t.Errorf("another item should start the detail at the top, off is %d", m.off[paneDetail])
	}
	if !shows(m, paneDetail, 0) {
		t.Errorf("the detail shows %q at the top, want its first line", drawnFirst(m, paneDetail))
	}
	// The line on screen is the first line of the body again, and the thumb
	// is back at the top of the column.
	if !isFirstLine(m, paneDetail, paneRows(m, paneDetail)[0]) {
		t.Errorf("the detail does not show its first line: %q", paneRows(m, paneDetail)[0])
	}
	if top := thumbAt(m, paneDetail); top != 0 {
		t.Errorf("the detail thumb sits on line %d after picking another item, want 0", top)
	}
}

func TestThumbAndTopLineFollowTheOffsetOnScreen(t *testing.T) {
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		m := paneModel(t, p)
		// Top: the thumb is on the first line and so is the content.
		if !shows(m, p, 0) {
			t.Fatalf("pane %d at the top shows %q, want its first line", p, drawnFirst(m, p))
		}
		if top := thumbAt(m, p); top != 0 {
			t.Errorf("pane %d at the top has its thumb on line %d, want 0", p, top)
		}
		// Middle: the thumb leaves both ends and the top of the box moves
		// with it. A list only follows its cursor once the cursor leaves the
		// window, so three pages are needed to get off the first one.
		m = press(m, "ctrl+d", "ctrl+d", "ctrl+d")
		first := m.off[p]
		if first <= 0 || first >= m.lastOff(p) {
			t.Fatalf("pane %d sits at %d of %d, which is neither the middle nor the end", p, first, m.lastOff(p))
		}
		b := scrollBox(m, p)
		if top, bottom := thumbSpan(m, p); top <= 0 || bottom >= b.inner-1 {
			t.Errorf("pane %d in the middle has its thumb on lines %d to %d of %d", p, top, bottom, b.inner)
		}
		if !shows(m, p, first) {
			t.Errorf("pane %d shows %q while its offset is %d", p, drawnFirst(m, p), first)
		}
		// End: the thumb is on the last line and the box starts the last
		// window its own content allows.
		m = press(m, "G")
		last := m.lastOff(p)
		if m.off[p] != last {
			t.Errorf("pane %d at the end sits at %d of %d", p, m.off[p], last)
		}
		if _, bottom := thumbSpan(m, p); bottom != b.inner-1 {
			t.Errorf("pane %d at the end has its thumb ending on line %d, want %d", p, bottom, b.inner-1)
		}
		if !shows(m, p, last) {
			t.Errorf("pane %d at the end shows %q, want the window at %d", p, drawnFirst(m, p), last)
		}
		// The last line on screen is the last line of the content.
		drawn := paneRows(m, p)
		if p == paneDetail {
			lines := m.detailLines(b.textW())
			if got, want := strings.TrimRight(drawn[len(drawn)-1], " "), plain(strings.TrimRight(lines[len(lines)-1], " ")); got != want {
				t.Errorf("pane %d ends on %q, the content ends on %q", p, got, want)
			}
		} else {
			rows, _, _ := m.slotOf(p)
			if r := rows[last+b.inner-1]; !isThatRow(m, r, drawn[len(drawn)-1]) {
				t.Errorf("pane %d ends on %q, the last row on screen is %q", p, drawn[len(drawn)-1], headOf(m, r))
			}
		}
	}
}

func TestOffsetStaysInsideThePaneAfterAResize(t *testing.T) {
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		m := press(paneModel(t, p), "G")
		if m.off[p] == 0 {
			t.Fatalf("pane %d never scrolled", p)
		}
		// A window so tall that the box holds all of its content: the place
		// the box was at is gone, so the offset goes with it. The five boxes
		// of the left column share the height, so the terminal has to be
		// five times as tall as the content it holds.
		m = sized(m, 120, 300)
		if m.off[p] != 0 {
			t.Errorf("pane %d kept the place %d with room for all of it", p, m.off[p])
		}
		if !isFirstLine(m, p, paneRows(m, p)[0]) {
			t.Errorf("pane %d does not start at its first line: %q", p, paneRows(m, p)[0])
		}
		if p == paneDetail {
			if strings.Contains(footOf(t, m, p), " of ") {
				t.Errorf("pane %d writes a counter for content that fits", p)
			}
		} else if rows, _, _ := m.slotOf(p); len(rows) > 0 {
			if got, want := footOf(t, m, p), itemCount(len(rows), len(rows)); got != want {
				t.Errorf("pane %d writes %q for content that fits, want %q", p, got, want)
			}
		}
		if at := thumbAt(m, p); at != -1 {
			t.Errorf("pane %d draws a thumb on line %d for content that fits", p, at)
		}
		// A window so short that the box shows a handful of lines still
		// draws its content, never a blank screen. The five boxes of the
		// left column share the height, so the terminal still has to be tall
		// enough to give each of them a border and a few rows.
		m = sized(m, 120, 40)
		if m.off[p] > m.lastOff(p) || m.off[p] < 0 {
			t.Errorf("pane %d sits at %d, the pane shows %d of %d", p, m.off[p], m.fitOf(p), m.linesOf(p))
		}
		if !shows(m, p, m.off[p]) {
			t.Errorf("pane %d shows %q after the resize, at offset %d", p, drawnFirst(m, p), m.off[p])
		}
		if strings.TrimSpace(paneRows(m, p)[0]) == "" {
			t.Errorf("pane %d draws a blank first line after the resize", p)
		}
		// The offset never leaves the pane, whichever way the keys go.
		m = press(m, "G", "j", "j", "ctrl+d")
		if m.off[p] > m.lastOff(p) || m.off[p] < 0 {
			t.Errorf("pane %d scrolled past its content: %d of %d", p, m.off[p], m.lastOff(p))
		}
		if !shows(m, p, m.lastOff(p)) {
			t.Errorf("pane %d shows %q after scrolling to the end, want the window at %d", p, drawnFirst(m, p), m.lastOff(p))
		}
		m = press(m, "g", "k", "k", "ctrl+u")
		if m.off[p] != 0 {
			t.Errorf("pane %d scrolled above its top: %d", p, m.off[p])
		}
		if !shows(m, p, 0) {
			t.Errorf("pane %d shows %q after scrolling back to the top, want its first line", p, drawnFirst(m, p))
		}
		// And growing the window again takes the box back to the top. The
		// five boxes of the left column share the height, so the terminal
		// has to be five times as tall as the content it holds.
		m = press(m, "G")
		m = sized(m, 120, 300)
		if m.off[p] != 0 {
			t.Errorf("pane %d kept the place %d after a window that fits it all", p, m.off[p])
		}
		if !isFirstLine(m, p, paneRows(m, p)[0]) {
			t.Errorf("pane %d does not start at its first line after growing: %q", p, paneRows(m, p)[0])
		}
	}
}

// A body that wraps has a different number of lines at every width, so the
// pane and the model have to measure it at the same one. When they do not,
// G stops a line short of the end the screen can show and the last line on
// screen is a line that never arrives.
func TestTheLastWindowArrivesWithAWrappingBody(t *testing.T) {
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		// The detail box needs an item under it before it has a body, so the
		// list is opened on the plans first.
		m := press(wrapModel(t), "3")
		m = press(m, keyOf(p), "G")
		b := scrollBox(m, p)
		if !shows(m, p, m.lastOff(p)) {
			t.Errorf("pane %d: after G the box shows %q, want the window at %d", p, drawnFirst(m, p), m.lastOff(p))
		}
		// The last line the box draws is the last line of its own content.
		drawn := paneRows(m, p)
		if p == paneDetail {
			lines := m.detailLines(b.textW())
			if got, want := strings.TrimRight(drawn[len(drawn)-1], " "), plain(strings.TrimRight(lines[len(lines)-1], " ")); got != want {
				t.Errorf("pane %d ends on %q, the body ends on %q", p, got, want)
			}
		} else {
			rows, _, _ := m.slotOf(p)
			if r := rows[m.off[p]+b.inner-1]; !isThatRow(m, r, drawn[len(drawn)-1]) {
				t.Errorf("pane %d ends on %q, the last row on screen is %q", p, drawn[len(drawn)-1], headOf(m, r))
			}
		}
		if _, bottom := thumbSpan(m, p); bottom != b.inner-1 {
			t.Errorf("pane %d: after G the thumb ends on line %d, want %d", p, bottom, b.inner-1)
		}
	}
}

// isFirstLine says whether a drawn line is the first line of the content of
// pane p, the row it starts on or the first line of the body.
func isFirstLine(m Model, p pane, line string) bool {
	b := scrollBox(m, p)
	if p == paneDetail {
		lines := m.detailLines(b.textW())
		return strings.TrimRight(line, " ") == plain(strings.TrimRight(lines[0], " "))
	}
	rows, _, _ := m.slotOf(p)
	return len(rows) > 0 && isThatRow(m, rows[0], line)
}

func TestPanesKeepTheirOwnPlace(t *testing.T) {
	m := press(longModel(t), "3", "G")
	open := screenTops(m)[paneSpecs]
	m = press(m, "5", "G")
	done := screenTops(m)[paneDone]
	if shows(m, paneDone, 0) {
		t.Fatalf("pane [2] did not scroll, it shows %q", drawnFirst(m, paneDone))
	}
	if got := screenTops(m)[paneSpecs]; got != open {
		t.Errorf("pane [1] moved to %q while pane [2] scrolled, was %q", got, open)
	}
	m = press(m, "0", "G")
	after := screenTops(m)
	if after[paneSpecs] != open || after[paneDone] != done {
		t.Errorf("the detail box moved the lists: %v, want %v and %v", after, open, done)
	}
	if shows(m, paneDetail, 0) {
		t.Error("the detail box did not scroll")
	}
	// Each pane has to end on the last line its own content allows, not on a
	// line that belongs to some other pane's body.
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		if !shows(m, p, m.lastOff(p)) {
			t.Errorf("pane %d ends on %q, its own content allows the window at %d", p, drawnFirst(m, p), m.lastOff(p))
		}
	}
	m = press(m, "2")
	if got := screenTops(m)[paneSpecs]; got != open {
		t.Errorf("coming back to pane [1] lost its place: %q, want %q", got, open)
	}
}

// A pane that gives its last cell to the scrollbar has to give a cell back
// when it is too narrow to spare one, or its walls and the line the view cuts
// to the window width end up a cell apart. A body whose last line is exactly
// the inner width, or that has no newline at the end, is the same problem
// from the other side: the line is already full when the pane pads it.
func TestNarrowPanesAndOddBodiesKeepTheirWalls(t *testing.T) {
	// Heights below three leave no room for a pane at all, which the frame
	// task covers on its own; what the scrollbar adds is a cell inside a pane.
	for _, w := range []int{1, 2, 3, 4, 5, 6, 8, 12, 20, 40} {
		for _, h := range []int{3, 4, 5, 8, 20, 40} {
			m := press(longModel(t), "3", "0", "G")
			v := sized(m, w, h).View()
			lines := strings.Split(v, "\n")
			if len(lines) > h {
				t.Errorf("%dx%d: %d lines", w, h, len(lines))
			}
			for i, ln := range lines {
				if lipgloss.Width(ln) > w {
					t.Errorf("%dx%d: line %d is %d cells wide", w, h, i, lipgloss.Width(ln))
				}
			}
		}
	}
	for _, p := range []pane{panePlans, paneDone, paneDetail} {
		// A window with no room for any row at all still leaves the offset
		// somewhere the pane can hold, so the keys never break it.
		m := press(paneModel(t, p), "G")
		for _, h := range []int{1, 2, 3} {
			m = sized(m, 120, h)
			m = press(m, "G", "j", "j", "ctrl+d", "g", "ctrl+u")
			if m.off[p] < 0 || m.off[p] > m.lastOff(p) {
				t.Errorf("pane %d at height %d sits at %d, its content allows %d", p, h, m.off[p], m.lastOff(p))
			}
		}
	}
	// A body with no newline at the end, and one whose last line is exactly
	// the inner width, both reach their end and stay inside the frame.
	for _, body := range []string{
		"# Plan\n\n### Task 1: One\n\n- [ ] **Step 1: Do it**\n\n" + strings.Repeat("\nA line of the body.\n", 60) + "last line, no newline",
		"# Plan\n\n### Task 1: One\n\n- [ ] **Step 1: Do it**\n\n" + strings.Repeat("\nA line of the body.\n", 60) + strings.Repeat("y", 82) + "\n",
	} {
		m := boardModel(t, body)
		m = press(m, "3", "0", "G")
		b := scrollBox(m, paneDetail)
		if !shows(m, paneDetail, m.lastOff(paneDetail)) {
			t.Errorf("a body ending %s shows %q after G, want the last window", tail(body), drawnFirst(m, paneDetail))
		}
		lines := m.detailLines(b.textW())
		drawn := paneRows(m, paneDetail)
		if got, want := strings.TrimRight(drawn[len(drawn)-1], " "), plain(strings.TrimRight(lines[len(lines)-1], " ")); got != want {
			t.Errorf("a body ending %q ends on screen on %q, its content ends on %q", tail(body), got, want)
		}
		for i, ln := range strings.Split(m.View(), "\n") {
			if lipgloss.Width(ln) > 120 {
				t.Errorf("a body ending %q draws line %d %d cells wide", tail(body), i, lipgloss.Width(ln))
			}
		}
	}
}

// tail names the end of a body in a message, so a failure says which one it
// was without printing the whole file.
func tail(body string) string {
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	return `"` + lines[len(lines)-1] + `"`
}

// A long list of specs with a group of untyped files under it. Folding that
// group open and shut changes the rows of pane [1] under the cursor, so the
// pane has to land on a window its new rows can fill and the screen has to
// show that window.
func TestFoldingTheGroupKeepsTheListReadable(t *testing.T) {
	files := map[string]string{}
	for i := range 40 {
		files[fmt.Sprintf(".acta/specs/2026-09-20-spec-%02d.md", i)] = "# Story " + fmt.Sprint(i) + "\n\nSome text.\n"
		files[fmt.Sprintf(".acta/2026-09-2%d-legacy.md", i%10)] = "# Legacy " + fmt.Sprint(i) + "\n\nSome text.\n"
	}
	cfg := treeCfg(t, files)
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := sized(New(cfg, b, true), 120, 40)
	m.render = func(md string, _ int) string { return md }
	m = press(m, "2", "G")
	if m.off[paneSpecs] == 0 {
		t.Fatal("the list did not overflow")
	}
	for _, open := range []bool{true, false, true} {
		m.groupOpen = open
		rows, sel, idx := m.slotOf(paneSpecs)
		if want := itemCount(cursorOf(rows, *sel, *idx)+1, len(rows)); footOf(t, m, paneSpecs) != want {
			t.Errorf("groupOpen %v: the pane writes %q, want %q", open, footOf(t, m, paneSpecs), want)
		}
		if !shows(m, paneSpecs, m.off[paneSpecs]) {
			t.Errorf("groupOpen %v: the pane shows %q, it sits at %d", open, drawnFirst(m, paneSpecs), m.off[paneSpecs])
		}
		if top := thumbAt(m, paneSpecs); top < 0 {
			t.Errorf("groupOpen %v: a long list draws no thumb", open)
		}
		if strings.TrimSpace(paneRows(m, paneSpecs)[0]) == "" {
			t.Errorf("groupOpen %v: the first line on screen is blank", open)
		}
	}
}
