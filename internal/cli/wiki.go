package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/gitc"
	"github.com/iyay/acta/internal/wiki"
)

var wikiUsage = usageLine("wiki")

// cmdWiki runs ls, match, check and ok. The first three only read: no file is
// written and nothing is committed. ok writes only the checked file and
// commits it, to record a re-check. A repo with no wiki folder has no pages,
// so ls, match and check print nothing and exit 0.
func cmdWiki(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, wikiUsage)
		return exitBadInput
	}
	sub := args[0]
	fs, root := flags("wiki "+sub, stderr)
	// Only ls has --type, so the flag parser refuses it for match, check and ok.
	kind := ""
	if sub == "ls" {
		fs.StringVar(&kind, "type", "", "list only the pages of this type")
	}
	pos, err := parseMixed(fs, args[1:])
	if err != nil || !wikiFits(sub, len(pos)) {
		fmt.Fprintln(stderr, wikiUsage)
		return exitBadInput
	}
	cfg, code := loadConfig(*root, stderr)
	if code != exitOK {
		return code
	}
	pages, loadErrs := wiki.Load(cfg.Root)
	switch sub {
	case "ls":
		for _, p := range pages {
			if kind == "" || p.Type == kind {
				fmt.Fprintln(stdout, wikiLine(cfg, p))
			}
		}
	case "match":
		wikiMatch(stdout, cfg, pages, pos)
	case "ok":
		return wikiOk(stdout, stderr, cfg, pages, pos)
	default:
		rng := ""
		if len(pos) == 1 {
			rng = pos[0]
		}
		return wikiCheck(stdout, cfg, pages, loadErrs, rng)
	}
	return exitOK
}

// wikiFits says whether n words after the command suit it: ls takes none, match
// takes a file or more, check takes at most one range, and ok takes a page id or
// more. A word that is not a command fits nothing.
func wikiFits(sub string, n int) bool {
	switch sub {
	case "ls":
		return n == 0
	case "match", "ok":
		return n > 0
	case "check":
		return n <= 1
	}
	return false
}

// wikiOk records that the pages were re-checked and nothing was wrong. The note
// is a new line for each page in the checked file, not an edit of the page.
// Git merges that file by union, so two branches that both re-check a page
// still merge without a conflict. Every id must name a page that loaded, or
// nothing is written. A repeated id counts once.
func wikiOk(stdout, stderr io.Writer, cfg config.Config, pages []wiki.Page, ids []string) int {
	known := map[string]bool{}
	for _, p := range pages {
		known[p.ID] = true
	}
	var unique, unknown []string
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
		if !known[id] {
			unknown = append(unknown, id)
		}
	}
	if len(unknown) > 0 {
		fmt.Fprintf(stderr, "unknown wiki page: %s\n", strings.Join(unknown, ", "))
		return exitBadInput
	}
	all := make([]string, 0, len(pages))
	for _, p := range pages {
		all = append(all, p.ID)
	}
	dir := filepath.Join(cfg.Root, "wiki")
	dirty, err := gitc.IsDirty(cfg.RepoRoot, filepath.Join(dir, wiki.CheckedFile))
	if err != nil {
		fmt.Fprintln(stderr, "acta wiki ok: "+oneLine(err.Error()))
		return exitOther
	}
	// The time down to the second, so a second check on the same day still
	// changes the line and makes a commit of its own.
	path, err := wiki.MarkChecked(dir, all, unique, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		fmt.Fprintln(stderr, "acta wiki ok: "+oneLine(err.Error()))
		return exitOther
	}
	r := gitc.CommitOrFoldIn(cfg.RepoRoot, cfg.Root, path, "chore(wiki): re-check "+strings.Join(unique, ", "), dirty)
	if !r.Committed {
		fmt.Fprintln(stderr, "acta wiki ok: written, not committed: "+oneLine(r.Reason))
		return exitOther
	}
	sha, err := gitc.ShortHead(cfg.RepoRoot)
	if err != nil {
		fmt.Fprintln(stderr, "acta wiki ok: "+oneLine(err.Error()))
		return exitOther
	}
	fmt.Fprintln(stdout, sha)
	return exitOK
}

// wikiMatch prints each page that covers one of the files, once. It goes in
// page order, not the order of the files, so the answer is the same however the
// files were listed.
func wikiMatch(stdout io.Writer, cfg config.Config, pages []wiki.Page, files []string) {
	hit := map[string]bool{}
	for _, file := range files {
		for _, p := range wiki.Match(pages, file) {
			hit[p.Path] = true
		}
	}
	for _, p := range pages {
		if hit[p.Path] {
			fmt.Fprintln(stdout, wikiLine(cfg, p))
		}
	}
}

// wikiCheck prints what is wrong, one line each: first the pages that cannot be
// loaded, then the problems of the pages that can. With a range, only the pages
// it touches are checked. Any line means exit 1, which is also the code a bad
// argument gets.
func wikiCheck(stdout io.Writer, cfg config.Config, pages []wiki.Page, loadErrs []error, rng string) int {
	var lines []string
	for _, err := range loadErrs {
		lines = append(lines, loadLine(cfg, err))
	}
	for _, p := range wiki.Check(cfg.RepoRoot, pages, rng) {
		// A problem with no page, like a range git cannot read, stands alone.
		if p.Page == "" {
			lines = append(lines, p.Msg)
		} else {
			lines = append(lines, p.Page+": "+p.Msg)
		}
	}
	for _, line := range lines {
		fmt.Fprintln(stdout, oneLine(line))
	}
	if len(lines) > 0 {
		return exitBadInput
	}
	return exitOK
}

// loadLine is one load error as one line: the file from the repo root, like a
// problem names its page, then the reason. An error that names no file stays as
// it is.
func loadLine(cfg config.Config, err error) string {
	var le *wiki.LoadError
	if !errors.As(err, &le) {
		return err.Error()
	}
	return repoPath(cfg, le.File) + ": " + le.Err.Error()
}

// wikiLine is one page as one line: its file from the repo root, then its
// description.
func wikiLine(cfg config.Config, p wiki.Page) string {
	return repoPath(cfg, p.Path) + ": " + oneLine(p.Description)
}

// repoPath is a file from the repo root, with forward slashes. A file outside
// the repo comes out as a path that climbs out with ../.
func repoPath(cfg config.Config, file string) string {
	if rel, err := filepath.Rel(cfg.RepoRoot, file); err == nil {
		return filepath.ToSlash(rel)
	}
	return file
}

// oneLine turns every break and run of spaces into one space, so a page or a
// problem never takes more than a line. A YAML error and a git error can run
// over several.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
