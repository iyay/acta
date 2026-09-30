package write

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSetValueTitle runs `acta set <id> title` over every path the command
// can take: a body heading, a frontmatter title field, no title at all, and
// the two texts it must refuse. Every accepted case checks the whole file
// byte for byte, so nothing but the title can move, and every refused case
// checks the file and HEAD are exactly as they were.
func TestSetValueTitle(t *testing.T) {
	cases := []struct {
		name    string
		file    string
		id      string // the id to use, when it is not the path without .md
		src     string
		value   string
		want    string // file after the set
		badWant bool   // refused as bad input, file untouched
	}{
		{
			name:  "heading rewritten",
			file:  ".acta/bugs/2026-09-24-crash.md",
			src:   "---\nref: B-1\nstatus: open\n---\n# Crash\n\n## Symptom\nIt crashes.\n",
			value: "Crash on save",
			want:  "---\nref: B-1\nstatus: open\n---\n# Crash on save\n\n## Symptom\nIt crashes.\n",
		},
		{
			name:  "frontmatter title rewritten when there is no heading",
			file:  ".acta/scratch/2026-09-24-idea.md",
			src:   "---\nid: SCRATCH-1\ntitle: Old idea\nstatus: raw\n---\ncatet dulu.\n",
			value: "New idea",
			want:  "---\nid: SCRATCH-1\ntitle: New idea\nstatus: raw\n---\ncatet dulu.\n",
		},
		{
			name:  "heading added when there is neither",
			file:  ".acta/scratch/2026-09-24-mood.md",
			src:   "---\nid: SCRATCH-2\nstatus: raw\n---\ncatet dulu.\n",
			value: "Beli bensin",
			want:  "---\nid: SCRATCH-2\nstatus: raw\n---\n# Beli bensin\ncatet dulu.\n",
		},
		{
			name:    "empty title refused",
			file:    ".acta/bugs/2026-09-24-crash.md",
			src:     "---\nref: B-1\nstatus: open\n---\n# Crash\n\n## Symptom\nIt crashes.\n",
			value:   "",
			badWant: true,
		},
		{
			name:    "blank title refused",
			file:    ".acta/bugs/2026-09-24-crash.md",
			src:     "---\nref: B-1\nstatus: open\n---\n# Crash\n\n## Symptom\nIt crashes.\n",
			value:   "   \t ",
			badWant: true,
		},
		{
			name:    "title with a line break refused",
			file:    ".acta/bugs/2026-09-24-crash.md",
			src:     "---\nref: B-1\nstatus: open\n---\n# Crash\n\n## Symptom\nIt crashes.\n",
			value:   "One\nTwo",
			badWant: true,
		},
		// A debt item is a line inside a debt file, so it has no heading of
		// its own to rewrite. Naming one must not touch the file that holds it.
		{
			name:    "debt item refused",
			file:    ".acta/debt/2026-10-01-d.md",
			id:      "debt/2026-10-01-d#item-1",
			src:     "---\nid: DBT-0001\n---\n# Review NOTEs: D\n\n- [ ] first\n- [x] second\n",
			value:   "Renamed item",
			badWant: true,
		},
		{
			name:    "crlf debt item refused",
			file:    ".acta/debt/2026-10-02-d.md",
			id:      "debt/2026-10-02-d#item-1",
			src:     "---\r\nid: DBT-0002\r\n---\r\n# Review NOTEs: D\r\n\r\n- [ ] first\r\n- [x] second\r\n",
			value:   "Renamed item",
			badWant: true,
		},
		// A task takes its title from its "### Task" heading inside a plan, so
		// it is refused the same way a debt item is.
		{
			name:    "plan task refused",
			file:    ".acta/plans/2026-10-03-p.md",
			id:      "plans/2026-10-03-p#task-1",
			src:     "---\nid: PLN-0001\n---\n# P\n\n### Task 1: T\n- [ ] a\n",
			value:   "Renamed task",
			badWant: true,
		},
		// The debt file itself is the item the board shows, so it keeps the
		// set and its checklist lines stay byte for byte.
		{
			name:  "debt file heading rewritten",
			file:  ".acta/debt/2026-10-04-d.md",
			src:   "---\nid: DBT-0003\n---\n# Review NOTEs: D\n\n- [ ] first\n- [x] second\n",
			value: "Debt notes",
			want:  "---\nid: DBT-0003\n---\n# Debt notes\n\n- [ ] first\n- [x] second\n",
		},
		{
			name:  "spec heading rewritten",
			file:  ".acta/specs/2026-10-05-s.md",
			src:   "---\nid: SPC-0001\n---\n# S\n\nDetail.\n",
			value: "Spec S",
			want:  "---\nid: SPC-0001\n---\n# Spec S\n\nDetail.\n",
		},
		// A CRLF file keeps its own ending on the rewritten heading, and every
		// other \r\n in the file stays where it was.
		{
			name:  "crlf heading rewritten keeps the line ending",
			file:  ".acta/bugs/2026-10-06-crash.md",
			src:   "---\r\nref: B-1\r\nstatus: open\r\n---\r\n# Crash\r\n\r\n## Symptom\r\nIt crashes.\r\n",
			value: "Crash on save",
			want:  "---\r\nref: B-1\r\nstatus: open\r\n---\r\n# Crash on save\r\n\r\n## Symptom\r\nIt crashes.\r\n",
		},
		{
			name:  "crlf heading added when there is neither",
			file:  ".acta/scratch/2026-10-07-mood.md",
			src:   "---\r\nid: SCRATCH-3\r\nstatus: raw\r\n---\r\ncatet dulu.\r\n",
			value: "Beli bensin",
			want:  "---\r\nid: SCRATCH-3\r\nstatus: raw\r\n---\r\n# Beli bensin\r\ncatet dulu.\r\n",
		},
		{
			name:  "crlf heading with no frontmatter keeps the line ending",
			file:  ".acta/scratch/2026-10-08-idea.md",
			src:   "# Old idea\r\n\r\ncatet dulu.\r\n",
			value: "New idea",
			want:  "# New idea\r\n\r\ncatet dulu.\r\n",
		},
		{
			name:  "crlf frontmatter title rewritten when there is no heading",
			file:  ".acta/scratch/2026-10-09-note.md",
			src:   "---\r\nid: SCRATCH-4\r\ntitle: Old note\r\n---\r\ncatet dulu.\r\n",
			value: "New note",
			want:  "---\r\nid: SCRATCH-4\r\ntitle: New note\r\n---\r\ncatet dulu.\r\n",
		},
		{
			name:  "lf heading with no frontmatter",
			file:  ".acta/specs/2026-10-10-s.md",
			src:   "# S\n\nDetail.\n",
			value: "Spec S",
			want:  "# Spec S\n\nDetail.\n",
		},
		// A plan file is the item the board shows, so it takes the set. Its
		// task sections below the heading are not the title and stay as they are.
		{
			name:  "plan file heading rewritten",
			file:  ".acta/plans/2026-10-09-p.md",
			src:   "---\nid: PLN-0001\n---\n# Old plan title\n\n### Task 1: T\n- [ ] a\n",
			value: "New plan title",
			want:  "---\nid: PLN-0001\n---\n# New plan title\n\n### Task 1: T\n- [ ] a\n",
		},
		{
			name:  "crlf plan file heading rewritten keeps the line ending",
			file:  ".acta/plans/2026-10-11-p.md",
			src:   "---\r\nid: PLN-0002\r\n---\r\n# Old plan title\r\n\r\n### Task 1: T\r\n- [ ] a\r\n",
			value: "New plan title",
			want:  "---\r\nid: PLN-0002\r\n---\r\n# New plan title\r\n\r\n### Task 1: T\r\n- [ ] a\r\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := repoWith(t, map[string]string{c.file: c.src})
			rel := strings.TrimPrefix(c.file, ".acta/")
			path := filepath.Join(cfg.Root, filepath.FromSlash(rel))
			id := c.id
			if id == "" {
				id = rel[:len(rel)-3] // the board id is the path without the .md suffix
			}
			// These tests commit, so keep the real user files out of it.
			t.Setenv("HOME", t.TempDir())
			t.Setenv("TMPDIR", t.TempDir())
			if mustLoad(t, cfg).Get(id) == nil {
				t.Fatalf("the board has no item %s", id)
			}
			head := gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD")

			_, err := SetValue(cfg, mustLoad(t, cfg), id, "title", c.value)
			got, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if c.badWant {
				if !errors.Is(err, ErrBadInput) {
					t.Fatalf("err = %v, want ErrBadInput", err)
				}
				if string(got) != c.src {
					t.Fatalf("refused set changed the file: %q", got)
				}
				if now := gitRun(t, cfg.RepoRoot, "rev-parse", "HEAD"); now != head {
					t.Fatalf("refused set moved HEAD from %s to %s", head, now)
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if string(got) != c.want {
				t.Fatalf("file = %q, want %q", got, c.want)
			}
			if title := mustLoad(t, cfg).Get(id).Title; title != c.value {
				t.Fatalf("board title = %q, want %q", title, c.value)
			}
		})
	}
}
