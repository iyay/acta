package wiki

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRewriteChecked(t *testing.T) {
	t.Parallel()
	known := []string{"a", "b", "sub/c"}
	cases := []struct {
		name, old string
		touched   []string
		want      string
	}{
		{"empty file gets the touched pages", "", []string{"b", "a"}, "a 2026-10-14\nb 2026-10-14\n"},
		{"a touched page gets today", "a 2026-01-01\nb 2026-01-01\n", []string{"a"}, "a 2026-10-14\nb 2026-01-01\n"},
		{"a duplicate keeps the newest date", "a 2026-01-01\na 2026-03-01\na 2026-02-01\n", nil, "a 2026-03-01\n"},
		{"an unknown page is dropped", "gone 2026-01-01\na 2026-01-01\n", nil, "a 2026-01-01\n"},
		{"lines come out sorted", "sub/c 2026-01-01\nb 2026-01-01\na 2026-01-01\n", nil, "a 2026-01-01\nb 2026-01-01\nsub/c 2026-01-01\n"},
		{"blank and broken lines are dropped", "\n  \na\nb 2026-02-01 extra\nb 2026-01-01\n", nil, "b 2026-01-01\n"},
		{"nothing left is an empty file", "gone 2026-01-01\n", nil, ""},
	}
	for _, c := range cases {
		if got := RewriteChecked(c.old, known, c.touched, "2026-10-14"); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
	// Every order of the same lines gives the same file.
	a := RewriteChecked("b 2026-01-01\na 2026-02-01\na 2026-01-01\n", known, nil, "2026-10-14")
	b := RewriteChecked("a 2026-01-01\na 2026-02-01\nb 2026-01-01\n", known, nil, "2026-10-14")
	if a != b || a != "a 2026-02-01\nb 2026-01-01\n" {
		t.Errorf("order changed the output: %q and %q", a, b)
	}
}

func TestMarkChecked(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p, err := MarkChecked(dir, []string{"a", "b"}, []string{"b"}, "2026-10-14")
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(dir, CheckedFile) {
		t.Errorf("path %q", p)
	}
	if err := os.WriteFile(p, []byte("a 2026-01-01\nb 2026-10-14\nb 2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := MarkChecked(dir, []string{"a", "b"}, []string{"a"}, "2026-10-15"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	if string(got) != "a 2026-10-15\nb 2026-10-14\n" {
		t.Errorf("file %q", got)
	}
}

func TestMarkCheckedRefusesALink(t *testing.T) {
	t.Parallel()
	dir, other := t.TempDir(), filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(other, []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, filepath.Join(dir, CheckedFile)); err != nil {
		t.Fatal(err)
	}
	if _, err := MarkChecked(dir, []string{"a"}, []string{"a"}, "2026-10-14T00:00:00Z"); err == nil {
		t.Error("MarkChecked wrote through a link")
	}
	if raw, _ := os.ReadFile(other); string(raw) != "keep me\n" {
		t.Errorf("the link target changed: %q", raw)
	}
}
