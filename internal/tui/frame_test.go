package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TestSplitAlwaysFillsTheHeight checks the boxes never leave a gap and never
// push each other off screen. Every height the terminal can hand us goes
// through, including the tiny ones that have no room to divide.
func TestSplitAlwaysFillsTheHeight(t *testing.T) {
	t.Parallel()

	for h := range 2 {
		// A terminal with no room yet must not panic, only split badly.
		split(h, sidePanes)
	}
	for h := 2; h <= 80; h++ {
		heights := split(h, sidePanes)
		sum := 0
		for i, v := range heights {
			if v < 0 {
				t.Errorf("split(%d) box %d is %d", h, i, v)
			}
			if h >= 2*sidePanes && v < 2 {
				t.Errorf("split(%d) box %d is %d, too short for its two walls", h, i, v)
			}
			sum += v
		}
		if sum != h {
			t.Errorf("split(%d) = %v, which adds up to %d", h, heights, sum)
		}
	}
}

// TestLeftHeightsFillAndExpand walks every height from 3 to 60 and the three
// widths a terminal reports, with each box of the open tab expanded in turn
// and with none of them expanded. Four claims at once. The left column is
// exactly as tall as the detail pane, a box that is not expanded keeps room
// lines once the screen is room*2+room = 9 lines tall and only its title bar
// below that, the view draws the heights the model hands it, and no line of
// the screen reaches past the right edge.
func TestLeftHeightsFillAndExpand(t *testing.T) {
	t.Parallel()

	// The Plans tab has both boxes, so the walk covers the two box column and
	// not the single box of Activities.
	m := press(newModel(t), tabKey(tabPlans))
	n := len(m.panes())
	others := n - 1
	// room is how many lines a pane that is not expanded keeps. Below
	// room*others+room lines of screen there is no such room, so it keeps
	// its title bar alone.
	const room = 3
	threshold := room*others + room
	for h := 3; h <= 60; h++ {
		for _, w := range []int{40, 80, 160} {
			for e := -1; e < n; e++ {
				s := sized(m, w, h)
				s.expanded = e
				// The tab box takes three lines and the status line one, so the
				// boxes share what is left of the screen.
				body := max(0, h-barRows-1)
				hs := s.leftHeights(body)
				sum := 0
				for _, x := range hs {
					sum += x
				}
				if sum != body {
					t.Fatalf("%dx%d expanded %d: heights %v add up to %d, want %d", w, h, e, hs, sum, body)
				}
				// The boxes on screen are the ones the view draws. Below 60
				// columns only the focused box reaches the screen, so the
				// column has no boxes to read there.
				g := s.geometry()
				if g.wide {
					left := 0
					for p, b := range g.side {
						left += b.h
						if b.h != hs[p] {
							t.Fatalf("%dx%d expanded %d: pane %d is %d lines, the heights say %d", w, h, e, p+1, b.h, hs[p])
						}
					}
					if left != g.detail.h {
						t.Fatalf("%dx%d expanded %d: the left column is %d lines, the detail %d", w, h, e, left, g.detail.h)
					}
				}
				for i, x := range hs {
					if e < 0 || i == e {
						continue
					}
					want := room
					if body < threshold {
						want = 1
					}
					// A screen with no room at all leaves nothing to keep, not
					// even a title bar.
					want = min(want, body)
					if x != want {
						t.Errorf("%dx%d expanded %d: pane %d keeps %d lines, want %d", w, h, e, i+1, x, want)
					}
				}
				// A box with a single line is its title bar and nothing
				// else, so a short screen still says which box is which.
				for p, b := range g.side {
					if b.h != 1 {
						continue
					}
					view := s.paneView(pane(p), b)
					drawn := strings.Split(view, "\n")
					if len(drawn) != 1 || drawn[0] != s.paneTop(pane(p), b, s.edge(pane(p))) {
						t.Errorf("%dx%d expanded %d: pane %d draws %d lines (%q), want its title bar alone", w, h, e, p+1, len(drawn), view)
					}
				}
				for i, ln := range strings.Split(s.View(), "\n") {
					if lipgloss.Width(ln) > w {
						t.Fatalf("%dx%d expanded %d: line %d is %d cells wide", w, h, e, i, lipgloss.Width(ln))
					}
				}
			}
		}
	}
}

// TestViewFitsEveryTerminalSize walks every size a terminal can report: each
// height from 3 up to 60, each width from 1 to 200. Three claims at once. The
// left column is exactly as tall as the detail pane, the screen has no more
// lines than the terminal has rows, and no line reaches past the right edge.
func TestViewFitsEveryTerminalSize(t *testing.T) {
	t.Parallel()

	// One subtest per height, each with a model of its own: the model holds
	// a cache behind a pointer, and parallel subtests must not share it.
	for h := 3; h <= 60; h++ {
		t.Run(fmt.Sprintf("h%d", h), func(t *testing.T) {
			t.Parallel()
			m := newModel(t)
			for w := 1; w <= 200; w++ {
				s := sized(m, w, h)
				lines := strings.Split(s.View(), "\n")
				if len(lines) > h {
					t.Fatalf("%dx%d: %d lines", w, h, len(lines))
				}
				for i, ln := range lines {
					if lipgloss.Width(ln) > w {
						t.Fatalf("%dx%d line %d is %d wide", w, h, i, lipgloss.Width(ln))
					}
				}
				g := s.geometry()
				left := 0
				for _, b := range g.side {
					left += b.h
				}
				if g.wide && left != g.detail.h {
					t.Fatalf("%dx%d: left %d, detail %d", w, h, left, g.detail.h)
				}
			}
		})
	}
}

