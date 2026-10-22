package wiki

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// pageFile is how a problem names the page file made by good: from the repo root.
const pageFile = ".acta/wiki/page.md"

// year is midnight UTC on the first of January. The test repo makes one commit
// a year, so a timestamp in the middle of 2022 sits between two commits.
func year(y int) time.Time { return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC) }

// gitAt runs git in repo with both git dates set to when, so a commit it makes
// has that time. The dates go to this one command only, so tests that run side
// by side cannot disturb each other.
func gitAt(t *testing.T, repo string, when time.Time, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	stamp := when.Format(time.RFC3339)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+stamp, "GIT_COMMITTER_DATE="+stamp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// history is a repo with one commit a year. What each year changed:
//
//	2020  src/a.go, src/sub/b.go, other/c.go, srcx/d.go and the page files page.md and quiet.md
//	2021  src/a.go
//	2022  other/c.go
//	2023  src/sub/b.go
//	2024  page.md, a page file
//	2025  srcx/d.go, in a folder whose name starts like src
type history struct {
	repo string
	sha  map[int]string // the commit made in each year
}

func newHistory(t *testing.T) history {
	t.Helper()
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitAt(t, repo, year(2020), "init", "-q", "-b", "main")
	gitAt(t, repo, year(2020), "config", "user.email", "test@example.com")
	gitAt(t, repo, year(2020), "config", "user.name", "test")
	h := history{repo: repo, sha: map[int]string{}}
	steps := []struct {
		year  int
		files []string
	}{
		{2020, []string{"src/a.go", "src/sub/b.go", "other/c.go", "srcx/d.go", ".acta/wiki/page.md", ".acta/wiki/quiet.md"}},
		{2021, []string{"src/a.go"}},
		{2022, []string{"other/c.go"}},
		{2023, []string{"src/sub/b.go"}},
		{2024, []string{".acta/wiki/page.md"}},
		{2025, []string{"srcx/d.go"}},
	}
	for _, s := range steps {
		for _, f := range s.files {
			p := filepath.Join(repo, filepath.FromSlash(f))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			// The year in the text makes sure the commit really changes the file.
			if err := os.WriteFile(p, []byte(fmt.Sprintf("%s %d\n", f, s.year)), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		gitAt(t, repo, year(s.year), "add", ".")
		gitAt(t, repo, year(s.year), "commit", "-q", "-m", fmt.Sprint(s.year))
		h.sha[s.year] = gitAt(t, repo, year(s.year), "rev-parse", "HEAD")
	}
	return h
}

// good is a page that keeps every rule in the repo of newHistory. Its file was
// last edited in 2024, after the last change of src/a.go, so it is not stale.
func good(repo string) Page {
	return Page{
		Path:        filepath.Join(repo, ".acta", "wiki", "page.md"),
		ID:          "page",
		Type:        "Gotcha",
		Title:       "A title",
		Description: "One short line.",
		Paths:       []string{"src/a.go"},
		Words:       10,
	}
}

// expect wants exactly one problem for each piece of message in want, in that
// order, all on the page file of good. No want means no problem at all.
func expect(t *testing.T, name string, got []Problem, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: got %d problems %+v, want %d", name, len(got), got, len(want))
		return
	}
	for i, w := range want {
		if got[i].Page != pageFile || !strings.Contains(got[i].Msg, w) {
			t.Errorf("%s: problem %d = %+v, want page %s and %q in the message", name, i, got[i], pageFile, w)
		}
	}
}

// Each rule gets a page that breaks it and a page that keeps it. A breaking
// page breaks that one rule only, so the answer says exactly what is wrong.
func TestCheckFieldRules(t *testing.T) {
	t.Parallel()
	h := newHistory(t)
	tests := []struct {
		name string
		edit func(*Page)
		want []string // a piece of each message expected, in order; none for a page that keeps the rule
	}{
		{"all rules kept", func(p *Page) {}, nil},

		{"type Decision", func(p *Page) { p.Type = "Decision" }, nil},
		{"type Runbook", func(p *Page) { p.Type = "Runbook" }, nil},
		{"type Reference", func(p *Page) { p.Type = "Reference" }, nil},
		{"type Glossary", func(p *Page) { p.Type = "Glossary" }, nil},
		{"type missing", func(p *Page) { p.Type = "" }, []string{"type is missing"}},
		{"type unknown", func(p *Page) { p.Type = "Note" }, []string{`type "Note" is not one of`}},
		{"type in the wrong case", func(p *Page) { p.Type = "gotcha" }, []string{`type "gotcha" is not one of`}},

		{"title missing", func(p *Page) { p.Title = "" }, []string{"title is missing"}},
		{"title only spaces", func(p *Page) { p.Title = "  " }, []string{"title is missing"}},

		{"description missing", func(p *Page) { p.Description = "" }, []string{"description is missing"}},
		{"description 120 characters", func(p *Page) { p.Description = strings.Repeat("x", 120) }, nil},
		{"description 121 characters", func(p *Page) { p.Description = strings.Repeat("x", 121) },
			[]string{"description is 121 characters, the limit is 120"}},
		{"description 120 characters of two bytes", func(p *Page) { p.Description = strings.Repeat("é", 120) }, nil},
		{"description 121 characters of two bytes", func(p *Page) { p.Description = strings.Repeat("é", 121) },
			[]string{"description is 121 characters"}},
		{"description on two lines", func(p *Page) { p.Description = "one\ntwo" }, []string{"description is more than one line"}},
		{"description with a line break at the end", func(p *Page) { p.Description = "one\n" }, []string{"description is more than one line"}},
		{"description with a windows line break", func(p *Page) { p.Description = "one\r\ntwo" }, []string{"description is more than one line"}},
		{"description with a lone carriage return", func(p *Page) { p.Description = "one\rtwo" }, []string{"description is more than one line"}},
		{"description on two lines and too long", func(p *Page) { p.Description = strings.Repeat("x", 70) + "\n" + strings.Repeat("y", 60) },
			[]string{"description is more than one line", "description is 131 characters"}},

		{"body 250 words", func(p *Page) { p.Words = 250 }, nil},
		{"body 251 words", func(p *Page) { p.Words = 251 }, []string{"body is 251 words, the limit is 250"}},

		{"no paths", func(p *Page) { p.Paths = nil }, nil},
		{"entries that exist", func(p *Page) { p.Paths = []string{"src/a.go", "src/", "src/sub/", "other/c.go"} }, nil},
		{"file that does not exist", func(p *Page) { p.Paths = []string{"src/gone.go"} }, []string{`paths entry "src/gone.go" does not exist`}},
		{"folder that does not exist", func(p *Page) { p.Paths = []string{"gone/"} }, []string{`paths entry "gone/" does not exist`}},
		{"one good entry and one that does not exist", func(p *Page) { p.Paths = []string{"src/a.go", "gone.go"} },
			[]string{`paths entry "gone.go" does not exist`}},
		{"folder written without its slash", func(p *Page) { p.Paths = []string{"src"} }, []string{`paths entry "src" is a folder, write it as "src/"`}},
		{"file written with a slash", func(p *Page) { p.Paths = []string{"src/a.go/"} }, []string{`paths entry "src/a.go/" is a file, write it as "src/a.go"`}},
		{"entry that starts with ./", func(p *Page) { p.Paths = []string{"./src/a.go"} },
			[]string{`paths entry "./src/a.go" is not a plain path, write it as "src/a.go"`}},
		{"entry with a double slash", func(p *Page) { p.Paths = []string{"src//"} }, []string{`paths entry "src//" is not a plain path, write it as "src/"`}},
		{"entry with a dot dot step", func(p *Page) { p.Paths = []string{"src/../other/"} },
			[]string{`paths entry "src/../other/" is not a plain path, write it as "other/"`}},
		{"entry that climbs out of the repo", func(p *Page) { p.Paths = []string{"../src/"} }, []string{`paths entry "../src/" is not inside the repo`}},
		{"entry from the disk root", func(p *Page) { p.Paths = []string{"/etc/"} }, []string{`paths entry "/etc/" is not inside the repo`}},
		{"entry that is the whole repo", func(p *Page) { p.Paths = []string{"./"} }, []string{`paths entry "./" is not inside the repo`}},
		{"entry that is a dot", func(p *Page) { p.Paths = []string{"."} }, []string{`paths entry "." is not inside the repo`}},
		{"entry that is empty", func(p *Page) { p.Paths = []string{""} }, []string{`paths entry "" is empty`}},
	}
	for _, tt := range tests {
		p := good(h.repo)
		tt.edit(&p)
		expect(t, tt.name, Check(h.repo, []Page{p}, ""), tt.want)
	}
}

// staleMsg is the whole message of a stale page.
func staleMsg(sha, subject, id string) string {
	return fmt.Sprintf(`stale: %s "%s" changed its paths after the page was last checked; review the page against relevant changes and update any incorrect facts`, sha, subject)
}

// ancestry is a repo for the stale rule. Its first commit adds src/a.go and
// other/c.go. Its second commit adds the page file, so the page was checked
// after src/a.go was written. Every commit below gets its own date, and the
// dates are chosen to be misleading: the rule must follow the commit graph.
type ancestry struct {
	t    *testing.T
	repo string
}

func newAncestry(t *testing.T) ancestry {
	t.Helper()
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitAt(t, repo, year(2020), "init", "-q", "-b", "main")
	gitAt(t, repo, year(2020), "config", "user.email", "test@example.com")
	gitAt(t, repo, year(2020), "config", "user.name", "test")
	a := ancestry{t: t, repo: repo}
	a.commit("src/a.go", "first", year(2020))
	a.commit("other/c.go", "other first", year(2020))
	a.commit(pageFile, "write the page", year(2021))
	return a
}

// commit writes the file with new text and commits it at when.
func (a ancestry) commit(file, subject string, when time.Time) string {
	a.t.Helper()
	p := filepath.Join(a.repo, filepath.FromSlash(file))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		a.t.Fatal(err)
	}
	old, _ := os.ReadFile(p)
	if err := os.WriteFile(p, append(old, []byte(subject+"\n")...), 0o644); err != nil {
		a.t.Fatal(err)
	}
	gitAt(a.t, a.repo, when, "add", file)
	gitAt(a.t, a.repo, when, "commit", "-q", "-m", subject)
	return a.head()
}

func (a ancestry) head() string {
	return gitAt(a.t, a.repo, year(2020), "rev-parse", "--short", "HEAD")
}

// recheck writes a fresh line for the page into the checked file and commits it.
func (a ancestry) recheck(id string) {
	a.t.Helper()
	dir := filepath.Join(a.repo, ".acta", "wiki")
	if _, err := MarkChecked(dir, []string{"page", "other", "page2"}, []string{id}, "2026-10-14"); err != nil {
		a.t.Fatal(err)
	}
	gitAt(a.t, a.repo, year(2024), "add", ".acta/wiki/"+CheckedFile)
	gitAt(a.t, a.repo, year(2024), "commit", "-q", "-m", "chore(wiki): re-check "+id)
}

// writeChecked puts text in the checked file and commits it, so a test can
// make any change to the file it needs.
func (a ancestry) writeChecked(text, subject string) {
	a.t.Helper()
	p := filepath.Join(a.repo, ".acta", "wiki", CheckedFile)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		a.t.Fatal(err)
	}
	gitAt(a.t, a.repo, year(2024), "add", ".acta/wiki/"+CheckedFile)
	gitAt(a.t, a.repo, year(2024), "commit", "-q", "-m", subject)
}

