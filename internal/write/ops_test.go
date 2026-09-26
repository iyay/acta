package write

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

func repoWith(t *testing.T, files map[string]string) config.Config {
	t.Helper()
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(k, "test")
	}
	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "test@example.com")
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for p, body := range files {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, dir, "init", "-q", "-b", "main")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-q", "-m", "init")
	cfg := config.Default(dir)
	cfg.IsGit = true
	return cfg
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

var baseFiles = map[string]string{
	".pm/bugs/2026-09-24-crash.md":             "---\nref: B-1\n---\n# Crash\n\n## Symptom\nIt crashes.\n",
	".pm/plans/2026-09-25-crash-fix.md":        "---\nparent: bugs/2026-09-24-crash\n---\n# Fix\n\n### Task 1: Fix\n- [ ] a\n",
	"docs/superpowers/specs/2026-01-01-old.md": "# Old\n",
}

func mustLoad(t *testing.T, cfg config.Config) *board.Board {
	t.Helper()
	b, err := board.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSetValueCommits(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	o, err := SetValue(cfg, mustLoad(t, cfg), "bugs/2026-09-24-crash", "status", "fixed")
	if err != nil || !o.Committed || o.Skipped {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	body, _ := os.ReadFile(filepath.Join(cfg.Root, "bugs/2026-09-24-crash.md"))
	if string(body) != "---\nref: B-1\nstatus: fixed\n---\n# Crash\n\n## Symptom\nIt crashes.\n" {
		t.Fatalf("file = %q", body)
	}
	if got := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); got != "pm: bugs/2026-09-24-crash status fixed" {
		t.Fatalf("commit message %q", got)
	}
}

func TestSetValueBadInput(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	b := mustLoad(t, cfg)
	cases := []struct{ name, id, field, value string }{
		{"unknown id", "bugs/nope", "status", "fixed"},
		{"task", "plans/2026-09-25-crash-fix#task-1", "status", "done"},
		{"legacy", "docs/superpowers/specs/2026-01-01-old", "status", "done"},
		{"status not allowed for a bug", "bugs/2026-09-24-crash", "status", "in-progress"},
		{"bad type", "bugs/2026-09-24-crash", "type", "task"},
		{"unknown field", "bugs/2026-09-24-crash", "owner", "rian"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := SetValue(cfg, b, c.id, c.field, c.value)
			if !errors.Is(err, ErrBadInput) {
				t.Fatalf("err = %v, want ErrBadInput", err)
			}
		})
	}
	if out := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); out != "" {
		t.Fatalf("bad input changed files: %s", out)
	}
}

