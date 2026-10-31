package write

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const debtNotes = "---\nparent: plans/2026-09-25-fix\n---\n# Notes\n\n" +
	"- [ ] (medium) retry loop has no cap\n" +
	"- [x] old line closed by hand (stale)\n" +
	"- [ ] (high) lock is never freed\n"

func debtRepo(t *testing.T) (cfgPath string, read func() string, closeLine func(id, status, reason string) error) {
	t.Helper()
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/plans/2026-09-25-fix.md":  "# Fix\n\n### Task 1: Fix\n- [x] a\n",
		".acta/debt/2026-09-26-notes.md": debtNotes,
	})
	path := filepath.Join(cfg.Root, "debt", "2026-09-26-notes.md")
	read = func() string {
		t.Helper()
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}
	closeLine = func(id, status, reason string) error {
		t.Helper()
		_, err := CloseDebtLine(cfg, mustLoad(t, cfg), id, status, reason)
		return err
	}
	return cfg.RepoRoot, read, closeLine
}

func TestCloseDebtLineWritesBoxAndReason(t *testing.T) {
	repo, read, closeLine := debtRepo(t)
	if err := closeLine("debt/2026-09-26-notes#item-1", "done", "done in abc1234"); err != nil {
		t.Fatal(err)
	}
	if err := closeLine("debt/2026-09-26-notes#item-3", "wontfix", "stale"); err != nil {
		t.Fatal(err)
	}
	want := "---\nparent: plans/2026-09-25-fix\n---\n# Notes\n\n" +
		"- [x] (medium) retry loop has no cap (done: done in abc1234)\n" +
		"- [x] old line closed by hand (stale)\n" +
		"- [-] (high) lock is never freed (wontfix: stale)\n"
	if got := read(); got != want {
		t.Fatalf("file:\n got %q\nwant %q", got, want)
	}
	if out := gitRun(t, repo, "status", "--porcelain"); out != "" {
		t.Fatalf("not committed: %s", out)
	}
}

// A second close of the same line swaps the reason this command wrote, so the
// board never shows two reasons on one line.
func TestCloseDebtLineReplacesItsOwnReason(t *testing.T) {
	_, read, closeLine := debtRepo(t)
	if err := closeLine("debt/2026-09-26-notes#item-1", "done", "done in abc1234"); err != nil {
		t.Fatal(err)
	}
	if err := closeLine("debt/2026-09-26-notes#item-1", "wontfix", "code is gone"); err != nil {
		t.Fatal(err)
	}
	if err := closeLine("debt/2026-09-26-notes#item-2", "wontfix", "not valid"); err != nil {
		t.Fatal(err)
	}
	want := "---\nparent: plans/2026-09-25-fix\n---\n# Notes\n\n" +
		"- [-] (medium) retry loop has no cap (wontfix: code is gone)\n" +
		"- [-] old line closed by hand (stale) (wontfix: not valid)\n" +
		"- [ ] (high) lock is never freed\n"
	if got := read(); got != want {
		t.Fatalf("file:\n got %q\nwant %q", got, want)
	}
}

func TestCloseDebtLineBadInput(t *testing.T) {
	repo, read, closeLine := debtRepo(t)
	cases := []struct{ name, id, status, reason string }{
		{"no reason", "debt/2026-09-26-notes#item-1", "done", ""},
		{"blank reason", "debt/2026-09-26-notes#item-1", "done", "  "},
		{"parentheses only", "debt/2026-09-26-notes#item-1", "done", "()"},
		{"newline", "debt/2026-09-26-notes#item-1", "done", "a\nb"},
		{"open is not a close", "debt/2026-09-26-notes#item-1", "open", "x"},
		{"other status", "debt/2026-09-26-notes#item-1", "fixed", "x"},
		{"unknown line", "debt/2026-09-26-notes#item-9", "done", "x"},
		{"not a debt line", "plans/2026-09-25-fix", "done", "x"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := closeLine(c.id, c.status, c.reason); !errors.Is(err, ErrBadInput) {
				t.Fatalf("err = %v, want ErrBadInput", err)
			}
		})
	}
	if got := read(); got != debtNotes {
		t.Fatalf("bad input changed the file: %q", got)
	}
	if out := gitRun(t, repo, "status", "--porcelain"); out != "" {
		t.Fatalf("bad input left changes: %s", out)
	}
}
