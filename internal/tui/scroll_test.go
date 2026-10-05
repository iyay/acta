package tui

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

// keyTo focuses a box of the open tab, the way a reader gets there.
func keyTo(p pane) string {
	switch p {
	case paneDetail:
		return "shift+tab"
	case paneDone:
		return "tab"
	}
	return tabKey(tabPlans)
}

// focusKeyFrom gives the key that moves the focus from one box of the open
// tab to another, so a test never has to count the ring itself.
func focusKeyFrom(from, to pane) string {
	if to == paneDetail {
		// The ring is List, Done, Detail: the detail is one step back from
		// the list and one step on from Done.
		if from == paneDone {
			return "tab"
		}
		return "shift+tab"
	}
	// The ring is List, Done, Detail, so walking forward from the List
	// reaches Done in one step and the detail box in two.
	steps := 1
	if to == paneDone {
		steps = 2
	}
	if from == paneDetail {
		steps = 3 - steps
	}
	if steps%2 == 1 {
		return "shift+tab"
	}
	return "tab"
}

// paneModel is the long board with box p focused and the Plans tab open,
// where both list boxes have rows to show.
func paneModel(t *testing.T, p pane) Model {
	t.Helper()
	// The plans list has rows to walk; the detail box needs one of them
	// selected before it has a body to scroll.
	m := press(longModel(t), tabKey(tabPlans))
	return press(m, keyTo(p))
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

// footOf gives the counter a pane writes in its bottom border, with the
// border and the sort word cut off, and nothing at all on the detail box.
func footOf(t *testing.T, m Model, p pane) string {
	t.Helper()
	foot := strings.Trim(bottomLine(t, m, p), "─└┘ ")
	if _, count, ok := strings.Cut(foot, " · "); ok {
		return count
	}
	return foot
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
// claims a place it has not drawn is caught here. The detail box keeps its
// header on top and its date line at the bottom when it has the room, so
// there the line of the index is drawn under the header, not on the top.
func shows(m Model, p pane, i int) bool {
	if p == paneDetail {
		b := scrollBox(m, p)
		head, mid, foot := m.detailParts(b.textW())
		if n := stickyMid(len(head), b.inner); n > 0 && foot != "" {
			first := min(max(i, 0), max(0, len(mid)-n))
			rows := paneRows(m, p)
			return i >= 0 && i < len(mid) && len(head) < len(rows) &&
				strings.TrimSpace(rows[len(head)]) == strings.TrimSpace(mid[first])
		}
		lines := m.detailLines(b.textW())
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	// The fixture fits in every box of a tall enough screen, so none of them
	// draws a thumb.
	tall := press(sized(newModel(t), 120, 80), tabKey(tabPlans))
	for _, p := range []pane{paneList, paneDone, paneDetail} {
		if at := thumbAt(tall, p); at != -1 {
			t.Errorf("pane %d draws a thumb for content that fits, on line %d", p, at)
		}
	}
	// The long board overflows in all three, and each one says so with a
	// thumb sitting at the top of its own wall.
	for _, p := range []pane{paneList, paneDone, paneDetail} {
		// The detail box needs an item under it before it has a body to
		// scroll, so it is read from the same board with one selected.
		m := paneModel(t, p)
		if p == paneDetail {
			m = press(paneModel(t, paneList), "shift+tab")
		}
		if at := thumbAt(m, p); at != 0 {
			t.Errorf("pane %d draws no thumb at the top, thumb on line %d", p, at)
		}
	}
}

func TestWheelScrollsOnlyTheFocusedPane(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		m := paneModel(t, p)
		before := m.Selected()
		shown := screenTops(m)
		b := scrollBox(m, p)
		m = wheel(m, b.x+1, b.y+2, false)
		step := min(wheelStep, m.lastOff(p))
		if m.off[p] != step {
			t.Errorf("pane %d: one notch should scroll it by %d, off is %d", p, step, m.off[p])
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
		if !shows(m, p, step) {
			t.Errorf("pane %d: after one notch the top of the box should be line %d, it shows %q", p, step, drawnFirst(m, p))
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

// A trackpad sends hundreds of notches a second. Drawing the screen after
// each one left the TUI far behind the wheel, so the notches of one frame are
// gathered and scrolled once. The first notch of a frame scrolls at once, and
// only the notches after it wait for the tick. A tick with nothing gathered
// must move nothing.
func TestWheelNotchesInOneFrameScrollOnce(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		m := paneModel(t, p)
		b := scrollBox(m, p)
		m = wheelOnly(m, b.x+1, b.y+2, false)
		first := min(wheelStep, m.lastOff(p))
		if m.off[p] != first {
			t.Errorf("pane %d: the first notch should scroll %d at once, off is %d", p, first, m.off[p])
		}
		m = wheelOnly(m, b.x+1, b.y+2, false)
		if m.off[p] != first {
			t.Errorf("pane %d: a gathered notch must wait for the frame tick, off is %d", p, m.off[p])
		}
		m = wheelTick(m)
		if want := min(2*wheelStep, m.lastOff(p)); m.off[p] != want {
			t.Errorf("pane %d: the gathered notch should take the pane to %d, off is %d", p, want, m.off[p])
		}
		// A second tick with nothing gathered must not move the pane again.
		before := m.off[p]
		m = wheelTick(m)
		if m.off[p] != before {
			t.Errorf("pane %d: an empty tick moved the pane from %d to %d", p, before, m.off[p])
		}
	}
}

// Scrolling must start on the notch itself. Waiting for the frame tick made
// every scroll start one frame late.
func TestWheelFirstNotchScrollsAtOnce(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		m := paneModel(t, p)
		b := scrollBox(m, p)
		m = wheelOnly(m, b.x+1, b.y+2, false)
		if want := min(wheelStep, m.lastOff(p)); m.off[p] != want {
			t.Errorf("pane %d: the first notch should scroll %d at once, off is %d", p, want, m.off[p])
		}
	}
}

// While scrolling goes on, the tick must come back each frame. When the wheel
// has stopped, the tick must stop too, so no tick runs for nothing.
func TestWheelTickRearmsOnlyWhileScrolling(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	next, cmd := m.Update(wheelTickMsg{})
	if cmd == nil {
		t.Fatal("a tick that scrolled must arm the next tick")
	}
	m = next.(Model)
	next, cmd = m.Update(wheelTickMsg{})
	if cmd != nil {
		t.Error("a tick with nothing gathered must not arm another tick")
	}
	m = next.(Model)
	before := m.off[paneDetail]
	m = wheelOnly(m, b.x+1, b.y+2, false)
	if m.off[paneDetail] == before {
		t.Error("the first notch after the wheel stopped must scroll at once")
	}
}

// Two ticks inside the time the renderer takes for two writes at 120fps
// keep the gap between writes at 16.7 ms or less.
func TestWheelFrameFitsTwoRendererWrites(t *testing.T) {
	if wheelFrame != 12*time.Millisecond {
		t.Errorf("wheelFrame is %v, the spec says 12ms", wheelFrame)
	}
}

// A wheel turned back and forth inside one frame ends where it started, so a
// reader who changes their mind leaves the pane where it was.
func TestWheelUpAndDownInOneFrameCancel(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	// The first notch of a frame scrolls at once, so the frame that cancels is
	// the next one: its first notch scrolls, and the down and the up after it
	// are gathered and cancel each other out.
	m = wheel(m, b.x+1, b.y+2, false)
	start := m.off[paneDetail]
	m = wheelOnly(m, b.x+1, b.y+2, false)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	m = wheelOnly(m, b.x+1, b.y+2, true)
	if want := start + wheelStep; m.off[paneDetail] != want {
		t.Fatalf("the gathered down and up should cancel before the tick, off is %d, want %d", m.off[paneDetail], want)
	}
	m = wheelTick(m)
	if want := start + wheelStep; m.off[paneDetail] != want {
		t.Errorf("down then up in one frame should leave the pane at %d, it is at %d", want, m.off[paneDetail])
	}
}

// Only the first notch of a frame starts a tick, so a fast wheel does not
// leave a queue of ticks behind that scroll the pane after the wheel stopped.
func TestWheelArmsOneTickPerFrame(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	b := scrollBox(m, paneList)
	next, first := m.Update(tea.MouseMsg{X: b.x + 1, Y: b.y + 2, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	if first == nil {
		t.Fatal("the first notch of a frame must start the frame tick")
	}
	if want := min(wheelStep, next.(Model).lastOff(paneList)); next.(Model).off[paneList] != want {
		t.Errorf("the first notch should scroll %d at once, off is %d", want, next.(Model).off[paneList])
	}
	_, second := next.Update(tea.MouseMsg{X: b.x + 1, Y: b.y + 2, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	if second != nil {
		t.Error("a notch in a frame that already has a tick must not start another one")
	}
}

// Whatever the wheel gathers, the scroll lands inside the content. A delta
// that runs past the end of the pane stops at the last line it has.
func TestWheelOverclampStaysInsideContent(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		m := paneModel(t, p)
		b := scrollBox(m, p)
		for range 500 {
			m = wheelOnly(m, b.x+1, b.y+2, false)
		}
		m = wheelTick(m)
		if want := m.lastOff(p); m.off[p] != want {
			t.Errorf("pane %d: 500 notches in one frame scrolled to %d, the content allows %d", p, m.off[p], want)
		}
		for range 500 {
			m = wheelOnly(m, b.x+1, b.y+2, true)
		}
		m = wheelTick(m)
		if m.off[p] != 0 {
			t.Errorf("pane %d: 500 notches up in one frame scrolled to %d, the top is 0", p, m.off[p])
		}
	}
}

// One delta belongs to one pane. A notch on the next pane starts a fresh
// frame for it, and the notches the reader already gave the old pane scroll
// that pane rather than the new one.
func TestWheelOnAnotherPaneKeepsEachNotchOnItsOwnPane(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	lb := scrollBox(m, paneList)
	m = wheelOnly(m, lb.x+1, lb.y+2, false)
	// The reader moves to the Done box and turns the wheel there, both before
	// the tick that closes the first frame.
	m = press(m, keyTo(paneDone))
	db := scrollBox(m, paneDone)
	m = wheelOnly(m, db.x+1, db.y+2, false)
	m = wheelTick(m)
	if want := min(wheelStep, m.lastOff(paneList)); m.off[paneList] != want {
		t.Errorf("the notch over the List pane scrolled it to %d, want %d", m.off[paneList], want)
	}
	if want := min(wheelStep, m.lastOff(paneDone)); m.off[paneDone] != want {
		t.Errorf("the notch over the Done pane scrolled it to %d, want %d", m.off[paneDone], want)
	}
}

// The wheel is off while the help sits over the panes, so a notch there
// gathers nothing and the pane underneath stays where the reader left it.
func TestWheelWhileHelpIsOpenGathersNothing(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m.help = true
	m = wheel(m, b.x+1, b.y+2, false)
	if m.off[paneDetail] != 0 || m.wheelDelta != 0 {
		t.Errorf("the wheel under the help popup moved the pane: off %d, pending %d", m.off[paneDetail], m.wheelDelta)
	}
}

// A popup, a slug and a search each sit over the panes and take the wheel
// themselves, so a notch there must not move a pane at all, not even on the
// first notch of a frame, which scrolls at once everywhere else.
func TestWheelUnderAPopupSlugOrSearchGathersNothing(t *testing.T) {
	t.Parallel()

	slug := "bug-"
	for _, c := range []struct {
		name string
		open func(m Model) Model
	}{
		{
			name: "a popup",
			open: func(m Model) Model {
				m.popup = &popup{field: "status", options: []string{"open", "done"}, idx: 0}
				return m
			},
		},
		{
			name: "a slug",
			open: func(m Model) Model { m.slug = &slug; return m },
		},
		{
			name: "a search",
			open: func(m Model) Model { m.searching = true; return m },
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := paneModel(t, paneDetail)
			b := scrollBox(m, paneDetail)
			m = c.open(m)
			m = wheelOnly(m, b.x+1, b.y+2, false)
			if m.off[paneDetail] != 0 || m.wheelDelta != 0 {
				t.Errorf("the wheel under %s moved the pane: off %d, pending %d", c.name, m.off[paneDetail], m.wheelDelta)
			}
			m = wheelTick(m)
			if m.off[paneDetail] != 0 {
				t.Errorf("the tick under %s moved the pane to %d", c.name, m.off[paneDetail])
			}
		})
	}
}

// One notch moves three lines, the usual terminal step, so a scroll covers
// ground. A step of one leaves the reader creeping down the pane.
func TestWheelNotchMovesThreeLines(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	b := scrollBox(m, paneList)
	m = wheel(m, b.x+1, b.y+2, false)
	if m.off[paneList] != 3 {
		t.Errorf("one notch should scroll three lines, off is %d", m.off[paneList])
	}
}

// longTabs puts forty items with a body long enough to scroll in every list of
// the Plans and the Bugs tab, under the open, finished and closed status, so a
// tab switch and a Done sub-tab switch both land on lists that overflow and on
// a detail box with room to move.
func longTabs(t *testing.T) Model {
	t.Helper()
	body := "# Item " + strings.Repeat("x", 40) + "\n" + strings.Repeat("\nA line of the body.\n", 60)
	files := map[string]string{}
	for _, f := range []struct{ dir, status string }{
		{"plans", ""},
		{"plans", "done"},
		{"plans", "dropped"},
		{"bugs", ""},
		{"bugs", "fixed"},
		{"bugs", "wontfix"},
	} {
		front := ""
		if f.status != "" {
			front = "---\nstatus: " + f.status + "\n---\n"
		}
		for i := range 40 {
			files[fmt.Sprintf(".acta/%s/2026-09-20-%s-%02d.md", f.dir, f.status, i)] = front + body
		}
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

// The notches of one frame belong to the screen the wheel turned on. Every
// path that moves the item on show, the open tab or the Done sub-tab before the
// tick must leave the new item at the top of its own box, and the notches of a
// frame where nothing changed must still scroll.
func TestWheelNotchesBelongToTheScreenTheyWereGatheredOn(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name  string
		start func(t *testing.T) Model
		box   pane
		leave func(m Model) Model
	}{
		{
			name: "a click on another row",
			start: func(t *testing.T) Model {
				return paneModel(t, paneDetail)
			},
			box: paneDetail,
			leave: func(m Model) Model {
				b := scrollBox(m, paneList)
				return click(m, b.x+1, b.y+2)
			},
		},
		{
			name: "a key that moves the cursor",
			start: func(t *testing.T) Model {
				return paneModel(t, paneDetail)
			},
			box:   paneDetail,
			leave: func(m Model) Model { return press(m, "esc", "j") },
		},
		{
			name: "another tab",
			start: func(t *testing.T) Model {
				return press(longTabs(t), tabKey(tabPlans), "0")
			},
			box:   paneDetail,
			leave: func(m Model) Model { return press(m, tabKey(tabBugs)) },
		},
		{
			name: "another Done sub-tab",
			start: func(t *testing.T) Model {
				return press(longTabs(t), tabKey(tabPlans), keyTo(paneDone))
			},
			box:   paneDone,
			leave: func(m Model) Model { return press(m, "]") },
		},
		{
			name: "a search for another item",
			start: func(t *testing.T) Model {
				return paneModel(t, paneDetail)
			},
			box:   paneDetail,
			leave: func(m Model) Model { return press(m, "/", "open-37") },
		},
		{
			name: "a reload that took the item away",
			start: func(t *testing.T) Model {
				return paneModel(t, paneDetail)
			},
			box:   paneDetail,
			leave: func(m Model) Model { return reloadedWithout(m, m.Selected().ID) },
		},
		{
			name: "nothing changed",
			start: func(t *testing.T) Model {
				return paneModel(t, paneDetail)
			},
			box:   paneDetail,
			leave: func(m Model) Model { return m },
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := c.start(t)
			b := scrollBox(m, c.box)
			// A first notch scrolls at once and a tick with nothing gathered
			// closes the frame, so it takes two notches and a tick to open a
			// frame that stays open. The two notches after them are the
			// gathered ones this case is about.
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m = wheelTick(m)
			opened := m.off[c.box]
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m = wheelOnly(m, b.x+1, b.y+2, false)
			if m.off[c.box] != opened {
				t.Fatalf("a gathered notch must wait for the frame tick, off is %d", m.off[c.box])
			}
			m = c.leave(m)
			// The change may have put the box back at the top, so the notches
			// are compared with where the box stood after it.
			before := m.off[c.box]
			m = wheelTick(m)
			if c.name == "nothing changed" {
				// The frame the reader never left must still scroll.
				if want := min(4*wheelStep, m.lastOff(c.box)); m.off[c.box] != want {
					t.Errorf("notches with nothing changed scrolled %d, want %d", m.off[c.box], want)
				}
				return
			}
			if m.off[c.box] != before {
				t.Errorf("the notches of the old screen scrolled the new one to %d, it stood at %d", m.off[c.box], before)
			}
			if !shows(m, c.box, before) {
				t.Errorf("the box shows %q, want line %d of what is on show", drawnFirst(m, c.box), before)
			}
		})
	}
}

// Notches that come in after the screen changed belong to the new screen, and
// the ones that came before it are dropped. A frame that gathers on both sides
// of a change scrolls what the new screen earned, no more.
func TestWheelNotchesAfterTheChangeBelongToTheNewScreen(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name  string
		start func(t *testing.T) Model
		box   pane
		leave func(m Model) Model
	}{
		{
			name: "a reload that took the item away",
			start: func(t *testing.T) Model {
				return paneModel(t, paneDetail)
			},
			box: paneDetail,
			leave: func(m Model) Model {
				return reloadedWithout(m, m.Selected().ID)
			},
		},
		{
			name: "another tab",
			start: func(t *testing.T) Model {
				// Both tabs remember the detail box as the focused one, so
				// the notch after the switch still lands on a detail box.
				return press(longTabs(t), tabKey(tabBugs), keyTo(paneDetail),
					tabKey(tabPlans), keyTo(paneDetail))
			},
			box:   paneDetail,
			leave: func(m Model) Model { return press(m, tabKey(tabBugs)) },
		},
		{
			name: "another Done sub-tab",
			start: func(t *testing.T) Model {
				return press(longTabs(t), tabKey(tabPlans), keyTo(paneDone))
			},
			box:   paneDone,
			leave: func(m Model) Model { return press(m, "]") },
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := c.start(t)
			b := scrollBox(m, c.box)
			// A first notch scrolls at once and a tick with nothing gathered
			// closes the frame, so it takes two notches and a tick to leave a
			// gathered notch behind. The change must drop that one.
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m = wheelTick(m)
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m = c.leave(m)
			// The pane on screen is where the next notch lands, which the
			// change may have moved.
			b = scrollBox(m, c.box)
			m = wheelOnly(m, b.x+1, b.y+2, false)
			m = wheelTick(m)
			want := min(wheelStep, m.lastOff(c.box))
			if m.off[c.box] != want {
				t.Errorf("one notch after the change scrolled %d, want %d", m.off[c.box], want)
			}
			if !shows(m, c.box, int(want)) {
				t.Errorf("the box shows %q, want line %d of what is on show", drawnFirst(m, c.box), want)
			}
		})
	}
}

// The next notch lands on the box under the pointer, so it belongs to that
// pane alone. The notches the reader gave the box they have already left are
// not scrolled there on the way out.
func TestWheelNotchOnAnotherPaneLeavesTheOldBoxWhereTheChangePutIt(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	b := scrollBox(m, paneDetail)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	m = wheelOnly(m, b.x+1, b.y+2, false)
	first := m.Selected().ID
	lb := scrollBox(m, paneList)
	m = click(m, lb.x+1, lb.y+2)
	if m.Selected().ID == first {
		t.Fatalf("the click stayed on %q, the test needs another item on show", first)
	}
	lb = scrollBox(m, paneList)
	m = wheelOnly(m, lb.x+1, lb.y+2, false)
	m = wheelTick(m)
	if m.off[paneDetail] != 0 {
		t.Errorf("the notches of the old item scrolled the new one to line %d, want 0", m.off[paneDetail])
	}
	if !shows(m, paneDetail, 0) {
		t.Errorf("the detail box shows %q, want the first line of the item on show", drawnFirst(m, paneDetail))
	}
	if want := min(wheelStep, m.lastOff(paneList)); m.off[paneList] != want {
		t.Errorf("the notch over the list scrolled it to %d, want %d", m.off[paneList], want)
	}
}

// A screen that never moved keeps its notches: the whole frame scrolls, or
// the wheel would swallow the turn the reader gave it.
func TestWheelNotchesScrollTheWholeFrameWhenNothingChanged(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name    string
		start   func(t *testing.T) Model
		box     pane
		notches int
		leave   func(m Model) Model
	}{
		{
			name:    "the list pane",
			start:   func(t *testing.T) Model { return paneModel(t, paneList) },
			box:     paneList,
			notches: 3,
		},
		{
			name:    "the detail pane",
			start:   func(t *testing.T) Model { return paneModel(t, paneDetail) },
			box:     paneDetail,
			notches: 2,
		},
		{
			name: "a reload that kept the item",
			start: func(t *testing.T) Model {
				m := paneModel(t, paneDetail)
				// The reader clicks the row they are on, so the cursor is
				// written down and the reload finds the same one again.
				lb := scrollBox(m, paneList)
				m = click(m, lb.x+1, lb.y+1)
				return press(m, keyTo(paneDetail))
			},
			box:     paneDetail,
			notches: 2,
			leave:   reloaded,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := c.start(t)
			b := scrollBox(m, c.box)
			for range c.notches {
				m = wheelOnly(m, b.x+1, b.y+2, false)
			}
			if c.leave != nil {
				m = c.leave(m)
			}
			m = wheelTick(m)
			want := min(c.notches*wheelStep, m.lastOff(c.box))
			if m.off[c.box] != want {
				t.Errorf("the frame scrolled %d, want the whole total %d", m.off[c.box], want)
			}
			if !shows(m, c.box, int(want)) {
				t.Errorf("the box shows %q, want line %d of what is on show", drawnFirst(m, c.box), want)
			}
		})
	}
}

// reloadedWithout is the reload message of a board the item id is gone from, so
// the cursor lands on another row the way a live reload lands it.
func reloadedWithout(m Model, id string) Model {
	fresh, err := m.load()
	if err != nil {
		return m
	}
	fresh.Items = slices.DeleteFunc(fresh.Items, func(it *board.Item) bool { return it.ID == id })
	next, _ := m.Update(reloadMsg{b: fresh})
	return next.(Model)
}

// reloaded is a reload message carrying the board as it stands, so the screen
// is rebuilt under the reader and the item on show stays the one they are on.
func reloaded(m Model) Model {
	fresh, err := m.load()
	if err != nil {
		return m
	}
	next, _ := m.Update(reloadMsg{b: fresh})
	return next.(Model)
}

// A tick closes the frame, and the notch after it opens the next one. A tick
// that left the frame armed forever would swallow every later notch, and the
// wheel would stop moving the pane.
func TestWheelScrollsAgainAfterTheFrameTick(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	b := scrollBox(m, paneList)
	first := wheel(m, b.x+1, b.y+2, false)
	next, cmd := first.Update(tea.MouseMsg{X: b.x + 1, Y: b.y + 2, Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	if cmd == nil {
		t.Fatal("a notch right after a tick must start the next frame")
	}
	// A second notch joins that frame, so the tick really covers both.
	second := wheelOnly(next.(Model), b.x+1, b.y+2, false)
	second = wheelTick(second)
	if want := min(3*wheelStep, second.lastOff(paneList)); second.off[paneList] != want {
		t.Errorf("three notches over two frames should scroll %d, off is %d", want, second.off[paneList])
	}
}

// A pane with less content than it has rows has nowhere to scroll, so every
// key and every notch leaves it at the top. An offset below zero would count
// from a line that does not exist.
func TestAPaneShorterThanItsContentNeverScrolls(t *testing.T) {
	t.Parallel()

	// The Plans tab has both list boxes, so the walk covers a tab with a
	// Done pane as well as the detail box.
	m := press(sized(newModel(t), 120, 40), tabKey(tabPlans))
	for _, p := range []pane{paneList, paneDone, paneDetail} {
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

// A wheel notch belongs to the box under the pointer, so it takes the focus
// the way a click does and then scrolls that box by one step, exactly as a
// notch on a box that already had the focus does. A notch that lands on no
// pane at all changes nothing: no focus, no offset, no word on the screen.
func TestWheelOverAnUnfocusedPaneFocusesIt(t *testing.T) {
	t.Parallel()

	// Every ordered pair of the three boxes of the Plans tab.
	for _, p := range []pane{paneList, paneDone, paneDetail} {
		for _, q := range []pane{paneList, paneDone, paneDetail} {
			if q == p {
				continue
			}
			t.Run(fmt.Sprintf("focus-pane%d-notches-pane%d", p+1, q+1), func(t *testing.T) {
				notchOverPair(t, tabPlans, p, q)
			})
		}
	}
	t.Run("a tab with no Done box", func(t *testing.T) {
		notchesOverATabWithoutADoneBox(t)
	})
	t.Run("a notch over no pane", func(t *testing.T) {
		notchesOverNoPane(t)
	})
	t.Run("a notch on a wall with no row", func(t *testing.T) {
		notchOnAWallWithNoRow(t)
	})
	t.Run("a wheel that takes the focus collapses a zoomed box", func(t *testing.T) {
		wheelFocusCollapsesAZoomedBox(t)
	})
}

// notchOverPair is the whole rule for one pair of a tab: the pointer sits over
// pane q while the focus is on pane p. A notch there must end with q focused
// and scrolled by one step, on the same offsets and the same screen a notch
// over q with q already focused ends on. Both runs start from the same screen,
// and the second one reaches q with the same key the reader would, so the only
// difference between them is the notch that takes the focus on the way.
func notchOverPair(t *testing.T, tab int, p, q pane) {
	t.Helper()

	// Opening a tab leaves the focus on its first box, so the walk to any box
	// starts there.
	open := func() Model { return press(longModel(t), tabKey(tab)) }
	on := func(m Model, from, to pane) Model {
		if to == from {
			return m
		}
		k := focusKeyFrom(from, to)
		if k == "" {
			t.Fatalf("pane %d is not part of tab %d, so there is no key from it to pane %d", from+1, tab+1, to+1)
		}
		return press(m, k)
	}
	onP := on(open(), paneList, p)
	onQ := on(on(open(), paneList, p), p, q)
	if onP.focus != p || onQ.focus != q {
		t.Fatalf("the two starts sit on panes %d and %d, want %d and %d", onP.focus+1, onQ.focus+1, p+1, q+1)
	}
	if !slices.Equal(onP.off, onQ.off) {
		t.Fatalf("the two starts hold offsets %v and %v, so they are not the same screen", onP.off, onQ.off)
	}
	// The offsets live in one slice the wheel writes into, so the screen as it
	// was needs a copy of its own to be compared against.
	start := onP
	start.off = slices.Clone(onP.off)
	b := scrollBox(onP, q)
	if b.w < 2 || b.h < 2 {
		t.Fatalf("pane %d is a box of %d by %d cells, too small to take a notch", q+1, b.w, b.h)
	}
	x, y := b.x+1, b.y+2
	if got, _, _ := onP.hit(x, y); got != q {
		t.Fatalf("the cell %d,%d answers pane %d, want pane %d", x, y, got+1, q+1)
	}
	for _, up := range []bool{false, true} {
		got := wheel(onP, x, y, up)
		want := wheel(onQ, x, y, up)
		if got.focus != q {
			t.Errorf("pane %d on pane %d: a notch over the unfocused pane left the focus on pane %d", p+1, q+1, got.focus+1)
		}
		if !slices.Equal(got.off, want.off) {
			t.Errorf("pane %d on pane %d: a notch over the unfocused pane holds offsets %v, the same notch on the focused pane holds %v",
				p+1, q+1, got.off, want.off)
		}
		if !slices.Equal(screenTops(got), screenTops(want)) {
			t.Errorf("pane %d on pane %d: a notch over the unfocused pane drew %v, the same notch on the focused pane drew %v",
				p+1, q+1, screenTops(got), screenTops(want))
		}
		// The pane under the pointer is the only one the notch moves.
		for _, r := range []pane{paneList, paneDone, paneDetail} {
			if r != q && got.off[r] != start.off[r] {
				t.Errorf("pane %d on pane %d: a notch over pane %d moved pane %d from %d to %d",
					p+1, q+1, q+1, r+1, start.off[r], got.off[r])
			}
		}
		if up {
			// Every pane starts on its first line, so a notch away from the
			// reader leaves it there.
			if got.off[q] != 0 {
				t.Errorf("pane %d on pane %d: a notch up scrolled pane %d to %d, the top is 0", p+1, q+1, q+1, got.off[q])
			}
			continue
		}
		if step := min(wheelStep, got.lastOff(q)); got.off[q] != step {
			t.Errorf("pane %d on pane %d: a notch scrolled pane %d to %d, one step from the top is %d",
				p+1, q+1, q+1, got.off[q], step)
		}
		// The offset moved, so the words on screen have to move with it.
		if q != paneDetail && got.lastOff(q) > 0 && drawnFirst(got, q) == drawnFirst(start, q) {
			t.Errorf("pane %d on pane %d: pane %d scrolled to %d but the screen still shows %q",
				p+1, q+1, q+1, got.off[q], drawnFirst(got, q))
		}
	}
}

// The Activity tab has no Done box. Only its two boxes can take a notch, and
// no cell of that screen ever answers the pane the tab does not have, so the
// two real pairs keep the rule and the missing box is named rather than
// skipped in silence.
func notchesOverATabWithoutADoneBox(t *testing.T) {
	t.Helper()

	onActivity := press(longModel(t), tabKey(tabActivity))
	if got := onActivity.panes(); len(got) != 1 || got[0] != paneList {
		t.Fatalf("the Activity tab holds panes %v, want only pane %d", got, paneList+1)
	}
	for y := range onActivity.height {
		for x := range onActivity.width {
			if p, _, _ := onActivity.hit(x, y); p == paneDone {
				t.Fatalf("the cell %d,%d answers pane %d, which the Activity tab does not have", x, y, paneDone+1)
			}
		}
	}
	notchOverPair(t, tabActivity, paneList, paneDetail)
	notchOverPair(t, tabActivity, paneDetail, paneList)
}

// A notch over no pane at all is not a notch of the scroll: the focus stays
// where it is and every offset stays where it is. Every spot here answers the
// focused pane with no row and no tab, which is how hit says nothing is there.
func notchesOverNoPane(t *testing.T) {
	for _, focus := range []pane{paneList, paneDetail} {
		t.Run(fmt.Sprintf("focus-pane%d", focus+1), func(t *testing.T) {
			// An item is selected first, so a box that has to hold still has
			// something in it that could move.
			m := paneModel(t, paneList)
			if focus != paneList {
				m = press(m, focusKeyFrom(paneList, focus))
			}
			if m.lastOff(focus) == 0 {
				t.Fatalf("the focused box has nothing to scroll, so a notch over no pane would prove nothing")
			}
			// The offsets live in one slice the wheel writes into, so the screen
			// as it was needs a copy of its own.
			before := slices.Clone(m.off)
			tops := screenTops(m)
			for _, spot := range []struct {
				name string
				x, y int
			}{
				{"the tab bar", 5, 1},
				{"the bottom border of the tab bar", 5, barRows - 1},
				{"the status line", 5, m.height - 1},
				{"a cell past the last column", m.width, 10},
			} {
				for _, up := range []bool{false, true} {
					if p, row, tab := m.hit(spot.x, spot.y); p != m.focus || row != -1 || tab != -1 {
						t.Fatalf("%s: the cell %d,%d answers pane %d, row %d, tab %d, so the spot is not on no pane",
							spot.name, spot.x, spot.y, p+1, row, tab)
					}
					got := wheel(m, spot.x, spot.y, up)
					if got.focus != m.focus {
						t.Errorf("%s: a notch moved the focus to pane %d", spot.name, got.focus+1)
					}
					if !slices.Equal(got.off, before) {
						t.Errorf("%s: a notch moved the offsets to %v, want %v", spot.name, got.off, before)
					}
					if !slices.Equal(screenTops(got), tops) {
						t.Errorf("%s: a notch changed the screen to %v, want %v", spot.name, screenTops(got), tops)
					}
				}
			}
		})
	}
}

// A cell with no row inside a box is not on no pane: it is on that box, on
// its wall. The notch belongs to the pane it is drawn on, so it scrolls it
// and leaves the focus alone, because the pane it lands on has the focus.
func notchOnAWallWithNoRow(t *testing.T) {
	t.Helper()

	m := paneModel(t, paneList)
	b := scrollBox(m, paneList)
	x, y := b.x+1, b.y+b.h-1
	if p, row, tab := m.hit(x, y); p != paneList || row != -1 || tab != -1 {
		t.Fatalf("the wall cell %d,%d answers pane %d, row %d, tab %d, want pane %d with no row and no tab",
			x, y, p+1, row, tab, paneList+1)
	}
	got := wheel(m, x, y, false)
	if got.focus != paneList {
		t.Errorf("a notch on the wall of the focused box moved the focus to pane %d", got.focus+1)
	}
	if step := min(wheelStep, got.lastOff(paneList)); got.off[paneList] != step {
		t.Errorf("a notch on the wall scrolled the box to %d, one step from the top is %d", got.off[paneList], step)
	}
}

// The room belongs to the box that has the focus, so a wheel that takes the
// focus collapses a zoomed list box the same way a click does.
func wheelFocusCollapsesAZoomedBox(t *testing.T) {
	t.Helper()

	m := press(paneModel(t, paneList), "z")
	if m.expanded != int(paneList) {
		t.Fatalf("the zoom left expanded at %d, want pane %d", m.expanded, paneList+1)
	}
	b := scrollBox(m, paneDetail)
	if b.w < 2 || b.h < 2 {
		t.Fatalf("with the list box zoomed the detail box is %d by %d cells", b.w, b.h)
	}
	got := wheel(m, b.x+1, b.y+1, false)
	if got.focus != paneDetail {
		t.Errorf("a notch over the detail box left the focus on pane %d", got.focus+1)
	}
	if got.expanded != -1 {
		t.Errorf("the wheel left the zoom on, expanded is %d, want -1", got.expanded)
	}
}

// Notches gathered for one box are that box's own, so a notch that takes the
// focus somewhere else hands them over instead of dropping them. The control
// run is the same two notches with no focus change, so the test says what
// nothing lost is rather than a number written down here.
func TestWheelKeepsGatheredNotchesWhenANotchTakesTheFocus(t *testing.T) {
	t.Parallel()

	// Every ordered pair of the three boxes of the Plans tab.
	for _, from := range []pane{paneList, paneDone, paneDetail} {
		for _, to := range []pane{paneList, paneDone, paneDetail} {
			if from == to {
				continue
			}
			t.Run(fmt.Sprintf("pane%d-gathers-pane%d-takes-the-focus", from+1, to+1), func(t *testing.T) {
				gatheredNotchesSurviveAFocusChange(t, from, to)
			})
		}
	}
}

// focusOnWithADoneRow is the long board with the Plans tab open, the Done box
// holding a row of its own and the focus on p. That row is what makes a notch
// which takes the focus there a real change of screen: with both list boxes on
// the same row, the screen does not move and nothing can be lost.
func focusOnWithADoneRow(t *testing.T, p pane) Model {
	t.Helper()
	m := press(longModel(t), tabKey(tabPlans), "tab", "j")
	if m.focus != paneDone || len(m.sel) <= int(paneDone) {
		t.Fatalf("the walk left the focus on pane %d and the selection %v, want pane %d holding a row of its own",
			m.focus+1, m.sel, paneDone+1)
	}
	// The two list boxes have to show a different selection, or a notch that
	// takes the focus moves no screen and there is nothing to lose.
	if m.sel[paneList] == m.sel[paneDone] {
		t.Fatalf("both list boxes show the selection %q, so a notch that takes the focus there changes no screen",
			m.sel[paneDone])
	}
	return press(m, focusKeyFrom(paneDone, p))
}

// twoNotchesOn turns the wheel twice over the same box. The first notch
// scrolls at once, the second one waits for the frame tick.
func twoNotchesOn(t *testing.T, m Model, p pane) Model {
	t.Helper()
	b := scrollBox(m, p)
	if b.w < 2 || b.h < 3 {
		t.Fatalf("pane %d is a box of %d by %d cells, too small to take a notch", p+1, b.w, b.h)
	}
	m = wheelOnly(m, b.x+1, b.y+2, false)
	return wheelOnly(m, b.x+1, b.y+2, false)
}

// gatheredNotchesSurviveAFocusChange walks one ordered pair: the reader
// gathers two notches on the box p and then turns the wheel over the box q,
// which takes the focus and closes the frame.
func gatheredNotchesSurviveAFocusChange(t *testing.T, from, to pane) {
	t.Helper()

	// The control: the reader keeps the focus where it is, so both notches
	// land on the box they were gathered for. It runs on a model of its own,
	// because the offsets live in one slice that every copy of the model
	// writes into.
	control := wheelTick(twoNotchesOn(t, focusOnWithADoneRow(t, from), from))
	want := min(2*wheelStep, control.lastOff(from))
	if control.off[from] != want {
		t.Fatalf("pane %d then pane %d: two notches with no focus change left pane %d at %d, two steps from the top is %d",
			from+1, to+1, from+1, control.off[from], want)
	}

	m := twoNotchesOn(t, focusOnWithADoneRow(t, from), from)
	if m.wheelDelta != wheelStep {
		t.Fatalf("pane %d then pane %d: the second notch left %d lines waiting for the tick, one notch is %d",
			from+1, to+1, m.wheelDelta, wheelStep)
	}
	b := scrollBox(m, to)
	if p, _, _ := m.hit(b.x+1, b.y+2); p != to {
		t.Fatalf("pane %d then pane %d: the cell %d,%d answers pane %d, want pane %d",
			from+1, to+1, b.x+1, b.y+2, p+1, to+1)
	}
	m = wheelOnly(m, b.x+1, b.y+2, false)
	if m.focus != to {
		t.Fatalf("the notch over pane %d left the focus on pane %d", to+1, m.focus+1)
	}
	m = wheelTick(m)
	if m.off[from] != want {
		t.Errorf("pane %d gathers and a notch over pane %d takes the focus: pane %d ends at %d, the same two notches with no focus change end at %d, so %d lines were dropped",
			from+1, to+1, from+1, m.off[from], want, want-m.off[from])
	}
	// The notch that took the focus is a notch of its own, so it moves the box
	// it landed on by one step.
	if step := min(wheelStep, m.lastOff(to)); m.off[to] != step {
		t.Errorf("pane %d then pane %d: the notch that took the focus left pane %d at %d, one step from the top is %d",
			from+1, to+1, to+1, m.off[to], step)
	}
}

// The two hints that name the focused box on the frame: a list box offers the
// key that opens the detail box, and the detail box offers the keys that
// scroll it. Both lead the hint line, so a narrow line cannot drop them.
const (
	listFocusHint   = "Detail: enter"
	detailFocusHint = "Scroll: j k"
)

// A notch that takes the focus shows a screen the reader has not seen, so it
// can never leave the frame from before on the screen. The hint line names the
// keys the focused box can use, so the frame says which box has the focus.
func TestANotchThatTakesTheFocusDrawsTheNewFocus(t *testing.T) {
	t.Parallel()

	// Both frames here leave the wheel armed with nothing waiting for the
	// tick, so the notch that lands on the detail box only gathers: there is
	// nothing to hand over to the box it leaves, and the frame it holds is
	// the only thing that can keep the old focus on screen.
	t.Run("nothing waiting for the tick", func(t *testing.T) {
		notchThatTakesTheFocusDrawsIt(t, false)
	})
	// Notches that cancel each other out leave the closing tick with nothing
	// to land either, so that tick keeps the frame the notch left behind.
	t.Run("notches that cancelled out", func(t *testing.T) {
		notchThatTakesTheFocusDrawsIt(t, true)
	})
}

// notchThatTakesTheFocusDrawsIt is one frame where the wheel is armed with
// nothing waiting, so the notch that lands on the detail box has no notches
// to hand over and the frame it holds is all it can keep. cancel adds notches
// that cancel out, so the tick that closes the frame has nothing to land.
func notchThatTakesTheFocusDrawsIt(t *testing.T, cancel bool) {
	t.Helper()

	// The program draws after every message, so the test does the same and the
	// frame cache holds what the reader really has on screen.
	draw := func(m Model) (Model, string) { return m, m.View() }

	m := focusOnWithADoneRow(t, paneList)
	lb := scrollBox(m, paneList)
	if lb.w < 2 || lb.h < 3 {
		t.Fatalf("the list box is %d by %d cells, too small to take a notch", lb.w, lb.h)
	}
	lx, ly := lb.x+1, lb.y+2
	m, stale := draw(m)
	if shown := plain(stale); !strings.Contains(shown, listFocusHint) || strings.Contains(shown, detailFocusHint) {
		t.Fatalf("the list box should have the focus before the notch, the frame says:\n%s", shown)
	}
	// Two notches on the list box and the tick that lands them, so the wheel
	// is still armed with nothing waiting.
	m, _ = draw(wheelOnly(m, lx, ly, false))
	m, _ = draw(wheelOnly(m, lx, ly, false))
	m, _ = draw(wheelTick(m))
	if !m.wheelArmed || m.wheelDelta != 0 {
		t.Fatalf("the frame landed with %d lines waiting and armed %v, want the wheel armed with nothing waiting",
			m.wheelDelta, m.wheelArmed)
	}
	if cancel {
		// A notch down and a notch up leave the wheel with nothing to land.
		m, _ = draw(wheelOnly(m, lx, ly, false))
		m, _ = draw(wheelOnly(m, lx, ly, true))
		if m.wheelDelta != 0 {
			t.Fatalf("a notch down and a notch up left %d lines waiting, want nothing", m.wheelDelta)
		}
	}
	db := scrollBox(m, paneDetail)
	if db.w < 2 || db.h < 3 {
		t.Fatalf("the detail box is %d by %d cells, too small to take a notch", db.w, db.h)
	}
	m, stale = draw(m)
	if shown := plain(stale); strings.Contains(shown, detailFocusHint) {
		t.Fatalf("the detail box has the focus before the notch, so the case proves nothing:\n%s", shown)
	}

	m, frame := draw(wheelOnly(m, db.x+1, db.y+2, false))
	if m.focus != paneDetail {
		t.Fatalf("the notch over the detail box left the focus on pane %d", m.focus+1)
	}
	if m.same {
		t.Errorf("the notch that took the focus kept the frame from before, so the next frame shows the old focus")
	}
	if frame == stale {
		t.Errorf("the frame drawn after the notch that took the focus is the one from before:\n%s", plain(frame))
	}
	drawn := plain(frame)
	if !strings.Contains(drawn, detailFocusHint) {
		t.Errorf("the frame drawn after the notch does not hold %q, so it does not show the detail box on focus:\n%s",
			detailFocusHint, drawn)
	}
	if strings.Contains(drawn, listFocusHint) {
		t.Errorf("the frame drawn after the notch still holds %q, so it still shows a list box on focus:\n%s",
			listFocusHint, drawn)
	}
	if cancel {
		// A notch back the other way leaves the closing tick with nothing to
		// land, so that tick keeps the frame the notch left on screen.
		m, _ = draw(wheelOnly(m, db.x+1, db.y+2, true))
		m, frame = draw(wheelTick(m))
		if drawn := plain(frame); !strings.Contains(drawn, detailFocusHint) {
			t.Errorf("the frame kept after notches that cancelled out does not hold %q, so the tick kept the old focus:\n%s",
				detailFocusHint, drawn)
		}
	}
}

// A click takes the focus, so the wheel over the same pane scrolls it right
// after, with no key in between.
func TestWheelAfterAClickScrollsThatPane(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		m := paneModel(t, paneList)
		b := scrollBox(m, p)
		m = click(m, b.x+1, b.y+2)
		m = wheel(m, b.x+1, b.y+2, false)
		if m.focus != p {
			t.Errorf("pane %d: the click should have taken the focus, focus is %d", p, m.focus)
		}
		if step := min(wheelStep, m.lastOff(p)); m.off[p] != step {
			t.Errorf("pane %d: the wheel after the click should scroll it by %d, off is %d", p, step, m.off[p])
		}
	}
}

func TestKeysActOnTheFocusedPaneOnly(t *testing.T) {
	t.Parallel()

	// The keys of the detail box scroll it and leave both lists where they
	// were, selection and offset.
	m := press(longModel(t), tabKey(tabPlans), "G")
	before := screenTops(m)
	m = press(m, "shift+tab", "ctrl+d", "ctrl+d")
	if m.off[paneDetail] != 2*pageLines {
		t.Fatalf("two pages should scroll the detail to %d, off is %d", 2*pageLines, m.off[paneDetail])
	}
	// The screen says the same: the detail box moved two pages down, the two
	// lists did not move at all.
	after := screenTops(m)
	if !shows(m, paneDetail, 2*pageLines) {
		t.Errorf("the detail box shows %q, it should show its line %d", after[paneDetail], 2*pageLines)
	}
	if after[paneList] != before[paneList] || after[paneDone] != before[paneDone] {
		t.Errorf("the keys in the detail box moved a list on screen: %v, want %v", after, before)
	}
	// Coming back to a list finds it exactly where it was left.
	m = press(m, "esc")
	if want := m.lastOff(paneList); m.off[paneList] != want {
		t.Errorf("the list scrolled to %d, want %d", m.off[paneList], want)
	}
	if got := screenTops(m)[paneList]; got != before[paneList] {
		t.Errorf("the list shows %q, it showed %q before the keys", got, before[paneList])
	}
	// A key in the list moves the list and no other box. Another item starts
	// the detail at its own top, which is the one place a key in a list
	// reaches into the detail box.
	m = press(m, "k")
	after = screenTops(m)
	if after[paneDone] != before[paneDone] {
		t.Errorf("k in the list moved the Done box on screen: %v, want %v", after[paneDone], before[paneDone])
	}
	if !shows(m, paneDetail, 0) {
		t.Errorf("another item should start the detail at its top, the screen shows %q", drawnFirst(m, paneDetail))
	}
	// Walking to the top of a list scrolls it back up, and still no other
	// pane moves.
	m = press(m, "g")
	after = screenTops(m)
	if !shows(m, paneList, 0) {
		t.Errorf("g should take the list back to its first row, the screen shows %q", after[paneList])
	}
	if after[paneDone] != before[paneDone] || !shows(m, paneDetail, 0) {
		t.Errorf("g in the list moved another box on screen: %v", after)
	}
}

func TestListKeepsTheSelectedRowVisible(t *testing.T) {
	t.Parallel()

	m := press(longModel(t), tabKey(tabPlans))
	rows, _, _ := m.slotOf(paneList)
	// Walk down until the cursor sits below the fold, then keep walking: the
	// pane has to follow the cursor, not the other way round.
	for m.cursor() < scrollBox(m, paneList).rows {
		m = press(m, "j")
	}
	below := m.cursor()
	m = press(m, "j")
	if m.cursor() == below {
		t.Fatal("the list should still move down")
	}
	b := scrollBox(m, paneList)
	cur := m.cursor()
	if cur < b.first || cur >= b.first+b.rows {
		t.Fatalf("row %d is off screen: the pane shows rows %d to %d", cur, b.first, b.first+b.rows-1)
	}
	// The row the pane says is on top is the row that is drawn on top.
	if !shows(m, paneList, b.first) {
		t.Fatalf("the top of the pane shows %q, the pane starts at row %d of %d", drawnFirst(m, paneList), b.first, len(rows))
	}
	drawn := paneRows(m, paneList)
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
	if m.cursor() < scrollBox(m, paneList).first {
		t.Error("k moved the cursor off the top of the pane")
	}
}

func TestSelectingAnotherItemPutsTheDetailBackAtTheTop(t *testing.T) {
	t.Parallel()

	m := press(longModel(t), tabKey(tabPlans), "shift+tab")
	for m.off[paneDetail] < 5 {
		m = press(m, "j")
	}
	if m.off[paneDetail] == 0 {
		t.Fatal("the detail never scrolled")
	}
	if !shows(m, paneDetail, m.off[paneDetail]) {
		t.Errorf("the detail shows %q at offset %d", drawnFirst(m, paneDetail), m.off[paneDetail])
	}
	m = press(m, "esc", "j")
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
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
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
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
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
	t.Parallel()

	for _, p := range []pane{paneList, paneDone, paneDetail} {
		// The detail box needs an item under it before it has a body, so the
		// list is opened on the plans first.
		m := press(wrapModel(t), tabKey(tabPlans))
		m = press(m, keyTo(p), "G")
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
	t.Parallel()

	m := press(longModel(t), tabKey(tabPlans), "G")
	open := screenTops(m)[paneList]
	m = press(m, "tab", "G")
	done := screenTops(m)[paneDone]
	if shows(m, paneDone, 0) {
		t.Fatalf("the Done box did not scroll, it shows %q", drawnFirst(m, paneDone))
	}
	if got := screenTops(m)[paneList]; got != open {
		t.Errorf("the list moved to %q while the Done box scrolled, was %q", got, open)
	}
	m = press(m, "tab", "G")
	after := screenTops(m)
	if after[paneList] != open || after[paneDone] != done {
		t.Errorf("the detail box moved the lists: %v, want %v and %v", after, open, done)
	}
	if shows(m, paneDetail, 0) {
		t.Error("the detail box did not scroll")
	}
	// Each box has to end on the last line its own content allows, not on a
	// line that belongs to some other box's body.
	for _, p := range []pane{paneList, paneDone, paneDetail} {
		if !shows(m, p, m.lastOff(p)) {
			t.Errorf("box %d ends on %q, its own content allows the window at %d", p+1, drawnFirst(m, p), m.lastOff(p))
		}
	}
	// Walking out of the detail box and back to the list finds the list
	// exactly where it was left.
	m = press(m, "esc", "shift+tab")
	if got := screenTops(m)[paneList]; got != open {
		t.Errorf("coming back to the list lost its place: %q, want %q", got, open)
	}
}

// A pane that gives its last cell to the scrollbar has to give a cell back
// when it is too narrow to spare one, or its walls and the line the view cuts
// to the window width end up a cell apart. A body whose last line is exactly
// the inner width, or that has no newline at the end, is the same problem
// from the other side: the line is already full when the pane pads it.
func TestNarrowPanesAndOddBodiesKeepTheirWalls(t *testing.T) {
	t.Parallel()

	// Heights below three leave no room for a pane at all, which the frame
	// task covers on its own; what the scrollbar adds is a cell inside a pane.
	for _, w := range []int{1, 2, 3, 4, 5, 6, 8, 12, 20, 40} {
		for _, h := range []int{3, 4, 5, 8, 20, 40} {
			m := press(longModel(t), tabKey(tabPlans), "shift+tab", "G")
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
	for _, p := range []pane{paneList, paneDone, paneDetail} {
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
		m = press(m, tabKey(tabPlans), "shift+tab", "G")
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
// group open and shut changes the rows of the list under the cursor, so the
// box has to land on a window its new rows can fill and the screen has to
// show that window.
func TestFoldingTheGroupKeepsTheListReadable(t *testing.T) {
	t.Parallel()

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
	m = press(m, tabKey(tabSpecs), "G")
	if m.off[paneList] == 0 {
		t.Fatal("the list did not overflow")
	}
	for _, open := range []bool{true, false, true} {
		m.groupOpen = open
		rows, sel, idx := m.slotOf(paneList)
		if want := itemCount(cursorOf(rows, *sel, *idx)+1, len(rows)); footOf(t, m, paneList) != want {
			t.Errorf("groupOpen %v: the box writes %q, want %q", open, footOf(t, m, paneList), want)
		}
		if !shows(m, paneList, m.off[paneList]) {
			t.Errorf("groupOpen %v: the box shows %q, it sits at %d", open, drawnFirst(m, paneList), m.off[paneList])
		}
		if top := thumbAt(m, paneList); top < 0 {
			t.Errorf("groupOpen %v: a long list draws no thumb", open)
		}
		if strings.TrimSpace(paneRows(m, paneList)[0]) == "" {
			t.Errorf("groupOpen %v: the first line on screen is blank", open)
		}
	}
}

// listLines draws the list box of the open tab the way the screen does.
func listLines(m Model) []string {
	p := m.listPane()
	b := m.geometry().at(p)
	return m.listView(p, b.textW(), b)
}

// listRows gives every drawn line the id of the row it draws, so a check that
// fails can name the row instead of counting lines.
func listRows(m Model) map[string]string {
	p := m.listPane()
	b := m.geometry().at(p)
	rows, _, _ := m.slotOf(p)
	out := map[string]string{}
	for i, ln := range listLines(m) {
		if n := b.first + i; n < len(rows) {
			out[rows[n].id] = ln
		}
	}
	return out
}

// listRow finds the drawn line of one row by its whole id, so a row is named
// and never counted.
func listRow(t *testing.T, m Model, id string) string {
	t.Helper()
	ln, ok := listRows(m)[id]
	if !ok {
		t.Fatalf("no row for %s in %q", id, plainLines(listLines(m)))
	}
	return ln
}

// A row of any kind is drawn at full brightness, and the id of every row the
// cursor is not on wears the color of its kind.
func TestListRowsAreNotFaintAndWearKindColors(t *testing.T) {
	withTrueColor(func() {
		// Three lists between them hold every kind of row: plans and their
		// tasks under Activity, a plan whose work has not begun on Plans,
		// and the folded group of untyped files on Specs.
		for _, m := range []Model{
			actModel(t).WithTheme("tokyo-night", true),
			press(actModel(t).WithTheme("tokyo-night", true), tabKey(tabPlans)),
			press(newModel(t).WithTheme("tokyo-night", true), tabKey(tabSpecs)),
		} {
			rows, sel, idx := m.slotOf(m.listPane())
			under := rows[cursorOf(rows, *sel, *idx)].id
			for id, ln := range listRows(m) {
				if sgrHas(ln, "2") {
					t.Errorf("%s row is faint: %q in %q", id, plain(ln), plainLines(listLines(m)))
				}
				if id == under {
					if !sgrHas(ln, "1") {
						t.Errorf("row under the cursor is not the bold band: %q", plain(ln))
					}
					continue
				}
				it := m.board.Get(id)
				if it == nil {
					continue // the group row holds no item, so it has no id
				}
				name := it.ShortID
				if name == "" {
					name = it.ID
				}
				// An id the pane had to cut gets no color of its own.
				if !strings.Contains(plain(ln), name) {
					continue
				}
				if want := m.styles.kind(it.Kind).Render(name); !strings.Contains(ln, want) {
					t.Errorf("%s id %s is not in its kind color: %q", id, name, plain(ln))
				}
			}
		}
	})
}

// A row whose work is under way reads in the plain foreground, and its tree
// dot is the pulse dot, so the view can make it pulse.
func TestListRowWithWorkIsPlainWithAPulseDot(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		ln := listRow(t, m, "plans/2026-09-23-q#task-1") // PLN-0004.01, under way, not the cursor row
		id := m.styles.kind(board.KindTask).Render("PLN-0004.01")
		// The id keeps the blue of its kind, so the words around it are the
		// only ones that could still wear it.
		if blue := sgr.FindString(m.styles.accent.Render("x")); strings.Contains(strings.Replace(ln, id, "", 1), blue) {
			t.Errorf("work row is still blue: %q", ln)
		}
		if sgrHas(ln, "2") {
			t.Errorf("work row is faint: %q", ln)
		}
		if !strings.Contains(ln, m.styles.goingDot) {
			t.Errorf("work row has no pulse dot: %q", ln)
		}
		if !strings.Contains(ln, id) {
			t.Errorf("work row id is not in its kind color: %q", ln)
		}
	})
}

func TestDoneListRowsAreGreen(t *testing.T) {
	withTrueColor(func() {
		cfg := treeCfg(t, map[string]string{
			".acta/plans/2026-09-25-d.md": "---\nid: PLN-0009\n---\n# Plan D\n\n### Task 1: Finished\n- [x] a\n\n### Task 2: Open\n- [ ] b\n",
		})
		b, err := board.Load(cfg)
		if err != nil {
			t.Fatal(err)
		}
		m := press(sized(New(cfg, b, true).WithTheme("tokyo-night", true), 160, 40), tabKey(tabPlans))
		m.setOpen("plans/2026-09-25-d", true)
		rows := listRows(m)
		done, ok := rows["plans/2026-09-25-d#task-1"]
		if !ok {
			t.Fatalf("no row for the done task in %q", plainLines(listLines(m)))
		}
		green := sgr.FindString(m.styles.done.Render("x"))
		// The mark sits inside the green span of the row, with the indent in
		// front of it, so the last code before the tick is the one that paints it.
		before := done[:strings.Index(done, dotDone)]
		if at := strings.LastIndex(before, green); at < 0 || strings.Contains(before[at:], "\x1b[0m") {
			t.Errorf("done tick is not green: %q", done)
		}
		if !strings.Contains(done, green) || !strings.Contains(plain(done), "Finished") {
			t.Errorf("done title is not green: %q", done)
		}
		if want := m.styles.kind(board.KindTask).Render("PLN-0009.01"); !strings.Contains(done, want) {
			t.Errorf("done row id is not in its kind color: %q", done)
		}
		open := rows["plans/2026-09-25-d#task-2"]
		id := m.styles.kind(board.KindTask).Render("PLN-0009.02")
		if strings.Contains(strings.Replace(open, id, "", 1), green) {
			t.Errorf("a row that is not done wears green: %q", open)
		}
	})
}

// A done plan keeps its + or - mark in the plain foreground, so the reader can
// still see the plan open or shut. Only the title and the id after it carry
// the green, and a done task still keeps its green tick.
func TestDonePlanRowKeepsItsMarkPlain(t *testing.T) {
	withTrueColor(func() {
		const name = "PLN-0010"
		cfg := treeCfg(t, map[string]string{
			".acta/plans/2026-09-26-e.md": "---\nid: PLN-0010\n---\n# Shipped E\n\n### Task 1: Shipped\n- [x] a\n\n### Task 2: Shipped too\n- [x] b\n",
		})
		b, err := board.Load(cfg)
		if err != nil {
			t.Fatal(err)
		}
		// The terminal theme has no hex colors, so every code is a plain ANSI
		// number a reader can check by eye.
		m := sized(New(cfg, b, true).WithTheme("terminal", true), 160, 40)
		m = press(m, tabKey(tabPlans), "tab") // the Done box under Plans
		m.setOpen("plans/2026-09-26-e", true)
		m = press(m, "j", "j") // the cursor leaves the head row and its first task
		rows := listRows(m)
		head, ok := rows["plans/2026-09-26-e"]
		if !ok {
			t.Fatalf("no row for the done plan in %q", plainLines(listLines(m)))
		}
		green := sgr.FindString(m.styles.done.Render("x"))
		// Between the start of the line and the id sit only the mark and the
		// space behind it, so the last code in front of the mark is the one
		// still in force when the mark is drawn.
		mark := head[:strings.Index(head, name)]
		if plain(mark) != "- " {
			t.Fatalf("the done plan mark is %q, want the tree mark and its space", plain(mark))
		}
		codes := sgr.FindAllString(mark[:strings.IndexAny(mark, "+-")], -1)
		if len(codes) > 0 && codes[len(codes)-1] == green {
			t.Errorf("the done plan mark is green: %q", head)
		}
		// The title after the id takes the green of finished work.
		tail := head[strings.Index(head, name)+len(name):]
		at := strings.Index(tail, "Shipped")
		if at < 0 {
			t.Fatalf("the done plan row has no title: %q", head)
		}
		before := tail[:at]
		if g := strings.LastIndex(before, green); g < 0 || strings.Contains(before[g:], "\x1b[0m") {
			t.Errorf("the done plan title is not green: %q", head)
		}
		if want := m.styles.kind(board.KindPlan).Render(name); !strings.Contains(head, want) {
			t.Errorf("the done plan id is not in its kind color: %q", head)
		}
		// A done task keeps its green tick: the last code before the tick is
		// still in force when the tick is drawn.
		task := listRow(t, m, "plans/2026-09-26-e#task-1")
		tick := task[:strings.Index(task, dotDone)]
		if g := strings.LastIndex(tick, green); g < 0 || strings.Contains(tick[g:], "\x1b[0m") {
			t.Errorf("the done task tick is not green: %q", task)
		}
	})
}

func TestCursorRowKeepsItsBandWithoutAPulseDot(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		const work = "plans/2026-09-23-q#task-1" // PLN-0004.01, under way
		for i := 0; i < 10 && (m.Selected() == nil || m.Selected().ID != work); i++ {
			m = press(m, "j")
		}
		if m.Selected() == nil || m.Selected().ID != work {
			t.Fatalf("the cursor never reached %s; rows are %q", work, rowIDs(m))
		}
		ln := listRow(t, m, work)
		if strings.Contains(ln, m.styles.goingDot) {
			t.Errorf("the cursor row pulses its dot inside the band: %q", ln)
		}
		if !sgrHas(ln, "1") {
			t.Errorf("the cursor row is not bold: %q", ln)
		}
	})
}

// A row with no item, and one whose id was cut off by the pane width, are all
// base: there is nothing to give a color of its own.
func TestPaintIDLeavesACutIDPlain(t *testing.T) {
	withTrueColor(func() {
		m := actModel(t).WithTheme("tokyo-night", true)
		it := &board.Item{ShortID: "PLN-0003", Kind: board.KindPlan}
		base := lipgloss.NewStyle()
		if got := m.paintID("PLN-0", it, base); got != base.Render("PLN-0") {
			t.Errorf("cut id got a color: %q", got)
		}
		if got := m.paintID("x", nil, base); got != base.Render("x") {
			t.Errorf("row with no item got a color: %q", got)
		}
	})
}
