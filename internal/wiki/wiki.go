// Package wiki reads the project wiki under <root>/wiki and finds the pages
// that cover a file.
package wiki

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Page is one wiki page.
type Page struct {
	Path        string // the page file on disk: <root>/wiki/<ID>.md
	ID          string // the page name under wiki/, with slashes and no .md
	Type        string
	Title       string
	Description string
	Paths       []string // files and folders the page covers; a folder ends with /
	Timestamp   time.Time
	Words       int // words in the body, which is the text after the frontmatter
}

// front is the frontmatter as written. The time stays text here so that a
// date with no clock time is refused, not read as midnight.
type front struct {
	Type        string   `yaml:"type"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Paths       []string `yaml:"paths"`
	Timestamp   string   `yaml:"timestamp"`
}

// Load reads every .md file under <root>/wiki, folders inside it too, in the
// order of the walk: by name, one folder after another. A file that cannot be
// read comes back as an error that names it, and the other pages still load.
// A missing wiki folder is no pages and no error.
func Load(root string) ([]Page, []error) {
	dir := filepath.Join(root, "wiki")
	var pages []Page
	var errs []error
	// The callback never stops the walk, so one bad file or folder cannot hide
	// the pages next to it. That is why the walk's own result is not checked.
	_ = fs.WalkDir(os.DirFS(dir), ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			if name != "." || !errors.Is(err, fs.ErrNotExist) {
				errs = append(errs, err)
			}
			return nil
		}
		if d.IsDir() || path.Ext(name) != ".md" {
			return nil
		}
		page, err := read(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		page.ID = strings.TrimSuffix(name, ".md")
		pages = append(pages, page)
		return nil
	})
	return pages, errs
}

// read loads one page file. Every error it gives names the file.
func read(file string) (Page, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return Page{}, err
	}
	head, body, ok := split(string(raw))
	if !ok {
		return Page{}, fmt.Errorf("%s: no frontmatter between two --- lines", file)
	}
	var f front
	if err := yaml.Unmarshal([]byte(head), &f); err != nil {
		return Page{}, fmt.Errorf("%s: %w", file, err)
	}
	page := Page{
		Path:        file,
		Type:        f.Type,
		Title:       f.Title,
		Description: f.Description,
		Paths:       f.Paths,
		Words:       len(strings.Fields(body)),
	}
	if f.Timestamp != "" {
		if page.Timestamp, err = time.Parse(time.RFC3339, f.Timestamp); err != nil {
			return Page{}, fmt.Errorf("%s: timestamp: %w", file, err)
		}
	}
	return page, nil
}

// split cuts a page into the lines between its first two --- lines and the
// text after them, the same way the board splits a planning file. A later ---
// line belongs to the body.
func split(src string) (head, body string, ok bool) {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	if lines[0] != "---" {
		return "", "", false
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return strings.Join(lines[1:i], "\n"), strings.Join(lines[i+1:], "\n"), true
		}
	}
	return "", "", false
}

// Match gives the pages that cover rel, in the order they were given. rel is a
// path from the repo root, the way an agent or a shell command wrote it. A
// folder entry covers what is inside it, and the folder itself with or without
// a slash at the end, since a shell word names a folder both ways, like
// `go vet ./internal/tui`.
func Match(pages []Page, rel string) []Page {
	rel, ok := clean(rel)
	if !ok {
		return nil
	}
	var out []Page
	for _, p := range pages {
		if slices.ContainsFunc(p.Paths, func(entry string) bool { return covers(entry, rel) }) {
			out = append(out, p)
		}
	}
	return out
}

// covers says whether one entry of a page takes in rel: it is the same path,
// it sits inside a folder entry, or it is that folder with its slash left off.
// Only an entry that ends in / is a folder, so src/parse.py does not take in
// src/parse.pyc.
func covers(entry, rel string) bool {
	folder, isFolder := strings.CutSuffix(entry, "/")
	return rel == entry || (isFolder && (rel == folder || strings.HasPrefix(rel, entry)))
}

// clean strips what a shell word adds to a path: quotes, a leading ./, doubled
// slashes and dot-dot steps. A path that leaves the repo, or starts at the disk
// root, is not a repo path, so it gives false and never matches a page. The Go
// pattern dir/... needs no step of its own: to the prefix rule, ... is just a
// name inside dir/.
func clean(rel string) (string, bool) {
	rel = strings.Trim(rel, `"'`)
	folder := strings.HasSuffix(rel, "/")
	rel = path.Clean(rel)
	if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") || strings.HasPrefix(rel, "/") {
		return "", false
	}
	if folder {
		rel += "/"
	}
	return rel, true
}