func TestSetValueDirtyFileIsWrittenNotCommitted(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	p := filepath.Join(cfg.Root, "bugs/2026-09-24-crash.md")
	if err := os.WriteFile(p, []byte("---\nref: B-1\n---\n# Crash\n\n## Symptom\nEdited by hand.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	o, err := SetValue(cfg, mustLoad(t, cfg), "bugs/2026-09-24-crash", "status", "fixing")
	if err != nil || o.Committed || !o.Skipped || !strings.Contains(o.Reason, "other uncommitted changes") {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	if body, _ := os.ReadFile(p); !strings.Contains(string(body), "status: fixing") || !strings.Contains(string(body), "Edited by hand.") {
		t.Fatalf("file = %q", body)
	}
}

func TestSetValueCRLFFileKeepsRef(t *testing.T) {
	cfg := repoWith(t, map[string]string{
		".pm/specs/2026-09-26-win.md": "---\r\nref: TICK-7\r\nstatus: draft\r\n---\r\n# Win spec\r\n",
	})
	b := mustLoad(t, cfg)
	if got := b.Get("specs/2026-09-26-win").Ref; got != "TICK-7" {
		t.Fatalf("ref before set = %q, want TICK-7", got)
	}
	if _, err := SetValue(cfg, b, "specs/2026-09-26-win", "status", "approved"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(cfg.Root, "specs/2026-09-26-win.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if strings.Count(strings.ReplaceAll(s, "\r\n", "\n"), "---\n") != 2 {
		t.Fatalf("two frontmatter blocks after set: %q", s)
	}
	if got := mustLoad(t, cfg).Get("specs/2026-09-26-win").Ref; got != "TICK-7" {
		t.Fatalf("ref after set = %q, want TICK-7 (file %q)", got, s)
	}
}

func TestSetValueAutoCommitOff(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	cfg.AutoCommit = false
	o, err := SetValue(cfg, mustLoad(t, cfg), "bugs/2026-09-24-crash", "type", "story")
	if err != nil || o.Committed || o.Skipped {
		t.Fatalf("outcome %+v err %v", o, err)
	}
}

func fixNow(t *testing.T) {
	t.Helper()
	old := Now
	Now = func() time.Time { return time.Date(2026, 9, 26, 10, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { Now = old })
}

func TestNewBug(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	fixNow(t)
	o, err := NewBug(cfg, "ack-dup", "", "New-261", []byte("## Symptom\nTwo ACKs.\n"))
	if err != nil || !o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	want := filepath.Join(cfg.Root, "bugs", "2026-09-26-ack-dup.md")
	if o.Path != want {
		t.Fatalf("path %s want %s", o.Path, want)
	}
	body, _ := os.ReadFile(want)
	got := mustLoad(t, cfg).Get("BUG-1")
	if got == nil || got.Path != want {
		t.Fatalf("BUG-1 = %+v, want path %s", got, want)
	}
	if !board.IsHash(strings.TrimPrefix(got.Hash, "BUG-")) {
		t.Fatalf("hash %q", got.Hash)
	}
	if !strings.Contains(string(body), "ref: New-261") || !strings.Contains(string(body), "## Symptom") {
		t.Fatalf("file lost content: %q", body)
	}
	if got := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); got != "pm: new bug 2026-09-26-ack-dup" {
		t.Fatalf("commit message %q", got)
	}
}

func TestNewBugBadInput(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	fixNow(t)
	cases := []struct{ name, slug, body string }{
		{"no symptom", "no-symptom", "## Repro\nx\n"},
		{"bad slug", "Bad Slug", "## Symptom\nx\n"},
		{"empty slug", "", "## Symptom\nx\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := NewBug(cfg, c.slug, "", "", []byte(c.body)); !errors.Is(err, ErrBadInput) {
				t.Fatalf("err = %v", err)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(cfg.Root, "bugs", "2026-09-26-no-symptom.md")); !os.IsNotExist(err) {
		t.Fatal("a bug with no Symptom was written")
	}
	if _, err := NewBug(cfg, "twice", "", "", []byte("## Symptom\nx\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := NewBug(cfg, "twice", "", "", []byte("## Symptom\nx\n")); !errors.Is(err, ErrBadInput) {
		t.Fatalf("second bug with the same name: err = %v", err)
	}
}

func TestNewBugCreatesRootFolder(t *testing.T) {
	cfg := repoWith(t, map[string]string{"README.md": "x\n"})
	fixNow(t)
	if _, err := NewBug(cfg, "first", "", "", []byte("## Symptom\nx\n")); err != nil {
		t.Fatal(err)
	}
}

func TestStartAndFinishBug(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	fixNow(t)

	path, tmpl, err := StartBug(cfg, "left-alone", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := FinishBug(cfg, path, tmpl); !errors.Is(err, ErrUnchanged) {
		t.Fatalf("unchanged: err = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("unchanged template was not removed")
	}

	path, tmpl, err = StartBug(cfg, "filled-in", "B-2")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tmpl), "ref: B-2") || !strings.Contains(string(tmpl), "# filled in") {
		t.Fatalf("template = %q", tmpl)
	}
	if err := os.WriteFile(path, append(tmpl, []byte("It broke.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	o, err := FinishBug(cfg, path, tmpl)
	if err != nil || !o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
}
func TestSetValueFixedInAndRef(t *testing.T) {
	cfg := repoWith(t, baseFiles)
	o, err := SetValue(cfg, mustLoad(t, cfg), "bugs/2026-09-24-crash", "fixed_in", "d2277688f")
	if err != nil || !o.Committed {
		t.Fatalf("fixed_in: %+v %v", o, err)
	}
	o, err = SetValue(cfg, mustLoad(t, cfg), "bugs/2026-09-24-crash", "ref", "New-261")
	if err != nil || !o.Committed {
		t.Fatalf("ref: %+v %v", o, err)
	}
	body, _ := os.ReadFile(filepath.Join(cfg.Root, "bugs/2026-09-24-crash.md"))
	if string(body) != "---\nref: New-261\nfixed_in: d2277688f\n---\n# Crash\n\n## Symptom\nIt crashes.\n" {
		t.Fatalf("file = %q", body)
	}
	if got := gitRun(t, cfg.RepoRoot, "log", "-1", "--format=%s"); got != "pm: bugs/2026-09-24-crash ref New-261" {
		t.Fatalf("commit message %q", got)
	}
}

func TestSetValueFixedInAndRefBadInput(t *testing.T) {
	files := map[string]string{
		".pm/bugs/2026-09-24-crash.md":  "# Crash\n\n## Symptom\nx\n",
		".pm/specs/2026-09-20-alpha.md": "# Alpha\n",
	}
	cfg := repoWith(t, files)
	b := mustLoad(t, cfg)
	cases := []struct{ name, id, field, value string }{
		{"fixed_in on a story", "specs/2026-09-20-alpha", "fixed_in", "abc1234"},
		{"sha too short", "bugs/2026-09-24-crash", "fixed_in", "abc12"},
		{"sha upper case", "bugs/2026-09-24-crash", "fixed_in", "ABC1234"},
		{"sha too long", "bugs/2026-09-24-crash", "fixed_in", strings.Repeat("a", 41)},
		{"sha not hex", "bugs/2026-09-24-crash", "fixed_in", "xyz1234"},
		{"ref with space", "bugs/2026-09-24-crash", "ref", "New 261"},
		{"ref empty", "bugs/2026-09-24-crash", "ref", ""},
		{"ref newline", "bugs/2026-09-24-crash", "ref", "A\nB"},
		{"ref too long", "bugs/2026-09-24-crash", "ref", strings.Repeat("a", 41)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := SetValue(cfg, b, c.id, c.field, c.value); !errors.Is(err, ErrBadInput) {
				t.Fatalf("err = %v, want ErrBadInput", err)
			}
		})
	}
	if out := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); out != "" {
		t.Fatalf("bad input changed files: %s", out)
	}
	if _, err := SetValue(cfg, b, "specs/2026-09-20-alpha", "ref", "S-1"); err != nil {
		t.Fatalf("ref on a story should work: %v", err)
	}
}

func TestSetValueFixedInAndRefKindsAndBounds(t *testing.T) {
	files := map[string]string{
		".pm/bugs/2026-09-24-crash.md":  "# Crash\n\n## Symptom\nx\n",
		".pm/specs/2026-09-20-alpha.md": "# Alpha\n",
		".pm/plans/2026-09-25-fix.md":   "# Fix\n\n### Task 1: Fix\n- [ ] a\n",
	}
	cfg := repoWith(t, files)
	b := mustLoad(t, cfg)
	// fixed_in is bugs-only: spec story, plan-as-story and task all refuse.
	for _, id := range []string{"specs/2026-09-20-alpha", "plans/2026-09-25-fix", "plans/2026-09-25-fix#task-1"} {
		if _, err := SetValue(cfg, b, id, "fixed_in", "d2277688f"); !errors.Is(err, ErrBadInput) {
			t.Fatalf("fixed_in on %s: err = %v, want ErrBadInput", id, err)
		}
	}
	// ref is stories and bugs only: tasks refuse.
	if _, err := SetValue(cfg, b, "plans/2026-09-25-fix#task-1", "ref", "S-1"); !errors.Is(err, ErrBadInput) {
		t.Fatalf("ref on task err = %v, want ErrBadInput", err)
	}
	// empty fixed_in is rejected; there is no clear path.
	if _, err := SetValue(cfg, b, "bugs/2026-09-24-crash", "fixed_in", ""); !errors.Is(err, ErrBadInput) {
		t.Fatalf("empty fixed_in err = %v, want ErrBadInput", err)
	}
	// ref must start with a letter or digit.
	for _, v := range []string{".a", "/a", "-a", "_a"} {
		if _, err := SetValue(cfg, b, "bugs/2026-09-24-crash", "ref", v); !errors.Is(err, ErrBadInput) {
			t.Fatalf("ref %q err = %v, want ErrBadInput", v, err)
		}
	}
	// lower bounds and 40-char upper bounds pass.
	if _, err := SetValue(cfg, b, "bugs/2026-09-24-crash", "fixed_in", strings.Repeat("a", 40)); err != nil {
		t.Fatalf("40-char sha: %v", err)
	}
	if _, err := SetValue(cfg, mustLoad(t, cfg), "specs/2026-09-20-alpha", "ref", strings.Repeat("a", 40)); err != nil {
		t.Fatalf("40-char ref: %v", err)
	}
	if out := gitRun(t, cfg.RepoRoot, "status", "--porcelain"); out != "" {
		// only the two boundary writes above committed; nothing dirty remains.
		t.Fatalf("dirty files after bounds: %s", out)
	}
}
