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
	if o.ShortID != "SCR-0001" {
		t.Errorf("short id %q want SCR-0001", o.ShortID)
	}
	got, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"id: SCR-0001", "status: raw", "title: newest-first", `created: "2026-09-26"`, `schema: "1"`} {
		if !strings.Contains(string(got), field) {
			t.Errorf("file lacks %q:\n%s", field, got)
		}
	}
	// The user's words sit under Words byte for byte, the accents included.
	if want := "## Words\n\n### 2026-09-26\n\n" + string(body); !strings.Contains(string(got), want) {
		t.Errorf("body changed:\ngot  %q\nwant %q", got, want)
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

func TestNewScratchWritesNewFormatIDAndHash(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	o, err := NewScratch(cfg, "short-id", "", []byte("note\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := frontField(t, o.Path, "id"); got != "SCR-0001" {
		t.Errorf("scratch id = %q, want SCR-0001", got)
	}
	if h := frontField(t, o.Path, "hash"); !board.IsHash(h) {
		t.Errorf("scratch hash = %q, want 7 characters", h)
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
	want := "schema: \"1\"\n---\n# rule-body\n\n## Words\n\n### 2026-09-26\n\n" + strings.TrimRight(string(body), "\n") + "\n\n## Context\n\n## Log\n\n## Open questions\n"
	if !strings.HasSuffix(string(got), want) {
		t.Errorf("file = %q, want it to end with %q", got, want)
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
	if want := "## Words\n\n### 2026-09-26\n\n" + string(body); !strings.Contains(string(got), want) {
		t.Errorf("body bytes changed:\ngot  %q\nwant %q", got, want)
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
	got, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", "", []byte("answer 1 ✓\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Committed || got.Skipped || got.ShortID != "SCR-0001" {
		t.Fatalf("outcome %+v", got)
	}
	body, _ := os.ReadFile(o.Path)
	if want := "# idea\n\n## Words\n\n### 2026-09-26\n\nfirst note ✓\n\n### 2026-09-26\n\nanswer 1 ✓\n\n## Context\n\n## Log\n\n## Open questions\n"; board.Parse(body).Body != want {
		t.Errorf("body = %q, want %q", body, want)
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
	if _, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", "", []byte("answer 2\n")); err != nil {
		t.Fatal(err)
	}
	body, _ = os.ReadFile(o.Path)
	if want := "# idea\n\n## Words\n\n### 2026-09-26\n\nfirst note ✓\n\n### 2026-09-26\n\nanswer 1 ✓\n\n### 2026-09-26\n\nanswer 2\n\n## Context\n\n## Log\n\n## Open questions\n"; board.Parse(body).Body != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

func TestAppendScratchAddsTheNewlineTheBodyLacks(t *testing.T) {
	fixNow(t)
	// A new item's body always ends with a newline, so this is the old
	// append path that has to add the missing one.
	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-01-old.md": "---\nid: SCR-0001\nhash: aaaaaaa\ntitle: old\nstatus: raw\n---\nno newline at the end",
	})
	o, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", "", []byte("next line\n"))
	if err != nil {
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
				_, err := AppendScratch(cfg, b, c.id, "", []byte(c.text))
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

func TestNewScratchWritesSkeleton(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	o, err := NewScratch(cfg, "idea", "Idea", []byte("kata user\n\n\n"))
	if err != nil {
		t.Fatal(err)
	}
	src, _ := os.ReadFile(o.Path)
	doc := board.Parse(src)
	if !board.HasSchema(doc.Front) {
		t.Errorf("no schema: 1 in %q", src)
	}
	if doc.Front["created"] != "2026-09-26" {
		t.Errorf("created = %v", doc.Front["created"])
	}
	want := "# Idea\n\n## Words\n\n### 2026-09-26\n\nkata user\n\n## Context\n\n## Log\n\n## Open questions\n"
	if doc.Body != want {
		t.Errorf("body\n%q\nwant\n%q", doc.Body, want)
	}
	if p := board.CheckBody(board.KindScratch, doc.Body); p != nil {
		t.Errorf("a new file fails its own schema: %v", p)
	}
}

func TestAppendScratchSections(t *testing.T) {
	const head = "# Idea\n\n## Words\n\n### 2026-09-26\n\nkata user\n\n"
	cases := []struct{ section, want string }{
		{"", head + "### 2026-09-26\n\nmore\n\n## Context\n\n## Log\n\n## Open questions\n"},
		{"words", head + "### 2026-09-26\n\nmore\n\n## Context\n\n## Log\n\n## Open questions\n"},
		{"context", head + "## Context\n\nmore\n\n## Log\n\n## Open questions\n"},
		{"log", head + "## Context\n\n## Log\n\n### 2026-09-26\n\nmore\n\n## Open questions\n"},
		{"questions", head + "## Context\n\n## Log\n\n## Open questions\n\nmore\n"},
	}
	for _, c := range cases {
		t.Run("section="+c.section, func(t *testing.T) {
			fixNow(t)
			cfg := repoWith(t, baseFiles)
			o, err := NewScratch(cfg, "idea", "Idea", []byte("kata user\n"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := AppendScratch(cfg, mustLoad(t, cfg), o.ShortID, c.section, []byte("more\n")); err != nil {
				t.Fatal(err)
			}
			src, _ := os.ReadFile(o.Path)
			body := board.Parse(src).Body
			if body != c.want {
				t.Errorf("body\n%q\nwant\n%q", body, c.want)
			}
			if p := board.CheckBody(board.KindScratch, body); p != nil {
				t.Errorf("the append broke the schema: %v", p)
			}
		})
	}
}

func TestAppendScratchBadSection(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	o, err := NewScratch(cfg, "idea", "Idea", []byte("x\n"))
	if err != nil {
		t.Fatal(err)
	}
	kept, _ := os.ReadFile(o.Path)
	refuse(t, cfg, `unknown section "notes"; use words, context, log or questions`, func() error {
		_, err := AppendScratch(cfg, mustLoad(t, cfg), o.ShortID, "notes", []byte("y\n"))
		return err
	})
	if got, _ := os.ReadFile(o.Path); string(got) != string(kept) {
		t.Errorf("the refused write changed the file: %q", got)
	}
}

func TestAppendScratchOldItem(t *testing.T) {
	// An old item has no schema field, so it has no sections to fill. No flag
	// keeps today's append; a flag is refused.
	fixNow(t)
	cfg := repoWith(t, map[string]string{
		".acta/scratch/2026-09-01-old.md": "---\nid: SCR-0001\nhash: aaaaaaa\ntitle: old\nstatus: raw\n---\nfree text\n",
	})
	path := filepath.Join(cfg.Root, "scratch", "2026-09-01-old.md")
	o, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", "", []byte("more\n"))
	if err != nil {
		t.Fatal(err)
	}
	if o.Path != path {
		t.Errorf("path %s want %s", o.Path, path)
	}
	src, _ := os.ReadFile(path)
	want := "---\nid: SCR-0001\nhash: aaaaaaa\ntitle: old\nstatus: raw\n---\nfree text\n\nmore\n"
	if string(src) != want {
		t.Errorf("the old item was reshaped:\ngot  %q\nwant %q", src, want)
	}
	refuse(t, cfg, "SCR-0001 is an old item with no sections", func() error {
		_, err := AppendScratch(cfg, mustLoad(t, cfg), "SCRATCH-1", "context", []byte("y\n"))
		return err
	})
	if after, _ := os.ReadFile(path); string(after) != want {
		t.Errorf("the refused write changed the file: %q", after)
	}
}

func TestAppendScratchPutsBackAMissingHeading(t *testing.T) {
	// A hand edit took a part away. The text still belongs there, so the
	// heading is written back where the schema wants it.
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	o, err := NewScratch(cfg, "idea", "Idea", []byte("kata user\n"))
	if err != nil {
		t.Fatal(err)
	}
	handEdit(t, cfg, o.Path, "## Log\n\n")
	if _, err := AppendScratch(cfg, mustLoad(t, cfg), o.ShortID, "log", []byte("more\n")); err != nil {
		t.Fatal(err)
	}
	src, _ := os.ReadFile(o.Path)
	want := "# Idea\n\n## Words\n\n### 2026-09-26\n\nkata user\n\n## Context\n\n## Log\n\n### 2026-09-26\n\nmore\n\n## Open questions\n"
	if body := board.Parse(src).Body; body != want {
		t.Errorf("body %q want %q", body, want)
	}
}

func TestAppendScratchRefusesABodyThatBreaksItsSchema(t *testing.T) {
	fixNow(t)
	cfg := repoWith(t, baseFiles)
	o, err := NewScratch(cfg, "idea", "Idea", []byte("kata user\n"))
	if err != nil {
		t.Fatal(err)
	}
	handEdit(t, cfg, o.Path, "## Words\n\n### 2026-09-26\n\nkata user\n\n")
	refuse(t, cfg, "scratch "+filepath.Base(o.Path)+": missing ## Words", func() error {
		_, err := AppendScratch(cfg, mustLoad(t, cfg), o.ShortID, "context", []byte("more\n"))
		return err
	})
}

// handEdit takes a piece out of a written file and commits it, so the tree
// is clean again before a refused write checks it.
func handEdit(t *testing.T, cfg config.Config, path, cut string) {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), cut) {
		t.Fatalf("%s does not hold %q", path, cut)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(src), cut, "", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, cfg.RepoRoot, "add", ".")
	gitRun(t, cfg.RepoRoot, "commit", "-q", "-m", "break the body by hand")
}
