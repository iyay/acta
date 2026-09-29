package board

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iyay/acta/internal/config"
)

func tree(t *testing.T, files map[string]string) config.Config {
	t.Helper()
	dir := t.TempDir()
	for p, body := range files {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return config.Default(dir)
}

const specA = "# Story A\n"
const planBehind = "# Plan A\n\n**Spec:** .acta/specs/2026-09-20-a.md\n\n### Task 1: One\n- [x] a\n- [ ] b\n"
const planAhead = "# Plan A\n\n**Spec:** .acta/specs/2026-09-20-a.md\n\n### Task 1: One\n- [x] a\n- [x] b\n"

func TestLoadTreesPicksTheWorktreeAhead(t *testing.T) {
	t.Parallel()

	main := tree(t, map[string]string{
		".acta/specs/2026-09-20-a.md":    specA,
		".acta/plans/2026-09-21-a.md":    planBehind,
		".acta/specs/2026-09-22-same.md": "# Same\n",
	})
	wt := tree(t, map[string]string{
		".acta/specs/2026-09-20-a.md":    specA,
		".acta/plans/2026-09-21-a.md":    planAhead,
		".acta/specs/2026-09-22-same.md": "# Same, edited in the worktree\n",
		".acta/specs/2026-09-25-new.md":  "# New story\n",
		".acta/plans/2026-09-25-new.md":  "# New plan\n\n**Spec:** .acta/specs/2026-09-25-new.md\n\n### Task 1: Start\n- [ ] a\n",
	})
	b, err := LoadTrees(main, []Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	task := b.Get("plans/2026-09-21-a#task-1")
	if task.Status != "done" || task.Worktree != "feat" || task.Path != filepath.Join(wt.Root, "plans", "2026-09-21-a.md") {
		t.Fatalf("task = %+v", task)
	}
	if s := b.Get("specs/2026-09-20-a"); s.Status != "done" || s.Worktree != "" {
		t.Fatalf("story A = %+v (status from the merged tasks, file from main)", s)
	}
	if s := b.Get("specs/2026-09-22-same"); s.Title != "Same" || s.Worktree != "" {
		t.Fatalf("equal ticks must keep the main copy: %+v", s)
	}
	if s := b.Get("specs/2026-09-25-new"); s == nil || s.Worktree != "feat" || s.Status != "approved" {
		t.Fatalf("worktree-only story = %+v", s)
	}
	if nt := b.Get("plans/2026-09-25-new#task-1"); nt == nil || nt.Worktree != "feat" || nt.Parent != "specs/2026-09-25-new" {
		t.Fatalf("worktree-only task = %+v", nt)
	}
}

func TestLoadTreesMainAheadWins(t *testing.T) {
	t.Parallel()

	main := tree(t, map[string]string{".acta/plans/2026-09-21-a.md": planAhead})
	wt := tree(t, map[string]string{".acta/plans/2026-09-21-a.md": planBehind})
	b, err := LoadTrees(main, []Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	if task := b.Get("plans/2026-09-21-a#task-1"); task.Worktree != "" || task.Status != "done" {
		t.Fatalf("task = %+v", task)
	}
}

func TestLoadTreesLegacyAndBrokenTrees(t *testing.T) {
	t.Parallel()

	main := tree(t, map[string]string{".acta/specs/2026-09-20-a.md": specA})
	wt := tree(t, map[string]string{
		"docs/superpowers/plans/2026-01-02-old.md": "# Old plan\n\n### Task 1: Old\n- [x] a\n",
	})
	missing := config.Default(filepath.Join(t.TempDir(), "gone"))
	b, err := LoadTrees(main, []Tree{{Cfg: missing, Branch: "gone"}, {Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	if it := b.Get("docs/superpowers/plans/2026-01-02-old#task-1"); it == nil || !it.Legacy || it.Worktree != "feat" {
		t.Fatalf("legacy task from a worktree = %+v", it)
	}
	if b.Get("specs/2026-09-20-a") == nil {
		t.Fatal("main item lost")
	}
}

func TestLoadTreesFromBranchFiles(t *testing.T) {
	t.Parallel()

	main := tree(t, map[string]string{
		".acta/specs/2026-09-20-a.md": specA,
		".acta/plans/2026-09-21-a.md": planBehind,
	})

	branchCfg := config.Default(filepath.Join(t.TempDir(), "not-on-disk"))
	files := map[string][]byte{
		".acta/plans/2026-09-21-a.md":              []byte(planAhead),
		".acta/specs/2026-09-26-x.md":              []byte("# Branch story\n"),
		"docs/superpowers/specs/2026-01-01-old.md": []byte("# Legacy on a branch is skipped\n"),
	}
	b, err := LoadTrees(main, []Tree{{Cfg: branchCfg, Branch: "feat-x", Files: files}})
	if err != nil {
		t.Fatal(err)
	}
	task := b.Get("plans/2026-09-21-a#task-1")
	if task.Status != "done" || task.Worktree != "feat-x" || task.OnDisk || task.Path != "feat-x:.acta/plans/2026-09-21-a.md" {
		t.Fatalf("task from a branch = %+v", task)
	}
	if s := b.Get("specs/2026-09-26-x"); s == nil || s.OnDisk || s.Worktree != "feat-x" {
		t.Fatalf("branch-only story = %+v", s)
	}
	if b.Get("docs/superpowers/specs/2026-01-01-old") != nil {
		t.Fatal("legacy folders must not be read from branches")
	}
	m := b.Get("plans/2026-09-21-a")
	if m == nil || m.Kind != KindPlan {
		t.Fatalf("plan from a branch = %+v", m)
	}
	if m.SpecID != "specs/2026-09-20-a" {
		t.Fatalf("plan spec = %q, want the spec it names", m.SpecID)
	}
	if m.Worktree != "feat-x" || m.OnDisk {
		t.Fatalf("plan from a branch is not marked as one: %+v", m)
	}
	for _, it := range b.Items {
		if it.Worktree == "" && !it.OnDisk {
			t.Fatalf("main item %s must be on disk", it.ID)
		}
	}
}

func TestLoadIsLoadTreesWithNoOthers(t *testing.T) {
	t.Parallel()

	main := tree(t, map[string]string{".acta/specs/2026-09-20-a.md": specA})
	a, _ := Load(main)
	b, _ := LoadTrees(main, nil)
	if len(a.Items) != 1 || len(b.Items) != 1 || a.Items[0].ID != b.Items[0].ID {
		t.Fatal("Load and LoadTrees(cfg, nil) differ")
	}
}
