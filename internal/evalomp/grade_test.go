package evalomp

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func workspace(t *testing.T) Workspace {
	t.Helper()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "old.md"), "from the scaffold")
	before, err := ListFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, ".acta/scratch/new.md"), "only happens in Safari")
	return Workspace{Dir: dir, Before: before, Result: Result{
		Reply: "Filed it. NONE of the rules leaked.",
		Calls: []Call{
			{Tool: "bash", Input: `{"command":"acta scratch new idea"}`},
			{Tool: "bash", Input: `{"command":"acta set scratch/x status brainstorming"}`},
			{Tool: "read", Input: `{"path":"skill://scratch"}`},
		},
	}}
}

func TestGrade(t *testing.T) {
	w := workspace(t)
	judge := func(p string) (string, error) {
		if !strings.Contains(p, "Filed it.") || !strings.Contains(p, "be polite") {
			return "", errors.New("prompt is missing the reply or the rubric")
		}
		return "**PASS** it is polite", nil
	}
	cases := []struct {
		name string
		g    Grader
		pass bool
	}{
		{"file made by the run", Grader{Type: "file_exists", Path: ".acta/scratch/*.md", Exists: new(true)}, true},
		{"file from the scaffold does not count", Grader{Type: "file_exists", Path: "old.md", Exists: new(true)}, false},
		{"exists defaults to true", Grader{Type: "file_exists", Path: ".acta/scratch/*.md"}, true},
		{"absent file wanted absent", Grader{Type: "file_exists", Path: "AGENTS.md", Exists: new(false)}, true},
		{"present file wanted absent", Grader{Type: "file_exists", Path: ".acta/scratch/*.md", Exists: new(false)}, false},
		{"regex on reply", Grader{Type: "regex", Pattern: "NONE"}, true},
		{"regex needs flag i for case", Grader{Type: "regex", Pattern: "none"}, false},
		{"regex flag i", Grader{Type: "regex", Pattern: "none", Flags: "i", Target: Target{Kind: "last_message"}}, true},
		{"regex not_contains", Grader{Type: "regex", Pattern: "Caveman talk", Match: "not_contains"}, true},
		{"regex not_contains found", Grader{Type: "regex", Pattern: "Filed", Match: "not_contains"}, false},
		{"regex on file", Grader{Type: "regex", Pattern: "only happens in Safari", Target: Target{Kind: "file", Path: ".acta/scratch/new.md"}}, true},
		{"regex on missing file", Grader{Type: "regex", Pattern: "x", Target: Target{Kind: "file", Path: "gone.md"}}, false},
		{"regex bad pattern", Grader{Type: "regex", Pattern: "("}, false},
		{"regex unknown target", Grader{Type: "regex", Pattern: "x", Target: Target{Kind: "trace"}}, false},
		{"regex unknown match", Grader{Type: "regex", Pattern: "Filed", Match: "count:1"}, false},
		{"tool used, case ignored", Grader{Type: "tool_used", Tool: "Bash", InputMatch: "acta scratch new"}, true},
		{"tool used, regex input", Grader{Type: "tool_used", Tool: "Bash", InputMatch: "status brainstorming$|status brainstorming\""}, true},
		{"tool not used", Grader{Type: "tool_used", Tool: "Bash", InputMatch: "acta scratch add"}, false},
		{"tool min", Grader{Type: "tool_used", Tool: "bash", Min: new(3)}, false},
		{"tool max 0 means never", Grader{Type: "tool_used", Tool: "bash", InputMatch: "scratch add", Min: new(0), Max: new(0)}, true},
		{"tool max broken", Grader{Type: "tool_used", Tool: "bash", Max: new(1)}, false},
		{"llm pass", Grader{Type: "llm", Body: "be polite"}, true},
		{"unknown type", Grader{Type: "tool_order"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.g.Name = "g"
			got := Grade(c.g, w, judge)
			if got.Pass != c.pass || got.Grader != "g" {
				t.Errorf("Grade = %+v, want pass %v", got, c.pass)
			}
			if !got.Pass && got.Why == "" {
				t.Error("a FAIL must say why")
			}
		})
	}
}