func (a ancestry) check(rng string) []Problem { return Check(a.repo, []Page{good(a.repo)}, rng) }

// A page is stale when a commit that changed one of its paths is not behind
// any of its check commits: a commit that changed the page file, or one that
// added a line for its id to the checked file. Dates play no part.
func TestCheckStale(t *testing.T) {
	t.Parallel()
	t.Run("a path change after the last page edit is stale", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		expect(t, "before the change", a.check(""), nil)
		sha := a.commit("src/a.go", "change the code", year(2022))
		expect(t, "after the change", a.check(""), []string{staleMsg(sha, "change the code", "page")})
	})
	t.Run("a check commit clears it and a later change brings it back", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.commit("src/a.go", "change the code", year(2022))
		a.recheck("page")
		expect(t, "after the re-check", a.check(""), nil)
		sha := a.commit("src/a.go", "change it again", year(2023))
		expect(t, "after a later change", a.check(""), []string{staleMsg(sha, "change it again", "page")})
	})
	t.Run("editing the page file is a check too", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.commit("src/a.go", "change the code", year(2022))
		a.commit(pageFile, "update the page", year(2023))
		expect(t, "after the page edit", a.check(""), nil)
	})
	t.Run("a check for another page or a longer id does not count", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		sha := a.commit("src/a.go", "change the code", year(2022))
		a.recheck("other")
		a.recheck("page2")
		expect(t, "after the wrong re-checks", a.check(""), []string{staleMsg(sha, "change the code", "page")})
	})
	t.Run("a commit that only removes the line is not a check", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.writeChecked("other 2026-01-01\npage 2026-01-01\n", "check both")
		sha := a.commit("src/a.go", "change the code", year(2025))
		a.writeChecked("other 2026-01-01\n", "drop the page line")
		expect(t, "after the removal", a.check(""), []string{staleMsg(sha, "change the code", "page")})
	})
	t.Run("two branches that both re-check merge clean and stay fresh", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.commit("src/a.go", "change the code", year(2022))
		if err := os.WriteFile(filepath.Join(a.repo, ".gitattributes"), []byte(".acta/wiki/.checked merge=union\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitAt(t, a.repo, year(2022), "add", ".gitattributes")
		gitAt(t, a.repo, year(2022), "commit", "-q", "-m", "union merge")
		gitAt(t, a.repo, year(2022), "checkout", "-q", "-b", "side")
		a.writeChecked("page 2026-10-13\n", "check on side")
		gitAt(t, a.repo, year(2022), "checkout", "-q", "main")
		a.writeChecked("page 2026-10-14\n", "check on main")
		gitAt(t, a.repo, year(2025), "merge", "-q", "--no-edit", "side")
		expect(t, "after the merge", a.check(""), nil)
	})
	t.Run("a change outside the paths is not stale", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.commit("other/c.go", "change something else", year(2022))
		expect(t, "other file", a.check(""), nil)
	})
	t.Run("a folder entry sees a deep change", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		sha := a.commit("src/sub/b.go", "deep change", year(2022))
		p := good(a.repo)
		p.Paths = []string{"src/"}
		expect(t, "folder", Check(a.repo, []Page{p}, ""), []string{staleMsg(sha, "deep change", "page")})
	})
	t.Run("the newest unseen change is the one named", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.commit("src/a.go", "older change", year(2022))
		sha := a.commit("other/c.go", "newer change", year(2023))
		p := good(a.repo)
		p.Paths = []string{"src/a.go", "other/c.go"}
		expect(t, "two entries", Check(a.repo, []Page{p}, ""), []string{staleMsg(sha, "newer change", "page")})
	})
	t.Run("a page with no usable paths is never stale", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.commit("src/a.go", "change the code", year(2022))
		p := good(a.repo)
		p.Paths = nil
		expect(t, "no paths", Check(a.repo, []Page{p}, ""), nil)
		p.Paths = []string{"../src/"}
		expect(t, "path outside the repo", Check(a.repo, []Page{p}, ""), []string{`paths entry "../src/" is not inside the repo`})
	})
	t.Run("a page with no check commit at all is stale", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		p := good(a.repo)
		p.ID, p.Path = "fresh", filepath.Join(a.repo, ".acta", "wiki", "fresh.md")
		got := Check(a.repo, []Page{p}, "")
		if len(got) != 1 || got[0].Page != ".acta/wiki/fresh.md" || !strings.HasPrefix(got[0].Msg, "stale: ") {
			t.Errorf("a page nobody committed: got %+v, want one stale problem", got)
		}
	})
	t.Run("an uncommitted edit to the page file is not stale", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.commit("src/a.go", "change the code", year(2022))
		if err := os.WriteFile(filepath.Join(a.repo, filepath.FromSlash(pageFile)), []byte("edited\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		expect(t, "dirty page", a.check(""), nil)
	})
	t.Run("an untracked page file is not stale", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		a.commit("src/a.go", "change the code", year(2022))
		p := good(a.repo)
		p.ID, p.Path = "new", filepath.Join(a.repo, ".acta", "wiki", "new.md")
		if err := os.WriteFile(p.Path, []byte("new page\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := Check(a.repo, []Page{p}, ""); len(got) != 0 {
			t.Errorf("untracked page: got %+v, want no problem", got)
		}
	})
	t.Run("git that cannot answer is a problem", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		p := good(dir)
		p.Paths = []string{"a.go"}
		got := Check(dir, []Page{p}, "")
		if len(got) != 1 || got[0].Warning || !strings.Contains(got[0].Msg, "cannot read the git history") {
			t.Errorf("a folder with no repo: got %+v, want one problem about the git history", got)
		}
	})
}

