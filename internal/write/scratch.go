package write

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// NewScratch writes a scratch item from the body an agent sent and commits
// it. The words go under a "## Words" heading, one empty section for each
// other part of the schema, so a reader always knows where to look. An
// empty body is refused before anything touches the disk.
func NewScratch(cfg config.Config, slug, title string, body []byte) (Outcome, error) {
	if strings.TrimSpace(string(body)) == "" {
		return Outcome{}, bad("scratch body is empty")
	}
	path, err := datedPath(cfg, cfg.Dirs.Scratch, slug)
	if err != nil {
		return Outcome{}, err
	}
	if title == "" {
		title = slug
	}
	b, err := board.Load(cfg)
	if err != nil {
		return Outcome{}, err
	}
	next, taken := scanIDs(b)
	prefix := board.Prefix(board.KindScratch, false)
	shortID := board.FormatID(prefix, next[prefix])
	// The frontmatter is built on its own and the body is added afterwards, so
	// a body that opens with a "---" rule cannot be read as frontmatter.
	var content []byte
	for _, f := range []struct{ key, value string }{
		{"id", shortID},
		{"hash", freeHash(taken)},
		{"title", title},
		{"status", "raw"},
		{"created", Now().Format(stampLayout)},
		{"schema", "1"},
	} {
		if content, err = SetField(content, f.key, f.value); err != nil {
			return Outcome{}, err
		}
	}
	day := Now().Format("2006-01-02")
	words := strings.TrimRight(string(body), "\n")
	content = append(content, fmt.Sprintf("# %s\n\n## Words\n\n### %s\n\n%s\n\n## Context\n\n## Log\n\n## Open questions\n", title, day, words)...)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Outcome{}, err
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return Outcome{}, err
	}
	o := finish(cfg, path, "acta: new scratch "+stem(path), false)
	o.ShortID = shortID
	return o, nil
}

// sections says which "## " heading a --section value fills. Words and log
// get a dated heading, so a reader sees when each entry came in.
var sections = map[string]struct {
	heading string
	dated   bool
}{
	"words":     {"Words", true},
	"context":   {"Context", false},
	"log":       {"Log", true},
	"questions": {"Open questions", false},
}

// AppendScratch puts text in one part of a scratch body, one blank line
// below what that part already holds. An old item has no parts, so any
// section lands as the plain append at the end of the file, and the text
// is never lost. Any status takes more text: an idea that was specced or
// dropped can still collect an answer.
func AppendScratch(cfg config.Config, b *board.Board, id, section string, text []byte) (Outcome, error) {
	it := b.Get(id)
	switch {
	case it == nil:
		return Outcome{}, bad("unknown id %s", id)
	case it.Kind != board.KindScratch:
		return Outcome{}, bad("%s is not a scratch item", id)
	}
	if strings.TrimSpace(string(text)) == "" {
		return Outcome{}, bad("scratch body is empty")
	}
	sec, known := sections[section]
	if section != "" && !known {
		return Outcome{}, bad("unknown section %q; use words, context, log or questions", section)
	}
	dirty, err := dirtyBefore(cfg, it.Path)
	if err != nil {
		return Outcome{}, err
	}
	src, err := os.ReadFile(it.Path)
	if err != nil {
		return Outcome{}, err
	}
	out := string(src)
	if !board.HasSchema(board.Parse(src).Front) {
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += "\n" + string(text)
	} else {
		if section == "" {
			sec = sections["words"]
		}
		_, body, _, _, err := frontOf(src)
		if err != nil {
			return Outcome{}, err
		}
		newBody := putInSection(body, sec.heading, sec.dated, text)
		if p := board.CheckBody(board.KindScratch, newBody); p != nil {
			return Outcome{}, bad("scratch %s: %s", filepath.Base(it.Path), p[0])
		}
		out = strings.TrimSuffix(out, body) + newBody
	}
	if err := os.WriteFile(it.Path, []byte(out), 0o644); err != nil {
		return Outcome{}, err
	}
	o := finish(cfg, it.Path, "acta: add to scratch "+stem(it.Path), dirty)
	o.ShortID = it.ShortID
	return o, nil
}

// putInSection adds text at the end of the "## heading" part of body. A
// heading that is not there any more is written where the schema wants it,
// so the parts stay in order whichever one the text is for.
func putInSection(body, heading string, dated bool, text []byte) string {
	lines := strings.SplitAfter(body, "\n")
	// SplitAfter leaves an empty piece after a final newline. Drop it, or text
	// at the end of the file would look like it already has a blank line
	// above it.
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	ins := strings.SplitAfter(strings.TrimRight(string(text), "\n")+"\n", "\n")
	if dated {
		ins = append([]string{"### " + Now().Format("2006-01-02") + "\n", "\n"}, ins...)
	}
	want := slices.Index(board.SectionNames(board.KindScratch), heading)
	at := -1
	for i, ln := range lines {
		if sectionName(ln) != heading {
			continue
		}
		// The text goes before the next "## " line, or at the end of the file.
		at = len(lines)
		for j := i + 1; j < len(lines); j++ {
			if sectionName(lines[j]) != "" {
				at = j
				break
			}
		}
		break
	}
	if at < 0 {
		ins = append([]string{"## " + heading + "\n", "\n"}, ins...)
		at = len(lines)
		for i, ln := range lines {
			if slices.Index(board.SectionNames(board.KindScratch), sectionName(ln)) > want {
				at = i
				break
			}
		}
	}
	// One blank line on each side keeps the text off the lines around it.
	if at > 0 && strings.TrimSpace(lines[at-1]) != "" {
		ins = append([]string{"\n"}, ins...)
	}
	if at < len(lines) {
		ins = append(ins, "\n")
	}
	lines = append(lines[:at], append(ins, lines[at:]...)...)
	return strings.Join(lines, "")
}

// sectionName gives the part a "## " line names, or "" for any other line.
func sectionName(ln string) string {
	s := strings.TrimRight(ln, " \t\r\n")
	if !strings.HasPrefix(s, "## ") {
		return ""
	}
	return strings.TrimPrefix(s, "## ")
}

// stem is the file name without ".md": the way a commit message names it.
func stem(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".md")
}
