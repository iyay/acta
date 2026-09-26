package write

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"pm-board/internal/board"
)

func TestAssignIDsGivesMissingOnly(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-20-a-design.md": "---\nid: SPEC-4\nhash: m2x9\n---\n# A\n",
		".pm/specs/2026-09-21-b-design.md": "# B\n",
		".pm/plans/2026-09-22-p.md":        "# P\n\n### Task 1: One\n- [ ] x\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	before, _ := os.ReadFile(filepath.Join(cfg.Root, "specs/2026-09-20-a-design.md"))
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed || len(changes) != 3 {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	b := mustLoad(t, cfg)
	if b.Get("SPEC-5") == nil || b.Get("PLAN-1") == nil || b.Get("BUG-1") == nil {
		t.Fatal("new numbers do not resolve")
	}
	after, _ := os.ReadFile(filepath.Join(cfg.Root, "specs/2026-09-20-a-design.md"))
	if string(before) != string(after) {
		t.Fatal("an existing id was rewritten")
	}
	for _, it := range b.Items {
		if it.Kind != "task" && !board.IsHash(strings.TrimPrefix(it.Hash, board.Prefix(it.Kind, strings.Contains(it.Path, "/plans/"))+"-")) {
			t.Errorf("%s hash %q", it.ID, it.Hash)
		}
	}
	if c, _, _ := AssignIDs(cfg, b, nil); len(c) != 0 {
		t.Fatalf("second run changed %v", c)
	}
}

func TestAssignIDsOneCommit(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-21-b-design.md": "# B\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	before, _ := strconv.Atoi(gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"))
	changes, out, err := AssignIDs(cfg, mustLoad(t, cfg), nil)
	if err != nil || !out.Committed || len(changes) != 2 {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	after, _ := strconv.Atoi(gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"))
	if after-before != 1 {
		t.Fatalf("commits went from %d to %d, want +1", before, after)
	}
}

func TestAssignIDsCountsOtherTrees(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/plans/2026-09-22-p.md": "# P\n\n### Task 1: One\n- [ ] x\n",
	})
	b, err := board.LoadTrees(cfg, []board.Tree{{Cfg: cfg, Branch: "x", Files: map[string][]byte{
		".pm/plans/2026-09-25-q.md": []byte("---\nid: PLAN-9\nhash: aaaa\n---\n# Q\n"),
	}}})
	changes, _, err := AssignIDs(cfg, b, nil)
	if err != nil || len(changes) != 1 {
		t.Fatalf("changes=%v err=%v", changes, err)
	}
	if b2 := mustLoad(t, cfg); b2.Get("PLAN-10") == nil {
		t.Fatal("new plan is not PLAN-10")
	}
}

func TestAssignIDsSkipsLegacy(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	legacy := filepath.Join(cfg.RepoRoot, "docs/superpowers/specs/2026-01-01-old.md")
	before, _ := os.ReadFile(legacy)
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(legacy)
	if string(before) != string(after) || strings.Contains(string(after), "id:") {
		t.Fatalf("legacy rewritten: %q", after)
	}
}

func TestAssignIDsOnlyNamed(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-21-b-design.md": "# B\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	changes, _, err := AssignIDs(cfg, mustLoad(t, cfg), []string{"bugs/2026-09-23-c"})
	if err != nil || len(changes) != 1 {
		t.Fatalf("changes=%v err=%v", changes, err)
	}
	spec, _ := os.ReadFile(filepath.Join(cfg.Root, "specs/2026-09-21-b-design.md"))
	if strings.Contains(string(spec), "id:") {
		t.Fatalf("spec touched: %q", spec)
	}
	if b := mustLoad(t, cfg); b.Get("BUG-1") == nil {
		t.Fatal("bug did not get BUG-1")
	}
}

func TestAssignIDsHashClash(t *testing.T) {
	old := randHash
	calls := 0
	randHash = func() string {
		calls++
		if calls == 1 {
			return "m2x9"
		}
		return "zz00"
	}
	defer func() { randHash = old }()
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-20-a-design.md": "---\nid: SPEC-4\nhash: m2x9\n---\n# A\n",
		".pm/bugs/2026-09-23-c.md":         "# C\n",
	})
	if _, _, err := AssignIDs(cfg, mustLoad(t, cfg), nil); err != nil {
		t.Fatal(err)
	}
	if got := mustLoad(t, cfg).Get("BUG-1"); got == nil || got.Hash != "BUG-zz00" {
		t.Fatalf("bug hash = %+v, want BUG-zz00", got)
	}
}

func TestFixDuplicates(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/bugs/2026-09-20-first.md": "---\nid: BUG-7\nhash: aaaa\n---\n# First\n",
	})
	time.Sleep(1100 * time.Millisecond)
	full := filepath.Join(cfg.Root, "bugs/2026-09-21-second.md")
	if err := os.WriteFile(full, []byte("---\nid: BUG-7\nhash: bbbb\n---\n# Second\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, cfg.RepoRoot, "add", ".")
	gitRun(t, cfg.RepoRoot, "commit", "-q", "-m", "second")
	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || !out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if len(changes) != 1 || changes[0] != "BUG-7 -> BUG-8 (bugs/2026-09-21-second.md)" {
		t.Fatalf("changes=%v", changes)
	}
	b := mustLoad(t, cfg)
	if b.Get("BUG-7") == nil || b.Get("BUG-8") == nil {
		t.Fatal("renumber did not resolve")
	}
	if b.Get("BUG-7").Hash != "BUG-aaaa" || b.Get("BUG-8").Hash != "BUG-bbbb" {
		t.Fatalf("hashes moved: %q %q", b.Get("BUG-7").Hash, b.Get("BUG-8").Hash)
	}
}

func TestFixDuplicatesNoOp(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	before := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD")
	changes, out, err := FixDuplicates(cfg, mustLoad(t, cfg))
	if err != nil || len(changes) != 0 || out.Committed {
		t.Fatalf("changes=%v out=%+v err=%v", changes, out, err)
	}
	if after := gitRun(t, cfg.RepoRoot, "rev-list", "--count", "HEAD"); after != before {
		t.Fatal("commit made with nothing to fix")
	}
}
