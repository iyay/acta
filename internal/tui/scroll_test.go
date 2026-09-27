package tui

import (
	"fmt"
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
	return press(longModel(t), "]", fmt.Sprint(p+1))
}

// scrollBox is the box pane p is drawn in, the detail pane included; the
// view tests only need the two list panes, so they answer for the open one
// whenever they are asked about anything else.
func scrollBox(m Model, p pane) box {
	g := m.geometry()
	if !g.wide {
		return g.full
	}
	return []box{g.open, g.done, g.detail}[p]
}

// paneRows gives the lines pane p draws between its walls, padding included,
// so a test can take the words and the scrollbar cell apart.
func paneRows(m Model, p pane) []string {
	b := scrollBox(m, p)
	col := column(m.View(), b.x, b.w)
	out := make([]string, 0, b.inner)
	for i := range b.inner {
		y := b.y + 1 + i
		if y >= len(col) {
			break
		}
		out = append(out, strings.TrimSuffix(strings.TrimPrefix(plain(col[y]), "│"), "│"))
	}
	return out
}

// lastCell gives the cell at the right end of every line of a pane, which is
// where the scrollbar sits, and a space when the pane has none.
func lastCell(rows []string) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		c := []rune(r)
		if len(c) == 0 {
			out = append(out, " ")
			continue
		}
		out = append(out, string(c[len(c)-1]))
	}
	return out
}

// thumbSpan gives the first and the last line the scrollbar thumb covers, or
// two -1 when the pane draws no thumb at all.
func thumbSpan(rows []string) (top, bottom int) {
	top, bottom = -1, -1
	for i, c := range lastCell(rows) {
		if c != "█" {
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
func thumbAt(rows []string) int {
	top, _ := thumbSpan(rows)
	return top
}

// countOf gives the n/m a pane writes in its bottom border.
func countOf(t *testing.T, m Model, p pane) (at, total int) {
	t.Helper()
	b := scrollBox(m, p)
	col := column(m.View(), b.x, b.w)
	if b.y+b.h-1 >= len(col) {
		t.Fatalf("pane %d draws no bottom border", p)
	}
	text := strings.Trim(plain(col[b.y+b.h-1]), "─└┘ ")
	if _, err := fmt.Sscanf(text, "%d/%d", &at, &total); err != nil {
		t.Fatalf("pane %d writes no count: %q (%v)", p, text, err)
	}
	return at, total
}

// screenCounts gives the n/m every pane writes in its bottom border, so a test
// can say what the screen shows and not only what the model holds. A pane
// that draws no count reads as zeroes.
func screenCounts(m Model) [3][2]int {
	var out [3][2]int
	view := m.View()
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		b := scrollBox(m, p)
		if b.w < 2 || b.h < 2 {
			continue
		}
		col := column(view, b.x, b.w)
		if b.y+b.h-1 >= len(col) {
			continue
		}
		text := strings.Trim(plain(col[b.y+b.h-1]), "─└┘ ")
		_, _ = fmt.Sscanf(text, "%d/%d", &out[p][0], &out[p][1])
	}
	return out
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
	return strings.TrimSpace(strings.Repeat("  ", r.depth) + name + "  " + it.Title)
}

// isThatRow says whether a drawn line is that row and nothing else, cut to
// the cells it had room for.
func isThatRow(m Model, r row, line string) bool {
	rest := strings.TrimRight(line, " ░█")
	return rest == cutTo(headOf(m, r), lipgloss.Width(rest))
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
			t.Fatalf("bar has %d cells, want %d", len(bar), c.vis)
		}
		if (bar[0] == "█") != c.wantTop || (bar[c.vis-1] == "█") != c.wantBottom {
			t.Errorf("%+v: bar %v", c, bar)
		}
	}
	if len(scrollbar(5, 10, 0, 10)) != 0 {
		t.Error("no scrollbar when the content fits")
	}
}