// Two branches. The check of one branch does not see the path change of the
// other, whatever the dates say, so the merge is stale until the page is
// checked again behind both.
func TestCheckStaleAcrossBranches(t *testing.T) {
	t.Parallel()
	// setup: main re-checks the page; side, made from the same base, changes
	// the code at a date before the re-check.
	setup := func(t *testing.T) (ancestry, string) {
		a := newAncestry(t)
		gitAt(t, a.repo, year(2022), "checkout", "-q", "-b", "side")
		sha := a.commit("src/a.go", "side changes the code", year(2019))
		gitAt(t, a.repo, year(2022), "checkout", "-q", "main")
		a.recheck("page")
		return a, sha
	}
	t.Run("merging the other branch is stale", func(t *testing.T) {
		t.Parallel()
		a, sha := setup(t)
		expect(t, "before the merge", a.check(""), nil)
		gitAt(t, a.repo, year(2030), "merge", "-q", "--no-ff", "-m", "merge side", "side")
		expect(t, "after the merge", a.check(""), []string{staleMsg(sha, "side changes the code", "page")})
	})
	t.Run("merging a branch that checked the page too is not stale", func(t *testing.T) {
		t.Parallel()
		a, _ := setup(t)
		gitAt(t, a.repo, year(2022), "checkout", "-q", "side")
		a.recheck("page")
		gitAt(t, a.repo, year(2022), "checkout", "-q", "main")
		gitAt(t, a.repo, year(2018), "merge", "-q", "--no-ff", "-m", "merge side", "side")
		expect(t, "after the merge", a.check(""), nil)
	})
	t.Run("a path change dated before the page edit but not behind it is stale", func(t *testing.T) {
		t.Parallel()
		a := newAncestry(t)
		gitAt(t, a.repo, year(2022), "checkout", "-q", "-b", "side")
		sha := a.commit("src/a.go", "old date, other branch", year(2000))
		gitAt(t, a.repo, year(2022), "checkout", "-q", "main")
		a.commit(pageFile, "edit the page", year(2040))
		gitAt(t, a.repo, year(2041), "merge", "-q", "--no-ff", "-m", "merge side", "side")
		expect(t, "after the merge", a.check(""), []string{staleMsg(sha, "old date, other branch", "page")})
	})
}

