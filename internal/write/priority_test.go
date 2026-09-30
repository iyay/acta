package write

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const prioDebt = "---\nid: DBT-0001\nhash: aaaaaaa\n---\n# Review NOTEs: D\n\n- [ ] first\n- [x] (low) second\n"

func prioRepo(t *testing.T) (cfgRoot string, set func(id, value string) error) {
	t.Helper()
	cfg := repoWith(t, map[string]string{
		".acta/bugs/2026-09-24-crash.md":   "---\nref: B-1\n---\n# Crash\n\n## Symptom\nIt crashes.\n",
		".acta/debt/2026-10-01-d.md":       prioDebt,
		".acta/specs/2026-10-01-s.md":      "---\nid: SPC-0001\n---\n# S\n",
		".acta/plans/2026-10-01-p.md":      "---\nid: PLN-0001\n---\n# P\n\n### Task 1: T\n- [ ] a\n",
		".acta/scratch/2026-10-01-idea.md": "---\nid: SCR-0001\n---\n# Idea\n",
	})
	return cfg.Root, func(id, value string) error {
		_, err := SetValue(cfg, mustLoad(t, cfg), id, "priority", value)
		return err
	}
}

func TestSetPriorityOnBug(t *testing.T) {
	root, set := prioRepo(t)
	path := filepath.Join(root, "bugs/2026-09-24-crash.md")
	if err := set("bugs/2026-09-24-crash", "high"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); !strings.Contains(got, "priority: high\n") {
		t.Fatalf("bug file = %q", got)
	}
	if err := set("bugs/2026-09-24-crash", "none"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); strings.Contains(got, "priority") {
		t.Fatalf("none left the field: %q", got)
	}
}

func TestSetPriorityOnDebtItem(t *testing.T) {
	root, set := prioRepo(t)
	path := filepath.Join(root, "debt/2026-10-01-d.md")
	if err := set("debt/2026-10-01-d#item-1", "medium"); err != nil {
		t.Fatal(err)
	}
	if err := set("debt/2026-10-01-d#item-2", "high"); err != nil {
		t.Fatal(err)
	}
	want := "- [ ] (medium) first\n- [x] (high) second\n"
	if got := readFile(t, path); !strings.HasSuffix(got, want) {
		t.Fatalf("debt file = %q, want it to end with %q", got, want)
	}
	if err := set("debt/2026-10-01-d#item-2", "none"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); !strings.HasSuffix(got, "- [x] second\n") {
		t.Fatalf("none left the tag: %q", got)
	}
}

func TestSetPriorityRefuses(t *testing.T) {
	root, set := prioRepo(t)
	before := map[string]string{}
	for _, p := range []string{"bugs/2026-09-24-crash.md", "debt/2026-10-01-d.md"} {
		before[p] = readFile(t, filepath.Join(root, p))
	}
	for _, c := range []struct{ id, value, want string }{
		{"specs/2026-10-01-s", "high", "priority is only for bugs and debt items"},
		{"plans/2026-10-01-p", "high", "priority is only for bugs and debt items"},
		{"plans/2026-10-01-p#task-1", "high", "priority is only for bugs and debt items"},
		{"scratch/2026-10-01-idea", "high", "priority is only for bugs and debt items"},
		{"bugs/2026-09-24-crash", "urgent", "priority must be high, medium, low or none"},
		{"debt/2026-10-01-d#item-1", "hgh", "priority must be high, medium, low or none"},
	} {
		err := set(c.id, c.value)
		if !errors.Is(err, ErrBadInput) || !strings.Contains(err.Error(), c.want) {
			t.Errorf("set %s priority %s: err = %v, want %q", c.id, c.value, err, c.want)
		}
	}
	for p, was := range before {
		if got := readFile(t, filepath.Join(root, p)); got != was {
			t.Errorf("%s changed after a refused set", p)
		}
	}
}

func TestNewBugPriority(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	body := []byte("# Hang\n\n## Symptom\nIt hangs.\n")
	if _, err := NewBug(cfg, "hang", "", "", "low", body); err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob(filepath.Join(cfg.Root, "bugs", "*-hang.md"))
	if len(matches) != 1 || !strings.Contains(readFile(t, matches[0]), "priority: low\n") {
		t.Fatalf("bug files %v, want one with priority: low", matches)
	}
	_, err := NewBug(cfg, "stall", "", "", "urgent", []byte("# Stall\n\n## Symptom\nIt stalls.\n"))
	if !errors.Is(err, ErrBadInput) || !strings.Contains(err.Error(), "priority must be high, medium or low") {
		t.Fatalf("bad priority err = %v", err)
	}
	if m, _ := filepath.Glob(filepath.Join(cfg.Root, "bugs", "*-stall.md")); len(m) != 0 {
		t.Fatalf("a refused bug left %v", m)
	}
}

func TestStartBugPriority(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	path, tmpl, err := StartBug(cfg, "slow", "", "high")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tmpl), "priority: high\n") {
		t.Fatalf("template = %q", tmpl)
	}
	if _, _, err := StartBug(cfg, "slower", "", "urgent"); !errors.Is(err, ErrBadInput) {
		t.Fatalf("bad priority err = %v", err)
	}
	_ = os.Remove(path)
}

func TestNewDebtKeepsTheTag(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	if _, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "", []byte("(high) big one\n(hgh) typo one\nplain one\n")); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md"))
	for _, want := range []string{"- [ ] (high) big one\n", "- [ ] (hgh) typo one\n", "- [ ] plain one\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	b := mustLoad(t, cfg)
	if it := b.Get("debt/2026-09-27-short-ids#item-1"); it == nil || it.Priority != "high" || it.Title != "big one" {
		t.Errorf("item-1 = %+v", it)
	}
}