func TestScrollbarThumbSitsBetweenTheEnds(t *testing.T) {
	// Half way down, the thumb sits inside the track, never on either end.
	bar := scrollbar(100, 10, 45, 10)
	if len(bar) != 10 {
		t.Fatalf("bar has %d cells, want 10", len(bar))
	}
	if bar[0] == "█" || bar[9] == "█" {
		t.Errorf("the thumb jumped to an end: %v", bar)
	}
	if n := strings.Count(strings.Join(bar, ""), "█"); n != 1 {
		t.Errorf("one cell should be the thumb: %v", bar)
	}
	if n := strings.Count(strings.Join(bar, ""), "░"); n != 9 {
		t.Errorf("the rest of the column should be track: %v", bar)
	}
	// A window that covers half the content gets a thumb of half the
	// column, so the thumb never lies about how much is on screen.
	bar = scrollbar(100, 50, 25, 10)
	if n := strings.Count(strings.Join(bar, ""), "█"); n != 5 {
		t.Errorf("half of the content should give a thumb of 5 cells: %v", bar)
	}
	if bar[1] != "░" || bar[2] != "█" || bar[6] != "█" || bar[7] != "░" {
		t.Errorf("the thumb should sit in the middle of the column: %v", bar)
	}
}

func TestScrollbarSurvivesOddNumbers(t *testing.T) {
	// An offset past the end reads as the end, and one before the start reads
	// as the start: a stale offset cannot leave the thumb off the column.
	for _, c := range []struct {
		first    int
		wantCell int
	}{
		{-5, 0}, {0, 0}, {45, 4}, {90, 9}, {999, 9},
	} {
		bar := scrollbar(100, 10, c.first, 10)
		if len(bar) != 10 {
			t.Fatalf("first %d: bar has %d cells, want 10", c.first, len(bar))
		}
		if n := strings.Count(strings.Join(bar, ""), "█"); n != 1 {
			t.Errorf("first %d: one cell should be the thumb: %v", c.first, bar)
		}
		if bar[c.wantCell] != "█" {
			t.Errorf("first %d: the thumb should sit on cell %d: %v", c.first, c.wantCell, bar)
		}
	}
	// A pane with no line to show still draws a column instead of crashing.
	if bar := scrollbar(10, 0, 0, 10); len(bar) != 10 {
		t.Errorf("a pane with no visible line has %d cells, want 10", len(bar))
	}
	// Content that exactly fills the pane needs no scrollbar, and neither
	// does a pane too short to draw one.
	if len(scrollbar(10, 10, 0, 10)) != 0 || len(scrollbar(10, 4, 0, 0)) != 0 {
		t.Error("a scrollbar where there is nothing to scroll")
	}
}

func TestScrollbarShowsOnlyOnOverflow(t *testing.T) {
	// The fixture fits in every pane, so no pane draws a scrollbar.
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		if at := thumbAt(paneRows(newModel(t), p)); at != -1 {
			t.Errorf("pane %d draws a scrollbar for content that fits, thumb on line %d", p, at)
		}
	}
	// The long board overflows in all three panes, and each one says so with
	// a scrollbar sitting at the top of the column.
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		if at := thumbAt(paneRows(paneModel(t, p), p)); at != 0 {
			t.Errorf("pane %d draws no scrollbar at the top, thumb on line %d", p, at)
		}
	}
}

func TestWheelScrollsOnlyTheFocusedPane(t *testing.T) {
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		m := paneModel(t, p)
		before := m.Selected()
		b := scrollBox(m, p)
		m = wheel(m, b.x+1, b.y+2, false)
		if m.off[p] != 1 {
			t.Errorf("pane %d: one notch should scroll it by one, off is %d", p, m.off[p])
		}
		for _, q := range []pane{paneOpen, paneDone, paneDetail} {
			if c := screenCounts(m)[q]; q != p && c[0] != 1 {
				t.Errorf("the wheel over pane %d moved pane %d to line %d on screen", p, q, c[0])
			}
		}
		// The wheel scrolls a pane, it never picks another item.
		if p != paneDetail {
			if got := m.Selected(); got == nil || before == nil || got.ID != before.ID {
				t.Errorf("the wheel over pane %d changed the selection to %v", p, got)
			}
		}
		// The count on screen follows the line that is on screen.
		if at, _ := countOf(t, m, p); at != 2 {
			t.Errorf("pane %d: after one notch the count should read 2, reads %d", p, at)
		}
		m = wheel(m, b.x+1, b.y+2, true)
		if m.off[p] != 0 {
			t.Errorf("pane %d: one notch back should be at the top, off is %d", p, m.off[p])
		}
		if at, _ := countOf(t, m, p); at != 1 {
			t.Errorf("pane %d: back at the top the count should read 1, reads %d", p, at)
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
		at, total := countOf(t, m, p)
		if at != total-b.inner+1 {
			t.Errorf("pane %d: at the end the count reads %d/%d, want %d", p, at, total, total-b.inner+1)
		}
		if strings.TrimSpace(paneRows(m, p)[0]) == "" {
			t.Errorf("pane %d: the wheel left the first line blank", p)
		}
	}
}

