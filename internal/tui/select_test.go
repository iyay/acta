package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// area5 is a text area of five cells by three rows at the top left, inside a
// one-cell wall, so the frame below has a wall on every side of it.
var area5 = rect{x0: 1, y0: 1, x1: 5, y1: 3}

const frame5 = "┌─────┐\n" +
	"│abcde│\n" +
	"│fg  h│\n" +
	"│ijklm│\n" +
	"└─────┘"

func sel(ax, ay, ex, ey int) drag {
	return drag{on: true, area: area5, ax: ax, ay: ay, ex: ex, ey: ey}
}

func TestDragTextIsAStream(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct {
		d    drag
		want string
	}{
		"one row":        {sel(2, 1, 4, 1), "bcd"},
		"one cell apart": {sel(1, 1, 2, 1), "ab"},
		"many rows":      {sel(3, 1, 2, 3), "cde\nfg  h\nij"},
		"reversed":       {sel(2, 3, 3, 1), "cde\nfg  h\nij"},
		"trailing space": {sel(1, 2, 4, 2), "fg"},
		"whole area":     {sel(1, 1, 5, 3), "abcde\nfg  h\nijklm"},
	} {
		if got := dragText(frame5, c.d); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}

func TestDragWithNoMoveSelectsNothing(t *testing.T) {
	t.Parallel()

	if d := sel(2, 1, 2, 1); d.shown() || dragText(frame5, d) != "" {
		t.Errorf("a press with no drag selected %q", dragText(frame5, d))
	}
}

func TestDragToStaysInsideTheArea(t *testing.T) {
	t.Parallel()

	d := sel(2, 2, 2, 2).to(40, -3)
	if d.ex != area5.x1 || d.ey != area5.y0 {
		t.Fatalf("end = (%d,%d), want (%d,%d)", d.ex, d.ey, area5.x1, area5.y0)
	}
	if got := dragText(frame5, d); strings.ContainsAny(got, "│─┌┐└┘") {
		t.Errorf("a drag past the edge copied a wall: %q", got)
	}
}

func TestDragTextKeepsWideCharactersWhole(t *testing.T) {
	t.Parallel()

	// 日 and 本 take two cells each: cells 1-2 and 3-4.
	frame := "│日本x│"
	area := rect{x0: 1, y0: 0, x1: 5, y1: 0}
	for x0 := 1; x0 <= 5; x0++ {
		for x1 := x0 + 1; x1 <= 5; x1++ {
			got := dragText(frame, drag{on: true, area: area, ax: x0, ay: 0, ex: x1, ey: 0})
			if !strings.Contains("日本x", got) {
				t.Errorf("cells %d-%d gave %q, not a whole piece of the row", x0, x1, got)
			}
		}
	}
}

func TestPaintDragOnlyTouchesTheSpans(t *testing.T) {
	t.Parallel()

	withColors(func() {
		brush := lipgloss.NewStyle().Reverse(true)
		got := paintDrag(frame5, sel(3, 1, 2, 3), brush)
		if plain(got) != frame5 {
			t.Fatalf("the words changed under the band:\n%s", plain(got))
		}
		lines := strings.Split(got, "\n")
		for _, i := range []int{0, 4} {
			if lines[i] != strings.Split(frame5, "\n")[i] {
				t.Errorf("line %d is outside the selection but changed: %q", i, lines[i])
			}
		}
		if !strings.Contains(lines[1], brush.Render("cde")) {
			t.Errorf("row 1 does not carry the band over cde: %q", lines[1])
		}
	})
}

func TestTextAreaLeavesOutWallsAndTitle(t *testing.T) {
	t.Parallel()

	a, ok := box{x: 10, y: 4, w: 6, h: 5}.textArea()
	if !ok || a != (rect{x0: 11, y0: 5, x1: 14, y1: 7}) {
		t.Fatalf("text area = %+v %v", a, ok)
	}
	if _, ok := (box{x: 0, y: 0, w: 2, h: 2}).textArea(); ok {
		t.Error("a box with no room for words has a text area")
	}
}

func TestAnchorOnlyInATextArea(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 40)
	b := m.geometry().at(paneList)
	if d := m.anchorAt(b.x+1, b.y+1); !d.on || !d.held {
		t.Error("a press on the first text cell set no anchor")
	}
	for name, c := range map[string][2]int{
		"left wall":   {b.x, b.y + 1},
		"right wall":  {b.x + b.w - 1, b.y + 1},
		"title line":  {b.x + 1, b.y},
		"bottom line": {b.x + 1, b.y + b.h - 1},
		"status line": {1, m.height - 1},
	} {
		if d := m.anchorAt(c[0], c[1]); d.on {
			t.Errorf("a press on the %s set an anchor", name)
		}
	}
}

func TestCopyPickedToast(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 40)
	var got string
	m.clip = func(s string) error { got = s; return nil }
	cmd := m.copyPicked("abc\ndef")
	if got != "abc\ndef" {
		t.Fatalf("clipboard = %q", got)
	}
	if m.status != "copied abc def" || cmd == nil {
		t.Errorf("status = %q, cmd nil = %v", m.status, cmd == nil)
	}
	// The toast clears the same way the y toast does.
	next, _ := m.Update(clearStatusMsg{text: m.status})
	if s := next.(Model).status; s != "" {
		t.Errorf("the toast stayed: %q", s)
	}

	m.width = 20
	m.copyPicked(strings.Repeat("x", 50))
	if !strings.HasSuffix(m.status, "…") || lipgloss.Width(m.status) > m.width {
		t.Errorf("a long copy was not cut to fit: %q", m.status)
	}

	m.clip = func(string) error { return errors.New("no clipboard") }
	if cmd := m.copyPicked("abc"); cmd != nil || m.status != "copy failed: no clipboard" {
		t.Errorf("failed copy: status %q, cmd nil = %v", m.status, cmd == nil)
	}
}

func TestCopyPickedOfSpacesCopiesNothing(t *testing.T) {
	t.Parallel()

	m := newModel(t)
	called := false
	m.clip = func(string) error { called = true; return nil }
	m.status = "before"
	for _, text := range []string{"", "   ", " \n  "} {
		if cmd := m.copyPicked(text); cmd != nil || called || m.status != "before" {
			t.Errorf("%q: copied %v, status %q, cmd nil = %v", text, called, m.status, cmd == nil)
		}
	}
}

func TestCopyDragCopiesTheScreenWithNoBand(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 40)
	a, _ := m.geometry().at(paneList).textArea()
	var got string
	m.clip = func(s string) error { got = s; return nil }
	m.drag = drag{on: true, area: a, ax: a.x0, ay: a.y0, ex: a.x1, ey: a.y0 + 1}
	bare := m
	bare.drag = drag{}
	want := dragText(bare.draw(), m.drag)
	m.copyDrag()
	if got == "" || got != want {
		t.Errorf("clipboard = %q, want %q", got, want)
	}
}
