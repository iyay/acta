package board

import (
	"slices"
	"testing"
)

// A plan waits exactly while one of the plans in its after list is not done.
// A plan with no after list never waits.
func TestAfterMarksWaiting(t *testing.T) {
	t.Parallel()

	plan := func(id, after, box string) string {
		front := "id: " + id + "\nhash: h" + id[4:] + "x\n"
		if after != "" {
			front += "after: [" + after + "]\n"
		}
		return "---\n" + front + "---\n# Plan " + id + "\n\n### Task 1: Do it\n\n- [" + box + "] Red: it fails\n"
	}

	cases := []struct {
		name    string
		files   map[string]string
		waiting bool
	}{
		{"after an open plan", map[string]string{
			"plans/2026-10-01-a.md": plan("PLN-0001", "", " "),
			"plans/2026-10-01-b.md": plan("PLN-0002", "PLN-0001", " "),
		}, true},
		{"after a done plan", map[string]string{
			"plans/2026-10-01-a.md": plan("PLN-0001", "", "x"),
			"plans/2026-10-01-b.md": plan("PLN-0002", "PLN-0001", " "),
		}, false},
		{"after one done and one open plan", map[string]string{
			"plans/2026-10-01-a.md": plan("PLN-0001", "", "x"),
			"plans/2026-10-01-c.md": plan("PLN-0003", "", " "),
			"plans/2026-10-01-b.md": plan("PLN-0002", "PLN-0001, PLN-0003", " "),
		}, true},
		{"no after list", map[string]string{
			"plans/2026-10-01-a.md": plan("PLN-0001", "", " "),
			"plans/2026-10-01-b.md": plan("PLN-0002", "", " "),
		}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := boardWith(t, c.files)
			p := b.Get("PLN-0002")
			if p == nil {
				t.Fatal("PLN-0002 missing")
			}
			if p.Waiting != c.waiting {
				t.Errorf("waiting %v, want %v (after %v)", p.Waiting, c.waiting, p.After)
			}
		})
	}
	b := boardWith(t, map[string]string{
		"plans/2026-10-01-a.md": plan("PLN-0001", "", " "),
		"plans/2026-10-01-b.md": plan("PLN-0002", "PLN-0001", " "),
	})
	if got := b.Get("PLN-0002").After; !slices.Equal(got, []string{"PLN-0001"}) {
		t.Errorf("after %v, want [PLN-0001]", got)
	}
}
