package tui

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
)

// paneKeys is every key that focuses a box: 1 to 5 the sidebar, 0 the detail.
var paneKeys = []string{"1", "2", "3", "4", "5", "0"}

// paneOfKey is the box a key focuses, so a test can walk the ring without
// repeating the arithmetic.
func paneOfKey(k string) pane {
	if k == "0" {
		return paneDetail
	}
	return pane(k[0] - '1')
}

// keyOf is the key that focuses a box, the way a reader presses it.
func keyOf(p pane) string {
	if p == paneDetail {
		return "0"
	}
	return strconv.Itoa(int(p) + 1)
}

// onTab is a model with the box p open on its tab, the way a reader gets there.
func onTab(t *testing.T, m Model, p pane, tab int) Model {
	t.Helper()
	m = press(m, keyOf(p))
	for range len(sidebar[p].tabs) {
		if m.tab[p] == tab {
			return m
		}
		m = press(m, "]")
	}
	t.Fatalf("pane %d has no tab %d", p, tab)
	return m
}

// TestNumberKeysFocusTheirBox walks every key from every starting box, so no
// key lands anywhere but its own box, from anywhere on the screen.
func TestNumberKeysFocusTheirBox(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	for _, k := range paneKeys {
		for _, start := range paneKeys {
			if got := press(press(m, start), k).focus; got != paneOfKey(k) {
				t.Errorf("from %s press %s: focus %d, want %d", start, k, got, paneOfKey(k))
			}
		}
	}
}

// TestTabCyclesSixBoxes walks the ring both ways from every box: tab moves on
// to the next box and comes back after a full turn, shift+tab walks the same
// ring the other way.
func TestTabCyclesSixBoxes(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	for _, k := range paneKeys {
		for _, key := range []string{"tab", "shift+tab"} {
			step := 1
			if key == "shift+tab" {
				step = -1
			}
			at := paneOfKey(k)
			walked := press(m, k)
			for i := 1; i <= len(paneKeys); i++ {
				walked = press(walked, key)
				at = pane((int(at) + step + len(paneKeys)) % len(paneKeys))
				if walked.focus != at {
					t.Errorf("from %s press %s %d times: focus %d, want %d", k, key, i, walked.focus, at)
					break
				}
			}
		}
	}
}

// TestSidebarTitles reads the titles off the drawn screen and pins the table to
// the keys: the box the reader focuses with a number is the box that number
// names.
func TestSidebarTitles(t *testing.T) {
	v := sized(newModel(t), 160, 50).View()
	for _, want := range []string{
		"[1]─Active",
		"[2]─Specs ─ Scratchpad",
		"[3]─Plans ─ Tasks",
		"[4]─Bugs ─ Debt",
		"[5]─Done",
		"[0]─Detail",
	} {
		if !strings.Contains(v, want) {
			t.Errorf("the view lacks the title %q", want)
		}
	}
	// The number a box wears is the key that focuses it.
	for p := pane(0); int(p) < boxes; p++ {
		if got, want := paneKey(p), "─["+keyOf(p)+"]─"; got != want {
			t.Errorf("box %d wears %q, want %q", p, got, want)
		}
	}
	// Every kind pane draws the two tab names its title names, so the reader
	// can see both kinds without pressing anything.
	for p := 1; p < int(paneDone); p++ {
		names := sidebar[p].tabs
		if got := sidebar[p].title; got != names[0].name+" ─ "+names[1].name {
			t.Errorf("pane %d draws %q, want the title of its own tabs %q", p, got, names[0].name+" ─ "+names[1].name)
		}
	}
}

// TestActiveHoldsOnlyWorkInProgress reads the Active pane of the real fixture
// and checks it holds every item whose work has begun, of every kind, and
// nothing that is only open, raw, todo or finished.
func TestActiveHoldsOnlyWorkInProgress(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	_, b := fixture(t)
	var want []string
	for _, it := range b.Items {
		if !it.Legacy && inProgress(it) {
			want = append(want, it.ID)
		}
	}
	got := ids(m.activeRows())
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("Active holds %q\nwant %q", strings.Join(got, " "), strings.Join(want, " "))
	}
	if len(got) == 0 {
		t.Fatal("the fixture holds no work in progress, so this test proves nothing")
	}
	// Every kind that has work under way is on show, so the pane never reads
	// as one kind of work only.
	kinds := map[board.Kind]bool{}
	for _, id := range got {
		kinds[b.Get(id).Kind] = true
	}
	for _, want := range []board.Kind{board.KindStory, board.KindPlan, board.KindTask, board.KindBug, board.KindScratch} {
		if !kinds[want] {
			t.Errorf("Active holds no %s, so it hides a kind of work in progress", want)
		}
	}
	// Nothing that is only waiting may sneak in.
	for _, r := range m.activeRows() {
		it := b.Get(r.id)
		if it == nil {
			t.Fatalf("Active holds the row %q, which is no item", r.id)
		}
		if !inProgress(it) {
			t.Errorf("Active holds %s with status %s, which is not work in progress", r.id, it.Status)
		}
		if it.Legacy {
			t.Errorf("Active holds the legacy item %s", r.id)
		}
	}
}

