package board

import (
	"reflect"
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
		"**Spec:** `docs/superpowers/specs/2026-09-25-a-design.md`\n":                              "docs/superpowers/specs/2026-09-25-a-design.md",
		"**Spec:** .pm/specs/2026-09-16-finished.md\n":                                             ".pm/specs/2026-09-16-finished.md",
		"**Spec:** `.pm/specs/x.md`, which builds on `.pm/specs/y.md`.\n":                          ".pm/specs/x.md",
		"Spec: `not-bold.md`\n":                                                                    "",
		"**Spec:** none (Bounded, approved in chat on 2026-09-26)\n":                               "",
		"**Spec:** No spec file. Source: memory `tick-fixes-review-notes`.\n":                      "",
		"**Spec:** Bounded design approved in chat on 2026-09-26 (no spec file):\n":                "",
		"**Spec:** `notes` then `.pm/specs/z.md`\n":                                                ".pm/specs/z.md",
		"**Spec:** see .pm/specs/typo-desing.md\n":                                                 ".pm/specs/typo-desing.md",
		"**Spec:** (.pm/specs/w.md).\n":                                                            ".pm/specs/w.md",
		"**Spec:** Design approved in chat (Bounded). Rulings: (1) CLAUDE.md or AGENTS.md wins.\n": "",
		"**Spec:** see README.md and .pm/specs/v.md\n":                                             ".pm/specs/v.md",
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

func TestParseTaskNumberAsWritten(t *testing.T) {
	cases := []struct{ heading, num, title string }{
		{"### Task 1: First step", "1", "First step"},
		{"### Task 1:", "1", ""},
		{"### Task 1", "1", ""},
		{"### Task 1.", "1", ""},
		{"### Task 1. Trailing dot space", "1", "Trailing dot space"},
		{"### Task 3. Dot title", "3", "Dot title"},
		{"### Task 2.1", "2.1", ""},
		{"### Task 2.1:", "2.1", ""},
		{"### Task 2.1 Title", "2.1", "Title"},
		{"### Task 1.2.3", "1.2.3", ""},
		{"### Task F1 (be): Fix it", "F1", "(be): Fix it"},
		{"### Task F-1: Drop the stored-column fallback", "F-1", "Drop the stored-column fallback"},
		{"### Task F-2: Re-check the quantity math", "F-2", "Re-check the quantity math"},
		{"### Task 2a: Suffix letter", "2a", "Suffix letter"},
		{"### Task F-10: Double digits", "F-10", "Double digits"},
		{"### Task 2", "2", ""},
		{"### Task 4", "4", ""},
		{"### Task 3 (be): x", "3", "(be): x"},
		{"### Task F-1 - Dash separator", "F-1", "- Dash separator"},
	}
	for _, c := range cases {
		d := Parse([]byte("# Plan\n\n" + c.heading + "\n"))
		if len(d.Tasks) != 1 {
			t.Errorf("%q: got %d tasks, want 1", c.heading, len(d.Tasks))
			continue
		}
		if g := d.Tasks[0]; g.Num != c.num || g.Title != c.title {
			t.Errorf("%q = num %q title %q, want num %q title %q", c.heading, g.Num, g.Title, c.num, c.title)
		}
	}
}

func TestParseDashTasksDoNotCollapse(t *testing.T) {
	d := Parse([]byte("# Plan\n\n### Task F-1: First\n\n### Task F-2: Second\n"))
	if len(d.Tasks) != 2 {
		t.Fatalf("got %d tasks, want 2: %+v", len(d.Tasks), d.Tasks)
	}
	if d.Tasks[0].Num != "F-1" || d.Tasks[1].Num != "F-2" {
		t.Fatalf("nums = %q, %q", d.Tasks[0].Num, d.Tasks[1].Num)
	}
}

func TestParseDottedTasksDoNotCollapse(t *testing.T) {
	d := Parse([]byte("# Plan\n\n### Task 2.1: First\n\n### Task 2.2: Second\n"))
	if len(d.Tasks) != 2 {
		t.Fatalf("got %d tasks, want 2: %+v", len(d.Tasks), d.Tasks)
	}
	if d.Tasks[0].Num != "2.1" || d.Tasks[1].Num != "2.2" {
		t.Fatalf("nums = %q, %q", d.Tasks[0].Num, d.Tasks[1].Num)
	}
	if d.Tasks[0].Title != "First" || d.Tasks[1].Title != "Second" {
		t.Fatalf("titles = %q, %q", d.Tasks[0].Title, d.Tasks[1].Title)
	}
}

func TestParseTitleIgnoresFence(t *testing.T) {
	d := Parse([]byte("```\n# not a title\n```\n# Real title\n"))
	if d.Title != "Real title" {
		t.Fatalf("title = %q", d.Title)
	}
}

func TestParseDebtChecklist(t *testing.T) {
	src := "---\nid: DEBT-3\n---\n# Review NOTEs: X\n\nintro line\n\n- [ ] open one\n- [x] done one\n- [X] done two\n- [-] skipped one\n\n```\n- [ ] inside fence\n```\n"
	doc := Parse([]byte(src))
	want := []ItemLine{
		{Num: 1, Text: "open one", Line: 8, State: ' '},
		{Num: 2, Text: "done one", Line: 9, State: 'x'},
		{Num: 3, Text: "done two", Line: 10, State: 'x'},
		{Num: 4, Text: "skipped one", Line: 11, State: '-'},
	}
	if !reflect.DeepEqual(doc.Items, want) {
		t.Fatalf("items = %#v, want %#v", doc.Items, want)
	}
}
