package tui

import "testing"

// A tab nobody has moved the cursor on still holds an empty cursor, and an
// empty cursor already shows row 0. This is the tab BUG-0010 was filed about.
func untouchedTab(t *testing.T) Model {
	t.Helper()
	m := paneModel(t, paneDetail)
	if sel := m.sel[paneList]; sel != "" {
		t.Fatalf("the tab should hold no cursor yet, it holds %q", sel)
	}
	return m
}

// scrollBody scrolls the detail box with the keys, which move the body and
// never the cursor, and gives back the place it scrolled to.
func scrollBody(m Model, times int) Model {
	for range times {
		m = press(m, "j")
	}
	return m
}

// walkCursor sends the focus back to the list and steps the cursor down, the
// way a reader walks to the next row. That is what writes a cursor down for
// the first time on a tab that had none.
func walkCursor(m Model) Model {
	return press(m, "esc", "j")
}

// readBody walks the cursor, hands the focus to the detail box and scrolls
// the body down, which is the screen the reader ends up reading.
func readBody(m Model, times int) Model {
	return scrollBody(press(walkCursor(m), "shift+tab"), times)
}

// The first reload must not throw away the scroll a reader made on a tab
// whose cursor was never moved: the item on show is the same one it was, so
// the detail box keeps the line it was on.
func TestReloadKeepsTheScrollOfATabWhoseCursorNeverMoved(t *testing.T) {
	t.Parallel()

	m := scrollBody(untouchedTab(t), 6)
	off := m.off[paneDetail]
	if off < 1 {
		t.Fatalf("the detail box should have scrolled, off is %d", off)
	}
	m = reloaded(m)
	if m.off[paneDetail] != off {
		t.Errorf("the reload moved the detail box to %d, the reader left it on line %d", m.off[paneDetail], off)
	}
}

// The wheel scrolls the same box, so a reload must keep that place too. The
// helper sends the frame tick as well, which is the moment the notches the
// wheel gathered land.
func TestReloadKeepsTheScrollMadeWithTheWheel(t *testing.T) {
	t.Parallel()

	m := untouchedTab(t)
	b := scrollBox(m, paneDetail)
	m = wheel(m, b.x+1, b.y+2, false)
	off := m.off[paneDetail]
	if off < 1 {
		t.Fatalf("the wheel should have scrolled the detail box, off is %d", off)
	}
	m = reloaded(m)
	if m.off[paneDetail] != off {
		t.Errorf("the reload moved the wheel-scrolled detail box to %d, the reader left it on line %d", m.off[paneDetail], off)
	}
}

// Everything else a reload does stays as it was. Only an item on show that
// really changes sends the detail box back to its own top.
func TestOnlyARealChangeOfTheItemOnShowResetsTheScroll(t *testing.T) {
	t.Parallel()

	t.Run("a cursor that was moved keeps the scroll when the item stays", func(t *testing.T) {
		t.Parallel()

		m := readBody(untouchedTab(t), 6)
		off := m.off[paneDetail]
		if off < 1 {
			t.Fatalf("the detail box should have scrolled, off is %d", off)
		}
		m = reloaded(m)
		if m.off[paneDetail] != off {
			t.Errorf("the reload moved the detail box to %d, the reader left it on line %d", m.off[paneDetail], off)
		}
	})

	t.Run("an item on show that is gone sends the box back to the top", func(t *testing.T) {
		t.Parallel()

		m := readBody(untouchedTab(t), 6)
		it := m.Selected()
		if it == nil {
			t.Fatal("the row under the cursor holds no item")
		}
		if m.off[paneDetail] < 1 {
			t.Fatalf("the detail box should have scrolled, off is %d", m.off[paneDetail])
		}
		m = reloadedWithout(m, it.ID)
		if m.off[paneDetail] != 0 {
			t.Errorf("a reload that took the item away left the new one on line %d, want 0", m.off[paneDetail])
		}
	})

	t.Run("a walk to another row on a tab with no cursor sends the box back to the top", func(t *testing.T) {
		t.Parallel()

		m := scrollBody(untouchedTab(t), 6)
		if m.off[paneDetail] < 1 {
			t.Fatalf("the detail box should have scrolled, off is %d", m.off[paneDetail])
		}
		m = walkCursor(m)
		if m.off[paneDetail] != 0 {
			t.Errorf("the walk left the new item on line %d of its body, want 0", m.off[paneDetail])
		}
	})

	t.Run("a walk to another row after the cursor moved sends the box back to the top", func(t *testing.T) {
		t.Parallel()

		m := readBody(untouchedTab(t), 6)
		off := m.off[paneDetail]
		if off < 1 {
			t.Fatalf("the detail box should have scrolled, off is %d", off)
		}
		m = press(m, "esc", "j", "j")
		if m.off[paneDetail] != 0 {
			t.Errorf("the walk left the new item on line %d, the reader was on line %d", m.off[paneDetail], off)
		}
	})
}
