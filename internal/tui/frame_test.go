package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TestSplitAlwaysFillsTheHeight checks the two left panes never leave a gap
// and never push each other off screen. Every height the terminal can hand us
// goes through, including the tiny ones that have no room to divide.
func TestSplitAlwaysFillsTheHeight(t *testing.T) {
	for h := range 2 {
		// A terminal with no room yet must not panic, only split badly.
		split(h)
	}
	for h := 2; h <= 80; h++ {
		top, bottom := split(h)
		if top+bottom != h || top < 1 || bottom < 1 {
			t.Errorf("split(%d) = %d, %d", h, top, bottom)
		}
	}
}

// TestViewFitsEveryTerminalSize walks every size a terminal can report: each
// height from 3 up to 60, each width from 1 to 200. Three claims at once. The
// left column is exactly as tall as the detail pane, the screen has no more
// lines than the terminal has rows, and no line reaches past the right edge.
func TestViewFitsEveryTerminalSize(t *testing.T) {
	m := newModel(t)
	for h := 3; h <= 60; h++ {
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
			if g.wide && g.open.h+g.done.h != g.detail.h {
				t.Fatalf("%dx%d: left %d+%d, detail %d", w, h, g.open.h, g.done.h, g.detail.h)
			}
		}
	}
}

// TestViewSurvivesTinyTerminals covers the sizes too small to hold a pane at
// all. A window squeezed to nothing must still draw without panicking.
func TestViewSurvivesTinyTerminals(t *testing.T) {
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
	m := sized(newModel(t), 120, 40)
	next, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if cmd == nil || fmt.Sprint(cmd()) != fmt.Sprint(tea.ClearScreen()) {
		t.Fatal("a size change must clear the screen so old rows do not stay")
	}
	got := next.(Model)
	if got.width != 100 || got.height != 30 {
		t.Fatalf("the model kept %dx%d, want 100x30", got.width, got.height)
	}
}