func TestGradeJudgeFailures(t *testing.T) {
	w := workspace(t)
	g := Grader{Name: "j", Type: "llm", Body: "be polite"}
	for name, judge := range map[string]Judge{
		"error":      func(string) (string, error) { return "", errors.New("down") },
		"no verdict": func(string) (string, error) { return "Looks fine to me.", nil },
		"empty":      func(string) (string, error) { return "", nil },
		"fail":       func(string) (string, error) { return "FAIL: rude", nil },
	} {
		t.Run(name, func(t *testing.T) {
			if got := Grade(g, w, judge); got.Pass {
				t.Errorf("Grade = %+v, want FAIL", got)
			}
		})
	}
}

func TestUnsupported(t *testing.T) {
	for _, g := range []Grader{
		{Type: "file_exists"}, {Type: "tool_used"}, {Type: "llm"},
		{Type: "regex"}, {Type: "regex", Target: Target{Kind: "file", Path: "a"}},
		{Type: "regex", Match: "not_contains", Target: Target{Kind: "last_message"}},
	} {
		if why := Unsupported(g); why != "" {
			t.Errorf("Unsupported(%+v) = %q, want empty", g, why)
		}
	}
	for _, g := range []Grader{
		{Type: "baseline"}, {Type: "regex", Target: Target{Kind: "files"}}, {Type: "regex", Match: "count:2"},
	} {
		if Unsupported(g) == "" {
			t.Errorf("Unsupported(%+v) is empty, want a reason", g)
		}
	}
}

// A grader this runner cannot do has to say which word stopped it, or the
// report leaves the reader guessing.
func TestGradeNamesTheUnknownPart(t *testing.T) {
	w := workspace(t)
	for _, c := range []struct {
		g    Grader
		want string
	}{
		{Grader{Type: "tool_order"}, "tool_order"},
		{Grader{Type: "baseline"}, "baseline"},
		{Grader{Type: "command"}, "command"},
		{Grader{Type: ""}, `""`},
		{Grader{Type: "regex", Target: Target{Kind: "trace"}}, "trace"},
		{Grader{Type: "regex", Target: Target{Kind: "files"}}, "files"},
		{Grader{Type: "regex", Target: Target{Kind: "mock_calls"}}, "mock_calls"},
		{Grader{Type: "regex", Match: "count:1"}, "count:1"},
		{Grader{Type: "regex", Match: "count:1", Target: Target{Kind: "trace"}}, "trace"},
	} {
		got := Grade(c.g, w, func(string) (string, error) { return "PASS", nil })
		if got.Pass || !strings.Contains(got.Why, c.want) {
			t.Errorf("Grade(%+v) = %+v, want a FAIL naming %q", c.g, got, c.want)
		}
	}
}

func TestGradeUnlistableWorkspace(t *testing.T) {
	w := workspace(t)
	w.Dir = filepath.Join(t.TempDir(), "gone")
	g := Grader{Name: "g", Type: "file_exists", Path: "*"}
	got := Grade(g, w, nil)
	if got.Pass || got.Why == "" {
		t.Errorf("Grade = %+v, want a FAIL saying why", got)
	}
}

func TestGradeBadInputMatch(t *testing.T) {
	w := workspace(t)
	g := Grader{Name: "g", Type: "tool_used", Tool: "bash", InputMatch: "("}
	got := Grade(g, w, nil)
	if got.Pass || !strings.Contains(got.Why, "(") {
		t.Errorf("Grade = %+v, want a FAIL naming the bad input_match", got)
	}
}

// Every verdict word the plan lists for a judge: the first word, uppercased
// and stripped of punctuation, decides.
func TestGradeJudgeVerdictWords(t *testing.T) {
	w := workspace(t)
	g := Grader{Name: "g", Type: "llm", Body: "be polite"}
	for _, c := range []struct {
		ans  string
		pass bool
		why  string
	}{
		{"PASS", true, ""}, {"pass it is polite", true, ""}, {"**PASS**", true, ""},
		{"PASS: it is polite", true, ""}, {"Pass.", true, ""}, {"\n\nPASS ok", true, ""},
		// A judge that said FAIL gets its own reason. It must not be the
		// "no verdict" one: the answer is right there.
		{"FAIL, rude", false, "judge: FAIL, rude"},
		{"maybe PASS", false, "no PASS or FAIL"},
		{"passed", false, "no PASS or FAIL"},
	} {
		got := Grade(g, w, func(string) (string, error) { return c.ans, nil })
		if got.Pass != c.pass {
			t.Errorf("Grade(judge = %q) = %+v, want pass %v", c.ans, got, c.pass)
		}
		if c.why != "" && !strings.Contains(got.Why, c.why) {
			t.Errorf("Grade(judge = %q).Why = %q, want it to mention %q", c.ans, got.Why, c.why)
		}
	}
}

