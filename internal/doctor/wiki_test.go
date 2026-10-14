package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/wiki"
)

// wikiEnv is env with a committed wiki page for each id. Each page covers
// src/<id>.go, and the first commit holds both the pages and those files.
func wikiEnv(t *testing.T, ids ...string) Env {
	t.Helper()
	e := env(t)
	gitT(t, e, "config", "user.email", "t@example.com")
	gitT(t, e, "config", "user.name", "t")
	for _, id := range ids {
		write(t, filepath.Join(e.ActaRoot, "wiki", id+".md"),
			"---\ntype: Gotcha\ntitle: Page\ndescription: Page "+id+"\npaths: [src/"+id+".go]\n---\nbody\n")
		write(t, filepath.Join(e.RepoRoot, "src", id+".go"), "package src\n")
	}
	gitT(t, e, "add", "-A")
	gitT(t, e, "commit", "-q", "-m", "start")
	return e
}

func gitT(t *testing.T, e Env, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", e.RepoRoot}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// change commits an edit to a file the page id covers.
func change(t *testing.T, e Env, id, msg string) {
	t.Helper()
	p := filepath.Join(e.RepoRoot, "src", id+".go")
	old, _ := os.ReadFile(p)
	write(t, p, string(old)+"// "+msg+"\n")
	gitT(t, e, "commit", "-q", "-am", msg)
}

// fixAndCommit runs Fix and commits what it wrote, the way the command does.
func fixAndCommit(t *testing.T, e Env) []string {
	t.Helper()
	paths, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) > 0 {
		gitT(t, e, append([]string{"add", "--"}, paths...)...)
		gitT(t, e, "commit", "-q", "-m", "chore: doctor fix")
	}
	return paths
}

func stalePages(t *testing.T, e Env) []string {
	t.Helper()
	pages, _ := wiki.Load(e.ActaRoot)
	var out []string
	for _, p := range wiki.Check(e.RepoRoot, pages, "") {
		if strings.HasPrefix(p.Msg, "stale: ") {
			out = append(out, filepath.Base(p.Page))
		}
	}
	slices.Sort(out)
	return out
}

func TestDoctorWikiWithNoWikiIsOK(t *testing.T) {
	e := env(t)
	wantLevel(t, byName(Run(e), "wiki"), OK, "")
	paths, err := Fix(e)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		if strings.Contains(p, "wiki") || strings.HasSuffix(p, ".gitattributes") {
			t.Errorf("Fix wrote %s with no wiki", p)
		}
	}
}

func TestDoctorWikiWarnsOnMissingUnionLine(t *testing.T) {
	e := wikiEnv(t, "a")
	fixAndCommit(t, e)
	write(t, filepath.Join(e.RepoRoot, ".gitattributes"), "*.png binary\n")
	gitT(t, e, "commit", "-q", "-am", "drop union")
	wantLevel(t, byName(Run(e), "wiki"), Warn, "acta doctor --fix")
	fixAndCommit(t, e)
	raw, _ := os.ReadFile(filepath.Join(e.RepoRoot, ".gitattributes"))
	if string(raw) != "*.png binary\n.acta/wiki/.checked merge=union\n" {
		t.Errorf(".gitattributes %q", raw)
	}
	wantLevel(t, byName(Run(e), "wiki"), OK, "")
}

func TestDoctorWikiWarnsOnDuplicateAndOrphanLines(t *testing.T) {
	for name, body := range map[string]string{
		"duplicate": "a 2026-01-01T00:00:00Z\na 2026-02-01T00:00:00Z\n",
		"orphan":    "a 2026-01-01T00:00:00Z\ngone 2026-01-01T00:00:00Z\n",
	} {
		t.Run(name, func(t *testing.T) {
			e := wikiEnv(t, "a")
			fixAndCommit(t, e)
			write(t, filepath.Join(e.ActaRoot, "wiki", wiki.CheckedFile), body)
			gitT(t, e, "commit", "-q", "-am", "mess")
			wantLevel(t, byName(Run(e), "wiki"), Warn, "acta doctor --fix")
			fixAndCommit(t, e)
			wantLevel(t, byName(Run(e), "wiki"), OK, "")
			if got := stalePages(t, e); len(got) != 0 {
				t.Errorf("stale after the fix: %v", got)
			}
		})
	}
}

// A repo checked with the old trailers moves to the checked file: a page fresh
// before stays fresh, a page stale before stays stale.
func TestDoctorFixMovesOldTrailerChecks(t *testing.T) {
	e := wikiEnv(t, "fresh", "stale", "edited")
	change(t, e, "fresh", "touch fresh")
	change(t, e, "stale", "touch stale")
	gitT(t, e, "commit", "-q", "--allow-empty", "-m", "chore(wiki): re-check fresh\n\nWiki-Checked: fresh")
	change(t, e, "stale", "touch stale again")
	wantLevel(t, byName(Run(e), "wiki"), Warn, "acta doctor --fix")
	fixAndCommit(t, e)
	raw, _ := os.ReadFile(filepath.Join(e.ActaRoot, "wiki", wiki.CheckedFile))
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		ids = append(ids, strings.Fields(line)[0])
	}
	if strings.Join(ids, " ") != "edited fresh" {
		t.Errorf("checked file %q; want lines for edited and fresh only", raw)
	}
	if got := stalePages(t, e); strings.Join(got, " ") != "stale.md" {
		t.Errorf("stale pages %v; want only stale.md", got)
	}
	if again := fixAndCommit(t, e); len(again) != 0 {
		t.Errorf("a second fix wrote %v", again)
	}
	// Staleness is for acta wiki check to say. The setup itself is fine now.
	wantLevel(t, byName(Run(e), "wiki"), OK, "")
}
