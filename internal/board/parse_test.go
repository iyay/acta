package board

import (
	"strings"
	"testing"
)

func TestParseFrontTitleAndBody(t *testing.T) {
	d := Parse([]byte("---\nref: New-261\nstatus: open\n---\n# The title\n\ntext\n"))
	if !d.HasFront || d.FrontErr != nil {
		t.Fatalf("front: has=%v err=%v", d.HasFront, d.FrontErr)
	}
	if d.Front["ref"] != "New-261" || d.Front["status"] != "open" {
		t.Fatalf("front = %v", d.Front)
	}
	if d.Title != "The title" {
		t.Fatalf("title = %q", d.Title)
	}
	if d.Body != "# The title\n\ntext\n" {
		t.Fatalf("body = %q", d.Body)
	}
}

func TestParseNoFront(t *testing.T) {
	d := Parse([]byte("# Only title\n"))
	if d.HasFront || d.Front != nil || d.Title != "Only title" {
		t.Fatalf("got %+v", d)
	}
}

func TestParseBrokenFront(t *testing.T) {
	d := Parse([]byte("---\nref: [unclosed\n---\n# Still read\n"))
	if !d.HasFront || d.FrontErr == nil {
		t.Fatalf("want a front error, got has=%v err=%v", d.HasFront, d.FrontErr)
	}
	if d.Title != "Still read" {
		t.Fatalf("title = %q", d.Title)
	}
}

func TestParseUnclosedFrontIsBody(t *testing.T) {
	d := Parse([]byte("---\nno closing line\n# Title\n"))
	if d.HasFront || d.Title != "Title" {
		t.Fatalf("got has=%v title=%q", d.HasFront, d.Title)
	}
}

func TestParseCRLF(t *testing.T) {
	d := Parse([]byte("---\r\nref: X\r\n---\r\n# T\r\n"))
	if d.Front["ref"] != "X" || d.Title != "T" {
		t.Fatalf("got %+v", d)
	}
}

func TestParseSpecLine(t *testing.T) {
	cases := map[string]string{
		"**Spec:** `docs/superpowers/specs/2026-09-25-a-design.md`\n":     "docs/superpowers/specs/2026-09-25-a-design.md",
		"**Spec:** .pm/specs/2026-09-16-finished.md\n":                    ".pm/specs/2026-09-16-finished.md",
		"**Spec:** `.pm/specs/x.md`, which builds on `.pm/specs/y.md`.\n": ".pm/specs/x.md",
		"Spec: `not-bold.md`\n":                                           "",
	}
	for in, want := range cases {
		if got := Parse([]byte("# P\n\n" + in)).SpecPath; got != want {
			t.Errorf("%q: SpecPath = %q, want %q", in, got, want)
		}
	}
}

func TestParseTasks(t *testing.T) {
	src := strings.Join([]string{
		"# Plan",                     // 1
		"",                           // 2
		"### Task 1: First step",     // 3
		"",                           // 4
		"- [x] **Step 1: test**",     // 5
		"- [X] **Step 2: code**",     // 6
		"",                           // 7
		"### Task F1 (be): Fix it",   // 8
		"- [x] one",                  // 9
		"- [ ] two",                  // 10
		"```text",                    // 11
		"- [ ] inside a fence",       // 12
		"### Task 9: inside a fence", // 13
		"```",                        // 14
		"## Next section",            // 15
		"- [ ] not in any task",      // 16
		"### Task 2",                 // 17
		"",
	}, "\n")
	d := Parse([]byte(src))
	if len(d.Tasks) != 3 {
		t.Fatalf("got %d tasks: %+v", len(d.Tasks), d.Tasks)
	}
	want := []TaskSec{
		{Num: "1", Title: "First step", Line: 3, Done: 2, Total: 2},
		{Num: "F1", Title: "(be): Fix it", Line: 8, Done: 1, Total: 2},
		{Num: "2", Title: "", Line: 17, Done: 0, Total: 0},
	}
	for i, w := range want {
		g := d.Tasks[i]
		if g.Num != w.Num || g.Title != w.Title || g.Line != w.Line || g.Done != w.Done || g.Total != w.Total {
			t.Errorf("task %d = %+v, want %+v", i, g, w)
		}
	}
	if !strings.HasPrefix(d.Tasks[1].Body, "### Task F1") || strings.Contains(d.Tasks[1].Body, "Next section") {
		t.Errorf("task F1 body = %q", d.Tasks[1].Body)
	}
}

func TestParseTitleIgnoresFence(t *testing.T) {
	d := Parse([]byte("```\n# not a title\n```\n# Real title\n"))
	if d.Title != "Real title" {
		t.Fatalf("title = %q", d.Title)
	}
}
