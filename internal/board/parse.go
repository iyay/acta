// Package board reads the planning files of a repo into one Board.
package board

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	taskRe = regexp.MustCompile(`^### Task ([^:\s]*[^:\s.])(?:[.:])?\s*(.*)$`)
	boxRe  = regexp.MustCompile(`^\s*[-*] \[([ xX])\]`)
	specRe = regexp.MustCompile(`^\*\*Spec:\*\*\s*(.+)$`)
	itemRe = regexp.MustCompile(`^\s*[-*] \[([ xX-])\] (.*)$`)
)

// ItemLine is one checklist line of a debt file, outside any task section.
type ItemLine struct {
	Num   int
	Text  string
	Line  int  // 1-based line in the file
	State byte // ' ' open, 'x' done, '-' wontfix
}

// TaskSec is one "### Task N" section of a plan.
type TaskSec struct {
	Num   string
	Title string
	Line  int // 1-based line of the heading in the file
	Done  int
	Total int
	Body  string // the section, heading included
}

// Doc is what one markdown file says about itself.
type Doc struct {
	Front    map[string]any
	FrontErr error
	HasFront bool
	Title    string
	SpecPath string
	Tasks    []TaskSec
	Items    []ItemLine // debt checklist lines outside any task section
	Body     string     // the file without its frontmatter block
}

// Parse reads frontmatter, the first "# " title, the "**Spec:**" line and the
// task sections. It never fails: a broken frontmatter lands in FrontErr.
func Parse(src []byte) Doc {
	lines := strings.Split(strings.ReplaceAll(string(src), "\r\n", "\n"), "\n")
	var d Doc
	start := 0
	if len(lines) > 0 && lines[0] == "---" {
		for i := 1; i < len(lines); i++ {
			if lines[i] != "---" {
				continue
			}
			d.HasFront = true
			var m map[string]any
			if err := yaml.Unmarshal([]byte(strings.Join(lines[1:i], "\n")), &m); err != nil {
				d.FrontErr = err
			} else {
				d.Front = m
			}
			start = i + 1
			break
		}
	}
	d.Body = strings.Join(lines[start:], "\n")

	var cur *TaskSec
	var curLines []string
	flush := func() {
		if cur != nil {
			cur.Body = strings.Join(curLines, "\n")
			d.Tasks = append(d.Tasks, *cur)
			cur, curLines = nil, nil
		}
	}
	inFence := false
	for i := start; i < len(lines); i++ {
		ln := lines[i]
		fence := strings.HasPrefix(strings.TrimSpace(ln), "```")
		if fence {
			inFence = !inFence
		}
		// Headings and boxes inside a code block are examples, not real ones.
		if !inFence && !fence {
			if d.Title == "" && strings.HasPrefix(ln, "# ") {
				d.Title = strings.TrimSpace(ln[2:])
			}
			if d.SpecPath == "" {
				if m := specRe.FindStringSubmatch(ln); m != nil {
					d.SpecPath = specPath(m[1])
				}
			}
			if m := taskRe.FindStringSubmatch(ln); m != nil {
				flush()
				cur = &TaskSec{Num: m[1], Title: strings.TrimSpace(m[2]), Line: i + 1}
				curLines = []string{ln}
				continue
			}
			if cur != nil && (strings.HasPrefix(ln, "### ") || strings.HasPrefix(ln, "## ")) {
				flush()
			}
			if cur != nil {
				if m := boxRe.FindStringSubmatch(ln); m != nil {
					cur.Total++
					if m[1] != " " {
						cur.Done++
					}
				}
			} else if m := itemRe.FindStringSubmatch(ln); m != nil {
				state := m[1][0]
				if state == 'X' {
					state = 'x'
				}
				d.Items = append(d.Items, ItemLine{Num: len(d.Items) + 1, Text: m[2], Line: i + 1, State: state})
			}
		}
		if cur != nil {
			curLines = append(curLines, ln)
		}
	}
	flush()
	return d
}

// specPath takes the spec path out of the text after "**Spec:**". Only a
// path ending in .md counts, so a plan with no spec ("none", a note in
// backticks) shows no false warning. A path in backticks wins over a bare
// word.
func specPath(s string) string {
	for rest := s; ; {
		i := strings.Index(rest, "`")
		if i < 0 {
			break
		}
		j := strings.Index(rest[i+1:], "`")
		if j < 0 {
			break
		}
		if p := rest[i+1 : i+1+j]; strings.HasSuffix(p, ".md") {
			return p
		}
		rest = rest[i+2+j:]
	}
	for _, f := range strings.Fields(s) {
		// Trim brackets at the start and punctuation at the end, but keep a
		// leading dot so ".pm/specs/x.md" stays whole.
		// A bare word must look like a path, so a file named in a sentence (CLAUDE.md) is not taken as the spec.
		if p := strings.TrimRight(strings.TrimLeft(f, "(["), ")],.:"); strings.HasSuffix(p, ".md") && strings.Contains(p, "/") {
			return p
		}
	}
	return ""
}