// A pane with less content than it has rows has nowhere to scroll, so every
// key and every notch leaves it at the top. An offset below zero would count
// from a line that does not exist.
func TestAPaneShorterThanItsContentNeverScrolls(t *testing.T) {
	m := sized(newModel(t), 120, 40)
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
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
		// Nothing on screen says the pane moved either: no count, no thumb,
		// and the first line of the content is still the first line drawn.
		if c := screenCounts(m)[p]; c != [2]int{} {
			t.Errorf("pane %d: content that fits writes a count %v", p, c)
		}
		if at := thumbAt(paneRows(m, p)); at != -1 {
			t.Errorf("pane %d: content that fits draws a thumb on line %d", p, at)
		}
		if p == paneDetail {
			lines := m.detailLines(m.textOf(p, scrollBox(m, p)))
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
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		for _, q := range []pane{paneOpen, paneDone, paneDetail} {
			if q == p {
				continue
			}
			// The pane under the pointer is scrolled away from the top first,
			// so a wheel that ignored the focus would show on screen instead of
			// hiding behind a clamp at zero.
			m := press(paneModel(t, p), fmt.Sprint(p+1), "G")
			m = press(m, fmt.Sprint(q+1))
			sel := m.Selected()
			before := m.off
			shown := screenCounts(m)
			if before[p] == 0 || shown[p][0] == 1 {
				t.Fatalf("pane %d did not scroll before the wheel over it", p)
			}
			g := m.geometry()
			target := []box{g.open, g.done, g.detail}[p]
			if target.w < 2 || target.h < 2 {
				continue
			}
			// Both notches, so neither direction of the wheel can slip past.
			for _, up := range []bool{false, true} {
				m = wheel(m, target.x+1, target.y+2, up)
				if m.focus != q {
					t.Errorf("a wheel over pane %d took the focus off pane %d", p, q)
				}
				if m.off != before {
					t.Errorf("a wheel over unfocused pane %d moved something: %v, want %v", p, m.off, before)
				}
				if screenCounts(m) != shown {
					t.Errorf("a wheel over unfocused pane %d moved the screen: %v, want %v", p, screenCounts(m), shown)
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
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		m := paneModel(t, paneOpen)
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
	// The keys of pane [3] scroll pane [3] and leave both lists where they
	// were, selection and offset.
	m := press(longModel(t), "]", "1", "G")
	before := screenCounts(m)
	m = press(m, "3", "ctrl+d", "ctrl+d")
	if m.off[paneDetail] != 2*pageLines {
		t.Fatalf("two pages should scroll the detail to %d, off is %d", 2*pageLines, m.off[paneDetail])
	}
	// The screen says the same: pane [3] moved, the two lists did not.
	after := screenCounts(m)
	if after[paneDetail][0] != before[paneDetail][0]+2*pageLines {
		t.Errorf("pane [3] reads %d, it read %d before the keys", after[paneDetail][0], before[paneDetail][0])
	}
	if after[paneOpen] != before[paneOpen] || after[paneDone] != before[paneDone] {
		t.Errorf("the keys in pane [3] moved a list on screen: %v, want %v", after, before)
	}
	// Coming back to a list finds it exactly where it was left.
	m = press(m, "1")
	if m.off[paneOpen] != before[paneOpen][0]-1 {
		t.Errorf("pane [1] lost its place: the screen reads %d, was %d", screenCounts(m)[paneOpen][0], before[paneOpen][0])
	}
	// A key in pane [1] moves pane [1] and no other pane. Another item
	// starts the detail at its own top, which is the one place a key in a
	// list reaches into pane [3].
	m = press(m, "k")
	after = screenCounts(m)
	if after[paneDone] != before[paneDone] {
		t.Errorf("k in pane [1] moved pane [2] on screen: %v, want %v", after[paneDone], before[paneDone])
	}
	if after[paneDetail][0] != 1 {
		t.Errorf("another item should start the detail at the top, the screen reads %d", after[paneDetail][0])
	}
	// Walking to the top of a list scrolls it back up, and still no other
	// pane moves.
	m = press(m, "g")
	after = screenCounts(m)
	if after[paneOpen][0] != 1 {
		t.Errorf("g should take pane [1] back to its first row, the screen reads %d", after[paneOpen][0])
	}
	if after[paneDone] != before[paneDone] || after[paneDetail][0] != 1 {
		t.Errorf("g in pane [1] moved another pane on screen: %v", after)
	}
}

func TestListKeepsTheSelectedRowVisible(t *testing.T) {
	m := press(longModel(t), "]", "1")
	rows, _, _ := m.slotOf(paneOpen)
	// Walk down until the cursor sits below the fold, then keep walking: the
	// pane has to follow the cursor, not the other way round.
	for m.cursor() < scrollBox(m, paneOpen).rows {
		m = press(m, "j")
	}
	below := m.cursor()
	m = press(m, "j")
	if m.cursor() == below {
		t.Fatal("the list should still move down")
	}
	b := scrollBox(m, paneOpen)
	cur := m.cursor()
	if cur < b.first || cur >= b.first+b.rows {
		t.Fatalf("row %d is off screen: the pane shows rows %d to %d", cur, b.first, b.first+b.rows-1)
	}
	// The row the count says is on screen is the row that is on screen.
	at, total := countOf(t, m, paneOpen)
	if at != b.first+1 || total != len(rows) {
		t.Fatalf("the count reads %d/%d, the pane starts at row %d of %d", at, total, b.first, len(rows))
	}
	drawn := paneRows(m, paneOpen)
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
	if m.cursor() < scrollBox(m, paneOpen).first {
		t.Error("k moved the cursor off the top of the pane")
	}
}

func TestSelectingAnotherItemPutsTheDetailBackAtTheTop(t *testing.T) {
	m := press(longModel(t), "]", "3")
	for screenCounts(m)[paneDetail][0] < 6 {
		m = press(m, "j")
	}
	if m.off[paneDetail] == 0 {
		t.Fatal("the detail never scrolled")
	}
	if at, _ := countOf(t, m, paneDetail); at != m.off[paneDetail]+1 {
		t.Errorf("the detail count reads %d at offset %d", at, m.off[paneDetail])
	}
	m = press(m, "1", "j")
	if m.off[paneDetail] != 0 {
		t.Errorf("another item should start the detail at the top, off is %d", m.off[paneDetail])
	}
	if at, _ := countOf(t, m, paneDetail); at != 1 {
		t.Errorf("the detail count should read 1 at the top, reads %d", at)
	}
	// The line on screen is the first line of the body again, and the thumb
	// is back at the top of the column.
	if !isFirstLine(m, paneDetail, paneRows(m, paneDetail)[0]) {
		t.Errorf("the detail does not show its first line: %q", paneRows(m, paneDetail)[0])
	}
	if top := thumbAt(paneRows(m, paneDetail)); top != 0 {
		t.Errorf("the detail thumb sits on line %d after picking another item, want 0", top)
	}
}

func TestThumbAndCountFollowTheOffsetOnScreen(t *testing.T) {
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		m := paneModel(t, p)
		// Top: the thumb is on the first line and the count reads 1.
		at, total := countOf(t, m, p)
		if at != 1 || total != m.linesOf(p) {
			t.Fatalf("pane %d at the top reads %d/%d, want 1/%d", p, at, total, m.linesOf(p))
		}
		if top := thumbAt(paneRows(m, p)); top != 0 {
			t.Errorf("pane %d at the top has its thumb on line %d, want 0", p, top)
		}
		// Middle: the thumb leaves both ends and the count moves with it. A
		// list only follows its cursor once the cursor leaves the window, so
		// three pages are needed to get off the first one.
		m = press(m, "ctrl+d", "ctrl+d", "ctrl+d")
		first := m.off[p]
		if first <= 0 || first >= m.lastOff(p) {
			t.Fatalf("pane %d sits at %d of %d, which is neither the middle nor the end", p, first, m.lastOff(p))
		}
		b := scrollBox(m, p)
		if top, bottom := thumbSpan(paneRows(m, p)); top <= 0 || bottom >= b.inner-1 {
			t.Errorf("pane %d in the middle has its thumb on lines %d to %d of %d", p, top, bottom, b.inner)
		}
		if at, _ := countOf(t, m, p); at != first+1 {
			t.Errorf("pane %d reads %d while its offset is %d", p, at, first)
		}
		// End: the thumb is on the last line and the count reads the last
		// window the pane can show.
		m = press(m, "G")
		last := m.lastOff(p)
		if m.off[p] != last {
			t.Errorf("pane %d at the end sits at %d of %d", p, m.off[p], last)
		}
		if _, bottom := thumbSpan(paneRows(m, p)); bottom != b.inner-1 {
			t.Errorf("pane %d at the end has its thumb ending on line %d, want %d", p, bottom, b.inner-1)
		}
		at, total = countOf(t, m, p)
		if at != last+1 || at != total-b.inner+1 {
			t.Errorf("pane %d at the end reads %d/%d, want %d/%d", p, at, total, last+1, total)
		}
		// The last line on screen is the last line of the content.
		drawn := paneRows(m, p)
		if p == paneDetail {
			lines := m.detailLines(m.textOf(p, b))
			if got, want := strings.TrimRight(drawn[len(drawn)-1], " ░█"), plain(strings.TrimRight(lines[len(lines)-1], " ")); got != want {
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
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		m := press(paneModel(t, p), "G")
		if m.off[p] == 0 {
			t.Fatalf("pane %d never scrolled", p)
		}
		// A window so tall that the pane holds all of its content: the place
		// the pane was at is gone, so the offset goes with it.
		m = sized(m, 120, 200)
		if m.off[p] != 0 {
			t.Errorf("pane %d kept the place %d with room for all of it", p, m.off[p])
		}
		if !isFirstLine(m, p, paneRows(m, p)[0]) {
			t.Errorf("pane %d does not start at its first line: %q", p, paneRows(m, p)[0])
		}
		if c := screenCounts(m)[p]; c != [2]int{} {
			t.Errorf("pane %d writes a count %v for content that fits", p, c)
		}
		if at := thumbAt(paneRows(m, p)); at != -1 {
			t.Errorf("pane %d draws a thumb on line %d for content that fits", p, at)
		}
		// A window so short that the pane shows a handful of lines still
		// draws its content, never a blank screen.
		m = sized(m, 120, 14)
		if m.off[p] > m.lastOff(p) || m.off[p] < 0 {
			t.Errorf("pane %d sits at %d, the pane shows %d of %d", p, m.off[p], m.fitOf(p), m.linesOf(p))
		}
		at, total := countOf(t, m, p)
		if at != m.off[p]+1 || at > total {
			t.Errorf("pane %d after the resize reads %d/%d at offset %d", p, at, total, m.off[p])
		}
		if strings.TrimSpace(paneRows(m, p)[0]) == "" {
			t.Errorf("pane %d draws a blank first line after the resize", p)
		}
		// The offset never leaves the pane, whichever way the keys go.
		m = press(m, "G", "j", "j", "ctrl+d")
		if m.off[p] > m.lastOff(p) || m.off[p] < 0 {
			t.Errorf("pane %d scrolled past its content: %d of %d", p, m.off[p], m.lastOff(p))
		}
		if at, total := countOf(t, m, p); at != total-scrollBox(m, p).inner+1 {
			t.Errorf("pane %d reads %d/%d after scrolling to the end, want %d", p, at, total, total-scrollBox(m, p).inner+1)
		}
		m = press(m, "g", "k", "k", "ctrl+u")
		if m.off[p] != 0 {
			t.Errorf("pane %d scrolled above its top: %d", p, m.off[p])
		}
		if at, _ := countOf(t, m, p); at != 1 {
			t.Errorf("pane %d reads %d after scrolling back to the top, want 1", p, at)
		}
		// And growing the window again takes the pane back to the top.
		m = press(m, "G")
		m = sized(m, 120, 200)
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
// G stops a line short of the end the screen can show and the n/m on screen
// reads a line that never arrives.
func TestCountReachesTheLastWindowWithAWrappingBody(t *testing.T) {
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		m := press(wrapModel(t), "]", fmt.Sprint(p+1), "G")
		b := scrollBox(m, p)
		at, total := countOf(t, m, p)
		if at != total-b.inner+1 {
			t.Errorf("pane %d: after G the count reads %d/%d, want %d", p, at, total, total-b.inner+1)
		}
		if _, bottom := thumbSpan(paneRows(m, p)); bottom != b.inner-1 {
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
		return strings.TrimRight(line, " ░█") == plain(strings.TrimRight(lines[0], " "))
	}
	rows, _, _ := m.slotOf(p)
	return len(rows) > 0 && isThatRow(m, rows[0], line)
}

func TestPanesKeepTheirOwnPlace(t *testing.T) {
	m := press(longModel(t), "]", "1", "G")
	open := screenCounts(m)[paneOpen]
	m = press(m, "2", "G")
	done := screenCounts(m)[paneDone]
	if done[0] == 1 {
		t.Fatal("pane [2] did not scroll")
	}
	if screenCounts(m)[paneOpen] != open {
		t.Errorf("pane [1] moved to %v while pane [2] scrolled, was %v", screenCounts(m)[paneOpen], open)
	}
	m = press(m, "3", "G")
	after := screenCounts(m)
	if after[paneOpen] != open || after[paneDone] != done {
		t.Errorf("pane [3] moved the lists: %v, want %v and %v", after, open, done)
	}
	if after[paneDetail][0] == 1 {
		t.Error("pane [3] did not scroll")
	}
	// Each pane has to end on the last line its own content allows, not on a
	// line that belongs to some other pane's body.
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
		c := screenCounts(m)[p]
		if c[0] != c[1]-scrollBox(m, p).inner+1 {
			t.Errorf("pane %d ends on line %d of %d, the last window starts at %d", p, c[0], c[1], c[1]-scrollBox(m, p).inner+1)
		}
	}
	m = press(m, "1")
	if screenCounts(m)[paneOpen] != open {
		t.Errorf("coming back to pane [1] lost its place: %v, want %v", screenCounts(m)[paneOpen], open)
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
			m := press(longModel(t), "]", "3", "G")
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
	for _, p := range []pane{paneOpen, paneDone, paneDetail} {
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
		m = press(m, "]", "3", "G")
		b := scrollBox(m, paneDetail)
		at, total := countOf(t, m, paneDetail)
		if at != total-b.inner+1 {
			t.Errorf("a body ending %q reaches %d of %d after G", tail(body), at, total)
		}
		lines := m.detailLines(m.textOf(paneDetail, b))
		drawn := paneRows(m, paneDetail)
		if got, want := strings.TrimRight(drawn[len(drawn)-1], " ░█"), plain(strings.TrimRight(lines[len(lines)-1], " ")); got != want {
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
// pane has to land on a window its new rows can fill and the n/m has to match
// that window.
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
	m = press(m, "1", "G")
	if m.off[paneOpen] == 0 {
		t.Fatal("the list did not overflow")
	}
	for _, open := range []bool{true, false, true} {
		m.groupOpen = open
		rows, _, _ := m.slotOf(paneOpen)
		at, total := countOf(t, m, paneOpen)
		if at < 1 || at > total || total != len(rows) {
			t.Errorf("groupOpen %v: the count reads %d/%d, the list holds %d rows", open, at, total, len(rows))
		}
		if at != m.off[paneOpen]+1 {
			t.Errorf("groupOpen %v: the count reads %d, the pane sits at %d", open, at, m.off[paneOpen])
		}
		if top := thumbAt(paneRows(m, paneOpen)); top < 0 {
			t.Errorf("groupOpen %v: a long list draws no scrollbar", open)
		}
		if strings.TrimSpace(paneRows(m, paneOpen)[0]) == "" {
			t.Errorf("groupOpen %v: the first line on screen is blank", open)
		}
	}
}
