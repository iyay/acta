// Package board reads the planning files of a repo into one Board.
package board

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	taskRe = regexp.MustCompile(`^### Task ([A-Za-z0-9]+)\b[:.]?\s*(.*)$`)
	boxRe  = regexp.MustCompile(`^\s*[-*] \[([ xX])\]`)
	specRe = regexp.MustCompile(`^\*\*Spec:\*\*\s*(.+)$`)
)

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
	Body     string // the file without its frontmatter block
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
			}
		}
		if cur != nil {
			curLines = append(curLines, ln)
		}
	}
	flush()
	return d
}

// specPath takes the path out of the text after "**Spec:**". A path in
// backticks wins; otherwise the first word is used.
func specPath(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "`"); i >= 0 {
		if j := strings.Index(s[i+1:], "`"); j >= 0 {
			return s[i+1 : i+1+j]
		}
	}
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	// Trim brackets at the start and punctuation at the end, but keep a
	// leading dot so ".pm/specs/x.md" stays whole.
	return strings.TrimRight(strings.TrimLeft(f[0], "(["), ")],.")
}
