package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// splitPlan is a one-task plan of spec s. started is its started stamp, or
// "" when the plan has not landed with one; box is " " or "x".
func splitPlan(id, started, box string) string {
	front := "id: " + id + "\nparent: specs/2026-10-01-s\n"
	if started != "" {
		front += "started: \"" + started + "\"\n"
	}
	return "---\n" + front + "---\n# Plan " + id + "\n\n### Task 1: Do it\n- [" + box + "] a\n"
}

func specFront(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// SpecDates gives a split spec the earliest started of its plans and one
// finished once every plan is done, whichever plan lands first. A later land
// may move started earlier, never later.
func TestSpecDatesBothLandOrders(t *testing.T) {
	const early, late = "2026-10-01 10:00:00", "2026-10-01 11:00:00"
	orders := []struct {
		name               string
		firstA, firstB     string // plan files on the parent after the first land
		firstStarted       string
		secondA, secondB   string
		wantFinishedBefore bool
	}{
		{
			// A started first and lands first.
			name:   "first started lands first",
			firstA: splitPlan("PLN-0001", early, "x"), firstB: splitPlan("PLN-0002", "", " "),
			firstStarted: early,
			secondA:      splitPlan("PLN-0001", early, "x"), secondB: splitPlan("PLN-0002", late, "x"),
		},
		{
			// A started first but lands last.
			name:   "first started lands last",
			firstA: splitPlan("PLN-0001", "", " "), firstB: splitPlan("PLN-0002", late, "x"),
			firstStarted: late,
			secondA:      splitPlan("PLN-0001", early, "x"), secondB: splitPlan("PLN-0002", late, "x"),
		},
	}
	for _, o := range orders {
		t.Run(o.name, func(t *testing.T) {
			cfg := repoWith(t, map[string]string{
				".acta/specs/2026-10-01-s.md": "---\nid: SPC-0009\n---\n# S\n",
				".acta/plans/2026-10-01-a.md": o.firstA,
				".acta/plans/2026-10-01-b.md": o.firstB,
			})
			spec := filepath.Join(cfg.RepoRoot, ".acta/specs/2026-10-01-s.md")
			if _, err := SpecDates(cfg, "PLN-0001"); err != nil {
				t.Fatal(err)
			}
			got := specFront(t, spec)
			if !strings.Contains(got, "started: \""+o.firstStarted+"\"") || strings.Contains(got, "finished:") {
				t.Fatalf("after the first land, want started %s and no finished:\n%s", o.firstStarted, got)
			}
			for path, body := range map[string]string{".acta/plans/2026-10-01-a.md": o.secondA, ".acta/plans/2026-10-01-b.md": o.secondB} {
				if err := os.WriteFile(filepath.Join(cfg.RepoRoot, path), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := SpecDates(cfg, "PLN-0002"); err != nil {
				t.Fatal(err)
			}
			got = specFront(t, spec)
			if !strings.Contains(got, "started: \""+early+"\"") || strings.Count(got, "finished:") != 1 {
				t.Errorf("after the second land, want started %s and one finished:\n%s", early, got)
			}
		})
	}
}

// A spec with one plan keeps getting its dates from tick, so SpecDates
// leaves it alone.
func TestSpecDatesLeavesAOnePlanSpecAlone(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".acta/specs/2026-10-01-s.md": "---\nid: SPC-0009\n---\n# S\n",
		".acta/plans/2026-10-01-a.md": splitPlan("PLN-0001", "2026-10-01 10:00:00", "x"),
	})
	wrote, err := SpecDates(cfg, "PLN-0001")
	if err != nil || len(wrote) != 0 {
		t.Fatalf("wrote %v err %v, want nothing", wrote, err)
	}
}
