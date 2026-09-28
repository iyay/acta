package write

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// unchanged fails when a refused call left a file written or a commit made.
func unchanged(t *testing.T, cfg config.Config, before string) {
	t.Helper()
	if s := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); s != "" {
		t.Errorf("tree changed: %s", s)
	}
	if after := gitRun(t, cfg.RepoRoot, "log", "--format=%H"); after != before {
		t.Error("a commit was made")
	}
}

// refuse runs fn and checks the error text and that nothing was written.
func refuse(t *testing.T, cfg config.Config, want string, fn func() error) {
	t.Helper()
	before := gitRun(t, cfg.RepoRoot, "log", "--format=%H")
	err := fn()
	if err == nil {
		t.Fatalf("want error %q, got none", want)
	}
	if !errors.Is(err, ErrBadInput) {
		t.Errorf("err = %v, want ErrBadInput", err)
	}
	if err.Error() != "bad input: "+want {
		t.Errorf("err = %q, want %q", err, "bad input: "+want)
	}
	unchanged(t, cfg, before)
}

func TestNewScratchWritesRawItemAndCommits(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	before := gitRun(t, cfg.RepoRoot, "log", "--format=%H")
	body := []byte("catet aja dulu: sort newest first ✓\n")
	o, err := NewScratch(cfg, "newest-first", "", body)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(cfg.Root, "scratch", "2026-09-26-newest-first.md")
	if o.Path != want {
		t.Fatalf("path %s want %s", o.Path, want)
	}
	if !o.Committed || o.Skipped {
		t.Fatalf("outcome %+v", o)
	}
	if o.ShortID != "SCRATCH-1" {
		t.Errorf("short id %q want SCRATCH-1", o.ShortID)
	}
	got, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"id: SCRATCH-1", "status: raw", "title: newest-first", `created: "2026-09-26"`} {
		if !strings.Contains(string(got), field) {
			t.Errorf("file lacks %q:\n%s", field, got)
		}
	}
	// The body comes back byte for byte, the accented text included.
	if !strings.HasSuffix(string(got), string(body)) {
		t.Errorf("body changed:\ngot  %q\nwant %q", got, body)
	}
	if msg := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); msg != "acta: new scratch 2026-09-26-newest-first" {
		t.Errorf("commit message %q", msg)
	}
	if n := strings.Count(gitRun(t, cfg.RepoRoot, "log", "--format=%H"), "\n") + 1; n != strings.Count(before, "\n")+2 {
		t.Errorf("commits went from %d to %d", strings.Count(before, "\n")+1, n)
	}
	it := mustLoad(t, cfg).Get("SCRATCH-1")
	if it == nil || it.Kind != board.KindScratch || it.Status != "raw" {
		t.Fatalf("board reads %+v, want a raw scratch item", it)
	}
	if s := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); s != "" {
		t.Errorf("tree dirty after a commit: %s", s)
	}
}

func TestNewScratchKeepsABodyThatStartsWithARule(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	body := []byte("---\nstill my body\n")
	o, err := NewScratch(cfg, "rule-body", "", body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(o.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(got), "created: \"2026-09-26\"\n---\n"+string(body)) {
		t.Errorf("file = %q", got)
	}
	if s := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); s != "" {
		t.Errorf("tree dirty after a commit: %s", s)
	}
}