// With a range, only the pages it touches are checked. A page is touched when
// a file changed in the range sits under one of its paths, or when the page
// file itself changed. Every page below is broken the same way, so a page in
// the answer is a page that was looked at. A page can have more than one
// problem, and it still shows once.
func TestCheckRange(t *testing.T) {
	t.Parallel()
	h := newHistory(t)
	broken := func(id string, paths ...string) Page {
		p := good(h.repo)
		p.ID, p.Path, p.Paths, p.Description = id, filepath.Join(h.repo, ".acta", "wiki", id+".md"), paths, ""
		return p
	}
	pages := []Page{
		broken("file", "src/a.go"),
		broken("folder", "src/"),
		broken("other", "other/c.go"),
		broken("sibling", "srcx/"),
		broken("nopaths"),
		// page.md is a page file that was committed, and changed in 2024.
		broken("page", "other/c.go"),
	}
	tests := []struct {
		name string
		rng  string
		want []string
	}{
		{"no range checks every page", "", []string{"file", "folder", "other", "sibling", "nopaths", "page"}},
		{"a file entry and a folder entry", h.sha[2020] + ".." + h.sha[2021], []string{"file", "folder"}},
		{"a file entry, and the page that lists it", h.sha[2021] + ".." + h.sha[2022], []string{"other", "page"}},
		{"only the page file changed", h.sha[2023] + ".." + h.sha[2024], []string{"page"}},
		{"a folder whose name starts like another", h.sha[2024] + ".." + h.sha[2025], []string{"sibling"}},
		{"a range with no change", h.sha[2021] + ".." + h.sha[2021], nil},
		{"a wide range, but not the page with no paths", h.sha[2020] + ".." + h.sha[2025], []string{"file", "folder", "other", "sibling", "page"}},
	}
	for _, tt := range tests {
		var got []string
		for _, pr := range Check(h.repo, pages, tt.rng) {
			name := strings.TrimSuffix(strings.TrimPrefix(pr.Page, ".acta/wiki/"), ".md")
			if len(got) == 0 || got[len(got)-1] != name {
				got = append(got, name)
			}
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("%s: Check(%q) looked at %v, want %v", tt.name, tt.rng, got, tt.want)
		}
	}

	// A stale page the range does not touch is left alone, and one it touches is
	// not. Neither of these two page files was ever committed, so each is stale
	// against any change of its paths.
	fresh := func(id, path string) Page {
		p := good(h.repo)
		p.ID, p.Path, p.Paths = id, filepath.Join(h.repo, ".acta", "wiki", id+".md"), []string{path}
		return p
	}
	rng := h.sha[2020] + ".." + h.sha[2021]
	got := Check(h.repo, []Page{fresh("inside", "src/a.go"), fresh("outside", "other/c.go")}, rng)
	if len(got) != 1 || got[0].Page != ".acta/wiki/inside.md" || !strings.HasPrefix(got[0].Msg, "stale:") {
		t.Errorf("stale pages in a range: got %+v, want only inside to be stale", got)
	}
	// page.md was edited in 2024, behind the last change of src/a.go.
	if got := Check(h.repo, []Page{good(h.repo)}, rng); len(got) != 0 {
		t.Errorf("a touched page that keeps every rule: got %+v, want no problem", got)
	}
}

