package tui

import (
	"strings"
	"testing"

	"github.com/iyay/acta/internal/commits"
	"github.com/iyay/acta/internal/theme"
)

// startCol is the screen column a line of text starts in, inside a box line
// that has its walls on.
func startCol(line string) int {
	r := []rune(line)
	return len(r) - len([]rune(strings.TrimLeft(string(r), " ")))
}

// The detail box pads every side by one cell: text starts under the D of the
// Detail title, one cell stays free before the right wall, and the first and
// last inner lines are empty. The body is drawn by the real markdown renderer,
// so its own margin is checked too.
func TestDetailBoxPadsEverySide(t *testing.T) {
	m := loadedModel(t, map[string][]commits.Commit{"abcdefg#1": fakeCommits("a", 2)})
	th, _ := theme.Builtin(theme.Default)
	m.render = newRenderer(th, true)
	m = onItem(t, m, m.board.Get("PLN-0001").ID)
	b := m.geometry().detail
	lines := column(m.View(), b.x, b.w)[b.y : b.y+b.h]
	col := strings.Index(string([]rune(lines[0])), "Detail")
	if col < 0 {
		t.Fatalf("no Detail title in %q", lines[0])
	}
	col = len([]rune(string([]rune(lines[0]))[:col]))
	inner := lines[1 : len(lines)-1]
	for _, i := range []int{0, len(inner) - 1} {
		if got := strings.TrimSpace(cutWalls(inner[i])); got != "" {
			t.Errorf("inner line %d is not empty: %q", i, got)
		}
	}
	seen := map[string]bool{}
	for _, ln := range inner {
		text := cutWalls(ln)
		if got := []rune(text); len(got) > 0 && got[len(got)-1] != ' ' {
			t.Errorf("a line reaches the last inner column: %q", ln)
		}
		fields := strings.Fields(text)
		if len(fields) == 0 {
			continue
		}
		// The left wall is cell 0, so cutWalls shifts every start by one.
		start := startCol(text) + 1
		for _, key := range []string{"ID", "COMMITS", "a020000", "Plan", "───", "created"} {
			if strings.HasPrefix(fields[0], key) {
				seen[key] = true
				if start != col {
					t.Errorf("%q starts at %d, want %d (the D of Detail)", strings.TrimSpace(text), start, col)
				}
			}
		}
	}
	for _, key := range []string{"ID", "COMMITS", "a020000", "Plan", "───", "created"} {
		if !seen[key] {
			t.Errorf("no line starting with %q on screen", key)
		}
	}
}

// A box too small for the padding draws without it, and no line is ever wider
// than the box, whatever the size.
func TestDetailBoxTooSmallForPaddingStaysInside(t *testing.T) {
	for _, w := range []int{1, 2, 3, 4, 5, 8, 10} {
		for _, h := range []int{1, 2, 3, 4, 5} {
			b := box{w: w + 2, h: h + 2, inner: h}
			dw, dh := b.detailArea()
			if dw > b.textW() || dh > b.inner {
				t.Errorf("w=%d h=%d: area %dx%d is bigger than the box", w, h, dw, dh)
			}
			if (w < 3 || h < 3) && (dw != w || dh != h) {
				t.Errorf("w=%d h=%d: a box too small for padding got %dx%d", w, h, dw, dh)
			}
		}
	}
}