func TestNewScratchTitleFlagAndUnicode(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	body := []byte("baru: fichier — naïve ✓ 日本語\n")
	o, err := NewScratch(cfg, "catatan", "Catatan Baru", body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(o.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "title: Catatan Baru") {
		t.Errorf("file lacks the title:\n%s", got)
	}
	if !strings.HasSuffix(string(got), string(body)) {
		t.Errorf("body bytes changed:\ngot  %q\nwant %q", got, body)
	}
}

func TestNewScratchRefusesAndWritesNothing(t *testing.T) {
	t.Run("empty body", func(t *testing.T) {
		fixNow(t)
		cfg := repoWith(t, baseFiles)
		refuse(t, cfg, "scratch body is empty", func() error {
			_, err := NewScratch(cfg, "idea", "", nil)
			return err
		})
		if _, err := os.Stat(filepath.Join(cfg.Root, "scratch")); !os.IsNotExist(err) {
			t.Error("the scratch folder must not exist after an empty body")
		}
	})
	t.Run("blank body", func(t *testing.T) {
		fixNow(t)
		cfg := repoWith(t, baseFiles)
		refuse(t, cfg, "scratch body is empty", func() error {
			_, err := NewScratch(cfg, "idea", "", []byte(" \n\t\n"))
			return err
		})
	})
	t.Run("bad slug", func(t *testing.T) {
		fixNow(t)
		cfg := repoWith(t, baseFiles)
		refuse(t, cfg, `slug "Bad Slug!" must be lower case words joined by -`, func() error {
			_, err := NewScratch(cfg, "Bad Slug!", "", []byte("x\n"))
			return err
		})
	})
	t.Run("empty slug", func(t *testing.T) {
		fixNow(t)
		cfg := repoWith(t, baseFiles)
		refuse(t, cfg, `slug "" must be lower case words joined by -`, func() error {
			_, err := NewScratch(cfg, "", "", []byte("x\n"))
			return err
		})
	})
	t.Run("existing file", func(t *testing.T) {
		fixNow(t)
		cfg := repoWith(t, baseFiles)
		path := filepath.Join(cfg.Root, "scratch", "2026-09-26-idea.md")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("mine\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitRun(t, cfg.RepoRoot, "add", ".")
		gitRun(t, cfg.RepoRoot, "commit", "-q", "-m", "add the file by hand")
		refuse(t, cfg, path+" already exists", func() error {
			_, err := NewScratch(cfg, "idea", "", []byte("x\n"))
			return err
		})
		if got, _ := os.ReadFile(path); string(got) != "mine\n" {
			t.Errorf("the file on disk changed: %q", got)
		}
	})
	t.Run("same slug twice on one day", func(t *testing.T) {
		fixNow(t)
		cfg := repoWith(t, baseFiles)
		first, err := NewScratch(cfg, "twice", "", []byte("first\n"))
		if err != nil {
			t.Fatal(err)
		}
		kept, _ := os.ReadFile(first.Path)
		refuse(t, cfg, first.Path+" already exists", func() error {
			_, err := NewScratch(cfg, "twice", "", []byte("second\n"))
			return err
		})
		if got, _ := os.ReadFile(first.Path); string(got) != string(kept) {
			t.Errorf("the first file changed: %q", got)
		}
	})
}

func TestAppendScratchAddsTextAfterOneBlankLine(t *testing.T) {
	fixNow(t)
	files := map[string]string{
		".acta/specs/2026-09-25-from-idea.md": "---\nparent: scratch/2026-09-26-idea\n---\n# From idea\n",
	}
	cfg := repoWith(t, files)
	o, err := NewScratch(cfg, "idea", "", []byte("first note ✓\n"))
	if err != nil {
		t.Fatal(err)
	}
	// A spec names the item as its parent, so the item is specced. Text can
	// still be added to it.
	it := mustLoad(t, cfg).Get("SCRATCH-1")
	if it == nil || it.Status != "specced" {
		t.Fatalf("item = %+v, want specced", it)
	}
	before := gitRun(t, cfg.RepoRoot, "log", "--format=%H")
	got, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", []byte("answer 1 ✓\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Committed || got.Skipped || got.ShortID != "SCRATCH-1" {
		t.Fatalf("outcome %+v", got)
	}
	body, _ := os.ReadFile(o.Path)
	if want := "first note ✓\n\nanswer 1 ✓\n"; !strings.HasSuffix(string(body), want) {
		t.Errorf("body = %q, want it to end with %q", body, want)
	}
	if msg := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); msg != "acta: add to scratch 2026-09-26-idea" {
		t.Errorf("commit message %q", msg)
	}
	if n := strings.Count(gitRun(t, cfg.RepoRoot, "log", "--format=%H"), "\n") + 1; n != strings.Count(before, "\n")+2 {
		t.Errorf("append made more than one commit")
	}
	if s := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); s != "" {
		t.Errorf("tree dirty after a commit: %s", s)
	}

	// A dropped item takes text too.
	if _, err := SetValue(cfg, mustLoad(t, cfg), "SCRATCH-1", "status", "dropped"); err != nil {
		t.Fatal(err)
	}
	if it := mustLoad(t, cfg).Get("SCRATCH-1"); it.Status != "dropped" {
		t.Fatalf("status %q want dropped", it.Status)
	}
	if _, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", []byte("answer 2\n")); err != nil {
		t.Fatal(err)
	}
	body, _ = os.ReadFile(o.Path)
	if want := "answer 1 ✓\n\nanswer 2\n"; !strings.HasSuffix(string(body), want) {
		t.Errorf("body = %q, want it to end with %q", body, want)
	}
}

