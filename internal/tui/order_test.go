package tui

import (
	"fmt"
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

func withID(id, short, date string) *board.Item {
	it := spec(id, date)
	it.ShortID = short
	return it
}

// The number in the short id wins over the date and over the file name, so
// items made on one day come out in the order they were made, and the input
// slice the board owns is never touched.
func TestOrderedSortsByIDNumberBothWays(t *testing.T) {
	t.Parallel()

	in := []*board.Item{
		withID("scratch/2026-09-29-a", "SCR-0018", "2026-09-29"),
		withID("scratch/2026-09-30-m", "SCR-0009", "2026-09-30"),
		withID("scratch/2026-09-29-z", "SCR-0017", "2026-09-29"),
	}
	if got, want := itemIDs(ordered(in, false)), []string{"scratch/2026-09-30-m", "scratch/2026-09-29-z", "scratch/2026-09-29-a"}; !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	if got, want := itemIDs(ordered(in, true)), []string{"scratch/2026-09-29-a", "scratch/2026-09-29-z", "scratch/2026-09-30-m"}; !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
	if want := []string{"scratch/2026-09-29-a", "scratch/2026-09-30-m", "scratch/2026-09-29-z"}; !slices.Equal(itemIDs(in), want) {
		t.Fatalf("ordered moved its input: %v", itemIDs(in))
	}
}

// Two items with no id have no number to sort by, so the date in the file
// name decides, the way it did before ids.
func TestOrderedFallsBackToFileDateWhenNoID(t *testing.T) {
	t.Parallel()

	in := []*board.Item{spec("specs/2026-09-20-b", "2026-09-20"), spec("specs/2026-09-01-a", "2026-09-01"), spec("specs/2026-09-10-c", "2026-09-10")}
	if got, want := itemIDs(ordered(in, false)), []string{"specs/2026-09-01-a", "specs/2026-09-10-c", "specs/2026-09-20-b"}; !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	if got, want := itemIDs(ordered(in, true)), []string{"specs/2026-09-20-b", "specs/2026-09-10-c", "specs/2026-09-01-a"}; !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
	if want := []string{"specs/2026-09-20-b", "specs/2026-09-01-a", "specs/2026-09-10-c"}; !slices.Equal(itemIDs(in), want) {
		t.Fatalf("ordered moved its input: %v", itemIDs(in))
	}
}

// A number can be the same while the day differs, so the day breaks the tie
// and not the file name. The two items sit in different folders, so the file
// order would put the newer one first.
func TestOrderedBreaksAnIDTieByDate(t *testing.T) {
	t.Parallel()

	in := []*board.Item{
		withID("specs/2026-09-01-a", "SPC-0005", "2026-09-01"),
		withID("plans/2026-09-02-b", "PLN-0005", "2026-09-02"),
	}
	if got, want := itemIDs(ordered(in, false)), []string{"specs/2026-09-01-a", "plans/2026-09-02-b"}; !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
	if got, want := itemIDs(ordered(in, true)), []string{"plans/2026-09-02-b", "specs/2026-09-01-a"}; !slices.Equal(got, want) {
		t.Fatalf("newest: got %v, want %v", got, want)
	}
}

// Same number and same day still need an order, so the file decides. Both
// kinds of item are covered: ones that carry the same id, and ones with no
// id at all.
func TestOrderedFallsBackToFileWhenNumberAndDateTie(t *testing.T) {
	t.Parallel()

	sameID := []*board.Item{withID("specs/2026-09-29-b", "SPC-0005", "2026-09-29"), withID("specs/2026-09-29-a", "SPC-0005", "2026-09-29")}
	if got, want := itemIDs(ordered(sameID, false)), []string{"specs/2026-09-29-a", "specs/2026-09-29-b"}; !slices.Equal(got, want) {
		t.Fatalf("same number oldest: got %v, want %v", got, want)
	}
	if got, want := itemIDs(ordered(sameID, true)), []string{"specs/2026-09-29-b", "specs/2026-09-29-a"}; !slices.Equal(got, want) {
		t.Fatalf("same number newest: got %v, want %v", got, want)
	}
	noID := []*board.Item{spec("specs/2026-09-29-b", "2026-09-29"), spec("specs/2026-09-29-a", "2026-09-29")}
	if got, want := itemIDs(ordered(noID, false)), []string{"specs/2026-09-29-a", "specs/2026-09-29-b"}; !slices.Equal(got, want) {
		t.Fatalf("no id: got %v, want %v", got, want)
	}
}

// An item with no id, or an id that holds no number, has nothing to sort by,
// so it goes after every item that has a number, whichever way the list runs.
// An item that has a number but no day still counts as having a number.
func TestOrderedPutsItemsWithNoIDLast(t *testing.T) {
	t.Parallel()

	in := []*board.Item{
		withID("specs/2026-01-01-none", "", "2026-01-01"),
		withID("specs/2026-01-02-prefix", "SPC-", "2026-01-02"),
		withID("specs/2026-09-01-word", "SPC-abc", "2026-09-01"),
		withID("specs/undated-numbered", "SPC-0030", ""),
		withID("specs/2026-09-20-b", "SPC-0020", "2026-09-20"),
		withID("specs/2026-09-10-a", "SPC-0010", "2026-09-10"),
	}
	for _, newest := range []bool{false, true} {
		got := itemIDs(ordered(in, newest))
		for _, want := range []string{"specs/2026-09-20-b", "specs/2026-09-10-a", "specs/undated-numbered"} {
			if !slices.Contains(got[:3], want) {
				t.Errorf("newest=%v: item with a number is not up front: %v", newest, got)
			}
		}
		for _, want := range []string{"specs/2026-01-01-none", "specs/2026-01-02-prefix", "specs/2026-09-01-word"} {
			if !slices.Contains(got[3:], want) {
				t.Errorf("newest=%v: item with a broken id is not last: %v", newest, got)
			}
		}
	}
}

// A task id carries its plan's number, so tasks sort with their plan.
func TestOrderedSortsTasksByPlanNumber(t *testing.T) {
	t.Parallel()

	task := func(id, short string) *board.Item {
		return &board.Item{ID: id, ShortID: short, Kind: board.KindTask, Date: "2026-09-29", Status: "in-progress"}
	}
	in := []*board.Item{
		task("plans/2026-09-29-a#task-1", "PLN-0033.01"),
		task("plans/2026-09-29-a#task-2", "PLN-0033.02"),
		task("plans/2026-09-29-z#task-1", "PLN-0002.01"),
	}
	want := []string{"plans/2026-09-29-z#task-1", "plans/2026-09-29-a#task-1", "plans/2026-09-29-a#task-2"}
	if got := itemIDs(ordered(in, false)); !slices.Equal(got, want) {
		t.Fatalf("oldest: got %v, want %v", got, want)
	}
}

// Every task of one plan shares a number, a day and a file, so nothing
// separates them and the sort must leave them in file order. The plan holds
// twelve tasks and another plan holds one more, which is the size an
// unstable sort really does shuffle.
func TestOrderedKeepsTasksUnderTheirPlanInFileOrder(t *testing.T) {
	t.Parallel()

	task := func(id, short string) *board.Item {
		return &board.Item{ID: id, ShortID: short, Kind: board.KindTask, Date: "2026-09-29", Status: "todo"}
	}
	in := make([]*board.Item, 0, 13)
	want := make([]string, 0, 13)
	for n := 1; n <= 12; n++ {
		id := fmt.Sprintf("plans/2026-09-29-a#task-%d", n)
		in = append(in, task(id, "PLN-0033.01"))
		want = append(want, id)
	}
	in = append(in, task("plans/2026-09-29-b#task-1", "PLN-0007.01"))
	want = append([]string{"plans/2026-09-29-b#task-1"}, want...)
	// Newest first puts the plan with the bigger number first, but the
	// twelve tasks of one plan stay in file order, because a stable sort
	// never swaps items it cannot tell apart.
	newest := append(slices.Clone(want[1:]), want[0])
	for _, c := range []struct {
		flip bool
		want []string
	}{{false, want}, {true, newest}} {
		if got := itemIDs(ordered(in, c.flip)); !slices.Equal(got, c.want) {
			t.Fatalf("newest=%v: got %v, want %v", c.flip, got, c.want)
		}
	}
}

func TestOrderedPutsItemsWithNoDateLast(t *testing.T) {
	t.Parallel()

	in := []*board.Item{spec("specs/undated", ""), spec("specs/2026-09-01-a", "2026-09-01"), spec("specs/2026-09-20-b", "2026-09-20")}
	for _, newest := range []bool{false, true} {
		got := itemIDs(ordered(in, newest))
		if got[len(got)-1] != "specs/undated" {
			t.Fatalf("newest=%v: undated item not last: %v", newest, got)
		}
	}
}

func TestIDNumReadsTheNumber(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		short string
		n     int
		ok    bool
	}{
		{"SCR-0017", 17, true}, {"PLN-0033.01", 33, true}, {"BUG-10000", 10000, true},
		{"", 0, false}, {"SCR-", 0, false}, {"SCR-abc", 0, false}, {"SCR0017", 0, false},
	} {
		n, ok := idNum(&board.Item{ShortID: c.short})
		if n != c.n || ok != c.ok {
			t.Errorf("idNum(%q) = %d %v, want %d %v", c.short, n, ok, c.n, c.ok)
		}
	}
}

func TestOFlipsOnlyTheFocusedBox(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	m := newModel(t)
	m = press(m, "0", "o")
	for p := range sidePanes {
		if m.newest[p] {
			t.Fatalf("o in the detail box flipped pane %d", p)
		}
	}
}

func TestDonePaneFollowsFileDateNotCommitTime(t *testing.T) {
	t.Parallel()

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