// TestDoneFollowsLastSidebarPane checks the Done pane reads the sidebar pane
// and tab that had the focus last: the Scratchpad tab brings Specced and
// Dropped, Bugs brings Fixed and Wontfix, and Active brings no tabs at all and
// every closed item of every kind.
func TestDoneFollowsLastSidebarPane(t *testing.T) {
	t.Run("Scratchpad", func(t *testing.T) {
		m := sized(press(newModel(t), "2", "]"), 160, 50)
		if m.tab[1] != 1 {
			t.Fatalf("the Scratchpad tab is not open: tab %d", m.tab[1])
		}
		m = press(m, "5")
		if got := strings.Join(m.doneTabNames(), " "); got != "Specced Dropped" {
			t.Errorf("after Scratchpad the finished tabs are %q, want Specced and Dropped", got)
		}
		if got := strings.Join(ids(m.doneRows()), " "); got != "scratch/2026-09-28-idea-used" {
			t.Errorf("Specced rows %q", got)
		}
		if got := strings.Join(ids(press(m, "]").doneRows()), " "); got != "scratch/2026-09-28-idea-dropped" {
			t.Errorf("Dropped rows %q", got)
		}
	})
	t.Run("Bugs", func(t *testing.T) {
		m := press(sized(newModel(t), 160, 50), "4", "5")
		if got := strings.Join(m.doneTabNames(), " "); got != "Fixed Wontfix" {
			t.Errorf("after Bugs the finished tabs are %q, want Fixed and Wontfix", got)
		}
		if got := strings.Join(ids(m.doneRows()), " "); got != "bugs/2026-09-24-crash" {
			t.Errorf("Fixed rows %q", got)
		}
		if got := ids(press(m, "]").doneRows()); len(got) != 0 {
			t.Errorf("no bug is wontfix, got %v", got)
		}
	})
	t.Run("Debt", func(t *testing.T) {
		m := press(sized(newModel(t), 160, 50), "4", "]", "5")
		if got := strings.Join(m.doneTabNames(), " "); got != "Done Wontfix" {
			t.Errorf("after Debt the finished tabs are %q, want Done and Wontfix", got)
		}
	})
	t.Run("Active", func(t *testing.T) {
		m := press(sized(newModel(t), 160, 50), "1", "5")
		if got := m.doneTabNames(); len(got) != 0 {
			t.Errorf("after Active the Done pane shows tabs %q, want none", got)
		}
		_, b := fixture(t)
		got := ids(m.doneRows())
		for _, it := range b.Items {
			if it.Legacy || !board.Closed(it.Status) {
				continue
			}
			if !slices.Contains(got, it.ID) {
				t.Errorf("Done after Active leaves out the finished item %s (%s)", it.ID, it.Status)
			}
		}
		if len(got) == 0 {
			t.Error("Done after Active holds no finished item at all")
		}
	})
}

// TestNoDividerRow walks every pane of every tab and checks no row is a rule:
// each row is a real item or the folded legacy group, and no pane draws a
// full-width line of dashes, even where in-progress and not-started items sit
// next to each other.
func TestNoDividerRow(t *testing.T) {
	m := sized(newModel(t), 160, 50)
	_, b := fixture(t)
	checked := 0
	for p := pane(1); p < paneDone; p++ {
		for tab := range sidebar[p].tabs {
			mm := onTab(t, m, p, tab)
			rows := mm.rowsOf(p)
			if len(rows) == 0 {
				continue
			}
			checked++
			for _, r := range rows {
				if r.id == groupRowID {
					continue
				}
				if b.Get(r.id) == nil {
					t.Errorf("pane %d tab %d holds a row %q that is no item", p, tab, r.id)
				}
			}
			// A pane that holds work begun and work not begun is where the old
			// rule sat, so that is the one to look at.
			var going, waiting bool
			for _, r := range rows {
				it := b.Get(r.id)
				if it == nil {
					continue
				}
				if inProgress(it) {
					going = true
				} else {
					waiting = true
				}
			}
			if !going || !waiting {
				continue
			}
			for _, line := range innerLines(mm, paneBox(mm, p)) {
				if strings.Trim(strings.TrimSpace(line), "─") == "" && strings.TrimSpace(line) != "" {
					t.Errorf("pane %d tab %d draws the divider row %q", p, tab, line)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no pane held a row, so this test proves nothing")
	}
}

// TestScratchDetailNamesTheLinkedSpec checks the detail of a scratch item
// names the spec that came out of it, the way a plan names the spec it
// carries. The first board gives the spec a number ID, so the line reads
// SPEC-9; the shared fixture's linked spec has none, so the same line names it
// by path there.
func TestScratchDetailNamesTheLinkedSpec(t *testing.T) {
	cfg := treeCfg(t, map[string]string{
		".acta/scratch/2026-09-28-idea.md":        "# Themes\n\nCatet aja dulu.\n",
		".acta/specs/2026-09-28-themes-design.md": "---\nid: SPEC-9\nparent: scratch/2026-09-28-idea\n---\n# Themes design\n\nThe spec the idea became.\n",
	})
	lines := plainLines(detailLines(t, cfg, "scratch/2026-09-28-idea"))
	if !strings.Contains(strings.Join(lines, "\n"), "SPEC-9") {
		t.Errorf("the detail of a scratch item does not name the spec it produced:\n%s", strings.Join(lines, "\n"))
	}
	fixtureCfg, _ := fixture(t)
	used := plainLines(detailLines(t, fixtureCfg, "scratch/2026-09-28-idea-used"))
	if !strings.Contains(strings.Join(used, "\n"), "specs/2026-09-28-from-scratch-design") {
		t.Errorf("the detail of the used idea does not name its spec:\n%s", strings.Join(used, "\n"))
	}
}
