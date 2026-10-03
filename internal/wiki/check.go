package wiki

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/iyay/acta/internal/gitc"
)

// The limits that keep a hint to one short line and a page cheap to read.
const (
	maxDescription = 120 // characters
	maxWords       = 250 // words in the body
)

// pageTypes are the five kinds of page there are.
var pageTypes = []string{"Decision", "Gotcha", "Runbook", "Reference", "Glossary"}

// Problem is one thing wrong with a page. Page is the page file from the repo
// root, with forward slashes, so an agent can open it from its checkout. It is
// empty when the problem is not about one page, like a range git cannot read.
type Problem struct {
	Page, Msg string
}

// Check says what is wrong with pages. repo is the checkout their paths are
// from. A page that keeps every rule gets no problem.
//
// With rng, a git range like main..HEAD, only the pages the range touches are
// checked. A page is touched when a file that changed in the range is covered
// by one of its paths, or when the page file itself changed, so a page the
// branch wrote always gets its checks. A range git cannot read is a problem of
// its own, so a bad range never passes as a clean check.
//
// The problems of one page come together: its fields first, then its paths,
// then whether it is stale.
func Check(repo string, pages []Page, rng string) []Problem {
	// With no pages there is nothing to check, so the range is not even read.
	if len(pages) == 0 {
		return nil
	}
	if rng != "" {
		changed, err := gitc.Changed(repo, rng)
		if err != nil {
			return []Problem{{Msg: fmt.Sprintf("cannot read the range %q: %v", rng, err)}}
		}
		pages = touched(repo, pages, changed)
	}
	var out []Problem
	for _, p := range pages {
		name := pageName(repo, p)
		for _, msg := range checkPage(repo, p) {
			out = append(out, Problem{Page: name, Msg: msg})
		}
	}
	return out
}

// pageName is the page file from the repo root. A file outside the repo comes
// out as a path that climbs out with ../.
func pageName(repo string, p Page) string {
	if rel, err := filepath.Rel(repo, p.Path); err == nil {
		return filepath.ToSlash(rel)
	}
	return p.Path
}

// touched keeps the pages that the changed files reach, in the order given. A
// page is reached when its own file changed, or when a changed file is covered
// by one of its paths.
func touched(repo string, pages []Page, changed []string) []Page {
	var out []Page
	for _, p := range pages {
		if slices.Contains(changed, pageName(repo, p)) || covered(p, changed) {
			out = append(out, p)
		}
	}
	return out
}

// covered says whether a changed file sits under a path of the page. It asks
// the same question as Match, so an entry counts only as it is written.
func covered(p Page, changed []string) bool {
	return slices.ContainsFunc(changed, func(file string) bool {
		return slices.ContainsFunc(p.Paths, func(entry string) bool { return covers(entry, file) })
	})
}

// checkPage gives the message for each rule a page breaks.
func checkPage(repo string, p Page) []string {
	msgs := fieldMsgs(p)
	var usable []string
	for _, entry := range p.Paths {
		msg, ok := entryMsg(repo, entry)
		if msg != "" {
			msgs = append(msgs, msg)
		}
		if ok {
			usable = append(usable, entry)
		}
	}
	// A page with no timestamp is stale against every commit, which says
	// nothing more than the missing field already does.
	if p.Timestamp.IsZero() || len(usable) == 0 {
		return msgs
	}
	// Only entries that point into the repo go to git: an entry that climbs out
	// of it would make git fail, and already has its own message.
	last, err := gitc.LastChange(repo, usable)
	switch {
	case err != nil:
		msgs = append(msgs, fmt.Sprintf("cannot read the git history: %v", err))
	case last.After(p.Timestamp):
		msgs = append(msgs, fmt.Sprintf("stale: its paths changed on %s, after its timestamp %s; re-check the page and bump the timestamp",
			last.UTC().Format(time.RFC3339), p.Timestamp.UTC().Format(time.RFC3339)))
	}
	return msgs
}

// fieldMsgs checks the fields that sit in the page itself, with no look at disk or git.
func fieldMsgs(p Page) []string {
	var msgs []string
	switch {
	case p.Type == "":
		msgs = append(msgs, "type is missing")
	case !slices.Contains(pageTypes, p.Type):
		msgs = append(msgs, fmt.Sprintf("type %q is not one of %s", p.Type, strings.Join(pageTypes, ", ")))
	}
	if strings.TrimSpace(p.Title) == "" {
		msgs = append(msgs, "title is missing")
	}
	msgs = append(msgs, descriptionMsgs(p.Description)...)
	if p.Timestamp.IsZero() {
		msgs = append(msgs, "timestamp is missing")
	}
	if p.Words > maxWords {
		msgs = append(msgs, fmt.Sprintf("body is %d words, the limit is %d", p.Words, maxWords))
	}
	return msgs
}

// descriptionMsgs checks the description, which is the hint text. Load keeps it
// raw, so a line break in it shows up here. The length counts characters, not bytes.
func descriptionMsgs(d string) []string {
	if strings.TrimSpace(d) == "" {
		return []string{"description is missing"}
	}
	var msgs []string
	if strings.ContainsAny(d, "\r\n") {
		msgs = append(msgs, "description is more than one line")
	}
	if n := utf8.RuneCountInString(d); n > maxDescription {
		msgs = append(msgs, fmt.Sprintf("description is %d characters, the limit is %d", n, maxDescription))
	}
	return msgs
}

// entryMsg checks one entry of paths. It gives the message for what is wrong
// with the entry, or an empty one. ok says whether the entry is a plain path
// inside the repo, which is the only kind git can be asked about.
//
// Match compares an entry as written. An entry that is not plain, or that ends
// the wrong way for what it is, would never match a file, so its page would
// never be shown. That is why each of those is a problem.
func entryMsg(repo, entry string) (msg string, ok bool) {
	if entry == "" {
		return `paths entry "" is empty`, false
	}
	plain := path.Clean(entry)
	if plain == "." || plain == ".." || strings.HasPrefix(plain, "../") || strings.HasPrefix(plain, "/") {
		return fmt.Sprintf("paths entry %q is not inside the repo", entry), false
	}
	folder := strings.HasSuffix(entry, "/")
	if folder {
		plain += "/"
	}
	if plain != entry {
		return fmt.Sprintf("paths entry %q is not a plain path, write it as %q", entry, plain), false
	}
	info, err := os.Stat(filepath.Join(repo, filepath.FromSlash(plain)))
	switch {
	case err != nil:
		return fmt.Sprintf("paths entry %q does not exist", entry), true
	case folder && !info.IsDir():
		return fmt.Sprintf("paths entry %q is a file, write it as %q", entry, strings.TrimSuffix(entry, "/")), true
	case !folder && info.IsDir():
		return fmt.Sprintf("paths entry %q is a folder, write it as %q", entry, entry+"/"), true
	}
	return "", true
}
