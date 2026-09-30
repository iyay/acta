package tui

import (
	"fmt"
	"strings"
	"testing"
)

// TestClickOnATopTabOpensIt reads the drawn bar the way a person does: every
// name it finds opens its tab when clicked, and the cells between names do not.
func TestClickOnATopTabOpensIt(t *testing.T) {
	t.Parallel()

	base := newModel(t)
	for w := 20; w <= 200; w++ {
		m := sized(base, w, 30)
		line := []rune(plain(strings.Split(m.View(), "\n")[1]))
		named := make([]bool, len(line))
		for i, tab := range topTabs {
			name := []rune(fmt.Sprintf("%d %s", i+1, tab.name))
			at := strings.Index(string(line), string(name))
			if at < 0 {
				continue
			}
			x := len([]rune(string(line)[:at]))
			for dx := range name {
				named[x+dx] = true
				for y := 0; y < barRows; y++ {
					if got := click(m, x+dx, y).top; got != i {
						t.Fatalf("at %d columns a click on %q (x %d, y %d) opened tab %d", w, string(name), x+dx, y, got)
					}
				}
			}
		}
		for x := range line {
			if named[x] {
				continue
			}
			for y := range barRows {
				if got := click(m, x, y).top; got != m.top {
					t.Fatalf("at %d columns a click on the blank cell %d on row %d opened tab %d", w, x, y, got)
				}
			}
		}
	}
}

// TestClickOnTheBarIsIgnoredUnderThePopups keeps the bar quiet while help or
// the search box is open, the same as every other click.
func TestClickOnTheBarIsIgnoredUnderThePopups(t *testing.T) {
	t.Parallel()

	m := sized(newModel(t), 120, 30)
	_, spans := m.barTabs(m.width - 2)
	x := 1 + spans[tabPlans].x
	for _, keys := range [][]string{{"?"}, {"/"}, {"n"}} {
		if got := click(press(m, keys...), x, 1).top; got != m.top {
			t.Errorf("after %v a click on Plans opened tab %d", keys, got)
		}
	}
}

// TestClickOnTheBarTakesTheHighlightOff proves the press that ends a pick
// also takes the band off the screen. The bar holds no words, so a press on
// one of its blank cells drops the pick, and the frame has to be drawn again
// or the band stays on with nothing picked under it.
func TestClickOnTheBarTakesTheHighlightOff(t *testing.T) {
	t.Parallel()

	m := paneModel(t, paneDetail)
	a, _ := scrollBox(m, paneDetail).textArea()
	before := m.View()
	m, _ = dragged(m, a.x0, a.y0, a.x0+4, a.y0)
	if !m.drag.shown() || m.View() == before {
		t.Fatal("the drag left no highlight on the screen to clear")
	}
	// The first cell after the bar's left wall is blank at every width, since
	// the first name starts one cell further in.
	m.View() // the frame on screen is the one with the band
	next := click(m, 1, 1)
	if next.drag.shown() {
		t.Error("the press left the highlight on")
	}
	if next.same {
		t.Error("the screen kept a highlight that is gone")
	}
	// A model that draws its own frame shows what the screen should be.
	shown := next.View()
	fresh := next
	fresh.same = false
	if want := fresh.View(); shown != want {
		t.Errorf("the screen still shows the band:\n%s", shown)
	}
}
