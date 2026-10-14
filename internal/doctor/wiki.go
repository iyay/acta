package doctor

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/wiki"
)

// checkWiki says whether the wiki check record is in good shape: the checked
// file exists, holds one line per page and nothing else, and git merges it by
// union so two branches never clash on it. Staleness itself is for acta wiki
// check to say.
func checkWiki(e Env) Result {
	r := Result{Name: "wiki", Level: OK, Msg: "wiki check record is in order"}
	pages, _ := wiki.Load(e.ActaRoot)
	if e.RepoRoot == "" || len(pages) == 0 {
		r.Msg = "no wiki pages"
		return r
	}
	if link := linkedWikiFile(e); link != "" {
		r.Level, r.Msg, r.Fix = Warn, link+" is a link, so acta will not write it", "replace it with a plain file"
		return r
	}
	var msgs []string
	if !slices.Contains(lines(filepath.Join(e.RepoRoot, ".gitattributes")), unionLine(e)) {
		msgs = append(msgs, ".gitattributes lacks "+unionLine(e))
	}
	old, err := os.ReadFile(checkedPath(e))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		msgs = append(msgs, "wiki checks are not in "+wiki.CheckedFile+" yet")
	case err != nil:
		msgs = append(msgs, err.Error())
	case wiki.RewriteChecked(string(old), pageIDs(pages), nil, "") != string(old):
		msgs = append(msgs, wiki.CheckedFile+" has duplicate, unknown or unsorted lines")
	}
	if len(msgs) > 0 {
		r.Level, r.Msg, r.Fix = Warn, strings.Join(msgs, "; "), "acta doctor --fix"
	}
	return r
}

// fixWiki adds the union line and rewrites the checked file. With no checked
// file yet, every page that is fresh by the old trailer rule gets a line, so
// moving to the file turns no fresh page stale. The git history is read before
// any wiki file is written, so a failed read writes neither of them. It gives
// the paths it changed.
func fixWiki(e Env) ([]string, error) {
	pages, _ := wiki.Load(e.ActaRoot)
	if len(pages) == 0 {
		return nil, nil
	}
	attrs := filepath.Join(e.RepoRoot, ".gitattributes")
	// A link carries the write to a file that is not this repo's. Skip only
	// this part, so the rest of --fix still runs; checkWiki says why.
	if linkedWikiFile(e) != "" {
		return nil, nil
	}
	old, err := os.ReadFile(checkedPath(e))
	missing := errors.Is(err, fs.ErrNotExist)
	if err != nil && !missing {
		return nil, err
	}
	var touched []string
	if missing {
		if touched, err = freshByTrailers(e, pages); err != nil {
			return nil, err
		}
	}
	var changed []string
	if !slices.Contains(lines(attrs), unionLine(e)) {
		// Append, so the user's own lines and blank lines stay as they are.
		raw, _ := os.ReadFile(attrs)
		if len(raw) > 0 && raw[len(raw)-1] != '\n' {
			raw = append(raw, '\n')
		}
		if err := os.WriteFile(attrs, append(raw, unionLine(e)+"\n"...), 0o644); err != nil {
			return nil, err
		}
		changed = append(changed, attrs)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	body := wiki.RewriteChecked(string(old), pageIDs(pages), touched, now)
	if missing || body != string(old) {
		if err := os.WriteFile(checkedPath(e), []byte(body), 0o644); err != nil {
			return nil, err
		}
		changed = append(changed, checkedPath(e))
	}
	return changed, nil
}

// freshByTrailers lists the pages that are fresh by the rule acta used before
// the checked file: a check was a page edit or an empty commit with a
// Wiki-Checked trailer. Only the move to the checked file reads it.
func freshByTrailers(e Env, pages []wiki.Page) ([]string, error) {
	raw, err := exec.Command("git", "-C", e.RepoRoot, "log", "--no-color", "--grep=^Wiki-Checked:",
		"--format=%H%x00%(trailers:key=Wiki-Checked,valueonly,separator=%x1f)").Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	out := string(raw)
	byID := map[string][]string{}
	for _, line := range strings.Split(out, "\n") {
		sha, values, _ := strings.Cut(line, "\x00")
		for _, v := range strings.Split(values, "\x1f") {
			if v = strings.TrimSpace(v); v != "" && sha != "" {
				byID[v] = append(byID[v], sha)
			}
		}
	}
	var fresh []string
	for _, p := range pages {
		file, err := filepath.Rel(e.RepoRoot, p.Path)
		if err != nil {
			return nil, err
		}
		checked, _ := filepath.Rel(e.RepoRoot, checkedPath(e))
		checks, err := gitc.CheckCommits(e.RepoRoot, filepath.ToSlash(file), filepath.ToSlash(checked), p.ID)
		if err != nil {
			return nil, err
		}
		// Keep the same paths acta wiki check keeps, so a page is fresh here
		// exactly when it was fresh there.
		var paths []string
		for _, entry := range p.Paths {
			if usable(entry) {
				paths = append(paths, entry)
			}
		}
		sha, _, err := gitc.Unseen(e.RepoRoot, append(checks, byID[p.ID]...), paths)
		if err != nil {
			return nil, err
		}
		if sha == "" {
			fresh = append(fresh, p.ID)
		}
	}
	return fresh, nil
}

// usable says whether a paths entry is a plain path inside the repo, the only
// kind acta wiki check asks git about.
func usable(entry string) bool {
	plain := path.Clean(entry)
	if entry == "" || plain == "." || plain == ".." || strings.HasPrefix(plain, "../") || strings.HasPrefix(plain, "/") {
		return false
	}
	if strings.HasSuffix(entry, "/") {
		plain += "/"
	}
	return plain == entry
}

// linkedWikiFile names .gitattributes or the checked file when either is a
// link or not a plain file, else "".
func linkedWikiFile(e Env) string {
	for _, p := range []string{filepath.Join(e.RepoRoot, ".gitattributes"), checkedPath(e)} {
		if fi, err := os.Lstat(p); err == nil && !fi.Mode().IsRegular() {
			return p
		}
	}
	return ""
}

func checkedPath(e Env) string { return filepath.Join(e.ActaRoot, "wiki", wiki.CheckedFile) }

// unionLine is the .gitattributes line that merges the checked file by union.
func unionLine(e Env) string {
	rel, err := filepath.Rel(e.RepoRoot, checkedPath(e))
	if err != nil {
		rel = checkedPath(e)
	}
	return filepath.ToSlash(rel) + " merge=union"
}

func pageIDs(pages []wiki.Page) []string {
	ids := make([]string, 0, len(pages))
	for _, p := range pages {
		ids = append(ids, p.ID)
	}
	return ids
}

// lines gives the lines of a file with the empty ones left out. A missing
// file has none.
func lines(path string) []string {
	raw, _ := os.ReadFile(path)
	var out []string
	for _, l := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}
