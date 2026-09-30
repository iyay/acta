package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func mouseAt(m Model, x, y int, a tea.MouseAction) Model {
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: a, Button: tea.MouseButtonLeft})
	return next.(Model)
}

// dragged presses at one cell, moves to another and lets go. It gives back
// the model and what reached the clipboard.
func dragged(m Model, x0, y0, x1, y1 int) (Model, string) {
	var got string
	m.clip = func(s string) error { got = s; return nil }
	m = mouseAt(m, x0, y0, tea.MouseActionPress)
	m = mouseAt(m, x1, y1, tea.MouseActionMotion)
	m = mouseAt(m, x1, y1, tea.MouseActionRelease)
	return m, got
}

// screenText cuts the cells a drag should copy straight out of the plain
// screen, rune by rune, so the test does not lean on the code it checks.
func screenText(m Model, x0, y0, x1, y1 int, a rect) string {
	lines := strings.Split(plain(m.View()), "\n")
	var rows []string
	for y := y0; y <= y1; y++ {
		r := []rune(lines[y])
		from, to := a.x0, a.x1
		if y == y0 {
			from = x0
		}
		if y == y1 {
			to = x1
		}
		rows = append(rows, strings.TrimRight(string(r[from:to+1]), " "))
	}
	return strings.Join(rows, "\n")
}

func TestDragCopiesThePaneText(t *testing.T) {
	t.Parallel()

	for _, p := range []pane{paneList, paneDetail} {
		m := paneModel(t, p)
		a, _ := scrollBox(m, p).textArea()
		want := screenText(m, a.x0+1, a.y0, a.x0+6, a.y0+1, a)
		m, got := dragged(m, a.x0+1, a.y0, a.x0+6, a.y0+1)
		if got == "" || got != want {
			t.Errorf("pane %d: copied %q, want %q", p, got, want)
		}
		if strings.ContainsAny(got, "│─╭╮╰╯┌┐└┘") {
			t.Errorf("pane %d: copy holds a wall: %q", p, got)
		}
		if !strings.HasPrefix(m.status, "copied ") {
			t.Errorf("pane %d: status = %q", p, m.status)
		}
	}
}

func TestDragBackwardsCopiesTheSame(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	a, _ := scrollBox(m, paneDetail).textArea()
	_, down := dragged(m, a.x0+2, a.y0, a.x0+5, a.y0+2)
	_, up := dragged(m, a.x0+5, a.y0+2, a.x0+2, a.y0)
	if down == "" || down != up {
		t.Errorf("down %q, up %q", down, up)
	}
}

func TestDragPastTheEdgeStaysInThePane(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	a, _ := scrollBox(m, paneList).textArea()
	m, got := dragged(m, a.x0, a.y0, m.width-1, m.height-1)
	if got != screenText(m, a.x0, a.y0, a.x1, a.y1, a) {
		t.Errorf("a drag off the pane copied %q", got)
	}
}

func TestClickWithNoDragCopiesNothing(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneList)
	b := scrollBox(m, paneList)
	want := click(paneModel(t, paneList), b.x+1, b.y+2).Selected().ID
	called := false
	m.clip = func(string) error { called = true; return nil }
	m = mouseAt(m, b.x+1, b.y+2, tea.MouseActionPress)
	m = mouseAt(m, b.x+1, b.y+2, tea.MouseActionRelease)
	if called {
		t.Error("a plain click copied")
	}
	// Two boards are two sets of rows, so the cursor is read by its id and
	// not by the pointer it happens to sit at.
	if m.Selected().ID != want {
		t.Errorf("a plain click left the cursor on %v, want %v", m.Selected().ID, want)
	}
}

func TestHighlightClears(t *testing.T) {
	t.Parallel()

	for name, end := range map[string]func(Model, rect) Model{
		"a key":    func(m Model, _ rect) Model { return press(m, "j") },
		"a press":  func(m Model, a rect) Model { return mouseAt(m, a.x0, a.y0, tea.MouseActionPress) },
		"a wheel":  func(m Model, a rect) Model { return wheel(m, a.x0, a.y0, false) },
		"a resize": func(m Model, _ rect) Model { return sized(m, m.width-1, m.height) },
	} {
		m := paneModel(t, paneDetail)
		a, _ := scrollBox(m, paneDetail).textArea()
		m, _ = dragged(m, a.x0, a.y0, a.x0+4, a.y0)
		if !m.drag.shown() {
			t.Fatalf("%s: the drag left no highlight to clear", name)
		}
		if m = end(m, a); m.drag.shown() {
			t.Errorf("%s: the highlight stayed", name)
		}
	}
}

