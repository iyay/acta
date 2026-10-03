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

// good is a page that keeps every rule in the repo of newHistory. Its
// timestamp is newer than every commit there.
func good(repo string) Page {
	return Page{
		Path:        filepath.Join(repo, ".acta", "wiki", "page.md"),
		ID:          "page",
		Type:        "Gotcha",
		Title:       "A title",
		Description: "One short line.",
		Paths:       []string{"src/a.go"},
		Timestamp:   year(2026),
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

		{"timestamp missing", func(p *Page) { p.Timestamp = time.Time{} }, []string{"timestamp is missing"}},

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

// A page is stale when a commit that touched a file under its paths is newer
// than its timestamp. Equal is not newer, and the time zone does not matter.
func TestCheckStale(t *testing.T) {
	t.Parallel()
	h := newHistory(t)
	at := func(s string) time.Time {
		ts, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return ts
	}
	tests := []struct {
		name  string
		paths []string
		stamp time.Time
		last  int // the year of the commit that makes the page stale; 0 for a page that is not
	}{
		{"file edited after the timestamp", []string{"src/a.go"}, at("2020-06-01T00:00:00Z"), 2021},
		{"file edited before the timestamp, others after", []string{"src/a.go"}, at("2021-06-01T00:00:00Z"), 0},
		{"timestamp at the very moment of the commit", []string{"src/a.go"}, year(2021), 0},
		{"timestamp one second before the commit", []string{"src/a.go"}, year(2021).Add(-time.Second), 2021},
		{"timestamp in another zone, before the commit", []string{"src/a.go"}, at("2021-01-01T00:30:00+07:00"), 2021},
		{"timestamp in another zone, the same moment", []string{"src/a.go"}, at("2021-01-01T07:00:00+07:00"), 0},
		{"folder, a file deep inside it edited after", []string{"src/"}, at("2022-06-01T00:00:00Z"), 2023},
		{"folder, every edit before the timestamp", []string{"src/"}, at("2023-06-01T00:00:00Z"), 0},
		{"folder, only a folder with a like name changed after", []string{"src/"}, at("2024-06-01T00:00:00Z"), 0},
		{"the newest of several entries decides", []string{"src/a.go", "other/c.go"}, at("2021-06-01T00:00:00Z"), 2022},
		{"no paths", nil, at("2000-01-01T00:00:00Z"), 0},
	}
	for _, tt := range tests {
		p := good(h.repo)
		p.Paths, p.Timestamp = tt.paths, tt.stamp
		var want []string
		if tt.last != 0 {
			want = []string{"stale: its paths changed on " + year(tt.last).Format(time.RFC3339) + ", after its timestamp " + tt.stamp.UTC().Format(time.RFC3339)}
		}
		expect(t, tt.name, Check(h.repo, []Page{p}, ""), want)
	}

	// A page with no timestamp has the missing field to fix first. Every commit
	// is newer than no time at all, which would only repeat itself.
	p := good(h.repo)
	p.Timestamp = time.Time{}
	expect(t, "no timestamp", Check(h.repo, []Page{p}, ""), []string{"timestamp is missing"})

	// Git that cannot answer is a problem, not a page that passes.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p = good(dir)
	p.Paths = []string{"a.go"}
	got := Check(dir, []Page{p}, "")
	if len(got) != 1 || !strings.Contains(got[0].Msg, "cannot read the git history") {
		t.Errorf("a folder with no repo: got %+v, want one problem about the git history", got)
	}
}

// With a range, only the pages it touches are checked. A page is touched when
// a file changed in the range sits under one of its paths, or when the page
// file itself changed. Every page below is broken the same way, so a page in
// the answer is a page that was looked at.
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
			got = append(got, strings.TrimSuffix(strings.TrimPrefix(pr.Page, ".acta/wiki/"), ".md"))
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("%s: Check(%q) looked at %v, want %v", tt.name, tt.rng, got, tt.want)
		}
	}

	// A stale page the range does not touch is left alone, and one it touches is
	// not. A bump of the timestamp is what clears it.
	old := func(id, path string) Page {
		p := good(h.repo)
		p.ID, p.Path, p.Paths, p.Timestamp = id, filepath.Join(h.repo, ".acta", "wiki", id+".md"), []string{path}, year(2000)
		return p
	}
	rng := h.sha[2020] + ".." + h.sha[2021]
	inside := old("inside", "src/a.go")
	got := Check(h.repo, []Page{inside, old("outside", "other/c.go")}, rng)
	if len(got) != 1 || got[0].Page != ".acta/wiki/inside.md" || !strings.HasPrefix(got[0].Msg, "stale:") {
		t.Errorf("stale pages in a range: got %+v, want only inside to be stale", got)
	}
	inside.Timestamp = year(2026)
	if got := Check(h.repo, []Page{inside}, rng); len(got) != 0 {
		t.Errorf("a touched page that keeps every rule: got %+v, want no problem", got)
	}
}

// A range git cannot read must not pass as a check of nothing. A range that
// starts with a dash is refused too, so it cannot be an option that writes a file.
func TestCheckRangeGitCannotRead(t *testing.T) {
	t.Parallel()
	h := newHistory(t)
	out := filepath.Join(t.TempDir(), "out.txt")
	for _, rng := range []string{"no-such-ref..HEAD", "--output=" + out} {
		got := Check(h.repo, []Page{good(h.repo)}, rng)
		if len(got) != 1 || got[0].Page != "" || !strings.Contains(got[0].Msg, rng) {
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
