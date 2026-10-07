package evalomp

import (
	"strings"
	"testing"
)

func TestPatternProblems(t *testing.T) {
	cases := []Case{{Name: "c1", Graders: []Grader{
		{Name: "good-regex", Type: "regex", Pattern: `a(?!b)`},
		{Name: "bad-regex", Type: "regex", Pattern: "("},
		{Name: "good-tool", Type: "tool_used", Tool: "Bash", InputMatch: `x(?=y)`},
		{Name: "bad-tool", Type: "tool_used", Tool: "Bash", InputMatch: "["},
		{Name: "judge", Type: "llm", Pattern: "("},
		{Name: "file", Type: "file_exists", Path: "a.md"},
	}}}
	got := PatternProblems(cases)
	if len(got) != 2 {
		t.Fatalf("PatternProblems = %q, want 2 lines", got)
	}
	if !strings.Contains(got[0], "c1") || !strings.Contains(got[0], "bad-regex") {
		t.Errorf("first line %q must name the case and grader", got[0])
	}
	if !strings.Contains(got[1], "bad-tool") {
		t.Errorf("second line %q must name bad-tool", got[1])
	}
	if p := PatternProblems(cases[:0]); len(p) != 0 {
		t.Errorf("no cases gave %q", p)
	}
	good := []Case{{Name: "g", Graders: []Grader{{Name: "ok", Type: "regex", Pattern: `a(?!b)`, Flags: "i"}}}}
	if p := PatternProblems(good); len(p) != 0 {
		t.Errorf("good patterns gave %q", p)
	}
}

func TestCompilePatternFlags(t *testing.T) {
	re, err := CompilePattern("none", "i")
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := re.MatchString("NONE"); !ok {
		t.Error("flag i must ignore case")
	}
	re, err = CompilePattern("none", "")
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := re.MatchString("NONE"); ok {
		t.Error("no flag must be case sensitive")
	}
	if _, err := CompilePattern("(", ""); err == nil {
		t.Error("a bad pattern must not compile")
	}
}