// The range form follows the same rule as the whole check: a page the branch
// touches is stale while a path change on it is not behind a check commit.
func TestCheckRangeStale(t *testing.T) {
	t.Parallel()
	a := newAncestry(t)
	gitAt(t, a.repo, year(2022), "checkout", "-q", "-b", "feature")
	sha := a.commit("src/a.go", "feature changes the code", year(2022))
	expect(t, "unseen change", a.check("main..HEAD"), []string{staleMsg(sha, "feature changes the code", "page")})
	a.recheck("page")
	expect(t, "after the re-check", a.check("main..HEAD"), nil)
	// A page edit on the branch is a check as well.
	gitAt(t, a.repo, year(2022), "checkout", "-q", "-b", "second", "main")
	a.commit("src/a.go", "second changes the code", year(2023))
	a.commit(pageFile, "second edits the page", year(2024))
	expect(t, "after the page edit", a.check("main..HEAD"), nil)
}

// A range git cannot read must not pass as a check of nothing. A range that
// starts with a dash is refused too, so it cannot be an option that writes a file.
func TestCheckRangeGitCannotRead(t *testing.T) {
	t.Parallel()
	h := newHistory(t)
	out := filepath.Join(t.TempDir(), "out.txt")
	for _, rng := range []string{"no-such-ref..HEAD", "--output=" + out} {
		got := Check(h.repo, []Page{good(h.repo)}, rng)
		if len(got) != 1 || got[0].Warning || got[0].Page != "" || !strings.Contains(got[0].Msg, rng) {
			t.Errorf("Check(%q) = %+v, want one problem with no page that names the range", rng, got)
		}
	}
	if _, err := os.Stat(out); err == nil {
		t.Error("a range that starts with a dash made git write a file")
	}
	// With no pages there is nothing to check, so a range is not read at all.
	if got := Check(h.repo, nil, "no-such-ref..HEAD"); len(got) != 0 {
		t.Errorf("Check with no pages = %+v, want no problem", got)
	}
}

