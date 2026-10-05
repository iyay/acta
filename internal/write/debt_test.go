package write

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
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
	for _, want := range []string{"id: DBT-0001\n", "parent: plans/2026-09-26-short-ids\n", "# Review NOTEs: Short IDs\n", "- [ ] first note\n", "- [ ] second note\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if n := strings.Count(gitRun(t, cfg.RepoRoot, "log", "--oneline"), "chore(debt): new debt 2026-09-27-short-ids"); n != 1 {
		t.Fatalf("commits = %d, want 1", n)
	}
}

func TestNewDebtWritesNewFormatIDAndHash(t *testing.T) {
	fixNowAt(t, "2026-09-27")
	cfg := repoWith(t, map[string]string{".acta/plans/2026-09-26-short-ids.md": debtPlan})
	if _, err := NewDebt(cfg, mustLoad(t, cfg), "PLAN-3", "", []byte("- one\n")); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(cfg.Root, "debt", "2026-09-27-short-ids.md")
	if got := frontField(t, file, "id"); got != "DBT-0001" {
		t.Errorf("debt id = %q, want DBT-0001", got)
	}
	if h := frontField(t, file, "hash"); !board.IsHash(h) {
		t.Errorf("debt hash = %q, want 7 characters", h)
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
	if n := strings.Count(gitRun(t, cfg.RepoRoot, "log", "--oneline"), "chore(debt): new debt 2026-09-27-short-ids"); n != 2 {
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
	t.Parallel()

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

// A tick and a NOTE landing on one debt file must both stay in the file,
// whatever order they run in, and the file must never be half written.
func TestAppendDebtAndTickLineKeepBoth(t *testing.T) {
	useLockBase(t)
	root := t.TempDir()
	// No git and no commit here: this test is about the file alone, and
	// fifty rounds of commits would only make it slow.
	cfg := config.Default(root)
	cfg.AutoCommit = false
	path := filepath.Join(root, "debt", "2026-09-29-locks.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// The box sits on the sixth line: two lines of front matter, the
	// heading, then a blank line.
	const boxLine = 6
	const first = "---\nid: DEBT-1\n---\n# Review NOTEs: Locks\n\n- [ ] first note\n"

	// start puts the file back to one open box and gives back the copy a
	// caller would already be holding.
	start := func(t *testing.T) []byte {
		t.Helper()
		if err := os.WriteFile(path, []byte(first), 0o644); err != nil {
			t.Fatal(err)
		}
		return []byte(readFile(t, path))
	}
	// noTempLeft proves no half written file is left behind for the next
	// run to pick up as junk.
	noTempLeft := func(t *testing.T) {
		t.Helper()
		if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
			t.Fatalf("a temp file was left behind at %s.tmp", path)
		}
	}
	wantBoth := func(t *testing.T, what string) {
		t.Helper()
		got := readFile(t, path)
		if !strings.Contains(got, "- [x] first note\n") || !strings.Contains(got, "- [ ] new note\n") {
			t.Fatalf("%s:\n%s", what, got)
		}
		noTempLeft(t)
	}

	t.Run("append then tick keeps both", func(t *testing.T) {
		src := start(t)
		if _, err := appendDebt(cfg, path, "2026-09-29-locks", src, []string{"new note"}); err != nil {
			t.Fatal(err)
		}
		if err := TickLine(path, boxLine, 'x'); err != nil {
			t.Fatal(err)
		}
		wantBoth(t, "the NOTE was lost")
	})

	t.Run("tick then append keeps both", func(t *testing.T) {
		// This is the copy NewDebt already read, before the tick landed.
		src := start(t)
		if err := TickLine(path, boxLine, 'x'); err != nil {
			t.Fatal(err)
		}
		if _, err := appendDebt(cfg, path, "2026-09-29-locks", src, []string{"new note"}); err != nil {
			t.Fatal(err)
		}
		wantBoth(t, "the tick was lost")
	})

	t.Run("nothing new leaves the file alone", func(t *testing.T) {
		src := start(t)
		out, err := appendDebt(cfg, path, "2026-09-29-locks", src, []string{"first note"})
		if err != nil || out.Committed {
			t.Fatalf("out=%+v err=%v, want no write and no commit", out, err)
		}
		if got := readFile(t, path); got != first {
			t.Fatalf("the file changed on a run with nothing new:\n%s", got)
		}
		// A NOTE that is already ticked is still on the file, so it must
		// not come back as a second open box.
		if err := TickLine(path, boxLine, 'x'); err != nil {
			t.Fatal(err)
		}
		ticked := readFile(t, path)
		if _, err := appendDebt(cfg, path, "2026-09-29-locks", []byte(ticked), []string{"first note"}); err != nil {
			t.Fatal(err)
		}
		if got := readFile(t, path); got != ticked {
			t.Fatalf("a ticked NOTE was added again:\n%s", got)
		}
		noTempLeft(t)
	})

	for round := range 50 {
		src := start(t)
		var wg sync.WaitGroup
		gate := make(chan struct{})
		failed := make(chan error, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-gate
			failed <- TickLine(path, boxLine, 'x')
		}()
		go func() {
			defer wg.Done()
			<-gate
			_, err := appendDebt(cfg, path, "2026-09-29-locks", src, []string{"new note"})
			failed <- err
		}()
		close(gate)
		wg.Wait()
		close(failed)
		for err := range failed {
			if err != nil {
				t.Fatalf("round %d: %v", round, err)
			}
		}
		wantBoth(t, fmt.Sprintf("round %d lost a change", round))
	}
}