func TestHighlightIsDrawnAndCleared(t *testing.T) {
	// No t.Parallel: withColors sets a color profile every other test reads,
	// so two of these at once would fight over it.

	withColors(func() {
		m := paneModel(t, paneDetail)
		a, _ := scrollBox(m, paneDetail).textArea()
		before := m.View()
		m, _ = dragged(m, a.x0, a.y0, a.x0+4, a.y0)
		// The release left a toast on the status line, so it goes before the
		// two screens are compared.
		m.status = ""
		shown := m.View()
		if shown == before || plain(shown) != plain(before) {
			t.Error("the band did not draw, or it changed the words")
		}
		// With the pick gone, the screen is the one from before the drag.
		cleared := m
		cleared.drag = drag{}
		cleared.same = false
		if cleared.View() != before {
			t.Error("the band is still drawn with no pick")
		}
	})
}

func TestTheBandIsDrawnOnANarrowScreen(t *testing.T) {
	// No t.Parallel here either, for the same reason.

	withColors(func() {
		// A screen too narrow for two panes draws one box, and that box has
		// its own way back out of draw().
		m := sized(paneModel(t, paneDetail), 40, 12)
		if m.geometry().wide {
			t.Skip("40 columns still fit two panes")
		}
		a, ok := scrollBox(m, paneDetail).textArea()
		if !ok {
			t.Fatal("the narrow box has no room for words")
		}
		before := m.View()
		m, _ = dragged(m, a.x0, a.y0, a.x0+4, a.y0)
		m.status = ""
		shown := m.View()
		if shown == before || plain(shown) != plain(before) {
			t.Error("the band did not draw on a narrow screen")
		}
	})
}

func TestNoDragWhileSomethingIsOpen(t *testing.T) {
	t.Parallel()

	for name, open := range map[string]func(Model) Model{
		"help":              func(m Model) Model { return press(m, "?") },
		"search":            func(m Model) Model { return press(m, "/") },
		"a slug":            func(m Model) Model { return press(m, "n") },
		"a popup":           func(m Model) Model { return press(m, "t") },
		"a popup of a kind": func(m Model) Model { return press(m, "s") },
	} {
		m := open(paneModel(t, paneDetail))
		a, _ := scrollBox(m, paneDetail).textArea()
		m, got := dragged(m, a.x0, a.y0, a.x0+4, a.y0)
		if got != "" || m.drag.on {
			t.Errorf("%s open: a drag picked %q", name, got)
		}
	}
}

func TestAWheelThatClearsTheBandDrawsAgain(t *testing.T) {
	t.Parallel()

	// A wheel over another pane is ignored, and the notches that gather for
	// the focused pane keep the frame. Both still have to draw, because the
	// band on it is gone.
	for name, turn := range map[string]func(Model, rect, rect) Model{
		"an ignored wheel": func(m Model, a, other rect) Model {
			return wheelOnly(m, other.x0, other.y0, false)
		},
		// The first notch of a frame arms the wheel and scrolls on the spot.
		// The loop arms it, so this one only gathers.
		"a gathered notch": func(m Model, a, _ rect) Model {
			return wheelOnly(m, a.x0, a.y0, false)
		},
	} {
		m := paneModel(t, paneDetail)
		a, _ := scrollBox(m, paneDetail).textArea()
		other, _ := scrollBox(m, paneList).textArea()
		m = armed(m, a)
		m, _ = dragged(m, a.x0, a.y0, a.x0+4, a.y0)
		withBand := m.View()
		m = turn(m, a, other)
		if m.drag.shown() {
			t.Errorf("%s: the band stayed on", name)
		}
		if m.same {
			t.Errorf("%s: the screen kept a band that is gone", name)
		}
		if m.View() == withBand {
			t.Errorf("%s: the band is still on the screen", name)
		}
	}
}

// armed turns one notch so the wheel is waiting for the notches that follow.
func armed(m Model, a rect) Model {
	return wheelOnly(m, a.x0, a.y0, false)
}

func TestReleaseWithNoDragReusesTheFrame(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	m.View()
	if next := mouseAt(m, 1, 1, tea.MouseActionRelease); !next.same {
		t.Error("a release with no drag drew the screen again")
	}
}

func TestAMotionThatMovesNothingReusesTheFrame(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	a, _ := scrollBox(m, paneDetail).textArea()
	m = mouseAt(m, a.x0, a.y0, tea.MouseActionPress)
	m = mouseAt(m, a.x0+4, a.y0, tea.MouseActionMotion)
	m.View()
	next := mouseAt(m, a.x0+4, a.y0, tea.MouseActionMotion)
	if !next.same {
		t.Error("a pointer that stood still drew the screen again")
	}
}

func TestAPlainClickReleaseReusesTheFrame(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	a, _ := scrollBox(m, paneDetail).textArea()
	m = mouseAt(m, a.x0, a.y0, tea.MouseActionPress)
	m.View()
	if next := mouseAt(m, a.x0, a.y0, tea.MouseActionRelease); !next.same {
		t.Error("a click that picked nothing drew the screen again")
	}
}
