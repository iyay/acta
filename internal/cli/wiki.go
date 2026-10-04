package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/wiki"
)

const wikiUsage = "usage: acta wiki ls [--type T] | match <file>... | check [<range>]"

// cmdWiki runs ls, match and check. They only read: no file is written and
// nothing is committed. A repo with no wiki folder has no pages, so all three
// print nothing and exit 0.
func cmdWiki(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, wikiUsage)
		return exitBadInput
	}
	sub := args[0]
	fs, root := flags("wiki "+sub, stderr)
	// Only ls has --type, so the flag parser refuses it for the other two.
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
// takes a file or more, and check takes at most one range. A word that is not a
// command fits nothing.
func wikiFits(sub string, n int) bool {
	switch sub {
	case "ls":
		return n == 0
	case "match":
		return n > 0
	case "check":
		return n <= 1
	}
	return false
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
		lines = append(lines, err.Error())
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

// wikiLine is one page as one line: its file from the repo root, then its
// description.
func wikiLine(cfg config.Config, p wiki.Page) string {
	name := p.Path
	if rel, err := filepath.Rel(cfg.RepoRoot, p.Path); err == nil {
		name = filepath.ToSlash(rel)
	}
	return name + ": " + oneLine(p.Description)
}

// oneLine turns every break and run of spaces into one space, so a page or a
// problem never takes more than a line. A YAML error and a git error can run
// over several.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