// TestCheckGitCallsDoNotGrowWithPages counts git processes with a wrapper
// first on PATH. On a long history each git log costs most of a second, so a
// run that asks git once per page takes minutes on a big wiki.
func TestCheckGitCallsDoNotGrowWithPages(t *testing.T) {
	h := newHistory(t)
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	logFile := filepath.Join(dir, "calls")
	script := "#!/bin/sh\necho x >> '" + logFile + "'\nexec '" + real + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	calls := func(n int) int {
		var pages []Page
		for i := range n {
			p := good(h.repo)
			if i > 0 {
				p.ID = fmt.Sprintf("page%d", i)
				p.Path = filepath.Join(h.repo, ".acta", "wiki", p.ID+".md")
			}
			pages = append(pages, p)
		}
		if err := os.Remove(logFile); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		Check(h.repo, pages, "")
		b, _ := os.ReadFile(logFile)
		return strings.Count(string(b), "\n")
	}
	if one, twenty := calls(1), calls(20); one != twenty {
		t.Errorf("git ran %d times for 1 page and %d times for 20 pages, want the same", one, twenty)
	}
}

func TestCheckWarningSeverity(t *testing.T) {
	a := newAncestry(t)
	a.commit("src/a.go", "change code", year(2022))
	p := good(a.repo)
	p.Description = ""
	got := Check(a.repo, []Page{p}, "")
	if len(got) != 2 || got[0].Warning || !got[1].Warning {
		t.Fatalf("mixed page diagnostics = %+v; want field error then stale warning", got)
	}
}