// A llm grader that is not a llm grader must not fall through to the judge.
func TestGradeNeverCallsTheJudgeForOtherTypes(t *testing.T) {
	w := workspace(t)
	for _, g := range []Grader{
		{Type: "file_exists", Path: "old.md"},
		{Type: "tool_used", Tool: "bash"},
		{Type: "regex", Pattern: "Filed"},
		{Type: "nonsense"},
	} {
		called := false
		Grade(g, w, func(string) (string, error) { called = true; return "PASS", nil })
		if called {
			t.Errorf("Grade(%+v) called the judge", g)
		}
	}
}

// probePattern is the real pattern from the probe-round case. It uses lookaround,
// which Go's own regexp cannot read.
const probePattern = `\*\*Q\d+\.[^\n*]*\*\*(?:(?!Recommended:)[\s\S])*?(?=\*\*Q\d+\.|$)`

func TestGradeJavaScriptPatterns(t *testing.T) {
	grade := func(g Grader, reply string, calls ...Call) Outcome {
		g.Name = "g"
		return Grade(g, Workspace{Dir: t.TempDir(), Result: Result{Reply: reply, Calls: calls}}, nil)
	}
	probe := Grader{Type: "regex", Pattern: probePattern, Match: "not_contains", Target: Target{Kind: "last_message"}}

	missing := "**Q1.** Which one?\nA or B.\n\n**Q2.** And this?\nRecommended: A"
	if got := grade(probe, missing); got.Pass || !strings.Contains(got.Why, "Q") {
		t.Errorf("a block with no Recommended: must FAIL with the pattern in the message, got %+v", got)
	}
	whole := "**Q1.** Which one?\nRecommended: B\n\n**Q2.** And this?\nRecommended: A"
	if got := grade(probe, whole); !got.Pass {
		t.Errorf("every block has Recommended:, want PASS, got %+v", got)
	}

	if got := grade(Grader{Type: "regex", Pattern: "none(?!x)", Flags: "i"}, "NONE here"); !got.Pass {
		t.Errorf("flag i must match across case, got %+v", got)
	}
	call := Call{Tool: "bash", Input: `{"command":"acta scratch new idea"}`}
	if got := grade(Grader{Type: "tool_used", Tool: "Bash", InputMatch: `acta scratch (?=new)`}, "", call); !got.Pass {
		t.Errorf("tool_used input_match with lookahead must match, got %+v", got)
	}
}

func TestGradeMatchErrorNeverPasses(t *testing.T) {
	// Real runs wait 2 seconds. Here a short wait keeps the test fast.
	oldTimeout := matchTimeout
	matchTimeout = 50 * time.Millisecond
	t.Cleanup(func() { matchTimeout = oldTimeout })
	// A pattern that backtracks forever hits the short timeout set above. The
	// grader must FAIL then, even when it wants "not found".
	g := Grader{Name: "g", Type: "regex", Pattern: `(a+)+$`, Match: "not_contains"}
	w := Workspace{Dir: t.TempDir(), Result: Result{Reply: strings.Repeat("a", 60) + "!"}}
	got := Grade(g, w, nil)
	if got.Pass || !strings.Contains(got.Why, "(a+)+$") {
		t.Errorf("a match error must FAIL and name the pattern, got %+v", got)
	}
	tool := Grader{Name: "t", Type: "tool_used", Tool: "bash", InputMatch: `(a+)+$`, Min: new(0), Max: new(0)}
	w.Result.Calls = []Call{{Tool: "bash", Input: strings.Repeat("a", 60) + "!"}}
	if got := Grade(tool, w, nil); got.Pass || !strings.Contains(got.Why, "(a+)+$") {
		t.Errorf("a tool match error must FAIL and name the pattern, got %+v", got)
	}
}
