package wiki

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// writePage puts one file under <root>/wiki and makes the folders it needs.
func writePage(t *testing.T, root, name, body string) {
	t.Helper()
	p := filepath.Join(root, "wiki", filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func ids(pages []Page) []string {
	var out []string
	for _, p := range pages {
		out = append(out, p.ID)
	}
	return out
}

func TestLoad(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Pages that load.
	writePage(t, root, "tui-wrap.md", `---
type: Gotcha
title: x/ansi Wrap overflows
description: Wrap can return lines wider than the limit after " -"; use Hardwrap(Wordwrap(...))
paths: [internal/tui/]
timestamp: 2026-10-04T14:05:00+07:00
---
one two three four five six seven
`)
	// This body has a --- line of its own. It must not cut the body short. It
	// counts as one word, like any chunk of text between spaces.
	writePage(t, root, "sub/deep.md", "---\ntype: Decision\ntitle: Deep one\ndescription: Lives in a subfolder\npaths:\n  - a.go\n  - b/\ntimestamp: 2026-10-05T01:00:00Z\n---\none two\n---\nthree four five\n")
	writePage(t, root, "no-paths.md", "---\ntype: Reference\ntitle: Loose\ndescription: Covers no file\ntimestamp: 2026-10-07T00:00:00Z\n---\nbody\n")
	writePage(t, root, "crlf.md", "---\r\ntype: Runbook\r\ntitle: Crlf page\r\ndescription: Written with Windows line ends\r\npaths: [x.go]\r\ntimestamp: 2026-10-06T00:00:00Z\r\n---\r\nalpha beta\r\n")

	// Files that are not pages, or cannot be loaded.
	writePage(t, root, "notes.txt", "---\ntype: [never read\n")
	writePage(t, root, "bad-yaml.md", "---\ntype: [unclosed\n---\nbody\n")
	writePage(t, root, "bad-time.md", "---\ntype: Gotcha\ntitle: Bad time\ndescription: d\npaths: [a.go]\ntimestamp: yesterday\n---\nbody\n")
	writePage(t, root, "date-only.md", "---\ntype: Gotcha\ntitle: No clock time\ndescription: d\npaths: [a.go]\ntimestamp: 2026-10-04\n---\nbody\n")
	writePage(t, root, "paths-scalar.md", "---\ntype: Gotcha\ntitle: Paths without brackets\ndescription: d\npaths: internal/tui/\ntimestamp: 2026-10-04T00:00:00Z\n---\nbody\n")
	writePage(t, root, "sub/inner/broken.md", "---\ntype: [unclosed\n---\nbody\n")
	writePage(t, root, "no-front.md", "just words, no frontmatter\n")
	writePage(t, root, "unclosed.md", "---\ntype: Gotcha\ntitle: Never closed\n")
	if err := os.MkdirAll(filepath.Join(root, "wiki", "empty.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("nowhere", filepath.Join(root, "wiki", "dead.md")); err != nil {
		t.Fatal(err)
	}

	pages, errs := Load(root)

	// Pages come back sorted by name. A subfolder sits where its own name falls.
	if got, want := ids(pages), []string{"crlf", "no-paths", "sub/deep", "tui-wrap"}; !slices.Equal(got, want) {
		t.Fatalf("page ids = %v, want %v", got, want)
	}
	wrap := pages[3]
	if wrap.Path != filepath.Join(root, "wiki", "tui-wrap.md") || wrap.ID != "tui-wrap" || wrap.Type != "Gotcha" ||
		wrap.Title != "x/ansi Wrap overflows" || !slices.Equal(wrap.Paths, []string{"internal/tui/"}) || wrap.Words != 7 ||
		wrap.Description != `Wrap can return lines wider than the limit after " -"; use Hardwrap(Wordwrap(...))` {
		t.Errorf("tui-wrap = %+v", wrap)
	}
	if !wrap.Timestamp.Equal(time.Date(2026, 10, 4, 7, 5, 0, 0, time.UTC)) {
		t.Errorf("tui-wrap timestamp = %v, want 2026-10-04T14:05:00+07:00", wrap.Timestamp)
	}

	deep := pages[2]
	if deep.Path != filepath.Join(root, "wiki", "sub", "deep.md") || deep.Type != "Decision" ||
		!slices.Equal(deep.Paths, []string{"a.go", "b/"}) || deep.Words != 6 {
		t.Errorf("sub/deep = %+v", deep)
	}
	if loose := pages[1]; len(loose.Paths) != 0 || loose.Title != "Loose" {
		t.Errorf("no-paths = %+v, want a page with no paths", loose)
	}
	if crlf := pages[0]; crlf.Description != "Written with Windows line ends" || crlf.Words != 2 {
		t.Errorf("crlf = %+v", crlf)
	}

	// Each file that cannot load comes back once, with its file in the message.
	wantErrs := []string{"bad-time.md", "bad-yaml.md", "date-only.md", "dead.md", "no-front.md", "paths-scalar.md", "sub/inner/broken.md", "unclosed.md"}
	if len(errs) != len(wantErrs) {
		t.Fatalf("got %d errors, want %d: %v", len(errs), len(wantErrs), errs)
	}
	for i, name := range wantErrs {
		if !strings.Contains(errs[i].Error(), filepath.Join(root, "wiki", name)) {
			t.Errorf("error %d = %q, want it to name %s", i, errs[i], name)
		}
	}
}

func TestLoadMissingFolder(t *testing.T) {
	t.Parallel()
	roots := map[string]string{
		"root with no wiki folder": t.TempDir(),
		"root that does not exist": filepath.Join(t.TempDir(), "nope"),
	}
	for name, root := range roots {
		pages, errs := Load(root)
		if len(pages) != 0 || len(errs) != 0 {
			t.Errorf("%s: pages = %v, errors = %v, want neither", name, pages, errs)
		}
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()
	pages := []Page{
		{ID: "tui", Paths: []string{"internal/tui/"}},
		{ID: "parser", Paths: []string{"src/parser.go", "docs/"}},
		{ID: "loose"},
		// Entries that point out of the repo or at nothing never match.
		{ID: "outside", Paths: []string{"../evil/", "/etc/", ""}},
	}
	tests := []struct {
		name string
		rel  string
		want []string
	}{
		// A path matches when it equals an entry, or sits under a folder entry.
		{"file equals a file entry", "src/parser.go", []string{"parser"}},
		{"file under a folder entry", "internal/tui/model.go", []string{"tui"}},
		{"file deep under a folder entry", "internal/tui/sub/deep/x.go", []string{"tui"}},
		{"folder equals its own entry", "internal/tui/", []string{"tui"}},
		{"second entry of a page", "docs/a.md", []string{"parser"}},
		{"second entry, folder itself", "docs/", []string{"parser"}},

		// What a shell word adds is cut first.
		{"double quotes", `"internal/tui/model.go"`, []string{"tui"}},
		{"single quotes", `'internal/tui/model.go'`, []string{"tui"}},
		{"leading ./", "./internal/tui/model.go", []string{"tui"}},
		{"leading ./ on a file entry", "./src/parser.go", []string{"parser"}},
		{"go pattern with ./ and /...", "./internal/tui/...", []string{"tui"}},
		{"go pattern with /...", "internal/tui/...", []string{"tui"}},
		{"quoted go pattern", `"./internal/tui/..."`, []string{"tui"}},
		{"double slash", "internal//tui/model.go", []string{"tui"}},
		{"dot dot that stays inside the repo", "internal/tui/../tui/model.go", []string{"tui"}},

		// A sibling name is not a match.
		{"sibling folder, longer name", "internal/tuix/x.go", nil},
		{"sibling folder, dashed name", "internal/tui-old/x.go", nil},
		{"sibling folder with slash", "internal/tuix/", nil},
		{"sibling folder, bare", "internal/tuix", nil},
		{"sibling folder before the page folder", "xinternal/tui/x.go", nil},
		{"sibling of a docs folder", "docs2/a.md", nil},
		{"file entry as prefix of a longer file", "src/parser.go.bak", nil},
		{"file entry as prefix, no dot", "src/parser.gox", nil},
		{"under a file entry", "src/parser.go/x", nil},
		{"dot dot that lands in a sibling", "internal/tui/../tuix/x.go", nil},

		// A parent folder is not a match.
		{"parent folder with slash", "internal/", nil},
		{"parent folder bare", "internal", nil},
		{"parent of a file entry", "src/", nil},
		{"folder named without its slash", "internal/tui", nil},

		// A path outside the repo is not a match, even when a page lists it.
		{"climbs out first", "../internal/tui/x.go", nil},
		{"climbs out later", "internal/tui/../../../etc/passwd", nil},
		{"absolute path", "/internal/tui/x.go", nil},
		{"absolute path a page lists", "/etc/passwd", nil},
		{"climbs out where a page lists it", "../evil/x.go", nil},
		{"dot dot alone", "..", nil},

		// Nothing to match.
		{"empty", "", nil},
		{"only quotes", `""`, nil},
		{"dot", ".", nil},
		{"dot slash", "./", nil},
		{"whole repo pattern", "./...", nil},
		{"unrelated file", "README.md", nil},
		{"unrelated file in a known folder", "src/other.go", nil},
	}
	for _, tt := range tests {
		got := ids(Match(pages, tt.rel))
		if !slices.Equal(got, tt.want) {
			t.Errorf("%s: Match(%q) = %v, want %v", tt.name, tt.rel, got, tt.want)
		}
	}
}

func TestMatchGivesEveryPageThatCovers(t *testing.T) {
	t.Parallel()
	pages := []Page{
		{ID: "wide", Paths: []string{"internal/"}},
		{ID: "narrow", Paths: []string{"internal/tui/", "x.go"}},
		{ID: "elsewhere", Paths: []string{"cmd/"}},
	}
	got := ids(Match(pages, "internal/tui/model.go"))
	if want := []string{"wide", "narrow"}; !slices.Equal(got, want) {
		t.Errorf("Match = %v, want %v, in the order of the pages given", got, want)
	}
}
