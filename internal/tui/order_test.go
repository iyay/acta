package tui

import (
	"slices"
	"testing"

	"github.com/iyay/acta/internal/board"
)

func itemIDs(items []*board.Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.ID
	}
	return out
}

func spec(id, date string) *board.Item {
	return &board.Item{ID: id, Kind: board.KindStory, Title: id, Date: date, Status: "open"}
}

func TestOrderedSortsByFileDateBothWays(t *testing.T) {
	in := []*board.Item{spec("specs/2026-09-20-b", "2026-09-20"), spec("specs/2026-09-01-a", "2026-09-01"), spec("specs/2026-09-10-c", "2026-09-10")}
	if got, want := itemIDs(ordered(in, false)), []string{"specs/2026-09-01-a", "specs/2026-09-10-c", "specs/2026-09-20-b"}; !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	if got, want := itemIDs(ordered(in, true)), []string{"specs/2026-09-20-b", "specs/2026-09-10-c", "specs/2026-09-01-a"}; !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
	if in[0].ID != "specs/2026-09-20-b" {
		t.Fatalf("ordered sorted its input in place")
	}
}

func TestOrderedSameDateFallsBackToIDBothWays(t *testing.T) {
	in := []*board.Item{spec("specs/2026-09-29-b", "2026-09-29"), spec("specs/2026-09-29-a", "2026-09-29")}
	if got, want := itemIDs(ordered(in, false)), []string{"specs/2026-09-29-a", "specs/2026-09-29-b"}; !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	if got, want := itemIDs(ordered(in, true)), []string{"specs/2026-09-29-b", "specs/2026-09-29-a"}; !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
}

func TestOrderedKeepsTasksUnderTheirPlanInFileOrder(t *testing.T) {
	task := func(id string) *board.Item {
		return &board.Item{ID: id, Kind: board.KindTask, Date: "2026-09-29", Status: "todo"}
	}
	in := []*board.Item{task("plans/2026-09-29-x#task-1"), task("plans/2026-09-29-x#task-2"), task("plans/2026-09-29-x#task-3")}
	want := []string{"plans/2026-09-29-x#task-1", "plans/2026-09-29-x#task-2", "plans/2026-09-29-x#task-3"}
	for _, newest := range []bool{false, true} {
		if got := itemIDs(ordered(in, newest)); !slices.Equal(got, want) {
			t.Fatalf("newest=%v: got %v, want %v", newest, got, want)
		}
	}
}

func TestOrderedPutsItemsWithNoDateLast(t *testing.T) {
	in := []*board.Item{spec("specs/undated", ""), spec("specs/2026-09-01-a", "2026-09-01"), spec("specs/2026-09-20-b", "2026-09-20")}
	for _, newest := range []bool{false, true} {
		got := itemIDs(ordered(in, newest))
		if got[len(got)-1] != "specs/undated" {
			t.Fatalf("newest=%v: undated item not last: %v", newest, got)
		}
	}
}

func TestOFlipsOnlyTheFocusedBox(t *testing.T) {
	m := newModel(t)
	m.board = &board.Board{Items: []*board.Item{
		spec("specs/2026-09-20-b", "2026-09-20"),
		spec("specs/2026-09-01-a", "2026-09-01"),
		{ID: "specs/2026-09-20-done", Kind: board.KindStory, Date: "2026-09-20", Status: "done"},
		{ID: "specs/2026-09-01-done", Kind: board.KindStory, Date: "2026-09-01", Status: "done"},
	}}
	m = press(m, tabKey(tabSpecs))
	if got, want := ids(m.rowsOf(paneList)), []string{"specs/2026-09-01-a", "specs/2026-09-20-b"}; !slices.Equal(got, want) {
		t.Fatalf("start: got %v, want oldest first %v", got, want)
	}
	m = press(m, "o")
	if got, want := ids(m.rowsOf(paneList)), []string{"specs/2026-09-20-b", "specs/2026-09-01-a"}; !slices.Equal(got, want) {
		t.Fatalf("after o: got %v, want newest first %v", got, want)
	}
	if got, want := ids(m.rowsOf(paneDone)), []string{"specs/2026-09-01-done", "specs/2026-09-20-done"}; !slices.Equal(got, want) {
		t.Fatalf("o in the List box moved the Done box: got %v, want %v", got, want)
	}
	m = press(m, "o")
	if got, want := ids(m.rowsOf(paneList)), []string{"specs/2026-09-01-a", "specs/2026-09-20-b"}; !slices.Equal(got, want) {
		t.Fatalf("second o: got %v, want oldest first again %v", got, want)
	}
}

// TestOKeepsTheSelectedItemSelected reads the selection by its id, because
// the row number moves when the list is reversed. The board is the real
// fixture, so Get can still find the item behind the cursor.
func TestOKeepsTheSelectedItemSelected(t *testing.T) {
	m := press(newModel(t), tabKey(tabSpecs))
	m = press(m, "j")
	before := m.Selected().ID
	m = press(m, "o")
	// The rows are reversed, so the cursor has to change its row number to
	// stay on the same item.
	if got, want := m.rowsOf(paneList)[m.cursor()].id, before; got != want {
		t.Fatalf("after o the cursor is on row %q, want %q", got, want)
	}
	if got := m.Selected().ID; got != before {
		t.Fatalf("o moved the selection from %q to %q", before, got)
	}
}

func TestOInTheDetailBoxChangesNoPane(t *testing.T) {
	m := newModel(t)
	m = press(m, "0", "o")
	for p := range sidePanes {
		if m.newest[p] {
			t.Fatalf("o in the detail box flipped pane %d", p)
		}
	}
}

func TestDonePaneFollowsFileDateNotCommitTime(t *testing.T) {
	m := newModel(t)
	m.board = &board.Board{Items: []*board.Item{
		{ID: "specs/2026-09-20-late", Kind: board.KindStory, Date: "2026-09-20", Status: "done"},
		{ID: "specs/2026-09-01-early", Kind: board.KindStory, Date: "2026-09-01", Status: "done"},
	}}
	m = press(m, tabKey(tabSpecs))
	if got, want := ids(m.rowsOf(paneDone)), []string{"specs/2026-09-01-early", "specs/2026-09-20-late"}; !slices.Equal(got, want) {
		t.Fatalf("Done pane: got %v, want oldest first %v", got, want)
	}
}
