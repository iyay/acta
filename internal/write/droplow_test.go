package write

import (
	"path/filepath"
	"strings"
	"testing"
)

// drop-low closes every open low or untagged debt line in one commit and
// leaves medium, high and already closed lines exactly as they were.
func TestDropLowDebtClosesOnlyLowAndUntagged(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-09-26-short-ids.md": debtPlan,
		".acta/debt/2026-09-27-a.md":          "---\nid: DBT-0001\n---\n# A\n\n- [ ] (low) one\n- [ ] (medium) two\n- [ ] plain three\n- [x] (low) done four\n",
		".acta/debt/2026-09-27-b.md":          "---\nid: DBT-0002\n---\n# B\n\n- [ ] (high) five\n- [-] six\n- [ ] (hgh) typo seven\n",
		".acta/debt/2026-09-27-c.md":          "---\nid: DBT-0003\n---\n# C\n\n- [ ] (medium) only\n",
	})
	beforeC := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-c.md"))
	beforeCount := strings.Count(gitRun(t, cfg.RepoRoot, "log", "--oneline"), "\n")

	out, err := DropLowDebt(cfg)
	if err != nil || !out.Committed {
		t.Fatalf("out=%+v err=%v", out, err)
	}
	a := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-a.md"))
	b := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-b.md"))
	for _, want := range []string{"- [-] (low) one\n", "- [ ] (medium) two\n", "- [-] plain three\n", "- [x] (low) done four\n"} {
		if !strings.Contains(a, want) {
			t.Errorf("a: missing %q in\n%s", want, a)
		}
	}
	for _, want := range []string{"- [ ] (high) five\n", "- [-] six\n", "- [-] (hgh) typo seven\n"} {
		if !strings.Contains(b, want) {
			t.Errorf("b: missing %q in\n%s", want, b)
		}
	}
	if got := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-c.md")); got != beforeC {
		t.Errorf("c changed:\n%s", got)
	}
	log := gitRun(t, cfg.RepoRoot, "log", "--oneline")
	if n := strings.Count(log, "\n") - beforeCount; n != 1 {
		t.Fatalf("commits added = %d, want 1:\n%s", n, log)
	}
	if !strings.Contains(gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%B"), "low, under new debt bar") {
		t.Fatal("commit message lacks the reason")
	}

	// A second run finds nothing and makes no commit.
	out, err = DropLowDebt(cfg)
	if err != nil || out.Committed {
		t.Fatalf("second run: out=%+v err=%v", out, err)
	}
	if gitRun(t, cfg.RepoRoot, "log", "--oneline") != log {
		t.Fatal("second run added a commit")
	}
}
