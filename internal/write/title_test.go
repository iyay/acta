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
// byte for byte, so nothing but the title can move.
func TestSetValueTitle(t *testing.T) {
	cases := []struct {
		name    string
		file    string
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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := repoWith(t, map[string]string{c.file: c.src})
			rel := strings.TrimPrefix(c.file, ".acta/")
			path := filepath.Join(cfg.Root, filepath.FromSlash(rel))
			id := rel[:len(rel)-3] // the board id is the path without the .md suffix
			// These tests commit, so keep the real user files out of it.
			t.Setenv("HOME", t.TempDir())
			t.Setenv("TMPDIR", t.TempDir())

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
