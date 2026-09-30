package board

import "testing"

func TestSplitPriority(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ in, level, rest string }{
		{"(high) a note", "high", "a note"},
		{"(medium) a note", "medium", "a note"},
		{"(low) a note", "low", "a note"},
		{"(hgh) a note", "", "(hgh) a note"},
		{"(high)a note", "", "(high)a note"},
		{"a note (high) here", "", "a note (high) here"},
		{"", "", ""},
	} {
		level, rest := SplitPriority(c.in)
		if level != c.level || rest != c.rest {
			t.Errorf("SplitPriority(%q) = %q, %q; want %q, %q", c.in, level, rest, c.level, c.rest)
		}
	}
}

func TestBugPriorityFromFrontmatter(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"bugs/2026-10-01-high.md":  "---\npriority: high\n---\n# High\n",
		"bugs/2026-10-01-none.md":  "---\nref: X-1\n---\n# None\n",
		"bugs/2026-10-01-wrong.md": "---\npriority: urgent\n---\n# Wrong\n",
	})
	if got := b.Get("bugs/2026-10-01-high").Priority; got != "high" {
		t.Errorf("high bug Priority = %q", got)
	}
	none := b.Get("bugs/2026-10-01-none")
	if none.Priority != "" || len(none.Problems) != 0 {
		t.Errorf("unset bug Priority = %q, Problems = %v; want empty and none", none.Priority, none.Problems)
	}
	wrong := b.Get("bugs/2026-10-01-wrong")
	if wrong.Priority != "" || !hasProblem(wrong, "unknown priority urgent") {
		t.Errorf("bad bug Priority = %q, Problems = %v", wrong.Priority, wrong.Problems)
	}
}

func TestSpecIgnoresPriority(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{"specs/2026-10-01-s.md": "---\npriority: high\n---\n# S\n"})
	if got := b.Get("specs/2026-10-01-s").Priority; got != "" {
		t.Errorf("spec Priority = %q, want empty", got)
	}
}

func TestDebtItemPriorityFromTag(t *testing.T) {
	t.Parallel()

	b := boardWith(t, map[string]string{
		"debt/2026-10-01-d.md": "---\nid: DBT-0001\nhash: aaaaaaa\n---\n# Review NOTEs: D\n\n- [ ] (high) first\n- [x] (low) second\n- [ ] (hgh) third\n- [ ] fourth\n",
	})
	for _, c := range []struct{ id, level, title string }{
		{"debt/2026-10-01-d#item-1", "high", "first"},
		{"debt/2026-10-01-d#item-2", "low", "second"},
		{"debt/2026-10-01-d#item-3", "", "(hgh) third"},
		{"debt/2026-10-01-d#item-4", "", "fourth"},
	} {
		it := b.Get(c.id)
		if it == nil {
			t.Fatalf("no %s on the board", c.id)
		}
		if it.Priority != c.level || it.Title != c.title {
			t.Errorf("%s Priority = %q Title = %q; want %q %q", c.id, it.Priority, it.Title, c.level, c.title)
		}
		if len(it.Problems) != 0 {
			t.Errorf("%s Problems = %v, want none", c.id, it.Problems)
		}
	}
}