// TestViewSurvivesTinyTerminals covers the sizes too small to hold a pane at
// all. A window squeezed to nothing must still draw without panicking.
func TestViewSurvivesTinyTerminals(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	for h := 0; h <= 2; h++ {
		for _, w := range []int{1, 2, 19, 20, 21} {
			sized(m, w, h).View()
		}
	}
}

// TestResizeClearsTheScreen checks the rows of the old size are wiped. The
// standard renderer only paints over the cells the new frame uses, so a window
// that shrinks leaves the rest of the old frame on screen until we say clear.
func TestResizeClearsTheScreen(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 40)
	// The first resize already armed the pulse, so the second one only
	// clears the screen and hands the next tick on.
	m.pulsing = true
	next, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if cmd == nil || fmt.Sprint(cmd()) != fmt.Sprint(tea.ClearScreen()) {
		t.Fatal("a size change must clear the screen so old rows do not stay")
	}
	got := next.(Model)
	if got.width != 100 || got.height != 30 {
		t.Fatalf("the model kept %dx%d, want 100x30", got.width, got.height)
	}
}

func TestTitleShowsTheSortOfEachPane(t *testing.T) {
	t.Parallel()

	// The Specs tab has both a List and a Done box, so both sort words show.
	m := press(newModel(t), tabKey(tabSpecs))
	view := m.View()
	if got := strings.Count(view, "oldest"); got != 2 {
		t.Fatalf("start: %d panes say oldest, want 2", got)
	}
	if strings.Contains(view, "newest") {
		t.Fatalf("start: a pane says newest before any o")
	}
	m = press(m, "o")
	view = m.View()
	if got := strings.Count(view, "newest"); got != 1 {
		t.Fatalf("after o: %d panes say newest, want 1", got)
	}
	if got := strings.Count(view, "oldest"); got != 1 {
		t.Fatalf("after o: %d panes say oldest, want 1", got)
	}
}

// TestSortWordSitsBeforeTheCountInTheBottomBorder checks both sort states at
// several widths, on every list box of the tab, top and bottom border both.
// The word belongs to the bottom border next to the count, so the top border
// never carries it.
func TestSortWordSitsBeforeTheCountInTheBottomBorder(t *testing.T) {
	t.Parallel()

	// The o key flips the box that has the focus, and tab walks from List to
	// Done, so each box gets the focus before it is flipped.
	for i, p := range []pane{paneList, paneDone} {
		for _, w := range []int{60, 100, 160} {
			for _, flip := range []bool{false, true} {
				m := press(sized(newModel(t), w, 40), tabKey(tabSpecs))
				for range i {
					m = press(m, "tab")
				}
				if flip {
					m = press(m, "o")
				}
				word := m.sortWord(p)
				if want := map[bool]string{false: "oldest", true: "newest"}[flip]; word != want {
					t.Fatalf("pane %d w=%d flip=%v: word %q, want %q", p, w, flip, word, want)
				}
				top := plain(m.paneTop(p, m.geometry().side[p], m.edge(p)))
				if strings.Contains(top, "oldest") || strings.Contains(top, "newest") {
					t.Errorf("pane %d w=%d flip=%v: the top border still says the sort: %q", p, w, flip, top)
				}
				rows, sel, idx := m.slotOf(p)
				count := itemCount(cursorOf(rows, *sel, *idx)+1, len(rows))
				if got := bottomLine(t, m, p); !strings.HasSuffix(got, " "+word+" · "+count+" ┘") {
					t.Errorf("pane %d w=%d flip=%v: bottom border %q, want it to end with %q", p, w, flip, got, word+" · "+count)
				}
			}
		}
	}
}

// TestNarrowBottomBorderDropsTheWordFirst walks every width from too narrow
// for the word up to room for both, so the count never goes first.
func TestNarrowBottomBorderDropsTheWordFirst(t *testing.T) {
	t.Parallel()

	count := "12 of 34"
	both := "newest · " + count
	for inner := range lipgloss.Width(" " + both + " ") {
		if got := sortFoot("newest", count, inner); got != count {
			t.Errorf("inner=%d: %q, want the count alone", inner, got)
		}
	}
	if got := sortFoot("newest", count, lipgloss.Width(" "+both+" ")); got != both {
		t.Errorf("room for both: %q, want %q", got, both)
	}
}

func TestHelpListsTheSortKey(t *testing.T) {
	t.Parallel()

	if !strings.Contains(helpLines, "oldest / newest") {
		t.Fatalf("help does not list o: %q", helpLines)
	}
}
