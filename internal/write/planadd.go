package write

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// fixRoundRe finds the number of a fix round heading, so the next round gets
// the number after the highest one.
var fixRoundRe = regexp.MustCompile(`^## Fix round (\d+)\s*$`)

// polishVerify is the verify line of every polish task. The review skill
// reads the task by it, so it never changes.
const polishVerify = "every NOTE below is applied, and nothing else changes."

// AddPlanSection adds a fix round, the polish task or review notes to a plan
// and commits the file, so an agent never reads the whole plan to add one.
// Each non-blank line of body is one item. Every byte already in the file
// stays as it was. It gives the line the command prints: the new task's id,
// or the plan id for review notes.
func AddPlanSection(cfg config.Config, b *board.Board, planID, section, title, verify string, body []byte) (Outcome, string, error) {
	if section != "fix-round" && section != "polish" && section != "review-notes" {
		return Outcome{}, "", bad("unknown section %q; use fix-round, polish or review-notes", section)
	}
	it := b.Get(planID)
	switch {
	case it == nil:
		return Outcome{}, "", bad("unknown id %s", planID)
	case it.Kind != board.KindPlan:
		return Outcome{}, "", bad("%s is not a plan", planID)
	}
	if section == "fix-round" {
		if strings.TrimSpace(title) == "" || strings.TrimSpace(verify) == "" {
			return Outcome{}, "", bad("fix-round needs --title and --verify")
		}
		if strings.ContainsAny(title+verify, "\r\n") {
			return Outcome{}, "", bad("--title and --verify must be one line each")
		}
	}
	items := addItems(body)
	if len(items) == 0 {
		return Outcome{}, "", bad("nothing on stdin; pipe in one line per item")
	}

	dirty, err := dirtyBefore(cfg, it.Path)
	if err != nil {
		return Outcome{}, "", err
	}
	src, err := os.ReadFile(it.Path)
	if err != nil {
		return Outcome{}, "", err
	}
	_, rest, nl, _, err := frontOf(src)
	if err != nil {
		return Outcome{}, "", bad("%s: %v", planID, err)
	}
	// A plan with no number id has no short id to name, so the commit and
	// the polish box say the id it was given.
	name := it.ShortID
	if name == "" {
		name = it.ID
	}
	lines := bodyLines(rest, nl)
	m := mapPlan(lines)

	var out []string
	printed := name
	switch section {
	case "review-notes":
		var notes []string
		for _, s := range items {
			notes = append(notes, "- "+s+nl)
		}
		if m.notes < 0 {
			out = putBlock(lines, m.state, append([]string{"## Review notes" + nl, nl}, notes...), nl)
		} else {
			out = addNotes(lines, m.notes, notes, nl)
		}
	default:
		n := strconv.Itoa(highestTask(src) + 1)
		var head []string
		var last string
		if section == "polish" {
			if m.polish {
				return Outcome{}, "", bad("%s already has ## Polish; one polish task holds every NOTE", planID)
			}
			head = []string{"## Polish" + nl, nl, "### Task " + n + ": Review polish" + nl, nl, "**verify:** " + polishVerify + nl, nl}
			last = "Commit: `polish: review notes for " + name + "`"
		} else {
			round := strconv.Itoa(m.rounds + 1)
			head = []string{"## Fix round " + round + nl, nl, "### Task " + n + ": " + strings.TrimSpace(title) + nl, nl, "**verify:** " + strings.TrimSpace(verify) + nl, nl}
			hash := it.Hash
			if hash == "" {
				hash = name
			}
			last = "Commit with the trailer `Task: " + hash + "#" + n + "`"
		}
		for _, s := range append(items, last) {
			head = append(head, "- [ ] "+s+nl)
		}
		out = putBlock(lines, m.tail, head, nl)
		printed = it.ID + "#task-" + n + " added"
	}

	full := append(append([]byte{}, src[:len(src)-len(rest)]...), strings.Join(out, "")...)
	if err := os.WriteFile(it.Path, full, 0o644); err != nil {
		return Outcome{}, "", err
	}
	return finish(cfg, it.Path, Subject(cfg, []string{it.Path}, section+" "+name), dirty), printed, nil
}

// addItems gives one item per non-blank line. A leading "- [ ] " or "- " is
// dropped, so a line pasted from a list does not get its box twice.
func addItems(body []byte) []string {
	var out []string
	for _, ln := range strings.Split(string(body), "\n") {
		s := strings.TrimSpace(ln)
		s = strings.TrimPrefix(s, "- [ ] ")
		s = strings.TrimPrefix(s, "- ")
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// bodyLines splits the body into lines that keep their line ending. A last
// line with no ending gets one, so text added after it starts on its own line.
func bodyLines(body, nl string) []string {
	lines := strings.SplitAfter(body, "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	if n := len(lines); n > 0 && !strings.HasSuffix(lines[n-1], "\n") {
		lines[n-1] += nl
	}
	return lines
}

// planMap says where the sections that matter sit in a plan body. Headings
// inside a code block are examples, so they are not counted.
type planMap struct {
	tail   int  // first "## Review notes" or "## State", or the line count
	state  int  // "## State", or the line count
	notes  int  // "## Review notes", or -1
	rounds int  // highest fix round number
	polish bool // a "## Polish" section is there
}

func mapPlan(lines []string) planMap {
	m := planMap{tail: len(lines), state: len(lines), notes: -1}
	inFence := false
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		switch sectionName(ln) {
		case "Review notes":
			if m.notes < 0 {
				m.notes = i
			}
			m.tail = min(m.tail, i)
		case "State":
			m.state = min(m.state, i)
			m.tail = min(m.tail, i)
		case "Polish":
			m.polish = true
		}
		if r := fixRoundRe.FindStringSubmatch(strings.TrimRight(ln, "\r\n")); r != nil {
			if n, err := strconv.Atoi(r[1]); err == nil && n > m.rounds {
				m.rounds = n
			}
		}
	}
	return m
}

// highestTask gives the highest "### Task N" number of the plan, read the way
// the board reads it, so the new task never takes a number already used.
func highestTask(src []byte) int {
	top := 0
	for _, t := range board.Parse(src).Tasks {
		if n, err := strconv.Atoi(t.Num); err == nil && n > top {
			top = n
		}
	}
	return top
}

// putBlock puts a new section in front of line cut, with one blank line on
// each side of it.
func putBlock(lines []string, cut int, block []string, nl string) []string {
	if cut > 0 && strings.TrimSpace(lines[cut-1]) != "" {
		block = append([]string{nl}, block...)
	}
	if cut < len(lines) {
		block = append(block, nl)
	}
	return append(lines[:cut:cut], append(block, lines[cut:]...)...)
}

// addNotes puts note lines after the last line of the Review notes section
// that starts at line at. Notes already there stay as they were.
func addNotes(lines []string, at int, notes []string, nl string) []string {
	end := len(lines)
	for i := at + 1; i < len(lines); i++ {
		if sectionName(lines[i]) != "" {
			end = i
			break
		}
	}
	last := at
	for i := at + 1; i < end; i++ {
		if strings.TrimSpace(lines[i]) != "" {
			last = i
		}
	}
	if last == at {
		// The section has only its heading, so the notes go one blank line
		// under it, and the blank lines that were there go before the next
		// section.
		notes = append([]string{nl}, notes...)
	}
	cut := last + 1
	return append(lines[:cut:cut], append(notes, lines[cut:]...)...)
}