func TestAppendScratchAddsTheNewlineTheBodyLacks(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	o, err := NewScratch(cfg, "no-newline", "", []byte("no newline at the end"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", []byte("next line\n")); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(o.Path)
	if want := "no newline at the end\n\nnext line\n"; !strings.HasSuffix(string(body), want) {
		t.Errorf("body = %q, want it to end with %q", body, want)
	}
}

func TestAppendScratchRefusesAndChangesNothing(t *testing.T) {
	fixNow(t)
	files := map[string]string{
		".acta/bugs/2026-09-24-crash.md": "---\nid: BUG-3\nhash: ab12\n---\n# Crash\n\n## Symptom\nx\n",
		".acta/plans/2026-09-25-fix.md":  "---\nid: PLAN-4\nhash: cd34\n---\n# Fix\n\n### Task 1: Fix\n- [ ] a\n",
	}
	cfg := repoWith(t, files)
	o, err := NewScratch(cfg, "idea", "", []byte("first note ✓\n"))
	if err != nil {
		t.Fatal(err)
	}
	kept, _ := os.ReadFile(o.Path)
	cases := []struct {
		name, id, text, want string
	}{
		{"unknown id", "SCRATCH-99", "text\n", "unknown id SCRATCH-99"},
		{"bug id", "BUG-3", "text\n", "BUG-3 is not a scratch item"},
		{"plan id", "PLAN-4", "text\n", "PLAN-4 is not a scratch item"},
		{"path id of a bug", "bugs/2026-09-24-crash", "text\n", "bugs/2026-09-24-crash is not a scratch item"},
		{"empty text", "SCRATCH-1", "", "scratch body is empty"},
		{"blank text", "SCRATCH-1", " \n\t\n", "scratch body is empty"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := mustLoad(t, cfg)
			refuse(t, cfg, c.want, func() error {
				_, err := AppendScratch(cfg, b, c.id, []byte(c.text))
				return err
			})
			if got, _ := os.ReadFile(o.Path); string(got) != string(kept) {
				t.Errorf("the scratch file changed: %q", got)
			}
		})
	}
}

func TestSetValueRefusesSpecced(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	if _, err := NewScratch(cfg, "idea", "", []byte("first note\n")); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cfg.Root, "scratch", "2026-09-26-idea.md")
	kept, _ := os.ReadFile(path)
	refuse(t, cfg, "specced comes from a spec's parent link", func() error {
		_, err := SetValue(cfg, mustLoad(t, cfg), "SCRATCH-1", "status", "specced")
		return err
	})
	if got, _ := os.ReadFile(path); string(got) != string(kept) {
		t.Errorf("the refused write changed the file: %q", got)
	}
	for _, v := range []string{"raw", "brainstorming", "dropped"} {
		if _, err := SetValue(cfg, mustLoad(t, cfg), "SCRATCH-1", "status", v); err != nil {
			t.Fatalf("status %q: %v", v, err)
		}
	}
	// A value no scratch status uses is still refused.
	refuse(t, cfg, `"fixing" is not a scratch status (raw, brainstorming, dropped)`, func() error {
		_, err := SetValue(cfg, mustLoad(t, cfg), "SCRATCH-1", "status", "fixing")
		return err
	})
}
