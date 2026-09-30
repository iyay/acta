package evalomp

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadCases(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "alpha/prompt.md"), "---\ntags: [scratch]\n# a comment\ntimeout_seconds: 60\n---\n\nDo the thing.\n")
	write(t, filepath.Join(dir, "alpha/case.yaml"), "schema_version: \"1.1\"\nname: alpha\ncontext:\n  scaffold_script: scaffold.sh\n")
	write(t, filepath.Join(dir, "alpha/graders/file.md"), "---\ntype: file_exists\npath: \".acta/scratch/*.md\"\nexists: true\n---\n")
	write(t, filepath.Join(dir, "alpha/graders/body.md"), "---\n# guards: x\ntype: regex\npattern: \"only here\"\ntarget:\n  source: file\n  path: \"a.md\"\n---\n")
	write(t, filepath.Join(dir, "beta/prompt.md"), "---\ntags: [brainstorm, claude-only]\n---\nSecond.")
	write(t, filepath.Join(dir, "beta/graders/judge.md"), "---\ntype: llm\n---\n\nPASS only when polite.\n")
	write(t, filepath.Join(dir, "beta/graders/tool.md"), "---\ntype: tool_used\ntool: Bash\ninput_match: \"acta scratch new\"\nmin: 2\n---")
	write(t, filepath.Join(dir, "results/report.html"), "not a case")

	cases, err := LoadCases(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 2 || cases[0].Name != "alpha" || cases[1].Name != "beta" {
		t.Fatalf("cases = %+v", cases)
	}
	a, b := cases[0], cases[1]
	if a.Prompt != "Do the thing." || a.TimeoutSeconds != 60 || a.ClaudeOnly() {
		t.Errorf("alpha = %+v", a)
	}
	if a.Scaffold != filepath.Join(dir, "alpha", "scaffold.sh") {
		t.Errorf("scaffold = %q", a.Scaffold)
	}
	if len(a.Graders) != 2 || a.Graders[0].Name != "body" || a.Graders[1].Name != "file" {
		t.Fatalf("alpha graders = %+v", a.Graders)
	}
	if g := a.Graders[0]; g.Target != (Target{Kind: "file", Path: "a.md"}) || g.Pattern != "only here" {
		t.Errorf("body grader = %+v", g)
	}
	if g := a.Graders[1]; g.Type != "file_exists" || g.Exists == nil || !*g.Exists || g.Path != ".acta/scratch/*.md" {
		t.Errorf("file grader = %+v", g)
	}
	if !b.ClaudeOnly() || b.TimeoutSeconds != 300 || b.Scaffold != "" || b.Prompt != "Second." {
		t.Errorf("beta = %+v", b)
	}
	if g := b.Graders[0]; g.Type != "llm" || g.Body != "PASS only when polite." {
		t.Errorf("judge grader = %+v", g)
	}
	if g := b.Graders[1]; g.Tool != "Bash" || g.InputMatch != "acta scratch new" || g.Min == nil || *g.Min != 2 || g.Max != nil {
		t.Errorf("tool grader = %+v", g)
	}
}

func TestLoadCasesScalarTarget(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "c/prompt.md"), "---\n---\nHi")
	write(t, filepath.Join(dir, "c/graders/r.md"), "---\ntype: regex\npattern: NONE\nflags: i\ntarget: last_message\n---\n")
	cases, err := LoadCases(dir)
	if err != nil {
		t.Fatal(err)
	}
	if g := cases[0].Graders[0]; g.Target.Kind != "last_message" || g.Flags != "i" {
		t.Errorf("grader = %+v", g)
	}
}

// A grader that is only frontmatter can end on the closing line with no
// newline behind it, and it still has to load.
func TestLoadCasesFrontmatterOnlyNoNewline(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "c/prompt.md"), "---\n---\nHi")
	write(t, filepath.Join(dir, "c/graders/r.md"), "---\ntype: regex\npattern: \"$^\"\nmatch: none\ntarget: last_message\n---")
	cases, err := LoadCases(dir)
	if err != nil {
		t.Fatal(err)
	}
	if g := cases[0].Graders[0]; g.Type != "regex" || g.Pattern != "$^" || g.Match != "none" || g.Body != "" {
		t.Errorf("grader = %+v", g)
	}
}

func TestLoadCasesErrors(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"no frontmatter": {"c/prompt.md": "Hi"},
		"bad yaml":       {"c/prompt.md": "---\ntags: [\n---\nHi"},
		"bad grader":     {"c/prompt.md": "---\n---\nHi", "c/graders/g.md": "no frontmatter"},
		"bad case.yaml":  {"c/prompt.md": "---\n---\nHi", "c/case.yaml": "context: ["},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			for p, text := range files {
				write(t, filepath.Join(dir, p), text)
			}
			if _, err := LoadCases(dir); err == nil {
				t.Error("want an error")
			}
		})
	}
	if _, err := LoadCases(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("a missing eval folder must be an error")
	}
}

// TestRealSuiteRunsInOmp keeps the shared suite honest: a new case either
// uses graders this runner can grade, or is tagged claude-only on purpose.
func TestRealSuiteRunsInOmp(t *testing.T) {
	cases, err := LoadCases("../../plugin/evals")
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no cases found in plugin/evals")
	}
	for _, c := range cases {
		if c.Name == "second-brainstorm-choices" && !c.ClaudeOnly() {
			t.Error("second-brainstorm-choices asks for claude --bg, so it must be tagged claude-only")
		}
		if c.ClaudeOnly() {
			continue
		}
		for _, g := range c.Graders {
			if why := Unsupported(g); why != "" {
				t.Errorf("%s/%s: %s", c.Name, g.Name, why)
			}
		}
	}
}
