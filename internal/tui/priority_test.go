package tui

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// prioCfg is a board with bugs and debt lines of every level, and some with
// none, so one fixture covers every sort group.
func prioCfg(t *testing.T) config.Config {
	t.Helper()
	bug := func(id, prio, extra string) string {
		fm := "---\nid: " + id + "\nhash: " + strings.ToLower(strings.ReplaceAll(id, "-", ""))[:7] + "\n"
		if prio != "" {
			fm += "priority: " + prio + "\n"
		}
		return fm + extra + "---\n# " + id + "\n"
	}
	cfg := treeCfg(t, map[string]string{
		".acta/bugs/2026-10-01-a.md": bug("BUG-0001", "", ""),
		".acta/bugs/2026-10-01-b.md": bug("BUG-0002", "low", ""),
		".acta/bugs/2026-10-01-c.md": bug("BUG-0003", "high", ""),
		".acta/bugs/2026-10-01-d.md": bug("BUG-0004", "high", ""),
		".acta/bugs/2026-10-01-e.md": bug("BUG-0005", "medium", ""),
		".acta/bugs/2026-10-01-f.md": bug("BUG-0006", "low", "status: fixed\n"),
		".acta/bugs/2026-10-01-g.md": bug("BUG-0007", "high", "status: fixed\n"),
		".acta/debt/2026-10-01-d.md": "---\nid: DBT-0001\nhash: dbt0001\n---\n# Review NOTEs: D\n\n- [ ] plain\n- [ ] (low) l\n- [ ] (high) h\n",
	})
	return cfg
}

func prioModel(t *testing.T) Model {
	t.Helper()
	cfg := prioCfg(t)
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, b, true)
	m.render = func(md string, _ int) string { return md }
	return m
}

func TestBugsSortByPriorityBothWays(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs))
	want := []string{"bugs/2026-10-01-c", "bugs/2026-10-01-d", "bugs/2026-10-01-e", "bugs/2026-10-01-b", "bugs/2026-10-01-a"}
	if got := rowIDs(m); !slices.Equal(got, want) {
		t.Fatalf("Bugs rows %v, want %v", got, want)
	}
	m = press(m, "o")
	want = []string{"bugs/2026-10-01-d", "bugs/2026-10-01-c", "bugs/2026-10-01-e", "bugs/2026-10-01-b", "bugs/2026-10-01-a"}
	if got := rowIDs(m); !slices.Equal(got, want) {
		t.Fatalf("Bugs rows after o %v, want %v", got, want)
	}
}

func TestBugsDonePaneKeepsIDOrder(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabBugs), "tab")
	if got, want := doneRowIDs(m), []string{"bugs/2026-10-01-f", "bugs/2026-10-01-g"}; !slices.Equal(got, want) {
		t.Fatalf("Fixed rows %v, want id order %v", got, want)
	}
}

func TestDebtSortByPriority(t *testing.T) {
	t.Parallel()

	m := press(prioModel(t), tabKey(tabDebt))
	want := []string{"debt/2026-10-01-d#item-3", "debt/2026-10-01-d#item-2", "debt/2026-10-01-d#item-1"}
	if got := rowIDs(m); !slices.Equal(got, want) {
		t.Fatalf("Debt rows %v, want %v", got, want)
	}
}

func TestRowShowsPriorityTag(t *testing.T) {
	t.Parallel()

	m := prioModel(t)
	for _, c := range []struct{ id, want string }{
		{"bugs/2026-10-01-c", "BUG-0003  H BUG-0003"},
		{"bugs/2026-10-01-e", "BUG-0005  M BUG-0005"},
		{"bugs/2026-10-01-b", "BUG-0002  L BUG-0002"},
		{"bugs/2026-10-01-a", "BUG-0001  BUG-0001"},
	} {
		if got := m.rowText(row{id: c.id}, m.board.Get(c.id), 80); got != c.want {
			t.Errorf("row %s = %q, want %q", c.id, got, c.want)
		}
	}
}

// The tag letter wears the brush of its level, so the eye finds a high bug
// without reading the letter. The test is not parallel because the color
// profile is one setting for the whole test binary.
func TestRowPaintsPriorityTag(t *testing.T) {
	m := prioModel(t)
	withTrueColor(func() {
		red := sgr.FindString(m.styles.priority["high"].Render("H"))
		got := m.paintID("BUG-0003  H BUG-0003", m.board.Get("bugs/2026-10-01-c"), lipgloss.NewStyle())
		if !strings.Contains(got, red+"H") {
			t.Errorf("the tag letter is not painted with the priority brush: %q", got)
		}
	})
}

func TestDetailShowsPriorityOnlyWhenSet(t *testing.T) {
	t.Parallel()

	cfg := prioCfg(t)
	if got := labelsOf(detailLines(t, cfg, "bugs/2026-10-01-c")); !slices.Contains(got, "PRIORITY") {
		t.Errorf("high bug labels %v, want PRIORITY", got)
	}
	if got := labelsOf(detailLines(t, cfg, "bugs/2026-10-01-a")); slices.Contains(got, "PRIORITY") {
		t.Errorf("unset bug labels %v, want no PRIORITY", got)
	}
}

func TestByPriorityIsStableAndCopies(t *testing.T) {
	t.Parallel()

	mk := func(id, p string) *board.Item { return &board.Item{ID: id, Priority: p} }
	in := []*board.Item{mk("a", ""), mk("b", "low"), mk("c", "high"), mk("d", ""), mk("e", "high")}
	if got, want := itemIDs(byPriority(in)), []string{"c", "e", "b", "a", "d"}; !slices.Equal(got, want) {
		t.Fatalf("byPriority %v, want %v", got, want)
	}
	if got := itemIDs(in); !slices.Equal(got, []string{"a", "b", "c", "d", "e"}) {
		t.Fatalf("input changed to %v", got)
	}
}
