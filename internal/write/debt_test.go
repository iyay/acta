package write

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A plan file with an id and hash already set, so NewDebt can resolve it by
// every id form the command line accepts.
const debtPlan = "---\nid: PLAN-3\nhash: k3f2\n---\n# Short IDs\n"

// fixNowAt pins Now() to one date for a test, so a written file name is
// predictable. Local to this file: the plan's debt tests want a different
// date than the bug tests in ops_test.go already fix.
func fixNowAt(t *testing.T, date string) {
	t.Helper()
	when, err := time.Parse("2006-01-02", date)
	if err != nil {
		t.Fatal(err)
	}
	old := Now
	Now = func() time.Time { return when }
	t.Cleanup(func() { Now = old })
}

func TestNewDebtWritesChecklist(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	out, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "", []byte("- first note\nsecond note\n\n"))
	if err != nil || !out.Committed {
		t.Fatalf("out=%+v err=%v", out, err)
	}
	got := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md"))
	for _, want := range []string{"id: DEBT-1\n", "parent: plans/2026-09-26-short-ids\n", "# Review NOTEs: Short IDs\n", "- [ ] first note\n", "- [ ] second note\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if n := strings.Count(gitRun(t, cfg.RepoRoot, "log", "--oneline"), "acta: new debt 2026-09-27-short-ids"); n != 1 {
		t.Fatalf("commits = %d, want 1", n)
	}
}

func TestNewDebtAppendsWithoutDuplicates(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	if _, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "", []byte("a\nb\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := NewDebt(cfg, mustLoad(t, cfg), "plans/2026-09-26-short-ids", "", []byte("b\nc\n")); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md"))
	if strings.Count(got, "- [ ] b\n") != 1 || !strings.Contains(got, "- [ ] c\n") {
		t.Fatalf("got\n%s", got)
	}
	// only the create commit and one append commit; the all-duplicate
	// pass in the next test must not add a third.
	if n := strings.Count(gitRun(t, cfg.RepoRoot, "log", "--oneline"), "acta: new debt 2026-09-27-short-ids"); n != 2 {
		t.Fatalf("commits = %d, want 2", n)
	}
}

func TestNewDebtAllDuplicatesIsNoOp(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	if _, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "", []byte("a\nb\n")); err != nil {
		t.Fatal(err)
	}
	before := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md"))
	beforeLog := gitRun(t, cfg.RepoRoot, "log", "--oneline")
	out, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "", []byte("a\nb\n"))
	if err != nil || out.Committed {
		t.Fatalf("out=%+v err=%v, want no commit", out, err)
	}
	after := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md"))
	if before != after {
		t.Fatalf("file changed on a no-op run:\nbefore=%s\nafter=%s", before, after)
	}
	if afterLog := gitRun(t, cfg.RepoRoot, "log", "--oneline"); afterLog != beforeLog {
		t.Fatalf("a no-op run must not add a commit:\nbefore=%s\nafter=%s", beforeLog, afterLog)
	}
}

func TestNewDebtRejectsBadInput(t *testing.T) {
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	for name, c := range map[string]struct{ id, notes string }{
		"empty":      {"PLAN-3", ""},
		"blank":      {"PLAN-3", "  \n\n"},
		"no plan":    {"PLAN-99", "x\n"},
		"not a plan": {"BUG-1", "x\n"},
	} {
		if _, err := NewDebt(cfg, mustLoad(t, cfg), c.id, "", []byte(c.notes)); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	if _, err := os.Stat(filepath.Join(cfg.Root, "debt")); !os.IsNotExist(err) {
		t.Fatal("debt folder must not exist after rejected input")
	}
}

func TestNewDebtTitleOverridesHeading(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	if _, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "Custom Title", []byte("x\n")); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md"))
	if !strings.Contains(got, "# Custom Title\n") {
		t.Fatalf("title not applied:\n%s", got)
	}
}

func TestNewDebtHashIDResolves(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	if _, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-k3f2", "", []byte("x\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md")); err != nil {
		t.Fatal(err)
	}
}

func TestNewDebtAutoCommitOff(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	cfg.AutoCommit = false
	out, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "", []byte("x\n"))
	if err != nil || out.Committed || out.Skipped {
		t.Fatalf("outcome %+v err %v", out, err)
	}
	if _, err := os.Stat(filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md")); err != nil {
		t.Fatal("file must still be written with auto_commit off")
	}
}
