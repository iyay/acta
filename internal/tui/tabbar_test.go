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
