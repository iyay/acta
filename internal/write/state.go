package write

import (
	"os"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// stateCap is the most lines one State subsection holds. The section is a
// note to the next session, not a log, so a longer one wants to be a findings
// file instead.
const stateCap = 10

// stateParts says which "### " heading each part name writes under, in the
// order the subsections sit in the file.
var stateParts = []struct{ part, heading string }{
	{"next", "Next"},
	{"findings", "Findings"},
	{"rulings", "Open rulings"},
}

// SetState writes one part of the State section of a plan and commits the
// file. Only the named subsection changes: every other byte of the file stays
// as it was. A plan with no State section, or a State section with no such
// subsection, gets one in the place the order above fixes. A blank body is
// refused without clear, so a forgotten pipe never wipes a part: pass clear
// to empty the subsection, never together with text on stdin.
func SetState(cfg config.Config, b *board.Board, planID, part string, body []byte, clear bool) (Outcome, error) {
	at := -1
	for i, p := range stateParts {
		if p.part == part {
			at = i
		}
	}
	if at < 0 {
		return Outcome{}, bad("unknown state part %q; use next, findings or rulings", part)
	}
	it := b.Get(planID)
	switch {
	case it == nil:
		return Outcome{}, bad("unknown id %s", planID)
	case it.Kind != board.KindPlan:
		return Outcome{}, bad("%s is not a plan", planID)
	}
	// A body of nothing but blank lines counts as blank too. Without clear a
	// blank body is a forgotten pipe, not an empty part; with clear, text on
	// stdin means the caller asked for two things at once.
	blank := strings.TrimSpace(string(body)) == ""
	switch {
	case blank && !clear:
		return Outcome{}, bad("nothing on stdin; pipe the lines in, or use --clear to empty %s", part)
	case !blank && clear:
		return Outcome{}, bad("text on stdin with --clear; pipe the lines in without it, or use --clear alone to empty %s", part)
	}
	trimmed := strings.Trim(string(body), "\r\n")
	var text []string
	if strings.TrimSpace(trimmed) != "" {
		text = strings.Split(trimmed, "\n")
	}
	if len(text) > stateCap {
		return Outcome{}, bad("a State part holds at most %d lines, got %d", stateCap, len(text))
	}

	dirty, err := dirtyBefore(cfg, it.Path)
	if err != nil {
		return Outcome{}, err
	}
	src, err := os.ReadFile(it.Path)
	if err != nil {
		return Outcome{}, err
	}
	// The frontmatter is a prefix of the file and the body is what comes
	// after it, so the new body written behind the old prefix keeps every
	// frontmatter byte.
	_, rest, nl, _, err := frontOf(src)
	if err != nil {
		return Outcome{}, bad("%s: %v", planID, err)
	}
	out := append(append([]byte{}, src[:len(src)-len(rest)]...), putState(rest, at, text, nl)...)
	if err := os.WriteFile(it.Path, out, 0o644); err != nil {
		return Outcome{}, err
	}
	// A plan with no number id has no short id to name, so the commit says
	// the id it was given.
	name := it.ShortID
	if name == "" {
		name = it.ID
	}
	return finish(cfg, it.Path, Subject(cfg, []string{it.Path}, "state "+name), dirty), nil
}

// putState writes text under the heading the State parts name at index at. A
// missing State section goes at the end of the file, and a missing subsection
// goes in the slot its place in the order fixes. A subsection that is there is
// replaced whole, so the lines around it stay as they were.
func putState(body string, at int, text []string, nl string) string {
	lines := strings.SplitAfter(body, "\n")
	// SplitAfter leaves an empty piece after a final newline. Drop it, or text
	// at the end of the file would look like it already has a blank line above
	// it.
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	sec := -1
	for i, ln := range lines {
		if sectionName(ln) == "State" {
			sec = i
			break
		}
	}
	if sec < 0 {
		out := lines
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, nl)
		}
		out = append(out, "## State"+nl, nl)
		return strings.Join(append(out, stateBlock(at, text, nl, false)...), "")
	}
	// The State section runs to the next "## " heading, or to the end of the
	// file.
	stop := len(lines)
	for i := sec + 1; i < len(lines); i++ {
		if sectionName(lines[i]) != "" {
			stop = i
			break
		}
	}
	sub := -1
	for i := sec + 1; i < stop; i++ {
		if subName(lines[i]) == stateParts[at].heading {
			sub = i
			break
		}
	}
	if sub < 0 {
		// A hand edit took the subsection away, so it goes back before the
		// part that reads after it and not at the end of the section.
		cut := stop
		for i := sec + 1; i < stop; i++ {
			if stateOrder(subName(lines[i])) > at {
				cut = i
				break
			}
		}
		block := stateBlock(at, text, nl, cut < len(lines))
		if cut > 0 && strings.TrimSpace(lines[cut-1]) != "" {
			block = append([]string{nl}, block...)
		}
		return strings.Join(append(lines[:cut:cut], append(block, lines[cut:]...)...), "")
	}
	// The subsection runs to the next heading of any level, or to the end of
	// the section, and that whole run is what the new body replaces.
	end := stop
	for i := sub + 1; i < stop; i++ {
		if sectionName(lines[i]) != "" || subName(lines[i]) != "" {
			end = i
			break
		}
	}
	block := stateBlock(at, text, nl, end < len(lines))
	return strings.Join(append(lines[:sub:sub], append(block, lines[end:]...)...), "")
}

// stateBlock gives the lines a subsection is written as: its heading, then a
// blank line, then the body. An empty body leaves the heading alone. A
// subsection with something after it gets a blank line behind it, so the next
// heading is not glued to it.
func stateBlock(at int, text []string, nl string, more bool) []string {
	out := []string{"### " + stateParts[at].heading + nl}
	if len(text) > 0 {
		out = append(out, nl)
		for _, ln := range text {
			out = append(out, strings.TrimSuffix(ln, "\r")+nl)
		}
	}
	if more {
		out = append(out, nl)
	}
	return out
}

// stateOrder gives where a subsection sits in the order the parts fix, or -1
// for a heading nobody wrote by hand. Such a heading is left where the person
// put it, so a new subsection never moves it.
func stateOrder(heading string) int {
	for i, p := range stateParts {
		if p.heading == heading {
			return i
		}
	}
	return -1
}

// subName gives the part a "### " line names, or "" for any other line.
func subName(ln string) string {
	s := strings.TrimRight(ln, " \t\r\n")
	if !strings.HasPrefix(s, "### ") {
		return ""
	}
	return strings.TrimPrefix(s, "### ")
}
